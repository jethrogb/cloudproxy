//  Copyright (c) 2015, Jethro G. Beekman.  All rights reserved.
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

/**
 * C bindings for the CloudProxy Tao
 *
 * Conventions:
 *   Naming functions:
 *     namespaces are UpperCamelCased and separated by underscores
 *     methods are separated by underscores
 *     constructors are treated as a function called `new'
 *     destructors are treated as a function called `delete'
 *     type-casting functions are declared as:
 *       Type2* Namespace_Type1_as_Type2(Type1* obj)
 *   Naming types:
 *     classes are defined as opaque structs of type struct Class_s
 *     namespaces are UpperCamelCased and not separated
 *   Converting types:
 *     bools are ints with value 0 or 1; conversion either way should happen with ternary operators
 *     strings are a tuple of (char*, size_t); the caller is responsible for freeing a non-NULL out parameter in a char**
 */

#include <stddef.h> //for size_t

#ifdef __cplusplus
extern "C" {
#endif //__cplusplus

// tao::MessageChannel
typedef struct TaoMessageChannel_s TaoMessageChannel;

// tao::FDMessageChannel
typedef struct TaoFDMessageChannel_s TaoFDMessageChannel;
TaoFDMessageChannel* Tao_FDMessageChannel_new(int readfd, int writefd);
TaoMessageChannel* Tao_FDMessageChannel_as_MessageChannel(TaoFDMessageChannel *obj);

// tao::Tao
typedef struct TaoTao_s TaoTao;
void Tao_Tao_delete(TaoTao *obj);
int Tao_Tao_GetTaoName(TaoTao *obj, char **name, size_t *name_size);
int Tao_Tao_GetRandomBytes(TaoTao *obj, size_t size, char **bytes, size_t *bytes_size);
int Tao_Tao_Attest(TaoTao *obj, const char *message, size_t message_size, char **attestation, size_t *attestation_size);
int Tao_Tao_Seal(TaoTao *obj, const char *data, size_t data_size, const char *policy, size_t policy_size, char **sealed, size_t *sealed_size);
int Tao_Tao_Unseal(TaoTao *obj, const char *sealed, size_t sealed_size, char **data, size_t *data_size, char **policy, size_t *policy_size);

#define Tao_Tao_SealPolicyDefault "self"
#define Tao_Tao_SealPolicyConservative "few"
#define Tao_Tao_SealPolicyLiberal "any"
#define Tao_Tao_AttestationSigningContext "tao::Attestation Version 1"

// tao::TaoRPC
typedef struct TaoTaoRPC_s TaoTaoRPC;
TaoTaoRPC* Tao_TaoRPC_new(TaoMessageChannel *channel);
TaoTao* Tao_TaoRPC_as_Tao(TaoTaoRPC *obj);

// tao::InitializeApp
int Tao_InitializeApp(int *argc, char ***argv, int remove_args);

// tao::MarshalSpeaksfor
int Tao_MarshalSpeaksfor(const char *key, size_t key_size, const char *binaryTaoName, size_t binaryTaoName_size, char **out, size_t *out_size);

#ifdef __cplusplus
} // extern "C"
#endif //__cplusplus
