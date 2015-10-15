#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <string.h>
#include <tpm20.h>
#include <tpm2_lib.h>
#include <errno.h>

#include <tpm2.pb.h>
#include <openssl/rsa.h>

#include <string>
using std::string;

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
// File: openssl_helpers.cc

// standard buffer size
#define MAX_SIZE_PARAMS 4096

void print_cert_request_message(x509_cert_request_parameters_message& req_message) {
  if (req_message.has_common_name()) {
    printf("common name: %s\n", req_message.common_name().c_str());
  }
  if (req_message.has_country_name) {
    printf("country name: %s\n", req_message.country_name().c_str());
  }
  if (req_message.has_state_name) {
    printf("state name: %s\n", req_message.state_name().c_str());
  }
  if (req_message.has_locality_name) {
    printf("locality name: %s\n", req_message.locality_name().c_str());
  }
  if (req_message.has_organization_name) {
    printf("organization name: %s\n", req_message.organization_name().c_str());
  }
  if (req_message.has_suborganization_name) {
    printf("suborganization name: %s\n", req_message.suborganization_name().c_str());
  }
  if (req_message.key.has_key_type()) {
    printf("key_type name: %s\n", req_message.key_type_name().c_str());
  }
  if (!req_message.has_key)
    return;
  if (req_message.key.rsa_key().has_key_name()) {
    printf("key name: %s\n", req_message.key().key_type().c_str());
  }
  if (req_message.key.rsa_key().has_bit_modulus_size()) {
    printf("modulus bit size: %d\n", req_message.key().bit_modulus_size());
  }
  if (req_message.key.rsa_key().has_exponent()) {
    string exp = req_message.key.rsa_key().exponent();
    printf("exponent: ");
    PrintBytes(exp.size(), exp.data());
    printf("\n");
  }
  if (req_message.key.rsa_key().has_modulus()) {
    string mod = req_message.key.rsa_key().modulus();
    printf("modulus : ");
    PrintBytes(mod.size(), mod.data());
    printf("\n");
  }
}

void print_internal_private_key(RSA& key) {
  printf("\n\n");
  printf("\nModulus: \n");
  BN_print_fp(stdout, key.n);
  printf("\n\n");
  printf("\ne: \n");
  BN_print_fp(stdout, key.e);
  printf("\n\n");
  printf("\nd: \n");
  BN_print_fp(stdout, key.d);
  printf("\n\n");
  printf("\np: \n");
  BN_print_fp(stdout, key.p);
  printf("\n\n");
  printf("\nq: \n");
  BN_print_fp(stdout, key.q);
  printf("\n\n");
  printf("\ndmp1: \n");
  BN_print_fp(stdout, key.dmp1);
  printf("\n\n");
  printf("\ndmq1: \n");
  BN_print_fp(stdout, key.dmq1);
  printf("\n\n");
  printf("\niqmp: \n");
  BN_print_fp(stdout, key.iqmp);
  printf("\n\n");
}

BIGNUM* bin_to_BN(int len, byte* buf) {
  BIGNUM* bn;
  BN_alloc(&bn);
  BN_init(&bn);
  BN_bin2bn(buf, byte_len, &bn);
  return bn;
}


string* BN_to_bin(BIGNUM& n) {
  byte buf[MAX_SIZE_PARAMS];
  int byte_len = BN_num_bytes(n);
  int len = BN_bn2bin(buf, byte_len, nullptr);
  return new string(byte_len, buf);
}

