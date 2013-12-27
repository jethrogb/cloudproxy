/****************************************************************************
* Copyright (c) 2013 Intel Corporation
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
****************************************************************************/

#ifndef _IMAGE_ACCESS_GEN_H_
#define _IMAGE_ACCESS_GEN_H_

#include  "image_access_env.h"

struct _GEN_IMAGE_ACCESS_S {
    size_t (*read)(struct _GEN_IMAGE_ACCESS_S *ia, void *dest, size_t src_offset, size_t bytes);
    size_t (*map_to_mem)(struct _GEN_IMAGE_ACCESS_S *ia, void **dest, size_t src_offset, size_t bytes);
    void   (*close)(struct _GEN_IMAGE_ACCESS_S *ia);    // close file | free memory
};

typedef struct _GEN_IMAGE_ACCESS_S GEN_IMAGE_ACCESS_S;

#endif // _IMAGE_ACCESS_GEN_H_

