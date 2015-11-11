#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <string.h>

#include <openssl/aes.h>
#include <openssl/rsa.h>
#include <openssl/x509.h>
#include <openssl_helpers.h>
#include <openssl/rand.h>
#include <openssl/hmac.h>
#include <openssl/sha.h>

#include <openssl_helpers.h>
#include <quote_protocol.h>

#include <tpm20.h>
#include <tpm2_lib.h>
#include <gflags/gflags.h>

//
// Copyright 2015 Google Corporation, All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// or in the the file LICENSE-2.0.txt in the top level sourcedirectory
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License
//
// Portions of this code were derived TPM2.0-TSS published
// by Intel under the license set forth in intel_license.txt
// and downloaded on or about August 6, 2015.
// Portions of this code were derived tboot published
// by Intel under the license set forth in intel_license.txt
// and downloaded on or about August 6, 2015.
// Portions of this code were derived from the crypto utility
// published by John Manferdelli under the Apache 2.0 license.
// See github.com/jlmucb/crypto.
// File: ServerSignProgramKeyRequest.cc


//  This program verifies endorsement cert, quote key and signature.
//  It then constructs and signs an x509 cert for the proposed
//  program key.  It encrypts the signed cert to the Endorsement Key
//  referencing the Quote Key and creates the decrypt information
//  required by ActivateCredential.  It saves the encrypted
//  information in the response file.

// Calling sequence: ServerSignProgramKeyRequest.exe
//    --program_cert_request_file=input-file-name
//    --program_cert_response_file=output-file-name


using std::string;


#define CALLING_SEQUENCE "ServerSignProgramKeyRequest.exe " \
"--signing_instructions_file=input-file" \
"--cloudproxy_key_file=input-file" \
"--program_cert_request_file=output-file-name " \
"--program_response_file=output-file-name"

void PrintOptions() {
  printf("Calling sequence: %s", CALLING_SEQUENCE);
}

DEFINE_string(signed_endorsement_cert_file, "", "input-file-name");
DEFINE_string(signing_instructions_file, "", "input-file-name");
DEFINE_string(program_cert_request_file, "", "input-file-name");
// TODO(jlm): policy file should contain list of approved pcrs
DEFINE_string(hash_alg, "sha1", "hash-function");
DEFINE_string(policy_file, "", "input-file-name");
DEFINE_string(policy_cert_file, "policy_cert_file", "input-file-name");
DEFINE_string(policy_identifier, "cloudproxy", "policy domain name");
DEFINE_string(cloudproxy_key_file, "", "input-file-name");
DEFINE_string(program_response_file, "", "output-file-name");

#ifndef GFLAGS_NS
#define GFLAGS_NS google
#endif

#define MAX_SIZE_PARAMS 8192
#define DEBUG

// magic constant for tpm generated
#define TpmMagicConstant 0xff544347

// Consults policy database to confirm pcr's are OK
bool ValidPCR(TPM_ALG_ID hash, byte* pcr_selection, byte* digest) {
  return true;
}

