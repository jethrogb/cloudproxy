// Copyright (c) 2014, Google Inc. All rights reserved.
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

// Package tpm supports direct communication with a tpm device under Linux.
package tpm

import (
	//"crypto"
	//"crypto/hmac"
	//"crypto/rand"
	//"crypto/rsa"
	//"crypto/sha1"
	//"crypto/subtle"
	//"bytes"
	//"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

// OpenTPM opens a channel to the TPM at the given path. If the file is a
// device, then it treats it like a normal TPM device, and if the file is a
// Unix domain socket, then it opens a connection to the socket.
func OpenTPM(path string) (io.ReadWriteCloser, error) {
	// If it's a regular file, then open it
	var rwc io.ReadWriteCloser
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.Mode()&os.ModeDevice != 0 {
		var f *os.File
		f, err = os.OpenFile(path, os.O_RDWR, 0600)
		if err != nil {
			return nil, err
		}
		rwc = io.ReadWriteCloser(f)
	} else if fi.Mode()&os.ModeSocket != 0 {
		uc, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: path, Net: "unix"})
		if err != nil {
			return nil, err
		}
		rwc = io.ReadWriteCloser(uc)
	} else {
		return nil, fmt.Errorf("unsupported TPM file mode %s", fi.Mode().String())
	}

	return rwc, nil
}

func SetLongPcrs(pcr_nums []int) ([]byte, error) {
	pcr := []byte{3,0,0,0}
	var byte_num int
	var byte_pos byte
	for _,e := range pcr_nums {
		byte_num = e / 8;
		byte_pos = 1 << uint16(e % 8)
		pcr[byte_num] |= byte_pos
	}
	return pcr, nil
}

func SetShortPcrs(pcr_nums []int) ([]byte, error) {
	return nil, nil
}

func ComputePcrDigest(alg uint16, in []byte) ([]byte, error) {
	return nil, nil
}

func SetOwnerHandle(handle Handle) ([]byte, error) {
	return nil, nil
}


// ----------------------------------------------------------------
// remove

// SetPasswordData(string& password, int size, byte* buf)

// int CreatePasswordAuthArea(string& password, int size, byte* buf)

// int CreateSensitiveArea(int size_in, byte* in, int size_data, byte* data, int size, byte* buf)

// CreateSensitiveArea(string& authString, int size_data, byte* data, int size, byte* buf)

// Marshal_AuthSession_Info(TPMI_DH_OBJECT& tpm_obj, TPMI_DH_ENTITY& bind_obj,
//                          TPM2B_NONCE& initial_nonce, TPM2B_ENCRYPTED_SECRET& salt,
//                          TPM_SE& session_type, TPMT_SYM_DEF& symmetric,
//                          TPMI_ALG_HASH& hash_alg, int size, byte* out_buf)

// FillPublicRsaTemplate(enc_alg, int_alg, flags, sym_alg,
//                        sym_key_size, sym_mode, sig_scheme,
//                        mod_size, exp, pub_key);
// Marshal_Public_Key_Info(TPM2B_PUBLIC& in, int size, byte* buf)
// GetReadPublicOut(uint16_t size_in, byte* input, TPM2B_PUBLIC& outPublic)
//int GetRsaParams(uint16_t size_in, byte* input, TPMS_RSA_PARMS& rsaParams,
//                 TPM2B_PUBLIC_KEY_RSA& rsa)
//GetCreateOut(int size, byte* in, int* size_public, byte* out_public,
//                  int* size_private, byte* out_private,
//                  TPM2B_CREATION_DATA* creation_out, TPM2B_DIGEST* digest_out,
//                  TPMT_TK_CREATION* creation_ticket)
//   FillKeyedHashTemplate(TPM_ALG_KEYEDHASH, int_alg, flags,
//                        size_policy_digest, policy_digest, keyed_hash);
//  n = Marshal_Keyed_Hash_Info(keyed_hash, space_left, in);

// ----------------------------------------------------------------


