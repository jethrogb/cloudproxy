//
//  File: logging.h
//  Description: Debugging support
//
//
//  Copyright (c) 2011, Intel Corporation. Some contributions
//    (c) John Manferdelli.  All rights reserved.
//
//  Redistribution and use in source and binary forms, with or without
//  modification, are permitted provided that the following conditions
//  are met:
//    Redistributions of source code must retain the above copyright notice,
//      this list of conditions and the disclaimer below.
//    Redistributions in binary form must reproduce the above copyright
//      notice, this list of conditions and the disclaimer below in the
//      documentation and/or other materials provided with the distribution.
//
//  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
//  "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
//  TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
//  PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
//  HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
//  SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
//  TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
//  PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
//  LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
//  NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//  SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//

// --------------------------------------------------------------------------

#ifndef _LOGGING__H
#define _LOGGING__H

#include "common.h"
#include <stdio.h>

#ifdef GLOGENABLED
#include "glog/logging.h"
#else
#include <fstream>
extern std::ostream *logFile;
#define INFO 1
#define WARNING 2
#define ERROR 3
#define FATAL 4

#define LOG(X) (*logFile)
#endif

bool initLog(const char* );
void closeLog();
void PrintBytes(const char* message, byte* pbData, int iSize, int col = 32);
void PrintBytesToConsole(const char* message, byte* pbData, int iSize, int col);

#endif

// --------------------------------------------------------------------
