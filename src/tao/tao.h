//  File: tao.h
//  Author: Tom Roeder <tmroeder@google.com>
//
//  Description: Interface used by hosted programs to access Tao services.
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

#include <string>

namespace tao {
using std::string;

/// Tao is the fundamental Trustworthy Computing interface provided by a host to
/// its hosted programs. Each level of a system can act as a host by exporting
/// the Tao interface and providing Tao services to higher-level hosted
/// programs.
///
/// In most cases, a hosted program will use a stub Tao that performs RPC over a
/// channel to its host. The details of such RPC depend on the specific
/// implementation of the host: some hosted programs may use pipes to
/// communicate with their host, others may use sockets, etc.
class Tao {
 public:
  Tao() {}
  virtual ~Tao() {}

  /// Serialize Tao parameters for passing across fork/exec or between
  /// processes, if possible. Not all Tao implementations must necessarily be
  /// serializable.
  /// @param params[out] The serialized parameters.
  virtual bool SerializeToString(string *params) const { return false; }

  /// Get the Tao principal name assigned to this hosted program. The name
  /// encodes the full path from the root Tao, through all intermediary Tao
  /// hosts, to this hosted program. The name will be globally unique: different
  /// hosted program (for some definition of "different") will be given
  /// different Tao names.
  /// @param[out] name The full, globally-unique name of this hosted program.
  virtual bool GetTaoName(string *name) = 0;

  /// Irreversibly extend the Tao principal name of this hosted program. In
  /// effect, the hosted program can drop privileges by taking on the identity
  /// of its subprincipal.
  /// @param subprin The subprincipal to append to the principal name.
  virtual bool ExtendTaoName(const string &subprin) = 0;

  /// Get random bytes.
  /// @param size The number of bytes to get.
  /// @param[out] bytes The random bytes.
  virtual bool GetRandomBytes(size_t size, string *bytes) = 0;

  /// Get a shared random secret, e.g. to be used as a shared secret key.
  /// Some levels (i.e. the TPM) do not implement this.
  /// @param size The number of bytes to get.
  /// @param policy A policy controlling which hosted programs can get this
  /// secret. The semantics of this value are host-specific, except
  /// all Tao hosts support at least the policies defined below.
  /// @param[out] sealed The encrypted data.
  virtual bool GetSharedSecret(size_t size, const string &policy,
                               string *bytes) = 0;

  /// Policy for shared secrets. Hosts may implement additional policies.
  /// @{

  /// The default shared secret, which corresponds roughly to "a past or future
  /// instance of a hosted program having a similar identity as the caller". The
  /// definition of "similar" is host-specific. For example, for a Linux OS, it
  /// may mean "has the same program binary (and is running on the same Linux OS
  /// platform)".
  constexpr static auto SharedSecretPolicyDefault = "self";

  /// The most conservative (but non-trivial) shared secret policy supported by
  /// the
  /// host. For example, a Linux OS may interpret this to mean "the same hosted
  /// program instance, including process ID, program hash and argument hash".
  constexpr static auto SharedSecretPolicyConservative = "few";

  /// The most liberal (but non-trivial) sealing policy supported by the host.
  /// For example, a Linux OS may interpret this to mean "any hosted program on
  /// the
  /// same Linux OS".
  constexpr static auto SharedSecretPolicyLiberal = "any";

  /// @}

  /// Request the Tao host sign a Statement on behalf of this hosted program.
  /// @param message A delegation statement to be signed. This statement is a
  /// serialized binary auth statement and can be created using MarshalSpeaksfor
  /// in util.h.
  /// @param[out] attestation The resulting signed attestation.
  virtual bool Attest(const string &message, string *attestation) = 0;

  /// Encrypt data so only certain hosted programs can unseal it.
  /// @param data The data to seal.
  /// @param policy A policy controlling which hosted programs can seal or
  /// unseal the data. The semantics of this value are host-specific, except:
  /// all Tao hosts support at least the policies defined below; and the policy
  /// must be satisfied both during Seal() and during Unseal().
  /// @param[out] sealed The encrypted data.
  /// TODO(kwalsh) Add expiration.
  virtual bool Seal(const string &data, const string &policy,
                    string *sealed) = 0;

  /// Decrypt data that has been sealed by the Seal() operation, but only
  /// if the policy specified during the Seal() operation is satisfied.
  /// @param sealed The sealed data to decrypt.
  /// @param[out] data The decrypted data, if the policy was satisfied.
  /// @param[out] policy The sealing policy, if it was satisfied.
  /// Note: The returned policy can be used as a limited integrity check, since
  /// only a hosted program that itself satisfies the policy could have
  /// performed the Seal() operation.
  virtual bool Unseal(const string &sealed, string *data, string *policy) = 0;

  // InitCounter initializes the rollback counter.
  virtual bool InitCounter(const string &label, int64_t& c) = 0;

  // GetCounter retrieved the rollback counter.
  virtual bool GetCounter(const string &label, int64_t* c) = 0;

  // RollbackProtectedSeal does a rollback protected seal.
  virtual bool RollbackProtectedSeal(const string& label, const string &data, const string &policy, string *sealed) = 0;

  // RollbackProtectedUnseal does a rollback protected unseal.
  virtual bool RollbackProtectedUnseal(const string &sealed, string *data, string *policy) = 0;

  /// Policy for sealing and unsealing. Hosts may implement additional policies.
  /// @{

  /// The default sealing policy, which corresponds roughly to "a past or future
  /// instance of a hosted program having a similar identity as the caller". The
  /// definition of "similar" is host-specific. For example, for a TPM, it may
  /// mean "has identical PCR values, for some subset of the PCRs". And for a
  /// Linux OS, it may mean "has the same program binary (and is running on the
  /// same Linux OS platform)".
  constexpr static auto SealPolicyDefault = "self";

  /// The most conservative (but non-trivial) sealing policy supported by the
  /// host. For example, a Linux OS may interpret this to mean "the same hosted
  /// program instance, including process ID, program hash and argument hash".
  constexpr static auto SealPolicyConservative = "few";

  /// The most liberal (but non-trivial) sealing policy supported by the host.
  /// For example, a TPM may interpret this to mean "any hosted program on the
  /// same platform". And a Linux OS might interpret this to mean "any hosted
  /// program on the same Linux OS".
  constexpr static auto SealPolicyLiberal = "any";

  /// @}

  /// Get most recent error message, or emptystring if there have been no
  /// errors.
  virtual string GetRecentErrorMessage() const = 0;

  // Clear the most recent error message and return the previous value, if any.
  virtual string ResetRecentErrorMessage() = 0;

  /// A context string for signed attestations.
  constexpr static auto AttestationSigningContext =
      "tao::Attestation Version 1";

  /// Default timeout for Attestation (= 1 year in seconds).
  static const int DefaultAttestationTimeout = 31556926;
};
}  // namespace tao

#endif  // TAO_TAO_H_
