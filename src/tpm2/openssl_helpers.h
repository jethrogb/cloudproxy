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

#ifndef __OPENSSL_HELPERS__
#define __OPENSSL_HELPERS__
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
#include <openssl/x509.h>
#include <openssl/x509v3.h>

#include <string>
using std::string;

bool GenerateX509CertificateRequest(x509_cert_request_parameters_message& params, X509_REQ* req);
bool GetPublicRsaParametersFromSSLKey(RSA& rsa, public_key_message* key_msg);
bool GetPrivateRsaParametersFromSSLKey(RSA& rsa, private_key_message* key_msg);
bool SignX509CertificateRequest(RSA& rsa, signing_instructions_message& signing_message, x509_req* req, X509* cert);
bool VerifyX509CertificateChain(certificate_chain_message& chain);
bool GetCertificateRequestParametersFromX509(X509_REQ& x509_req, cert_parameters* cert_params);
bool GetCertificateParametersFromX509(X509& x509_cert, cert_parameters* cert_params);
bool GetPublicRsaKeyFromParametersFromParameters(rsa_public_key_message& key_msg, RSA* rsa);
bool GetPrivateRsaKeyFromParameters(rsa_private_key_message& key_msg, RSA* rsa);
#endif