// ConstructGetRandom constructs a GetRandom command.
func ConstructGetRandom(size uint32) ([]byte, error) {
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdGetRandom)
	if err != nil {
		return nil, errors.New("ConstructGetRandom failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
}

// DecodeGetRandom decodes a GetRandom response.
func DecodeGetRandom(in []byte) ([]byte, error) {
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode GetRandom response")
        }

        return rand_bytes, nil
}

// GetRandom gets random bytes from the TPM.
func GetRandom(rw io.ReadWriteCloser, size uint32) ([]byte, error) {
	// Construct command
	x, err:= ConstructGetRandom(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeGetRandom(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeGetRandom %s\n", err)
		return nil,err
	}
	return rand, nil
}

// ConstructFlushContext constructs a FlushContext command.
func ConstructFlushContext(handle Handle) ([]byte, error) {
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdFlushContext)
	if err != nil {
		return nil, errors.New("ConstructFlushContext failed")
	}
	cmd_text :=  []interface{}{uint32(handle)}
	x, _ := packWithHeader(cmdHdr, cmd_text)
	return x, nil
}

// FlushContext
func FlushContext(rw io.ReadWriter, handle Handle) (error) {
	// Construct command
	x, err:= ConstructFlushContext(handle)
	if err != nil {
		return errors.New("ConstructFlushContext fails") 
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
                return errors.New("DecodeCommandResponse fails")
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)
	if status != errSuccess {
		return errors.New("FlushContext unsuccessful")
	}
	return nil
}

// ConstructReadPcrs constructs a ReadPcr command.
func ConstructReadPcrs(num_spec int, num_pcr byte, pcrs []byte) ([]byte, error) {
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdPCR_Read)
	if err != nil {
		return nil, errors.New("ConstructReadPcrs failed")
	}
	num := uint32(num_spec)
	out := []interface{}{&num, &pcrs}
	x, _ := packWithHeader(cmdHdr, out)
	return x, nil
}

// DecodeReadPcrs decodes a ReadPcr response.
func DecodeReadPcrs(in []byte) (uint32, []byte, uint16, []byte, error) {
        var pcr []byte
        var digest []byte
        var updateCounter uint32
        var t uint32
        var s uint32

        out :=  []interface{}{&t, &updateCounter, &pcr, &s, &digest}
        err := unpack(in, out)
        if err != nil {
                return 1, nil, 0, nil, errors.New("Can't decode ReadPcrs response")
        }
	return updateCounter, pcr, uint16(t), digest, nil
}

// ReadPcr reads a PCR value from the TPM.
//	Output: updatecounter, selectout, digest
func ReadPcrs(rw io.ReadWriter, num_byte byte, pcrSelect []byte) (uint32, []byte, uint16, []byte, error) {
	// Construct command
	x, err:= ConstructReadPcrs(1, 4, pcrSelect)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return 1, nil, 0, nil, errors.New("MakeCommandHeader failed") 
	}
	fmt.Printf("ReadPcrs command: %x", x)

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return 0, nil, 0, nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return 0, nil, 0, nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return 0, nil, 0, nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		return 0, nil, 0, nil, errors.New("DecodeCommandResponse fails")
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
		return 0, nil, 0, nil, errors.New("ReadPcr command failed")
	}
	counter, pcr, alg, digest, err := DecodeReadPcrs(resp[10:])
	if err != nil {
		return 0, nil, 0, nil, errors.New("DecodeReadPcrsfails")
	}
	return counter, pcr, alg, digest, err 
}

// ConstructReadClock constructs a ReadClock command.
func ConstructReadClock() ([]byte, error) {
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdReadClock)
	if err != nil {
		return nil, errors.New("ConstructGetRandom failed")
	}
	x, _ := packWithHeader(cmdHdr, nil)
	return x, nil
}

// DecodeReadClock decodes a ReadClock response.
func DecodeReadClock(in []byte) (uint64, uint64, error) {
        var current_time, current_clock uint64

        out :=  []interface{}{&current_time, &current_clock}
        err := unpack(in, out)
        if err != nil {
                return 0, 0, errors.New("Can't decode DecodeReadClock response")
        }

	return current_time, current_clock, nil
}

