/*
 * Copyright (c) 2013 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include "vmm_defs.h"
#include "local_apic.h"
#include "em64t_defs.h"
#include "hw_utils.h"
#include "vmm_dbg.h"
#include "host_memory_manager_api.h"
#include "memory_allocator.h"
#include "file_codes.h"
#include "hw_vmx_utils.h"
#ifdef JLMDEBUG
#include "jlmdebug.h"
#endif


UINT64 hw_rdtsc(void)
{
    UINT64      out;
    UINT64*     pout= &out;

    __asm__ volatile (
        "\trdtsc\n"
        "\tmovq     %[pout],%%rcx\n"
        "\tmovl     %%eax, (%%rcx)\n"
        "\tmovl     %%edx, 4(%%rcx)\n"
    :[out] "=g" (out)
    :[pout] "m" (pout): "%rcx");
    return out;
}


UINT8 hw_read_port_8( UINT16 port )
{
    UINT8 out;

    __asm__ volatile(
        "\tinb      %[port], %[out]\n"
    :[out] "=a" (out)
    :[port] "Nd" (port)
    :);
    return out;
}


UINT16 hw_read_port_16( UINT16 port )
{
    UINT16 out;

    __asm__ volatile(
        "\tinw      %[port], %[out]\n"
    :[out] "=a" (out)
    :[port] "Nd" (port) :);
    return out;
}


UINT32 hw_read_port_32( UINT16 port )
{
    UINT32 out;

    __asm__ volatile(
        "\tinl      %[port], %[out]\n"
    :[out] "=a" (out)
    :[port] "Nd" (port) :);
    return out;
}


void hw_write_port_8(UINT16 port, UINT8 val)
{
    __asm__ volatile(
        "\toutb     %[val], %[port]\n"
    ::[val] "a" (val), [port] "Nd" (port) :);
    return;
}


void hw_write_port_16( UINT16 port, UINT16 val)
{
    __asm__ volatile(
        "\toutw     %[val], %[port]\n"
    ::[val] "a" (val), [port] "Nd" (port) :);
    return;
}


void hw_write_port_32( UINT16 port, UINT32 val)
{
    __asm__ volatile(
        "\toutl     %[val], %[port]\n"
    ::[val] "a" (val), [port] "Nd" (port) :);
    return;
}


void hw_write_msr(UINT32 msr_id, UINT64 val)
{
    UINT32 low = (val & (UINT32)-1);
    UINT32 high = (UINT32)(val >> 32);
    __asm__ volatile (
        "\twrmsr\n"
    :: "a" (low), "d" (high), "c" (msr_id):);
    return;
}


UINT64 hw_read_msr(UINT32 msr_id)
{
    UINT32 high;
    UINT32 low;

    // RDMSR reads the processor (MSR) whose index is stored in ECX, 
    // and stores the result in EDX:EAX. 
    __asm__ volatile (
        "\trdmsr\n"
    : "=a" (low), "=d" (high)
    : "c" (msr_id):);

    return ((UINT64)high << 32) | low;
}


UINT64 hw_read_cr0(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%cr0, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


UINT64 hw_read_cr2(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%cr2, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


UINT64 hw_read_cr3(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%cr3, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


UINT64 hw_read_cr4(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%cr4, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


UINT64 hw_read_cr8(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%cr8, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


void hw_write_cr0(UINT64 data)
{
    __asm__ volatile (
        "\tmovq     %[data], %%cr0\n"
    ::[data] "r" (data):); 
    return;
}


void hw_write_cr3(UINT64 data)
{
    __asm__ volatile (
        "\tmovq     %[data], %%cr3\n"
    ::[data] "r" (data):); 
    return;
}


void hw_write_cr4(UINT64 data)
{
    __asm__ volatile (
        "\tmovq     %[data], %%cr4\n"
    ::[data] "r" (data):); 
    return;
}


void hw_write_cr8(UINT64 data)
{
    __asm__ volatile (
        "\tmovq     %[data], %%cr8\n"
    ::[data] "r" (data):); 
    return;
}


UINT64 hw_read_dr0(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%dr0, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


UINT64 hw_read_dr1(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%dr1, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


UINT64 hw_read_dr2(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%dr2, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


UINT64 hw_read_dr3(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%dr3, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


UINT64 hw_read_dr4(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%dr4, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


UINT64 hw_read_dr5(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%dr5, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


UINT64 hw_read_dr6(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%dr6, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


UINT64 hw_read_dr7(void)
{
    UINT64  out;
    __asm__ volatile (
        "\tmovq     %%dr7, %[out]\n"
    :[out] "=r" (out) ::); 
    return out;
}


void hw_write_dr0(UINT64 value)
{
    __asm__ volatile (
        "\tmovq     %[value], %%dr0\n"
    ::[value] "r" (value):); 
    return;
}


void hw_write_dr1(UINT64 value)
{
    __asm__ volatile (
        "\tmovq     %[value], %%dr1\n"
    ::[value] "r" (value):); 
    return;
}


void hw_write_dr2(UINT64 value)
{
    __asm__ volatile (
        "\tmovq     %[value], %%dr2\n"
    ::[value] "r" (value):); 
    return;
}


void hw_write_dr3(UINT64 value)
{
    __asm__ volatile (
        "\tmovq     %[value], %%dr3\n"
    ::[value] "r" (value):); 
    return;
}


void hw_write_dr4(UINT64 value)
{
    __asm__ volatile (
        "\tmovq     %[value], %%dr4\n"
    ::[value] "r" (value):); 
    return;
}


void hw_write_dr5(UINT64 value)
{
    __asm__ volatile (
        "\tmovq     %[value], %%dr5\n"
    ::[value] "r" (value):); 
    return;
}


void hw_write_dr6(UINT64 value)
{
    __asm__ volatile (
        "\tmovq     %[value], %%dr6\n"
    ::[value] "r" (value):); 
    return;
}


void hw_write_dr7(UINT64 value)
{
    __asm__ volatile (
        "\tmovq     %[value], %%dr7\n"
    ::[value] "r" (value):); 
    return;
}


void hw_invlpg(void *address)
{
    __asm__ volatile (
        "\tinvlpg   %[address]\n"
    ::[address] "m" (address):); 
    return;
}


void hw_wbinvd(void)
{
    __asm__ volatile(
        "\twbinvd\n"
    : : :);
    return;
}


void hw_halt( void )
{
    __asm__ volatile(
        "\thlt\n"
    :::);
    return;
}


void hw_lidt(void *source)
{
    __asm__ volatile(
        "\tlidt     (%[source])\n"
    ::[source] "p" (source):);
    return;
}


void hw_sidt(void *destination)
{
    __asm__ volatile(
        "\tsidt (%[destination])\n"
    ::[destination] "p" (destination) 
    :);
    return;
}


INT32  hw_interlocked_increment(INT32 *addend)
{
    __asm__ volatile(
      "\tlock; incl (%[addend])\n"
    :
    :[addend] "p" (addend)
    :"memory");
    return *addend;
}


UINT64 hw_interlocked_increment64(INT64* addend)
{
    __asm__ volatile(
        "\tlock; incq (%[addend])\n"
    : :[addend] "p" (addend)
    :"memory");
    return *addend;
}

INT32 hw_interlocked_decrement(INT32 * minuend)
{
    __asm__ volatile(
      "\tlock; decl (%[minuend])\n"
    : :[minuend] "p" (minuend)
    :"memory");
    return *minuend;
}

INT32 hw_interlocked_add(INT32 volatile * addend, INT32 value)
{
    __asm__ volatile(
        "\tmovq     %[addend], %%rbx\n"
        "\tmovl     %[value], %%eax\n"
        "\tlock;    addl %%eax, (%%rbx)\n"
    : 
    : [addend] "p" (addend), [value] "r" (value)
    : "%eax", "%rbx");
    return *addend;
}

INT32 hw_interlocked_or(INT32 volatile * value, INT32 mask)
{
    __asm__ volatile(
        "\tmovq     %[value], %%rbx\n"
        "\tmovl     %[mask], %%eax\n"
        "\tlock;    orl %%eax, (%%rbx)\n"
    : 
    : [mask] "m" (mask), [value] "r" (value)
    : "%eax", "%rbx");
    return *value;
}

INT32 hw_interlocked_xor(INT32 volatile * value, INT32 mask)
{
    __asm__ volatile(
        "\tmovq     %[value], %%rbx\n"
        "\tmovl     %[mask], %%eax\n"
        "\tlock;    xorl %%eax, (%%rbx)\n"
    : 
    : [mask] "m" (mask), [value] "r" (value)
    : "%eax", "%rbx");
    return *value;
}

void hw_store_fence(void)
{
    __asm__ volatile(
        "\tsfence\n"
    :::);
    return;
}


INT32 hw_interlocked_compare_exchange(INT32 volatile * destination,
                                      INT32 expected, INT32 comperand)
{
#ifdef JLMDEBUG1
    bprint("expected: %d, new: %d --- ", expected, comperand);
#endif
    INT32 old = *destination;
    __asm__ volatile(
        "\tmovq     %[destination], %%r15\n"
        "\tmovl     (%%r15), %%edx\n"
        "\tmovl     %[expected], %%eax\n"
        "\tmovl     %[comperand], %%ecx\n"
        "\tcmpxchgl %%ecx, %%edx\n"
        "\tmovl     %%edx, (%%r15)\n"
    :
    : [expected] "m" (expected), [comperand] "m" (comperand),
      [destination] "m" (destination)
    :"%eax", "%ecx", "%edx", "%rbx", "%r15");
#ifdef JLMDEBUG1
    bprint("destination: %d\n", *destination);
#endif
    return old;
}


INT8 hw_interlocked_compare_exchange_8(INT8 volatile * destination,
            INT8 expected, INT8 comperand)
{
    __asm__ volatile(
        "\tmovq     %[destination], %%r15\n"
        "\tmovb     (%%r15), %%dl\n"
        "\tmovb     %[expected], %%al\n"
        "\tmovb     %[comperand], %%cl\n"
        "\tcmpxchgb %%cl, %%dl\n"
        "\tmovb     %%dl, (%%r15)\n"
    :
    : [expected] "m" (expected), [comperand] "m" (comperand),
      [destination] "m" (destination)
    :"%rax", "%rcx", "%rbx", "%r15");
    return *destination;
}


INT64 hw_interlocked_compare_exchange_64(INT64 volatile * destination,
            INT64 expected, INT64 comperand)
{
    __asm__ volatile(
        "\tmovq     %[destination], %%r15\n"
        "\tmovq     (%%r15), %%rdx\n"
        "\tmovq     %[expected], %%rax\n"
        "\tmovq     %[comperand], %%rcx\n"
        "\tcmpxchgq %%rcx, %%rdx\n"
        "\tmovq     %%rdx, (%%r15)\n"
    :
    : [expected] "m" (expected), [comperand] "m" (comperand),
      [destination] "m" (destination)
    :"%rax", "%rcx", "%rbx", "%r15");
    return *destination;
}


INT32 hw_interlocked_assign(INT32 volatile * target, INT32 new_value)
{
    __asm__ volatile(
        "\tmovq     %[target], %%rbx\n"
        "\tmovl     %[new_value], %%eax\n"
        "\txchgl %%eax, (%%rbx)\n"
    : 
    : [new_value] "m" (new_value), [target] "r" (target)
    : "%eax", "%rbx");
    return *target;
}


// find first bit set
//  forward: LSB->MSB
//  backward: MSB->LSB
// Return 0 if no bits set
// Fills "bit_number" with the set bit position zero based
// BOOLEAN hw_scan_bit_forward( UINT32& bit_number, UINT32 bitset )
// BOOLEAN hw_scan_bit_backward( UINT32& bit_number, UINT32 bitset )
// BOOLEAN hw_scan_bit_forward64( UINT32& bit_number, UINT64 bitset )
// BOOLEAN hw_scan_bit_backward64( UINT32& bit_number, UINT64 bitset )


BOOLEAN hw_scan_bit_forward(UINT32 *bit_number_ptr, UINT32 bitset)
{
    __asm__ volatile(
        "\tbsfl %[bitset], %%eax\n"
        "\tmovq %[bit_number_ptr], %%rbx\n"
        "\tmovl %%eax, (%%rbx)\n"
    :: [bit_number_ptr] "p" (bit_number_ptr), [bitset] "g" (bitset)
    : "%rax", "%eax", "%rbx");
    return bitset ? TRUE : FALSE;
}

BOOLEAN hw_scan_bit_forward64(UINT32 *bit_number_ptr, UINT64 bitset)
{
    __asm__ volatile(
        "\tbsfq %[bitset], %%rax\n"
        "\tmovq %[bit_number_ptr], %%rbx\n"
        "\tmovl %%eax, (%%rbx)\n"
    :
    : [bit_number_ptr] "p" (bit_number_ptr), [bitset] "g" (bitset)
    : "%rax", "%eax", "%rbx");
    return bitset ? TRUE : FALSE;
}

BOOLEAN hw_scan_bit_backward(UINT32 *bit_number_ptr, UINT32 bitset)
{
    __asm__ volatile(
        "\tbsrl %[bitset], %%eax\n"
        "\tmovq %[bit_number_ptr], %%rbx\n"
        "\tmovl %%eax, (%%rbx)\n"
    :
    : [bit_number_ptr] "p" (bit_number_ptr), [bitset] "g" (bitset)
    : "%eax", "%rbx");
    return bitset ? TRUE : FALSE;
}


BOOLEAN hw_scan_bit_backward64(UINT32 *bit_number_ptr, UINT64 bitset)
{
    __asm__ volatile(
        "\tbsrq %[bitset], %%rax\n"
        "\tmovq %[bit_number_ptr], %%rbx\n"
        "\tmovl %%eax, (%%rbx)\n"
    : :[bit_number_ptr] "p" (bit_number_ptr), [bitset] "m" (bitset)
    :"%rax", "%eax", "%rbx");
    return bitset ? TRUE : FALSE;
}


// from fpu2

// Read FPU status word, this doesnt seem to be called
void hw_fnstsw (UINT16* loc) {
    __asm__ volatile(
        "\tmovq %[loc], %%rax\n" 
        "\tfnstsw (%%rax)\n"
        : : [loc] "m"(loc)
        :"%rax");
    return;
}


// Read FPU control word
void hw_fnstcw ( UINT16 * loc )
{
    __asm__ volatile(
        "\tmovq %[loc], %%rax\n"
        "\tfnstcw (%%rax)\n"
        :
        : [loc] "m"(loc)
        :"%rax");
    return;
}


// Init FP Unit
void hw_fninit()
{
    __asm__ volatile(
        "\tfninit\n"
        :::);
    return;
}


// from em64t_utils2.c
typedef struct {
    unsigned long P_RAX;
    unsigned long P_RBX;
    unsigned long P_RCX;
    unsigned long P_RDX;
    unsigned long P_RSI;
    unsigned long P_RDI;
    unsigned long P_RFLAGS;
} PACKED SMI_PORT_PARAMS;  
SMI_PORT_PARAMS spp;
CPUID_PARAMS cp;

void  hw_lgdt (void *gdtr) {
     __asm__ volatile(
        "lgdt (%[gdtr])\n"
     : :[gdtr] "p" (gdtr)
     :);
    return;
}

void hw_sgdt (void * gdtr) {
    //  Store GDTR (to buffer pointed by RCX)
    __asm__ volatile(
        "\tsgdt (%[gdtr])\n"
    : :[gdtr] "p" (gdtr)
    :);
        return;
}


//  Read Command Segment Selector
//  Stack offsets on entry:
//  ax register will contain result
UINT16 hw_read_cs () {
                
    UINT16 ret = 0;

    __asm__ volatile(
        "\txor %%rax, %%rax\n"
        "\tmovw %%cs, %%ax\n"
        "\tmovw %%ax, %[ret]\n"
    :"=rm" (ret)
    :[ret] "rm" (ret)
    :"cc", "rax", "memory");
    return ret;
}


void hw_write_cs (UINT16 i) { 
    // push segment selector
    __asm__ volatile (
        "\txor %%rax, %%rax\n"
        "\tmovw %[i], %%ax\n"
        "\tshlq $32, %%rax\n"
        "\tlea L_CONT_WITH_NEW_CS, %%rdx\n"
        "\tadd %%rdx, %%rax\n"
        "\tpush %%rax\n"
        "\tlret\n" //brings IP to CONT_WITH_NEW_CS
        "L_CONT_WITH_NEW_CS:\n"
        "\tret\n"
    : :[i] "m" (i)
    :"rax", "rdx");
}


//  UINT16 hw_read_ds ( void);
//  Read Data Segment Selector
//  Stack offsets on entry:
//  ax register will contain result
UINT16 hw_read_ds () {
    UINT16 ret = 0;

    __asm__ volatile(
        "\txor %%rax, %%rax\n"
        "\tmovw %%ds, %%ax\n"
        "\tmovw %%ax, %[ret]\n"
    :[ret] "=g" (ret)
    : :"cc", "memory");
    return ret;
}


//  void hw_write_ds ( UINT16);
//  Write to Data Segment Selector
void hw_write_ds(UINT16 i) {
    __asm__ volatile(
        "\tmovw %[i], %%ds\n"
    :
    :[i] "g" (i) :);
    return;
}


//  UINT16 hw_read_es ( void);
//  Read ES Segment Selector
//  Stack offsets on entry:
//  ax register will contain result
UINT16 hw_read_es() {

    UINT16 ret = 0;

     __asm__ volatile(
        "\txor %%rax, %%rax\n"
        "\tmovw %%es, %%ax\n"
        "\tmovw %%ax, %[ret]\n"
    :[ret] "=g" (ret)
    ::);
    return ret;
}


//  void hw_write_es ( UINT16);
//  Write to ES Segment Selector
void hw_write_es (UINT16 i) { 
    __asm__ volatile(
        "\tmovw %[i], %%es\n"
    :
    :[i] "g" (i)
    :);
    return;
}


//  UINT16 hw_read_ss ( void);
//  Read Stack Segment Selector
//  ax register will contain result
UINT16 hw_read_ss() {
    UINT16 ret = 0;

    __asm__ volatile(
        "\txor %%rax, %%rax\n"
        "\tmovw %%es, %%ax\n"
        "\tmovw %%ax, %[ret]\n"
    :[ret] "=g" (ret)
    ::);
    return ret;
}


//  void hw_write_ss ( UINT16);
//  Write to Stack Segment Selector
void hw_write_ss (UINT16 i) { 
    __asm__ volatile(
        "\tmovw %[i], %%ss\n"
    : :[i] "g" (i)
    :);
    return;
}


//  UINT16 hw_read_fs ( void);
//  Read FS
//  ax register will contain result
UINT16 hw_read_fs() {
    UINT16 ret = 0;

    __asm__ volatile(
        "\txor %%rax, %%rax\n"
        "\tmovw %%fs, %%ax\n"
        "\tmovw %%ax, %[ret]\n"
    :[ret] "=g" (ret)
    :
    :"rax");
    return ret;
}


//  void hw_write_fs ( UINT16);
//  Write to FS
void hw_write_fs (UINT16 i) { 
    __asm__ volatile(
        "\tmovw %[i], %%fs\n"
    :
    :[i] "r" (i)
    :);
    return;
}


//  UINT16 hw_read_gs ( void);
//  Read GS
//  ax register will contain result
UINT16 hw_read_gs() {
    UINT16 ret = 0;

    __asm__ volatile(
        "\txor %%rax, %%rax\n"
        "\tmovw %%gs, %%ax\n"
        "\tmovw %%ax, %[ret]\n"
    :[ret] "=rm" (ret) 
    ::"rax");
    return ret;
}


//  void hw_write_gs ( UINT16);
//  Write to GS
void hw_write_gs (UINT16 i) { 
    __asm__ volatile(
        "\tmovw %[i], %%gs\n"
    :
    :[i] "r" (i)
    :);
    return;
}


//  UINT64 hw_read_rsp (void);
UINT64 hw_read_rsp () {
    UINT64 ret = 0;
    __asm__ volatile(
        "\tmovq %%rsp, %%rax\n"
        "\tadd $8,%%rax\n"
        "\tmovq %%rax, %[ret]\n"
    :[ret] "=rm"(ret) 
    :: "cc", "memory");
    return ret;
}


// CHECK the args/offsets need to be double-checked
void hw_write_to_smi_port(
    UINT64 * p_rax,     // rcx
    UINT64 * p_rbx,     // rdx
    UINT64 * p_rcx,     // r8
    UINT64 * p_rdx,     // r9
    UINT64 * p_rsi,     // on the stack
    UINT64 * p_rdi,     // on the stack
    UINT64 * p_rflags) // on the stack
{
#ifdef JLMDEBUG
    bprint("hw_write_to_smi_port\n");
    LOOP_FOREVER
#endif
      (void)p_rax;
    (void)p_rbx;
    (void)p_rcx;
    (void)p_rdx;
    (void)p_rsi;
    (void)p_rdi;
    (void)p_rflags;
    // save callee saved registers
     __asm__ volatile(
        "\tpush %%rbp\n"
        "\tmovq %%rbp, %%rsp\n" //setup stack frame pointer
        "\tpush %%rbx\n"
        "\tpush %%rdi\n"
        "\tpush %%rsi\n"
        "\tpush %%r12\n"
        "\tpush %%r13\n"
        "\tpush %%r14\n"
        "\tpush %%r15\n"
        "\tlea 16(%%rbp), %%r15\n"//set r15 to point to SMI_PORT_PARAMS struct
        // normalize stack\n"\t
        "\tmovq %%rcx, (%%r15)\n"
        "\tmovq %%rdx, 8(%%r15)\n"
        "\tmovq %%r8, 16(%%r15)\n"
        "\tmovq %%r9, 24(%%r15)\n"
        //copy emulator registers into CPU
        "\tmovq (%%r15), %%r8\n"
        "\tmovq (%%r8), %%rax\n"
        "\tmovq 8(%%r15), %%r8\n"
        "\tmovq (%%r8), %%rbx\n"
        "\tmovq 16(%%r15), %%r8\n"
        "\tmovq (%%r8), %%rcx\n"
        "\tmovq 24(%%r15), %%r8\n"
        "\tmovq (%%r8), %%rdx\n"
        "\tmovq 32(%%r15), %%r8\n"
        "\tmovq (%%r8), %%rsi\n"
        "\tmovq 40(%%r15), %%r8\n"
        "\tmovq (%%r8), %%rdi\n"
        "\tmovq 48(%%r15), %%r8\n"
        "\tpush (%%r8)\n"
        "\tpopfq\n" //rflags = *p_rflags

        //we assume that sp will not change after SMI

        "\tpush %%rbp\n"
        "\tpush %%r15\n"
        //  "\tout %%dx, %%al\n"
        "\tout %%al, %%dx\n"
        "\tpop %%r15\n"
        "\tpop %%rbp\n"
        //fill emulator registers from CPU
        "\tmovq (%%r15), %%r8\n"
        "\tmovq %%rax, (%%r8)\n"
        "\tmovq 8(%%r15), %%r8\n"
        "\tmovq %%rbx, (%%r8)\n"
        "\tmovq 16(%%r15), %%r8\n"
        "\tmovq %%rcx, (%%r8)\n"
        "\tmovq 24(%%r15), %%r8\n"
        "\tmovq %%rdx, (%%r8)\n"
        "\tmovq 32(%%r15), %%r8\n"
        "\tmovq %%rsi, (%%r8)\n"
        "\tmovq 40(%%r15), %%r8\n"
        "\tmovq %%rdi, (%%r8)\n"
        "\tmovq 48(%%r15), %%r8\n"
        "\tpushfq\n"
        "\tpop (%%r8)\n" // *p_rflags = rflags
        //restore callee saved registers
        "\tpop %%r15\n"
        "\tpop %%r14\n"
        "\tpop %%r13\n"
        "\tpop %%r12\n"
        "\tpop %%rsi\n"
        "\tpop %%rdi\n"
        "\tpop %%rbx\n"
        "\tpop %%rbp\n"
    :::);
    return;
}

//  void hw_enable_interrupts (void);
void hw_enable_interrupts () {
    __asm__ volatile("\tsti\n");
    return;
}

//  void hw_disable_interrupts (void);
void hw_disable_interrupts () {
    __asm__ volatile("\tcli\n");
    return;
}

//  void hw_fxsave (void* buffer);
void hw_fxsave (void *buffer) {
    __asm__ volatile(
        "\tmovq   %[buffer], %%rbx\n"
        "\tfxsave (%%rbx)\n"
    :
    :[buffer] "g" (buffer)
    :"%rbx");
    return;
}


//  void hw_fxrestore (void* buffer);
void hw_fxrestore (void *buffer) {
    __asm__ volatile(
        "\tmovq   %[buffer], %%rbx\n"
        "\tfxrstor (%%rbx)\n"
    :
    :[buffer] "m" (buffer)
    : "%rbx");
    return;
}


//  void hw_write_cr2 (UINT64 value);
void hw_write_cr2 (UINT64 value) {
    __asm__ volatile(
        "\tmovq %%cr2, %[value]\n"
    :[value] "=g" (value)
    : :"cc", "memory");
    return;
}


// UINT16 * hw_cpu_id ( void * );
//  Read TR and calculate cpu_id
//  ax register will contain result
//  IMPORTANT NOTE: only RAX regsiter may be used here !!!!
//  This assumption is used in gcpu_regs_save_restore.asm
#define CPU_LOCATOR_GDT_ENTRY_OFFSET 32
#define TSS_ENTRY_SIZE_SHIFT 4

__asm__(
".text\n"
".globl hw_cpu_id\n"
".type hw_cpu_id,@function\n"
"hw_cpu_id:\n"
	"\txor %rax, %rax\n"
	"\tstr %ax\n"
	"\tsubw $32, %ax\n" // CPU_LOCATOR_GDT_ENTRY_OFFSET == 32
	"\tshrw $4, %ax\n" // TSS_ENTRY_SIZE_SHIFT == 4
	"\tret\n"
);


// UINT16 hw_read_tr ( void);
//  Read Task Register
//  ax register will contain result
UINT16 hw_read_tr() {
    UINT16 ret = 0;

   __asm__ volatile(
        "\tstr %%ax\n"
        "\tmovw %%ax, %[ret]\n"
    :[ret] "=g" (ret)
    : :"%rax");
    return ret;
}


//  void hw_write_tr ( UINT16);
//  Write Task Register
void hw_write_tr (UINT16 i) {
    __asm__ volatile(
        "\tltr %[i]\n"
    : :[i] "g" (i)
    :);
    return;
}


//  UINT16 hw_read_ldtr ( void);
//  Read LDT Register
//  ax register will contain result
UINT16 hw_read_ldtr () {
    UINT16 ret = 0;
    __asm__ volatile (
        "\tsldt %[ret]\n"
    :[ret] "=g" (ret)
    : :);
    return ret;
}


//  void hw_write_ldtr ( UINT16);
//  Write LDT Register
void hw_write_ldtr (UINT16 i) {
    __asm__ volatile(
        "\tlldt %[i]\n"
    :
    :[i] "r" (i) :);
    return;
}


//  void hw_cpuid (CPUID_PARAMS *)
//  Execute cpuid instruction
void hw_cpuid (CPUID_PARAMS *cp) {
    __asm__ volatile(
        "\tmovq %[cp], %%r8\n" 
        //# fill regs for cpuid
        "\tmovq (%%r8), %%rax\n"
        "\tmovq 8(%%r8), %%rbx\n"
        "\tmovq 16(%%r8), %%rcx\n"
        "\tmovq 24(%%r8), %%rdx\n"
        "\tcpuid\n"
        "\tmovq %%rax, (%%r8)\n"
        "\tmovq %%rbx, 8(%%r8)\n"
        "\tmovq %%rcx, 16(%%r8)\n"
        "\tmovq %%rdx, 24(%%r8)\n"
        :
        :[cp] "g" (cp)
        :"%r8", "%rax", "%rbx", "%rcx", "%rdx", "memory");
        return;
}


/*
 *  void
 *  hw_perform_asm_iret(void);
 * Transforms stack from entry to regular procedure: 
 *
 * [       RIP        ] <= RSP
 *
 * To stack  to perform iret instruction:
 * 
 * [       SS         ]
 * [       RSP        ]
 * [      RFLAGS      ]
 * [       CS         ]
 * [       RIP        ] <= RSP should point prior iret
 */
