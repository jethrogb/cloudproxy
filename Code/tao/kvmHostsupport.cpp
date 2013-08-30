//  File: kvmHostsupport.cpp
//      John Manferdelli
//  Description:  Support for KVM host
//
//  Copyright (c) 2012, John Manferdelli
//  Some contributions copyright (c) 2012, Intel Corporation
//
// Use, duplication and disclosure of this file and derived works of
// this file are subject to and licensed under the Apache License dated
// January, 2004, (the "License").  This License is contained in the
// top level directory originally provided with the CloudProxy Project.
// Your right to use or distribute this file, or derived works thereof,
// is subject to your being bound by those terms and your use indicates
// consent to those terms.
//
// If you distribute this file (or portions derived therefrom), you must
// include License in or with the file and, in the event you do not include
// the entire License in the file, the file must contain a reference
// to the location of the License.

#include "jlmTypes.h"
#include "logging.h"
#include "jlmcrypto.h"
#include "jlmUtility.h"
#include "modesandpadding.h"
#include "sha256.h"
#include "tao.h"
#include "bignum.h"
#include "mpFunctions.h"
#include "cryptoHelper.h"
#include "trustedKeyNego.h"
#include "tcIO.h"
#include "buffercoding.h"
#include "tcService.h"
#include "kvmHostsupport.h"
#include "kvmHostsupport.h"
#include <libvirt/libvirt.h>
#include <string.h>
#include <time.h>


// -------------------------------------------------------------------------


bool startKvmVM(const char* szvmimage, const char* systemname,
                const char* xmldomainstring)
{
    virConnectPtr   vmconnection;
    virDomainPtr    vmdomain;

    if(szvmimage==NULL || systemname==NULL || xmldomainstring==NULL) {
        fprintf(g_logFile, "startKvm: Bad input arguments\n");
        return false;
    }

#ifdef TEST
    fprintf(g_logFile, "startKvm: %s, %s, %s\n", szvmimage, systemname, xmldomainstring);
#endif
    
    vmconnection= virConnectOpen(systemname);
    if(vmconnection==NULL) {
        fprintf(g_logFile, "startKvm: couldn't connect\n");
        return false;
    }
   
    vmdomain= virDomainCreateXML(vmconnection, xmldomainstring, 0); 
    if(vmdomain==NULL) {
        fprintf(g_logFile, "startKvm: couldn't start domain\n");
        return false;
    }
    return true;
}


// -------------------------------------------------------------------------