// ReadClock
//	Output: current time, current clock
func ReadClock(rw io.ReadWriter) (uint64, uint64, error) {
	// Construct command
	x, err:= ConstructReadClock()
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return 0 ,0, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return 0, 0, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return 0, 0, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return 0, 0, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return 0, 0, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	current_time, current_clock, err :=  DecodeReadClock(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeReadClock %s\n", err)
		return 0, 0,err
	}
	return current_time, current_clock, nil
}

// ConstructGetCapabilities constructs a GetCapabilities command.
func ConstructGetCapabilities(cap uint32, count uint32, property uint32) ([]byte, error) {
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdGetCapability)
	if err != nil {
		return nil, errors.New("GetCapability failed")
	}
	cap_bytes:=  []interface{}{&cap, &property, &count}
	x, _ := packWithHeader(cmdHdr, cap_bytes)
	return x, nil
}

// DecodeGetCapabilities decodes a GetCapabilities response.
func DecodeGetCapabilities(in []byte) (uint32, []uint32, error) {
        var num_handles uint32
        var cap_reported uint32

        out :=  []interface{}{&cap_reported,&num_handles}
        err := unpack(in[1:9], out)
        if err != nil {
                return 0, nil, errors.New("Can't decode GetCapabilities response")
        }
	// only ordTPM_CAP_HANDLES handled
        if cap_reported !=  ordTPM_CAP_HANDLES {
                return 0, nil, errors.New("Only ordTPM_CAP_HANDLES supported")
        }
	var handles []uint32
	var handle uint32
        handle_out :=  []interface{}{&handle}
	for i:= 0; i < int(num_handles); i++ {
		err := unpack(in[8 + 4 * i:18:12 + 4 * i], handle_out)
		if err != nil {
			return 0, nil, errors.New("Can't decode GetCapabilities handle")
		}
		handles = append(handles, handle)
	}

        return cap_reported, handles, nil
}

// GetCapabilities 
//	Output: output buf
func GetCapabilities(rw io.ReadWriter, cap uint32, count uint32, property uint32) ([]uint32, error) {
	// Construct command
	x, err:= ConstructGetCapabilities(cap, count, property)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	_, handles, err :=  DecodeGetCapabilities(resp[10:read])
	if err != nil {
		return nil,err
	}
	return handles, nil
}

// Flushall
func Flushall(rw io.ReadWriter) (error) {
	handles, err := GetCapabilities(rw, ordTPM_CAP_HANDLES, 1, 0x80000000)
	if err != nil {
		return err
	}
	for _, e := range handles {
		_ = FlushContext(rw, Handle(e))
	}
	return nil
}