void hw_perform_asm_iret () {
    __asm__ volatile(
        "\tsubq $0x20, %%rsp\n"     //prepare space for "interrupt stack"
        "\tpush %%rax\n"            //save scratch registers
        "\tpush %%rbx\n"
        "\tpush %%rcx\n"
        "\tpush %%rdx\n"
        "\taddq $0x40, %%rsp\n"   // get rsp back to RIP
        "\tpop %%rax\n"          //RIP -> RAX
        "\tmovq %%cs, %%rbx\n"   //CS  -> RBX
        "\tmovq %%rsp, %%rcx\n"  //good RSP -> RCX
        "\tmovq %%ss, %%rdx\n"   //CS  -> RDX
        "\tpush %%rdx\n"         //[       SS         ]
        "\tpush %%rcx\n"         //[       RSP        ]
        "\tpushfq\n"             //[      RFLAGS      ]
        "\tpush %%rbx\n"         //[       CS         ]
        "\tpush %%rax\n"         //[       RIP        ]

        "\tsubq $0x20, %%rsp\n"   //restore scratch registers
        "\tpop %%rdx\n"
        "\tpop %%rcx\n"
        "\tpop %%rbx\n"
        "\tpop %%rax\n"          // now RSP is in right position 
        "\tiretq "                  //perform IRET
    :::);
} 


