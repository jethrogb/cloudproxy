//  File: cloud_client.cc
//  Author: Tom Roeder <tmroeder@google.com>
//
// Description: Client-side stub that communicates with CloudServer instances
// over a TLS+Tao authenticated channel.
//
//  Copyright (c) 2013, Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
#include "cloudproxy/cloud_client.h"

#include <glog/logging.h>

#include "cloudproxy/cloud_channel.h"
#include "tao/keys.h"
#include "tao/tao.h"

using tao::Tao;
using tao::Keys;

namespace cloudproxy {

bool CloudClient::Init() {
  Tao *host_tao = Tao::GetHostTao();
  if (host_tao == nullptr) {
    LOG(ERROR) << "No host tao available";
    return false;
  }
  if (tls_key_.get() == nullptr) {
    tls_key_.reset(new Keys("CloudClient", Keys::Signing));
    if (!tls_key_->InitTemporaryHosted(host_tao)) {
      LOG(ERROR) << "Could not initialize CloudClient keys";
      return false;
    }
  }
  if (tls_self_cert_.empty()) {
    // x509 details are mostly not used by peers, so we use arbitrary constants
    // here. However, commonname must match the Key nickname, above.
    string nickname = tao::quotedString(tls_key_->Nickname());
    string details = string("country: \"US\" "
                            "state: \"Washington\" "
                            "organization: \"Google\" ") +
                     "commonname: " + nickname;
    if (!tls_key_->CreateSelfSignedX509(details, &tls_self_cert_)) {
      LOG(ERROR) << "Could not create self signed x509";
      return false;
    }
  }
  if (tls_delegation_.empty()) {
    if (!tls_key_->GetHostDelegation(&tls_delegation_)) {
      LOG(ERROR) << "Could not load delegation for attestation key";
      return false;
    }
  }
  if (!SetUpSSLClientCtx(*tls_key_, tls_self_cert_, &tls_context_)) {
    LOG(ERROR) << "Could not set up client TLS";
    return false;
  }
  return true;
}

bool CloudClient::Connect(const string &server, const string &port) {
  int sock = -1;
  if (!tao::ConnectToTCPServer(server, port, &sock)) {
    LOG(ERROR) << "Could not connect to the server at " << server << ":"
               << port;
    return false;
  }
  scoped_ptr<CloudChannel> chan(new CloudChannel(tls_context_.get(), sock));
  if (!chan->TLSClientHandshake()) {
    LOG(ERROR) << "TLS handshake failed";
    return false;
  }
  if (!chan->TaoHandshake(tls_delegation_)) {
    LOG(ERROR) << "Tao handshake failed";
    return false;
  }
  chan_.reset(chan.release());
  return true;
}
}  // namespace cloudproxy