// ConstructCreatePrimary constructs a CreatePrimary command.
func ConstructCreatePrimary(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdCreatePrimary)
	if err != nil {
		return nil, errors.New("ConstructCreatePrimary failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeCreatePrimary decodes a CreatePrimary response.
func DecodeCreatePrimary(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode CreatePrimary response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// CreatePrimary
func CreatePrimary(rw io.ReadWriter, keyBlob []byte) ([]byte, error) {
/*
	TPM_HANDLE owner, string& authString,
        TPML_PCR_SELECTION& pcr_selection,
        TPM_ALG_ID enc_alg, TPM_ALG_ID int_alg,
        TPMA_OBJECT& flags, TPM_ALG_ID sym_alg,
        TPMI_AES_KEY_BITS sym_key_size,
        TPMI_ALG_SYM_MODE sym_mode, TPM_ALG_ID sig_scheme,
        int mod_size, uint32_t exp,
        TPM_HANDLE* handle, TPM2B_PUBLIC* pub_out

	// Construct command
	x, err:= ConstructCreatePrimary(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeCreatePrimary(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeCreatePrimary %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

// ConstructReadPublic constructs a ReadPublic command.
func ConstructReadPublic(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdReadPublic)
	if err != nil {
		return nil, errors.New("ConstructReadPublic failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeReadPublic decodes a ReadPublic response.
func DecodeReadPublic(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode ReadPublic response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// ReadPublic
func ReadPublic(rw io.ReadWriter, keyBlob []byte) ([]byte, error) {
/*
	TPM_HANDLE handle,
        uint16_t* pub_blob_size, byte* pub_blob,
        TPM2B_PUBLIC& outPublic, TPM2B_NAME& name,
        TPM2B_NAME& qualifiedName
	// Construct command
	x, err:= ConstructReadPublic(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeReadPublic(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeReadPublic %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

// CreateKey

// ConstructCreateKey constructs a CreateKey command.
func ConstructCreateKey(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdGetRandom)
	if err != nil {
		return nil, errors.New("ConstructGetRandom failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeCreateKey decodes a CreateKey response.
func DecodeCreateKey(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode CreateKey response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

func CreateKey(rw io.ReadWriter, keyBlob []byte) ([]byte, error) {
/*
 	TPM_HANDLE parent_handle,
        string& parentAuth, string& authString,
        TPML_PCR_SELECTION& pcr_selection,
        TPM_ALG_ID enc_alg, TPM_ALG_ID int_alg,
        TPMA_OBJECT& flags, TPM_ALG_ID sym_alg,
        TPMI_AES_KEY_BITS sym_key_size,
        TPMI_ALG_SYM_MODE sym_mode, TPM_ALG_ID sig_scheme,
        int mod_size, uint32_t exp,
        int* size_public, byte* out_public,
        int* size_private, byte* out_private,
        TPM2B_CREATION_DATA* creation_out,
        TPM2B_DIGEST* digest_out, TPMT_TK_CREATION* creation_ticket
	// Construct command
	x, err:= ConstructCreateKey(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeCreateKey(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeCreateKey %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

// ConstructLoad constructs a Load command.
func ConstructLoad(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdLoad)
	if err != nil {
		return nil, errors.New("ConstructLoad failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeLoad decodes a Load response.
func DecodeLoad(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode Load response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// Load
//	Output: handle, name
func Load(rw io.ReadWriter, parentHandle Handle, parentAuth string,
	     inPublic []byte, inPrivate []byte) (Handle, []byte, error) {
/*
	// Construct command
	x, err:= ConstructGetRandom(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeLoad(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeLoad %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return 1, nil, nil
}

// ConstructPolicyPassword constructs a PolicyPassword command.
func ConstructPolicyPassword(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdPolicyPassword)
	if err != nil {
		return nil, errors.New("ConstructPolicyPassword failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodePolicyPassword decodes a PolicyPassword response.
func DecodePolicyPassword(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode PolicyPassword response")
        }

        return rand_bytes, nil
 */
	return nil, nil
}

// PolicyPassword
func PolicyPassword(rw io.ReadWriter, handle Handle) (error) {
/*
	// Construct command
	x, err:= ConstructGetRandom(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodePolicyPassword(resp[10:read])
	if err != nil {
		fmt.Printf("DecodePolicyPassword %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil
}

// ConstructPolicyGetDigest constructs a PolicyGetDigest command.
func ConstructPolicyGetDigest(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdPolicyGetDigest)
	if err != nil {
		return nil, errors.New("ConstructPolicyGetDigest failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodePolicyGetDigest decodes a PolicyGetDigest response.
func DecodePolicyGetDigest(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode PolicyGetDigest response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// PolicyGetDigest
//	Output: digest
func PolicyGetDigest(rw io.ReadWriter, handle Handle) ([]byte, error) {
/*
	// Construct command
	x, err:= ConstructPolicyGetDigest(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodePolicyGetDigest(resp[10:read])
	if err != nil {
		fmt.Printf("DecodePolicyGetDigest %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

// ConstructStartAuthSession constructs a StartAuthSession command.
func ConstructStartAuthSession(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdStartAuthSession)
	if err != nil {
		return nil, errors.New("ConstructStartAuthSession failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeStartAuthSession decodes a StartAuthSession response.
func DecodeStartAuthSession(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode StartAuthSession response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// StartAuthSession
func StartAuthSession(rw io.ReadWriter, keyBlob []byte) ([]byte, error) {
/*
 	TPM_RH tpm_obj, TPM_RH bind_obj,
        TPM2B_NONCE& initial_nonce,
        TPM2B_ENCRYPTED_SECRET& salt,
        TPM_SE session_type, TPMT_SYM_DEF& symmetric,
        TPMI_ALG_HASH hash_alg, TPM_HANDLE* session_handle,
        TPM2B_NONCE* nonce_obj
	// Construct command
	x, err:= ConstructStartAuthSession(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeStartAuthSession(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeStartAuthSession %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

// ConstructCreateSealed constructs a CreateSealed command.
func ConstructCreateSealed(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdCreateSealed)
	if err != nil {
		return nil, errors.New("ConstructCreateSealed failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeCreateSealed decodes a CreateSealed response.
func DecodeCreateSealed(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode CreateSealed response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// CreateSealed
func CreateSealed(rw io.ReadWriter, keyBlob []byte) ([]byte, error) {
/*
	TPM_HANDLE parent_handle,
        int size_policy_digest, byte* policy_digest,
        string& parentAuth,
        int size_to_seal, byte* to_seal,
        TPML_PCR_SELECTION& pcr_selection,
        TPM_ALG_ID int_alg,
        TPMA_OBJECT& flags, TPM_ALG_ID sym_alg,
        TPMI_AES_KEY_BITS sym_key_size,
        TPMI_ALG_SYM_MODE sym_mode, TPM_ALG_ID sig_scheme,
        int mod_size, uint32_t exp,
        int* size_public, byte* out_public,
        int* size_private, byte* out_private,
        TPM2B_CREATION_DATA* creation_out,
        TPM2B_DIGEST* digest_out,
        TPMT_TK_CREATION* creation_ticket
	// Construct command
	x, err:= ConstructCreateSealed(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeCreateSealed(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeCreateSealed %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

// ConstructUnseal constructs a Unseal command.
func ConstructUnseal(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdUnseal)
	if err != nil {
		return nil, errors.New("ConstructUnseal failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeUnseal decodes a Unseal response.
func DecodeUnseal(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode Unseal response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// Unseal
func Unseal(rw io.ReadWriter, keyBlob []byte) ([]byte, error) {
/*
 	TPM_HANDLE item_handle, string& parentAuth,
        TPM_HANDLE session_handle, TPM2B_NONCE& nonce,
        byte session_attributes, TPM2B_DIGEST& hmac_digest,
        int* out_size, byte* unsealed
	// Construct command
	x, err:= ConstructUnseal(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeUnseal(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeUnseal %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

// ConstructQuote constructs a Quote command.
func ConstructQuote(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdQuote)
	if err != nil {
		return nil, errors.New("ConstructQuote failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeQuote decodes a Quote response.
func DecodeQuote(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode Quote response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// Quote
func Quote(rw io.ReadWriter, keyBlob []byte) ([]byte, error) {
/*
	TPM_HANDLE signingHandle, string& parentAuth,
        int quote_size, byte* toQuote,
        TPMT_SIG_SCHEME scheme, TPML_PCR_SELECTION& pcr_selection,
        TPM_ALG_ID sig_alg, TPM_ALG_ID hash_alg,
        int* attest_size, byte* attest, int* sig_size, byte* sig
	// Construct command
	x, err:= ConstructQuote(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeQuote(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeQuote %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

// ConstructActivateCredential constructs a ActivateCredential command.
func ConstructActivateCredential(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdActivateCredential)
	if err != nil {
		return nil, errors.New("ConstructActivateCredential failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeActivateCredential decodes a ActivateCredential response.
func DecodeActivateCredential(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode ActivateCredential response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// ActivateCredential
func ActivateCredential(rw io.ReadWriter, keyBlob []byte) ([]byte, error) {
/*
 	TPM_HANDLE activeHandle,
        TPM_HANDLE keyHandle,
        string& activeAuth, string& keyAuth,
        TPM2B_ID_OBJECT& credentialBlob,
        TPM2B_ENCRYPTED_SECRET& secret,
        TPM2B_DIGEST* certInfo
	// Construct command
	x, err:= ConstructActivateCredential(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeActivateCredential(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeActivateCredential %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

// ConstructEvictControl constructs a EvictControl command.
func ConstructEvictControl(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdEvictControl)
	if err != nil {
		return nil, errors.New("ConstructEvictControl failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeEvictControl decodes a EvictControl response.
func DecodeEvictControl(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode EvictControl response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// EvictControl
func EvictControl(rw io.ReadWriter, keyBlob []byte) ([]byte, error) {
/*
	TPMI_RH_PROVISION owner,
        TPM_HANDLE handle, string& authString,
        TPMI_DH_PERSISTENT persistantHandle
	// Construct command
	x, err:= ConstructEvictControl(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeEvictControl(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeEvictControl %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

// ConstructSaveContext constructs a SaveContext command.
func ConstructSaveContext(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdSaveContext)
	if err != nil {
		return nil, errors.New("ConstructSaveContext failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeSaveContext decodes a SaveContext response.
func  DecodeSaveContext(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode SaveContext response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// SaveContext
func SaveContext(rw io.ReadWriter, keyBlob []byte) ([]byte, error) {
/*
	TPM_HANDLE handle, int* size, byte* saveArea
	// Construct command
	x, err:= ConstructSaveContext(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeSaveContext(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeSaveContext %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

// ConstructLoadContext constructs a LoadContext command.
func ConstructLoadContext(keyBlob []byte) ([]byte, error) {
/*
	cmdHdr, err := MakeCommandHeader(tagNO_SESSIONS, 0, cmdLoadContext)
	if err != nil {
		return nil, errors.New("ConstructLoadContext failed")
	}
	num_bytes :=  []interface{}{uint16(size)}
	x, _ := packWithHeader(cmdHdr, num_bytes)
	return x, nil
*/
	return nil, nil
}

// DecodeLoadContext decodes a LoadContext response.
func  DecodeLoadContext(in []byte) ([]byte, error) {
/*
        var rand_bytes []byte

        out :=  []interface{}{&rand_bytes}
        err := unpack(in, out)
        if err != nil {
                return nil, errors.New("Can't decode LoadContext response")
        }

        return rand_bytes, nil
*/
	return nil, nil
}

// LoadContext
func LoadContext(rw io.ReadWriter, keyBlob []byte) ([]byte, error) {
/*
	int size, byte* saveArea, TPM_HANDLE* handle
	// Construct command
	x, err:= ConstructLoadContext(size)
	if err != nil {
		fmt.Printf("MakeCommandHeader failed %s\n", err)
		return nil, err
	}

	// Send command
	_, err = rw.Write(x)
	if err != nil {
		return nil, errors.New("Write Tpm fails") 
	}

	// Get response
	var resp []byte
	resp = make([]byte, 1024, 1024)
	read, err := rw.Read(resp)
        if err != nil {
                return nil, errors.New("Read Tpm fails")
        }

	// Decode Response
        if read < 10 {
                return nil, errors.New("Read buffer too small")
	}
	tag, size, status, err := DecodeCommandResponse(resp[0:10])
	if err != nil {
		fmt.Printf("DecodeCommandResponse %s\n", err)
		return nil, err
	}
	fmt.Printf("Tag: %x, size: %x, error code: %x\n", tag, size, status)  // remove
	if status != errSuccess {
	}
	rand, err :=  DecodeLoadContext(resp[10:read])
	if err != nil {
		fmt.Printf("DecodeLoadContext %s\n", err)
		return nil,err
	}
	return rand, nil
*/
	return nil, nil
}

/*
bool ComputeQuotedValue(TPM_ALG_ID alg, int credInfo_size, byte* credInfo,
                        int* size_quoted, byte* quoted)

ComputeMakeCredentialData()
 */