void hw_set_stack_pointer (HVA new_stack_pointer, main_continue_fn func, 
                           void *params) 
{
    __asm__ volatile(
        "L1:\n"
        "\tmovq %[new_stack_pointer], %%rsp\n"
        "\tmovq %[params], %[new_stack_pointer]\n"
        "\tsubq $32, %%rsp\n" // allocate home space for 4 input params
        "\tcall *%[func]\n" 
        "\tjmp L1\n"
    :
    :[new_stack_pointer] "g"(new_stack_pointer),
     [func] "g" (func), [params] "p"(params)
    :"cc");
    return;
}


// from em64t_interlocked2.c

// Execute assembler 'pause' instruction
void hw_pause( void ) {
    __asm__ volatile(
        "\tpause\n"
        :::);
    return;
}


// Execute assembler 'monitor' instruction
// CHECK
void hw_monitor( void* addr, UINT32 extension, UINT32 hint) {
#ifdef JLMDEBUG
    bprint("hw_monitor\n");
    LOOP_FOREVER
#endif
    __asm__ volatile(
        "\tmovq %[addr], %%rcx\n" 
        "\tmovq %[extension], %%rdx\n" 
        "\tmovq %[hint], %%r8\n" 
        "\tmovq %%rcx, %%rax\n" 
        "\tmovq %%rdx, %%rcx\n"
        "\tmovq %%r8, %%rdx\n"
        "\tmonitor\n"
        : : [addr] "m" (addr), [extension] "m" (extension), [hint] "m" (hint)
        :"rax", "rcx", "rdx", "r8");
        return;
}

