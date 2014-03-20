//  File: unix_domain_socket_tao_child_channel.h
//  Author: Kevin Walsh <kwalsh@holycross.edu>
//
//  Description: A child channel class that communicates over unix domain
//  sockets.
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

#ifndef TAO_UNIX_DOMAIN_SOCKET_TAO_CHILD_CHANNEL_H_
#define TAO_UNIX_DOMAIN_SOCKET_TAO_CHILD_CHANNEL_H_

#include <string>

#include "tao/tao_child_channel.h"
#include "tao/util.h"

namespace tao {
/// A child channel that communicates with a host Tao over unix domain sockets.
class UnixDomainSocketTaoChildChannel : public TaoChildChannel {
 public:
  /// Construct a new tao child channel.
  /// @param host_socket_path The location of the host server socket.
  explicit UnixDomainSocketTaoChildChannel(const string &host_socket_path);

  virtual ~UnixDomainSocketTaoChildChannel() {}

  virtual bool Init();
  virtual bool Destroy();

 protected:
  /// The location of the host server socket.
  const string host_socket_path_;

  /// The socket for read and write.
  ScopedFd sock_;

  virtual bool ReceiveMessage(google::protobuf::Message *m) const;
  virtual bool SendMessage(const google::protobuf::Message &m) const;

 private:
  DISALLOW_COPY_AND_ASSIGN(UnixDomainSocketTaoChildChannel);
};
}  // namespace tao

#endif  // TAO_UNIX_DOMAIN_SOCKET_TAO_CHILD_CHANNEL_H_