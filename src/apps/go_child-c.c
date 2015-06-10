//  Copyright (c) 2014, Google Inc.  All rights reserved.
//    and
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
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#include "tao-c.h"

#define LOG(level,args...) {fprintf(stderr,#level ": " args);fputc('\n',stderr);if(0==strcmp("FATAL",#level))exit(1);}

#define s(name) char* name;size_t name##_size
#define p(name) name, name##_size

static int hexencode(const char *in, size_t in_size, char **out, size_t *out_size)
{
	static char hextable[16]={'0','1','2','3','4','5','6','7','8','9','a','b','c','d','e','f'};
	*out_size=in_size*2+1;
	if (! ((*out)=(char*)malloc(*out_size))) return 0;
	size_t i;
	for (i=0;i<in_size;i++)
	{
		(*out)[i*2  ]=hextable[((unsigned char)in[i])>>4];
		(*out)[i*2+1]=hextable[((unsigned char)in[i])&0xf];
	}
	(*out)[i*2]=0;
	return 1;
}

int main(int argc, char **argv) {
	Tao_InitializeApp(&argc, &argv, 0);

	// This code expects fd 3 and 4 to be the pipes from and to the Tao, so it
	// doesn't need to take any parameters. It will establish a Tao Child Channel
	// directly with these fds.
	TaoFDMessageChannel* msg=Tao_FDMessageChannel_new(3, 4);
	TaoTao* tao=Tao_TaoRPC_as_Tao(Tao_TaoRPC_new(Tao_FDMessageChannel_as_MessageChannel(msg)));
	s(bytes);
	if (!Tao_Tao_GetRandomBytes(tao,10, p(&bytes))) {
		LOG(FATAL,"Couldn't get 10 bytes from the Tao RPC channel");
	}

	if (bytes_size == 10) {
		LOG(INFO,"Got 10 bytes from the Tao RPC channel");
	} else {
		LOG(FATAL,"Got %zu bytes from the channel, but expected 10", bytes_size);
	}

	s(encodedBytes);
	if (!hexencode(p(bytes), p(&encodedBytes))) {
		LOG(FATAL,"Couldn't encode 10 bytes in Base64W");
	}
	LOG(INFO,"Encoded bytes: %s",encodedBytes);

	s(sealed);
	if (!Tao_Tao_Seal(tao,p(bytes), Tao_Tao_SealPolicyDefault, sizeof(Tao_Tao_SealPolicyDefault)-1, p(&sealed))) {
		LOG(FATAL,"Couldn't seal bytes across the channel");
	}

	s(encodedSealed);
	if (!hexencode(p(sealed), p(&encodedSealed))) {
		LOG(FATAL,"Couldn't encode the sealed bytes");
	}
	LOG(INFO,"Encoded sealed bytes: %s",encodedSealed);

	s(unsealed);
	s(policy);
	if (!Tao_Tao_Unseal(tao,p(sealed), p(&unsealed), p(&policy))) {
		LOG(FATAL,"Couldn't unseal the tao-sealed data");
	}
	LOG(INFO,"Got a seal policy '%s'",policy);

	if (strcmp(policy,Tao_Tao_SealPolicyDefault) != 0) {
		LOG(FATAL,"The policy returned by Unseal didn't match the Seal policy");
	}

	if (unsealed_size!=bytes_size || memcmp(unsealed,bytes,bytes_size) != 0) {
		LOG(FATAL,"The unsealed data didn't match the sealed data");
	}

	s(encodedUnsealed);
	if (!hexencode(p(unsealed), p(&encodedUnsealed))) {
		LOG(FATAL,"Couldn't encoded the unsealed bytes");
	}

	LOG(INFO,"Encoded unsealed bytes: %s",encodedUnsealed);

	// Set up a fake attestation using a fake key.
	s(taoName);
	if (!Tao_Tao_GetTaoName(tao,p(&taoName))) {
		LOG(FATAL,"Couldn't get the name of the Tao");
	}

	s(msf);
	if (!Tao_MarshalSpeaksfor("This is a fake key", sizeof("This is a fake key")-1, p(taoName), p(&msf))) {
		LOG(FATAL,"Couldn't marshal a speaksfor statement");
	}

	s(attest);
	if (!Tao_Tao_Attest(tao, p(msf), p(&attest))) {
		LOG(FATAL,"Couldn't attest to a fake key delegation");
	}

	s(encodedAttest);
	if (!hexencode(p(attest), p(&encodedAttest))) {
		LOG(FATAL,"Couldn't encode the attestation");
	}
	LOG(INFO,"Got attestation %s",encodedAttest);

	LOG(INFO,"All Go Tao tests pass");
	return 0;
}