// Execute assembler 'mwait' instruction
void hw_mwait( UINT32 extension, UINT32 hint ) {
#ifdef JLMDEBUG
    bprint("hw_mwait\n");
    LOOP_FOREVER
#endif
    __asm__ volatile(
        "\tmovq %[extension], %%rcx\n"
        "\tmovq %[hint], %%rdx\n"
        "\tmovq %%rdx, %%rax\n"
        // changed the macro .byte... to mwait instruction for better portability.
        "\tmwait %%rax, %%rcx\n"
    : : [extension] "m" (extension), [hint] "m" (hint)
    :"rax", "rbx", "rcx", "rdx");
    return;
}

// from em64t_gcpu_regs_save_restore.c
#include "guest_save_area.h"

// pointer to the array of pointers to the GUEST_CPU_SAVE_AREA
extern GUEST_CPU_SAVE_AREA** g_guest_regs_save_area;

// Utility function for getting the save area pointer into rbx, using the host cpu id
// from a call to hw_cpu_id
__asm__(
".text\n"
".globl load_save_area_into_rbx\n"
".type load_save_area_into_rbx,@function\n"
"load_save_area_into_rbx:\n"
        "\tpush %rax\n" // save rax, since it's used by hw_cpu_id
        "\tcall hw_cpu_id\n" // no arguments, and this only uses rax
        "\tmov g_guest_regs_save_area, %rbx\n" // get g_guest_regs_save_area
         // the original code has this, but it's one indirection too many
        // "\tmov (%rbx), %rbx\n"
        "\tmov (%rbx, %rax, 8), %rbx\n" // SIZEOF QWORD == 8 for multiplier
        "\tpop %rax\n"
        "\tret\n"
);