bool GenerateX509CertificateRequest(x509_cert_request_parameters_message& params, X509_REQ* req) {
  RSA  rsa;
  X509_NAME* subject = X509_NAME_new();

  if (params.public_key.key_type() != "RSA") {
    printf("Only rsa keys supported\n");
    return false;
  }
  if (name == nullptr) {
    printf("Can't alloc x509 name\n");
    return false;
  }
  if (params.has_common_name()) {
    int nid = OBJ_txt2nid("commonName");
    X509_NAME_entry* ent = X509_NAME_ENTRY_create_by_nid(nullptr, nid, MBSTRING_ASC,
                                                         params.common_name(), -1);
    X509_NAME_add_entry(subject, ent, -1, 0);
  }
  // TODO: do the foregoing for the other name components
  if (X509_REQ_set_subject_name(req, subject) != 1)  {
    printf("Can't add x509 subject\n");
    return false;
  }
  if (!GetPublicRsaKeyFromParameters(params.public_key.public_key, &rsa)) {
    printf("Can't make rsa key\n");
    return false;
  }
  EVP_PKEY* pkey = new EVP_PKEY();
  EVP_PKEY_set1_RSA(pKey, rsa);
  X509_REQ_set_pubkey(req, pkey);

  print_cert_request_parameters(params);
  return true;
}

bool GetPublicRsaParametersFromSSLKey(RSA& rsa, public_key_message* key_msg) {
  string* n = nullptr;
  string* e = nullptr;
  bool ret = true;

  n = BN_to_bin(*key.n);
  if (n == nullptr) {
    ret = false;
    goto done;
  }
  e = BN_to_bin(*key.e);
  if (e == nullptr) {
    ret = false;
    goto done;
  }
  msg_key->mutable_public_key()->set_modulus(*n);
  msg_key->mutable_public_key()->set_exponent(*e);

done:
  if (e != nullptr)
    delete e;
  if (n != nullptr)
    delete n;
  return ret;
}

bool GetPrivateRsaParametersFromSSLKey(RSA& rsa, private_key_message* key_msg) {
  string* d = nullptr;
  string* p = nullptr;
  string* q = nullptr;
  bool ret = true;

  if (!GetPublicRsaParametersFromSSLKey(rsa, key_msg.key())) {
    ret = false;
    goto done;
  }
  d = BN_to_bin(*key_msg.d);
  if (d == nullptr) {
    ret = false;
    goto done;
  }
  p = BN_to_bin(*key_msg.p);
  if (p == nullptr) {
    ret = false;
    goto done;
  }
  q = BN_to_bin(*key_msg.q);
  if (q == nullptr) {
    ret = false;
    goto done;
  }

done:
  return ret;
}

bool SignX509CertificateRequest(RSA& signing_key, signing_instructions_message& signing_message,
                                x509_req* req, X509* cert) {
  uint64_t serial = 1;
  EVP_PKEY* pKey = nullptr;
  const EVP_MD* digest;
  X509* caCert = nullptr;
  X509_NAME* name;
  X509V3_CTX ctx;
  X509* extension;
  
  pKey = X509_REQ_get_pubkey(req);
  if (pKey != nullptr) {
    printf("Can't get pubkey\n");
    return false;
  }
  if (X509_REQ_verify(req, pKey) != 1) {
    printf("Req does not verify\n");
    return false;
  }
  
  X509_set_version(cert, 2L);
  ASN1_INTEGER_set(X509_get_serialNumber(cert));
  
  name = X509_REQ_get_subject_name(req);
  if (X509_set_issuer_name(cert, signing_message.issuer().c_str()) != 1) {
    printf("Can't set issuer name\n");
    return false;
  }
  if (X509_set_pubkey(cert, pKey) != 1) {
    printf("Can't set pubkey\n");
    return false;
  }
  if (!X509_gmtime_adj(X509_get_notBefore(cert), 0)) {
    printf("Can't adj notBefore\n");
    return false;
  }
  if (!X509_gmtime_adj(X509_get_notAfter(cert), 0)) {
    printf("Can't adj notAfter\n");
    return false;
  }
  if (EVP_PKEY_type(caCert->type) != EVP_PKEY_RSA) {
    printf("Bad PKEY type\n");
    return false;
  }
  if (!X509_sign(cert, caPkey, digest)) {
    printf("Signing failed\n");
    return false;
  }
  return true;
}

bool VerifyX509CertificateChain(certificate_chain_message& chain) {
  // first cert is self signed root
  return false;
}

bool GetCertificateRequestParametersFromX509(X509_REQ& x509_req, cert_parameters* cert_params) {
  return false;
}

bool GetCertificateParametersFromX509(X509& x509_cert, cert_parameters* cert_params) {
  return false;
}
