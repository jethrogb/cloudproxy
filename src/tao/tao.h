//  File: tao.h
//  Author: Tom Roeder <tmroeder@google.com>
//
//  Description: The Tao interface for Trusted Computing
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

#ifndef TAO_TAO_H_
#define TAO_TAO_H_

#include <list>
#include <string>

#include <keyczar/base/basictypes.h>  // DISALLOW_COPY_AND_ASSIGN

using std::list;
using std::string;

namespace tao {

/// The Tao is the fundamental interface for Trustworthy Computing in
/// CloudProxy. Each level of a system can implement a Tao interface and provide
/// Tao services to higher-level hosted programs.
///
/// For example, a Linux system installed on hardware with a TPM might work as
/// follows: TPMTaoChildChannel <-> LinuxTao <-> PipeTaoChannel. The
/// TPMTaoChildChannel implements a shim for the TPM hardware to convert Tao
/// operations into TPM commands. LinuxTao implements the Tao for Linux, and it
/// holds a PipeTaoChannel that it uses to communicate with hosted programs
/// running as processes. A hosted program called CloudServer would have the
/// following interactions: PipeTaoChildChannel <-> CloudServer. The
/// PipeTaoChildChannel and the PipeTaoChannel communicate over Unix pipes to
/// send Tao messages between LinuxTao and CloudServer. See the apps/ folder for
/// applications that implement exactly this setup: apps/linux_tao_service.cc
/// implements the LinuxTao, and apps/server.cc implements CloudServer.
///
/// Similarly, the LinuxTao could start KVM Guests as hosted programs
/// (using the KvmVmFactory instead of the ProcessFactory). In this case, the
/// interaction would be: TPMTaoChildChannel <-> LinuxTao <-> KvmUnixTaoChannel.
///
/// And the guest OS would have another instance of the LinuxTao that would have
/// the following interactions:
/// KvmUnixTaoChildChannel <-> LinuxTao <-> PipeTaoChannel. This version of
/// the LinuxTao in the Guest OS would use the ProcessFactory to start hosted
/// programs as processes in the guest.
///
/// In summary: each level of the Tao can have a TaoChildChannel to communicate
/// with its host Tao and has a TaoChannel to communicated with hosted programs.
/// Hosts use implementations of HostedProgramFactory to instantiate hosted
/// programs.
class Tao {
 public:
  Tao() {}
  virtual ~Tao() {}

  /// Initialize the Tao.
  virtual bool Init() = 0;

  /// Clean up an resources that were allocated in Init().
  virtual bool Destroy() = 0;

  /// Start a hosted program with a given set of arguments. These arguments
  /// might not always be the arguments passed directly to a process. For
  /// example, they might be kernel, initrd, and disk for starting a virtual
  /// machine.
  /// @param name The name of the hosted program. This can sometimes be a path
  /// to a process, but it is not always.
  /// @param args A list of arguments for starting the hosted program.
  /// @param identifier An identifier for the started program (e..g, a PID)
  virtual bool StartHostedProgram(const string &name, const list<string> &args,
                                  string *identifier) = 0;

  /// Remove the hosted program from the running programs. Note that this does
  /// not necessarily stop the hosted program itself.
  /// @param child_hash The hash of the program to remove.
  virtual bool RemoveHostedProgram(const string &child_hash) = 0;

  /// Get a random string of a given size.
  /// @param size The size of string to get, in bytes.
  /// @param[out] bytes The random bytes generated by this call.
  virtual bool GetRandomBytes(size_t size, string *bytes) const = 0;

  /// Encrypt the data in a way that prevents any other hosted program from
  /// reading it. The identity of the hosted program is specified by child_hash.
  /// @param child_hash The identity of the hosted program that should be able
  /// to access the data.
  /// @param data The data to seal.
  /// @param[out] sealed The result of the sealing operation.
  virtual bool Seal(const string &child_hash, const string &data,
                    string *sealed) const = 0;

  /// Decrypt data that has been sealed by the Seal() operation. This decryption
  /// should only succeed if the hash of the program matches child_hash.
  /// @param child_hash A hosted-program hash to compare against the sealed
  /// data.
  /// @param sealed The sealed data to decrypt.
  /// @param[out] data The decrypted data, if the hash check passes.
  virtual bool Unseal(const string &child_hash, const string &sealed,
                      string *data) const = 0;

  /// Produce a signed statement that asserts that a given program produced a
  /// given data string.
  /// @param child_hash The identity of the program to attest to. The caller
  /// must ensure that this is the identity of the program making the request.
  /// See subclasses of TaoChannel for an implementation.
  /// @param data The data produced by the hosted program.
  /// @param[out] attestation The resulting signed message. For verification see
  /// TaoAuth and its implementations.
  virtual bool Attest(const string &child_hash, const string &data,
                      string *attestation) const = 0;

  constexpr static auto AttestationSigningContext =
      "tao::Attestation Version 1";

  // the timeout for an Attestation (= 1 year in seconds)
  const static int DefaultAttestationTimeout = 31556926;

 private:
  DISALLOW_COPY_AND_ASSIGN(Tao);
};
}

#endif  // TAO_TAO_H_