/*
 * This functions are part of the GUEST_CPU class.  They are called by
 * assembler-lever VmExit/VmResume functions to save all registers that are not
 * saved in VMCS but may be used immediately by C-language VMM code.
 * The following registers are NOT saved here
 *   RIP            part of VMCS
 *   RSP            part of VMCS
 *   RFLAGS         part of VMCS
 *   segment regs   part of VMCS
 *   control regs   saved in C-code later
 *   debug regs     saved in C-code later
 *   FP/MMX regs    saved in C-code later
 *
 * Assumptions:
 *   No free registers except for RSP/RFLAGS.
 *   All are saved on return.
 */
__asm__(
".text\n"
".globl gcpu_save_registers\n"
".type gcpu_save_registers,@function\n"
"gcpu_save_registers:\n"
        "\tpush   %rbx\n" // get rbx out of the way so it can be used as a base
        "\tcall   load_save_area_into_rbx\n"
        "\tmovq   %rax, (%rbx)\n"
        "\tpop    %rax\n" // get the original rbx into rax to save it
        "\tmovq   %rax, 8(%rbx)\n" // save original rbx
        "\tmovq   %rcx, 16(%rbx)\n"
        "\tmovq   %rdx, 24(%rbx)\n"
        "\tmovq   %rdi, 32(%rbx)\n"
        "\tmovq   %rsi, 40(%rbx)\n"
        "\tmovq   %rbp, 48(%rbx)\n"
        "\tmovq   %r8, 64(%rbx)\n"
        "\tmovq   %r9, 72(%rbx)\n"
        "\tmovq   %r10, 80(%rbx)\n"
        "\tmovq   %r11, 88(%rbx)\n"
        "\tmovq   %r12, 96(%rbx)\n"
        "\tmovq   %r13, 104(%rbx)\n"
        "\tmovq   %r14, 112(%rbx)\n"
        "\tmovq   %r15, 120(%rbx)\n"
        // skip RIP and RFLAGS here (16 missing bytes)
        // Note that the XMM registers require 16-byte alignment
        "\tmovaps %xmm0, 144(%rbx)\n"
        "\tmovaps %xmm1, 160(%rbx)\n"
        "\tmovaps %xmm2, 176(%rbx)\n"
        "\tmovaps %xmm3, 192(%rbx)\n"
        "\tmovaps %xmm4, 208(%rbx)\n"
        "\tmovaps %xmm5, 224(%rbx)\n"
        "\tret\n"
);


