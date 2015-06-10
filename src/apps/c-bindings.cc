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

#include "tao-c.h"

#include <string>

#include "tao/fd_message_channel.h"
#include "tao/tao_rpc.h"
#include "tao/util.h"

using tao::FDMessageChannel;
using tao::MessageChannel;
using tao::InitializeApp;
using tao::MarshalSpeaksfor;
using tao::Tao;
using tao::TaoRPC;

using std::string;

#define define_string_out(name) if (!name || !name##_size) return 0; string _##name;
#define define_string_in(name) string _##name(name,name##_size);

#define string_to_bytes(name,additional_frees) { \
	*name##_size=_##name.size(); \
	if (!((*name)=(char*)malloc(*name##_size))) { additional_frees; return 0; } \
	memcpy(*name,_##name.data(),*name##_size); \
}

static void strfree(char** s)
{
	if (*s)
	{
		free(*s);
		*s=NULL;
	}
}

// tao::FDMessageChannel
TaoFDMessageChannel* Tao_FDMessageChannel_new(int readfd, int writefd)
{
	return (TaoFDMessageChannel*)new FDMessageChannel(readfd, writefd);
}

TaoMessageChannel* Tao_FDMessageChannel_as_MessageChannel(TaoFDMessageChannel *obj)
{
	return (TaoMessageChannel*)static_cast<MessageChannel*>((FDMessageChannel*)obj);
}

// tao::Tao
void Tao_Tao_delete(TaoTao *obj)
{
	delete (Tao*)obj;
}

int Tao_Tao_GetTaoName(TaoTao *obj, char **name, size_t *name_size)
{
	define_string_out(name);

	if (((Tao*)obj)->GetTaoName(&_name))
	{
		string_to_bytes(name,);
		return 1;
	}
	return 0;
}

int Tao_Tao_GetRandomBytes(TaoTao *obj, size_t size, char **bytes, size_t *bytes_size)
{
	define_string_out(bytes);

	if (((Tao*)obj)->GetRandomBytes(size,&_bytes))
	{
		string_to_bytes(bytes,);
		return 1;
	}
	return 0;
}

int Tao_Tao_Attest(TaoTao *obj, const char *message, size_t message_size, char **attestation, size_t *attestation_size)
{
	define_string_in(message);
	define_string_out(attestation);

	if (((Tao*)obj)->Attest(_message,&_attestation))
	{
		string_to_bytes(attestation,);
		return 1;
	}
	return 0;
}

int Tao_Tao_Seal(TaoTao *obj, const char *data, size_t data_size, const char *policy, size_t policy_size, char **sealed, size_t *sealed_size)
{
	define_string_in(data);
	define_string_in(policy);
	define_string_out(sealed);

	if (((Tao*)obj)->Seal(_data,_policy,&_sealed))
	{
		string_to_bytes(sealed,);
		return 1;
	}
	return 0;
}


int Tao_Tao_Unseal(TaoTao *obj, const char *sealed, size_t sealed_size, char **data, size_t *data_size, char **policy, size_t *policy_size)
{
	define_string_in(sealed);
	define_string_out(data);
	define_string_out(policy);

	if (((Tao*)obj)->Unseal(_sealed,&_data,&_policy))
	{
		string_to_bytes(data,);
		string_to_bytes(policy,strfree(data));
		return 1;
	}
	return 0;
}

// tao::TaoRPC
TaoTaoRPC* Tao_TaoRPC_new(TaoMessageChannel *channel)
{
	return (TaoTaoRPC*)new TaoRPC((MessageChannel*)channel);
}

TaoTao* Tao_TaoRPC_as_Tao(TaoTaoRPC *obj)
{
	return (TaoTao*)static_cast<Tao*>((TaoRPC*)obj);
}

// tao::InitializeApp
int Tao_InitializeApp(int *argc, char ***argv, int remove_args)
{
	return InitializeApp(argc,argv,remove_args ? true : false) ? 1 : 0;
}

// tao::MarshalSpeaksfor
int Tao_MarshalSpeaksfor(const char *key, size_t key_size, const char *binaryTaoName, size_t binaryTaoName_size, char **out, size_t *out_size)
{
	define_string_in(key);
	define_string_in(binaryTaoName);
	define_string_out(out);

	if (MarshalSpeaksfor(_key,_binaryTaoName,&_out))
	{
		string_to_bytes(out,);
		return 1;
	}
	return 0;
}