int main(int an, char** av) {
  int ret_val = 0;

  printf("\nServerSignProgramKeyRequest\n\n");

  GFLAGS_NS::ParseCommandLineFlags(&an, &av, true);

  int size_cert_request = MAX_SIZE_PARAMS;
  byte cert_request_buf[MAX_SIZE_PARAMS];
  x509_cert_request_parameters_message cert_request;
 
#ifdef DEBUG
  int test_size = MAX_SIZE_PARAMS;
  byte test_buf[MAX_SIZE_PARAMS];
#endif
  
  int in_size = MAX_SIZE_PARAMS;
  byte in_buf[MAX_SIZE_PARAMS];

  X509_REQ* req = nullptr;

  int size_seed = 16;
  byte seed[32];

  TPM2B_DIGEST unmarshaled_credential;
  TPM2B_DIGEST marshaled_credential;
  TPM2B_NAME unmarshaled_name;
  TPM2B_NAME marshaled_name;

  byte symKey[MAX_SIZE_PARAMS];
  TPM_ALG_ID hash_alg_id;
  if (FLAGS_hash_alg == "sha1") {
    hash_alg_id = TPM_ALG_SHA1;
  } else if (FLAGS_hash_alg == "sha256") {
    hash_alg_id = TPM_ALG_SHA256;
  } else {
    printf("Unknown hash algorithm\n");
    return 1;
  }

  int size_hmacKey = SizeHash(hash_alg_id);
  byte hmacKey[MAX_SIZE_PARAMS];

  string key;
  string label;
  string contextU;
  string contextV;
  HMAC_CTX hctx;
  int size_encIdentity;
  byte encIdentity[MAX_SIZE_PARAMS];
  byte decrypted_quote[MAX_SIZE_PARAMS];
  byte outerHmac[128];

  string cert_key_seed;
  int size_derived_keys;
  byte derived_keys[128];

  byte* p;

  int size_active_out;

  byte* der_program_cert = nullptr;
  int der_program_cert_size = 0;
  byte der_policy_cert[MAX_SIZE_PARAMS];
  int der_policy_cert_size = MAX_SIZE_PARAMS;

  int endorsement_blob_size;
  X509* program_cert = X509_new();
  X509* endorsement_cert = nullptr;
  X509* policy_cert = nullptr;

  byte program_key_quoted_hash[256];

  int encrypted_data_size = 0;
  byte encrypted_data[MAX_SIZE_PARAMS];
  int encrypted_data_hmac_size = 0;
  byte encrypted_data_hmac[MAX_SIZE_PARAMS];

  int signed_quote_hash_size = 0;
  byte signed_quote_hash[MAX_SIZE_PARAMS];

  byte zero_iv[32];
  memset(zero_iv, 0, 32);
  
  int encrypted_secret_size = 0;
  byte encrypted_secret[MAX_SIZE_PARAMS];
  int quote_struct_size = 0;
  byte quote_struct[MAX_SIZE_PARAMS];

  private_key_blob_message private_key;
  program_cert_request_message request;
  program_cert_response_message response;
  signing_instructions_message signing_message;
  x509_cert_request_parameters_message cert_parameters;

  X509_STORE* store = nullptr;
  X509_STORE_CTX* verify_ctx = nullptr;
  int cert_OK = 0;
  byte* p_byte = nullptr;

  string name;
  string input;
  string output;
  string private_key_blob;
  TPM2B_DIGEST marshaled_integrityHmac;

  uint16_t size_in = 0;
  SHA_CTX sha1;
  SHA256_CTX sha256;
  RSA* signing_key = nullptr;
  RSA* active_key = RSA_new();
  TPMS_ATTEST attested_quote;

  EVP_PKEY* protector_evp_key = nullptr;
  RSA* protector_key = nullptr;

  string serialized_program_key;

  if (FLAGS_signing_instructions_file == "") {
    printf("signing_instructions_file is empty\n");
    ret_val = 1;
    goto done;
  }
  if (FLAGS_program_cert_request_file == "") {
    printf("program_cert_request_file is empty\n");
    ret_val = 1;
    goto done;
  }
  if (FLAGS_cloudproxy_key_file == "") {
    printf("cloudproxy_key_file is empty\n");
    ret_val = 1;
    goto done;
  }
  if (FLAGS_program_response_file == "") {
    printf("program_response_file is empty\n");
    ret_val = 1;
    goto done;
  }

  OpenSSL_add_all_algorithms();

  // Get request
  if (!ReadFileIntoBlock(FLAGS_program_cert_request_file, &size_cert_request,
                         cert_request_buf)) {
    printf("Can't read cert request\n");
    ret_val = 1;
    goto done;
  }

#ifdef DEBUG
  printf("Program cert request (%d): ", size_cert_request);
  PrintBytes(size_cert_request, cert_request_buf);
  printf("\n");
#endif

  input.assign((const char*)cert_request_buf, size_cert_request);
  if (!request.ParseFromString(input)) {
    printf("Can't parse cert request\n");
    ret_val = 1;
    goto done;
  }

  // Get signing instructions
  if (!ReadFileIntoBlock(FLAGS_signing_instructions_file, &in_size, in_buf)) {
    printf("Can't read signing instructions %s\n",
           FLAGS_signing_instructions_file.c_str());
    ret_val = 1;
    goto done;
  }
  input.assign((const char*)in_buf, in_size);
  if (!signing_message.ParseFromString(input)) {
    printf("Can't parse signing instructions\n");
    ret_val = 1;
    goto done;
  }
  printf("issuer: %s, duration: %ld, purpose: %s, hash: %s\n",
         signing_message.issuer().c_str(), (long)signing_message.duration(),
         signing_message.purpose().c_str(), signing_message.hash_alg().c_str());
  if (!signing_message.can_sign()) {
    printf("Signing is invalid\n");
    ret_val = 1;
    goto done;
  } 

  // Get cloudproxy key
  in_size = MAX_SIZE_PARAMS;
  if (!ReadFileIntoBlock(FLAGS_cloudproxy_key_file, &in_size, in_buf)) {
    printf("Can't read private key\n");
    printf("    %s\n", FLAGS_cloudproxy_key_file.c_str());
  }
  input.assign((const char*)in_buf, in_size);
  if (!private_key.ParseFromString(input)) {
    printf("Can't parse private key\n");
  }

  private_key_blob = private_key.blob();
  PrintBytes(private_key_blob.size(), (byte*)private_key_blob.data());
  printf("\n");
  p = (byte*)private_key_blob.data();
  signing_key = d2i_RSAPrivateKey(nullptr, (const byte**)&p,
                                  private_key_blob.size());
  if (signing_key == nullptr) {
    printf("Can't translate private key\n");
    ret_val = 1;
    goto done;
  }
#ifdef DEBUG
  print_internal_private_key(*signing_key);
#endif

  // Extract program key request
  if (!request.has_cred()) {
    printf("No information to construct cred\n");
    ret_val = 1;
    goto done;
  }

  // Get Policy cert
  if (!ReadFileIntoBlock(FLAGS_policy_cert_file, &der_policy_cert_size,
                         der_policy_cert)) {
    printf("Can't read policy cert \n");
    ret_val = 1;
    goto done;
  }

  // Get endorsement cert
  p = (byte*)request.endorsement_cert_blob().data();
  endorsement_blob_size = request.endorsement_cert_blob().size();
  endorsement_cert = d2i_X509(nullptr, (const byte**)&p, endorsement_blob_size);

  // TODO(jlm): make sure self-signed policy key signed endorsement
  if ((store = X509_STORE_new()) == nullptr) {
    printf("Can't new X509_STORE\n");
    ret_val = 1;
    goto done;
  }
  if ((verify_ctx = X509_STORE_CTX_new()) == nullptr) {
    printf("Can't new X509_STORE_CTX\n");
    ret_val = 1;
    goto done;
  }
  p_byte = der_policy_cert;
  policy_cert = d2i_X509(nullptr, (const byte**)&p_byte,
                         der_policy_cert_size);
  if (policy_cert == nullptr) {
    printf("Can't convert policy cert\n");
    ret_val = 1;
    goto done;
  }

  cert_OK = X509_verify(endorsement_cert, X509_get_pubkey(policy_cert));
  if (cert_OK <= 0) {
    printf("Endorsement cert does not verivy\n");
    ret_val = 1;
    goto done;
  }

  // Generate request for program cert
  cert_parameters.set_common_name(request.program_key().program_name());
  cert_parameters.mutable_key()->set_key_type(request.program_key().program_key_type());
  cert_parameters.mutable_key()->mutable_rsa_key()->set_bit_modulus_size(
      request.program_key().program_bit_modulus_size());
  cert_parameters.mutable_key()->mutable_rsa_key()->set_exponent(
      request.program_key().program_key_exponent());
  cert_parameters.mutable_key()->mutable_rsa_key()->set_modulus(
       request.program_key().program_key_modulus());
  print_cert_request_message(cert_parameters); printf("\n");

  req = X509_REQ_new();
  if (!GenerateX509CertificateRequest(cert_parameters, false, req)) {
    printf("Can't generate certificate request\n");
    ret_val = 1;
    goto done;
  }

  // sign program key
  if (!SignX509Certificate(signing_key, signing_message, nullptr, req,
                           false, program_cert)) {
    printf("Can't sign x509 request for program key\n");
    ret_val = 1;
    goto done;
  }
  printf("\nmessage signed\n");

  // Serialize program cert
  der_program_cert = nullptr;
  der_program_cert_size = i2d_X509(program_cert, &der_program_cert);

#ifdef DEBUG
  printf("Program cert: ");
  PrintBytes(der_program_cert_size, der_program_cert); printf("\n");
  X509_print_fp(stdout, program_cert);
  printf("\n");
#endif

  // Hash request
  serialized_program_key = request.program_key().DebugString();
  if (hash_alg_id == TPM_ALG_SHA1) {
    SHA1_Init(&sha1);
    SHA1_Update(&sha1, (byte*)serialized_program_key.data(),
                      serialized_program_key.size());
    SHA1_Final(program_key_quoted_hash, &sha1);
  } else if (hash_alg_id == TPM_ALG_SHA256) {
    SHA256_Init(&sha256);
    SHA256_Update(&sha256, (byte*)serialized_program_key.data(),
                           serialized_program_key.size());
    SHA256_Final(program_key_quoted_hash, &sha256);
  } else {
    printf("Unknown hash alg\n");
    ret_val = 1;
    goto done;
  }

#ifdef DEBUG
  printf("\nprogram_key_quoted_hash: ");
  PrintBytes(SizeHash(hash_alg_id), program_key_quoted_hash);
  printf("\n");
#endif

  // "Encrypt" with quote key
  if (!request.cred().has_public_key()) {
    printf("no quote key\n");
    ret_val = 1;
    goto done;
  }

  quote_struct_size = request.quoted_blob().size();
  memcpy(quote_struct, request.quoted_blob().data(), quote_struct_size);

#ifdef DEBUG_EXTRA
  printf("\nmodulus size: %d\n",
      (int)request.cred().public_key().rsa_key().modulus().size());
  printf("exponent size: %d\n",
      (int)request.cred().public_key().rsa_key().exponent().size());
  printf("modulus: ");
  PrintBytes(request.cred().public_key().rsa_key().modulus().size(),
             (byte*)request.cred().public_key().rsa_key().modulus().data());
  printf("\n");
  printf("exponent: ");
  PrintBytes(request.cred().public_key().rsa_key().exponent().size(),
             (byte*)request.cred().public_key().rsa_key().exponent().data());
  printf("\n");
  printf("quote_struct: ");
  PrintBytes(quote_struct_size, quote_struct);
  printf("\n");
#endif

  // Decode quote structure
  if (!UnmarshalCertifyInfo(quote_struct_size, quote_struct, &attested_quote)) {
#ifdef DEBUG_EXTRA
    print_quote_certifyinfo(attested_quote);
#endif
    printf("Invalid attested structure\n");
    ret_val = 1;
    goto done;
  }
#ifdef DEBUG_EXTRA
  print_quote_certifyinfo(attested_quote);
#endif
  if (attested_quote.magic !=  TpmMagicConstant) {
    printf("Invalid magic number\n");
    ret_val = 1;
    goto done;
  }

  if(!ValidPCR(attested_quote.attested.quote.pcrSelect.pcrSelections[0].hash,
               &attested_quote.attested.quote.pcrSelect.pcrSelections[0].sizeofSelect,
               attested_quote.attested.quote.pcrDigest.buffer)) {
    printf("Invalid pcr\n");
    ret_val = 1;
    goto done;
  }

  // Set quote key exponent and modulus
  active_key->n = bin_to_BN(
      request.cred().public_key().rsa_key().modulus().size(),
      (byte*)request.cred().public_key().rsa_key().modulus().data());
  active_key->e = bin_to_BN(
      request.cred().public_key().rsa_key().exponent().size(),
      (byte*)request.cred().public_key().rsa_key().exponent().data());
  size_active_out = RSA_public_encrypt(request.active_signature().size(),
                        (const byte*)request.active_signature().data(),
                        decrypted_quote, active_key, RSA_NO_PADDING);
  if (size_active_out > MAX_SIZE_PARAMS) {
    printf("active signature is too big\n");
    ret_val = 1;
    goto done;
  }
  signed_quote_hash_size = MAX_SIZE_PARAMS;
  if (!ComputeQuotedValue(hash_alg_id, quote_struct_size, quote_struct,
                          &signed_quote_hash_size, signed_quote_hash)) {
    printf("Cant compute ComputeQuotedValue\n");
    ret_val = 1;
    goto done;
    }

  // Check hash of request
  if (memcmp(attested_quote.extraData.buffer, program_key_quoted_hash, 
             attested_quote.extraData.size) != 0) {
    printf("Program key hash does not match\n");
    ret_val = 1;
    goto done;
  }

#ifdef DEBUG
  printf("\nactive signature size: %d\n", size_active_out);
  printf("Quote structure: ");
  PrintBytes(quote_struct_size, quote_struct);
  printf("\n");
  printf("Quote hash: ");
  PrintBytes(signed_quote_hash_size, signed_quote_hash);
  printf("\n");
  printf("Decrypted hash: ");
  PrintBytes(size_active_out, decrypted_quote);
  printf("\n");
#endif

  // recover pcr hash and magic number and check them
  // Compare signature and computed hash
  if (memcmp(signed_quote_hash,
             decrypted_quote + size_active_out - SizeHash(hash_alg_id),
             SizeHash(hash_alg_id)) != 0) {
    printf("quote signature is wrong\n");
    PrintBytes(SizeHash(hash_alg_id), signed_quote_hash); printf("\n");
    PrintBytes(size_active_out, decrypted_quote); printf("\n");
    ret_val = 1;
    goto done;
  }

  // Generate encryption key for signed program cert
  // This is the "credential."
  unmarshaled_credential.size = 16;
  RAND_bytes(unmarshaled_credential.buffer, unmarshaled_credential.size);

  ChangeEndian16(&unmarshaled_credential.size, &marshaled_credential.size);
  memcpy(marshaled_credential.buffer, unmarshaled_credential.buffer,
         unmarshaled_credential.size);

  // Encrypt signed program cert
  size_derived_keys = 128;
  label = "PROTECT";
  cert_key_seed.assign((const char*)unmarshaled_credential.buffer,
                       unmarshaled_credential.size);
  if (!KDFa(hash_alg_id, cert_key_seed, label, contextV, contextV, 256,
            size_derived_keys, derived_keys)) {
    printf("Can't derive cert protection keys\n");
    ret_val = 1;
    goto done;
  }
  if (!AesCtrCrypt(128, derived_keys, der_program_cert_size,
                   der_program_cert, encrypted_data)) {
    printf("Can't encrypt cert\n");
    ret_val = 1;
    goto done;
  }
  HMAC_CTX_init(&hctx);
  if (hash_alg_id == TPM_ALG_SHA1) {
    HMAC_Init_ex(&hctx, &derived_keys[16], 16, EVP_sha1(), nullptr);
    encrypted_data_hmac_size = 20;
  } else {
    HMAC_Init_ex(&hctx, &derived_keys[16], 16, EVP_sha256(), nullptr);
    encrypted_data_hmac_size = 32;
  }
  HMAC_Update(&hctx, encrypted_data, der_program_cert_size);
  HMAC_Final(&hctx, encrypted_data_hmac, (uint32_t*)&encrypted_data_hmac_size);
  HMAC_CTX_cleanup(&hctx);

#ifdef DEBUG_EXTRA
  printf("\ncredential secret: ");
  PrintBytes(16, unmarshaled_credential.buffer);
  printf("\n");
  printf("\nderived keys: ");
  PrintBytes(32, derived_keys);
  printf("\n");
  printf("\nder_program_cert: ");
  PrintBytes(der_program_cert_size, der_program_cert);
  printf("\n");
  printf("\nencrypted der_program_cert: ");
  PrintBytes(der_program_cert_size, encrypted_data);
  printf("\n");
  if (!AesCtrCrypt(128, derived_keys, der_program_cert_size,
                   encrypted_data, test_buf)) {
    printf("\nCan't decrypt cert\n");
    ret_val = 1;
    goto done;
  }
  printf("\ndecrypted der_program_cert: ");
  PrintBytes(der_program_cert_size, test_buf);
  printf("\n");
  printf("\nencrypted_dert_hmac: ");
  PrintBytes(encrypted_data_hmac_size, encrypted_data_hmac);
  printf("\n");
#endif
  encrypted_data_size = der_program_cert_size;
  response.set_encrypted_cert(encrypted_data, encrypted_data_size);
  response.set_encrypted_cert_hmac(encrypted_data_hmac, encrypted_data_hmac_size);

  // Generate seed for MakeCredential protocol
  RAND_bytes(seed, size_seed);

#ifdef DEBUG
  printf("\nseed: ");
  PrintBytes(size_seed, seed); printf("\n");
#endif

  // Protector_key is endorsement key
  protector_evp_key = X509_get_pubkey(endorsement_cert);
  protector_key = EVP_PKEY_get1_RSA(protector_evp_key);
  RSA_up_ref(protector_key);

  // Secret= E(protector_key, seed || "IDENTITY")
  //   args: to, from, label, len
  size_in = 256;
  RSA_padding_add_PKCS1_OAEP(in_buf, 256, seed, size_seed, 
      (byte*)"IDENTITY", strlen("IDENTITY")+1);

#ifdef DEBUG
  printf("After RSAPad: ");
  PrintBytes(size_in, in_buf);
  printf("\n");
#endif

  encrypted_secret_size = RSA_public_encrypt(size_in, in_buf, encrypted_secret,
                              protector_key, RSA_NO_PADDING);
                              // RSA_PKCS1_OAEP_PADDING);
  response.set_secret(encrypted_secret, encrypted_secret_size);

#ifdef DEBUG_EXTRA
  printf("\nEndorsement modulus: ");
  BN_print_fp(stdout, protector_key->n);
  printf("\n");
  printf("endorsement exponent: ");
  BN_print_fp(stdout, protector_key->e);
  printf("\n");
  printf("\nEncrypted secret (%d): ", encrypted_secret_size);
  PrintBytes(encrypted_secret_size, encrypted_secret); printf("\n");
  printf("\nname: "); 
  PrintBytes(request.cred().name().size(),
             (byte*)request.cred().name().data()); printf("\n");
  printf("\n");
#endif

  // Calculate symKey
  label = "STORAGE";
  key.assign((const char*)seed, size_seed);
  contextV.clear();
  name.assign(request.cred().name().data(), request.cred().name().size());
  if (!KDFa(hash_alg_id, key, label, name, contextV, 128, 32, symKey)) {
    printf("Can't KDFa symKey\n");
    ret_val = 1;
    goto done;
  }

#ifdef DEBUG
  printf("\nsymKey: "); PrintBytes(16, symKey); printf("\n");
  printf("marshaled_credential: ");
  PrintBytes(unmarshaled_credential.size + sizeof(uint16_t),
             (byte*)&marshaled_credential);
  printf("\n");
#endif

  // Calculate encIdentity
  //   Note: We need to encrypt the entire marshaled_credential.
  size_encIdentity = MAX_SIZE_PARAMS;
  if (!AesCFBEncrypt(symKey, unmarshaled_credential.size + sizeof(uint16_t),
                     (byte*)&marshaled_credential, 16, zero_iv,
                     &size_encIdentity, encIdentity)) {
    printf("Can't AesCFBEncrypt\n");
    ret_val = 1;
    goto done;
  }

#ifdef DEBUG
  printf("size_encIdentity: %d\n", size_encIdentity);
  test_size = MAX_SIZE_PARAMS;
  if (!AesCFBDecrypt(symKey, size_encIdentity,
                     (byte*)encIdentity, 16, zero_iv,
                     &test_size, test_buf)) {
    printf("Can't AesCFBDecrypt\n");
    ret_val = 1;
    goto done;
  }
  printf("Decrypted secret (%d): ", test_size);
  PrintBytes(test_size, test_buf); printf("\n");
#endif

  response.set_encidentity(encIdentity, size_encIdentity);

  // hmacKey= KDFa(hash, seed, "INTEGRITY", nullptr, nullptr, 8*hashsize);
  label = "INTEGRITY";
  if (!KDFa(hash_alg_id, key, label, contextV,
            contextV, 8 * SizeHash(hash_alg_id), 32, hmacKey)) {
    printf("Can't KDFa hmacKey\n");
    ret_val = 1;
    goto done;
  }
  
  // Calculate outerMac = HMAC(hmacKey, encIdentity || name);
  unmarshaled_name.size = name.size();
  ChangeEndian16(&unmarshaled_name.size, &marshaled_name.size);
  memcpy(unmarshaled_name.name, (byte*)name.data(), unmarshaled_name.size);
  memcpy(marshaled_name.name, (byte*)name.data(), unmarshaled_name.size);

  HMAC_CTX_init(&hctx);
  if (hash_alg_id == TPM_ALG_SHA1) {
    HMAC_Init_ex(&hctx, hmacKey, size_hmacKey, EVP_sha1(), nullptr);
  } else {
    HMAC_Init_ex(&hctx, hmacKey, size_hmacKey, EVP_sha256(), nullptr);
  }
  HMAC_Update(&hctx, (const byte*)encIdentity, size_encIdentity);
  HMAC_Update(&hctx, (const byte*)&marshaled_name.name, name.size());
  HMAC_Final(&hctx, outerHmac, (uint32_t*)&size_hmacKey);
  HMAC_CTX_cleanup(&hctx);

  // integrityHMAC
  ChangeEndian16((uint16_t*)&size_hmacKey, &marshaled_integrityHmac.size);
  memcpy(marshaled_integrityHmac.buffer, outerHmac, size_hmacKey);
  response.set_integrityhmac((byte*)&marshaled_integrityHmac,
                             size_hmacKey + sizeof(uint16_t));

  // Serialize output
  response.SerializeToString(&output);
  if (!WriteFileFromBlock(FLAGS_program_response_file,
                          output.size(),
                          (byte*)output.data())) {
    printf("Can't write endorsement cert\n");
    ret_val = 1;
    goto done;
  }

done:
  return ret_val;
}