__asm__(
".globl gcpu_restore_registers\n"
".type gcpu_restore_registers,@function\n"
"gcpu_restore_registers:\n"
        "\tcall load_save_area_into_rbx\n"
        // restore XMM registers first
        // These are aligned on 16-byte boundaries
        "\tmovaps 144(%rbx), %xmm0\n"
        "\tmovaps 160(%rbx), %xmm1\n"
        "\tmovaps 176(%rbx), %xmm2\n"
        "\tmovaps 192(%rbx), %xmm3\n"
        "\tmovaps 208(%rbx), %xmm4\n"
        "\tmovaps 224(%rbx), %xmm5\n"

        "\tmovq   (%rbx), %rax\n"
        // rbx is restored at the end
        "\tmovq   16(%rbx), %rcx\n"
        "\tmovq   24(%rbx), %rdx\n"
        "\tmovq   32(%rbx), %rdi\n"
        "\tmovq   40(%rbx), %rsi\n"
        "\tmovq   48(%rbx), %rbp\n"
        // rsp is not restored
        "\tmovq   64(%rbx), %r8\n"
        "\tmovq   72(%rbx), %r9\n"
        "\tmovq   80(%rbx), %r10\n"
        "\tmovq   88(%rbx), %r11\n"
        "\tmovq   96(%rbx), %r12\n"
        "\tmovq   104(%rbx), %r13\n"
        "\tmovq   112(%rbx), %r14\n"
        "\tmovq   120(%rbx), %r15\n"
        // skip RIP and RFLAGS
        
        // restore rbx now that we're done using it as a base register
        "\tmovq   8(%rbx), %rbx\n"
        "\tret\n"
);



#if 0   // never ported
void hw_leave_64bit_mode (unsigned int compatibility_segment,
    unsigned short int port_id,
    unsigned short int value,
    unsigned int cr3_value) 
{

        jmp $
        shl rcx, 32             ;; prepare segment:offset pair for retf by shifting
                                ;; compatibility segment in high address
        lea rax, compat_code    ;; and
        add rcx, rax            ;; placing offset into low address
        push rcx                ;; push ret address onto stack
        mov  rsi, rdx           ;; rdx will be used during EFER access
        mov  rdi, r8            ;; r8 will be unaccessible, so use rsi instead
        mov  rbx, r9            ;; save CR3 in RBX. this function is the last called, so we have not to save rbx
        retf                    ;; jump to compatibility mode
compat_code:                    ;; compatibility mode starts right here

        mov rax, cr0            ;; only 32-bit are relevant
        btc eax, 31             ;; disable IA32e paging (64-bits)
        mov cr0, rax            ;;

        ;; now in protected mode
        mov ecx, 0C0000080h     ;; EFER MSR register
        rdmsr                   ;; read EFER into EAX
        btc eax, 8              ;; clear EFER.LME
        wrmsr                   ;; write EFER back

;        mov cr3, rbx            ;; load CR3 for 32-bit mode
;
;        mov rax, cr0            ;; use Rxx notation for compiler, only 32-bit are valuable
;        bts eax, 31             ;; enable IA32 paging (32-bits)
;        mov cr0, rax            ;;
;        jmp @f

;; now in 32-bit paging mode
        mov rdx, rsi
        mov rax, rdi
        out dx, ax              ;; write to PM register
        ret                     ;; should never get here
} //hw_leave_64bit_mode
#endif // never ported

