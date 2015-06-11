// rust-tao - Rust bindings for the Cloudproxy Tao
// Copyright (C) 2015  Jethro G. Beekman
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software Foundation,
// Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

extern crate libc;

use std::ptr;
use std::os::unix::io::RawFd;
use std::str;
use std::slice::from_raw_parts;
use libc::{c_void, c_int, c_char, size_t};

unsafe fn free<T>(ptr: *mut T) {
	libc::free(ptr as *mut c_void);
}

pub use native::Tao_Tao_SealPolicyDefault as TAO_SEAL_POLICY_DEFAULT;
pub use native::Tao_Tao_SealPolicyConservative as TAO_SEAL_POLICY_CONSERVATIVE;
pub use native::Tao_Tao_SealPolicyLiberal as TAO_SEAL_POLICY_LIBERAL;
pub use native::Tao_Tao_AttestationSigningContext as TAO_SEAL_POLICY_ATTESTATION_SIGNING_CONTEXT;

pub trait MessageChannel {
	fn native_ptr(&self) -> *mut native::TaoMessageChannel;
}

pub struct FDMessageChannel {
	native_ptr: *mut native::TaoFDMessageChannel,
}

impl FDMessageChannel {
	pub fn new(readfd: RawFd, writefd: RawFd) -> FDMessageChannel {
		FDMessageChannel{native_ptr:unsafe{native::Tao_FDMessageChannel_new(readfd,writefd)}}
	}
}

impl MessageChannel for FDMessageChannel {
	fn native_ptr(&self) -> *mut native::TaoMessageChannel {
		unsafe{native::Tao_FDMessageChannel_as_MessageChannel(self.native_ptr)}
	}
}

pub trait Tao: Drop {
	fn get_tao_name(&mut self) -> Option<Vec<u8>> {
		let mut name: *mut c_char=ptr::null_mut();
		let mut name_size: size_t=0;
		match unsafe{native::Tao_Tao_GetTaoName(self.native_ptr(),&mut name,&mut name_size)} {
			0 => None,
			_ => {
				let r_name=unsafe{from_raw_parts(name as *const u8,name_size as usize)}.to_owned();
				unsafe{free(name)};
				Some(r_name)
			},
		}
	}

	fn get_random_bytes(&mut self, size: usize) -> Option<Vec<u8>> {
		let mut bytes: *mut c_char=ptr::null_mut();
		let mut bytes_size: size_t=0;
		match unsafe{native::Tao_Tao_GetRandomBytes(self.native_ptr(),size as size_t,&mut bytes,&mut bytes_size)} {
			0 => None,
			_ => {
				let r_bytes=unsafe{from_raw_parts(bytes as *const u8,bytes_size as usize)}.to_owned();
				unsafe{free(bytes)};
				Some(r_bytes)
			},
		}
	}

	fn attest(&mut self, message: &[u8]) -> Option<Vec<u8>> {
		let mut attestation: *mut c_char=ptr::null_mut();
		let mut attestation_size: size_t=0;
		match unsafe{native::Tao_Tao_Attest(self.native_ptr(),message.as_ptr() as *const c_char,message.len() as size_t,&mut attestation,&mut attestation_size)} {
			0 => None,
			_ => {
				let r_attestation=unsafe{from_raw_parts(attestation as *const u8,attestation_size as usize)}.to_owned();
				unsafe{free(attestation)};
				Some(r_attestation)
			},
		}
	}

	fn seal(&mut self, data: &[u8], policy: &str) -> Option<Vec<u8>> {
		let mut sealed: *mut c_char=ptr::null_mut();
		let mut sealed_size: size_t=0;
		match unsafe{native::Tao_Tao_Seal(self.native_ptr(),data.as_ptr() as *const c_char,data.len() as size_t,policy.as_ptr() as *const c_char,policy.len() as size_t,&mut sealed,&mut sealed_size)} {
			0 => None,
			_ => {
				let r_sealed=unsafe{from_raw_parts(sealed as *const u8,sealed_size as usize)}.to_owned();
				unsafe{free(sealed)};
				Some(r_sealed)
			},
		}
	}

	fn unseal(&mut self, sealed: &[u8]) -> Option<(Vec<u8>,String)> {
		let mut data: *mut c_char=ptr::null_mut();
		let mut data_size: size_t=0;
		let mut policy: *mut c_char=ptr::null_mut();
		let mut policy_size: size_t=0;
		match unsafe{native::Tao_Tao_Unseal(self.native_ptr(),sealed.as_ptr() as *const c_char,sealed.len() as size_t,&mut data,&mut data_size,&mut policy,&mut policy_size)} {
			0 => None,
			_ => {
				let r_data=unsafe{from_raw_parts(data as *const u8,data_size as usize)}.to_owned();
				unsafe{free(data)};
				let r_policy=str::from_utf8(unsafe{from_raw_parts(policy as *const u8,policy_size as usize)}).unwrap().to_owned();
				unsafe{free(policy)};
				Some((r_data,r_policy))
			},
		}
	}

	fn native_ptr(&self) -> *mut native::TaoTao;
}

pub struct TaoRPC {
	native_ptr: *mut native::TaoTaoRPC,
}

impl TaoRPC {
	pub fn new(channel: &mut MessageChannel) -> TaoRPC {
		TaoRPC{native_ptr:unsafe{native::Tao_TaoRPC_new(channel.native_ptr())}}
	}
}

impl Tao for TaoRPC {
	fn native_ptr(&self) -> *mut native::TaoTao {
		unsafe{native::Tao_TaoRPC_as_Tao(self.native_ptr)}
	}
}

impl Drop for TaoRPC {
	fn drop(&mut self) {
		unsafe{native::Tao_Tao_delete(self.native_ptr())};
	}
}

pub fn marshal_speaksfor(key: &[u8], binary_tao_name: &[u8]) -> Option<Vec<u8>> {
	let mut out: *mut c_char=ptr::null_mut();
	let mut out_size: size_t=0;
	match unsafe{native::Tao_MarshalSpeaksfor(key.as_ptr() as *const c_char,key.len() as size_t,binary_tao_name.as_ptr() as *const c_char,binary_tao_name.len() as size_t,&mut out,&mut out_size)} {
		0 => None,
		_ => {
			let r_out=unsafe{from_raw_parts(out as *const u8,out_size as usize)}.to_owned();
			unsafe{free(out)};
			Some(r_out)
		},
	}
}

use std::ffi::OsString;
use std::os::unix::ffi::OsStrExt;
use std::collections::HashMap;
use std::mem::size_of;

fn pointer_difference<T>(a: *const T, b: *const T) -> isize {
	((a as isize)-(b as isize))/(size_of::<T>() as isize)
}

pub fn initialize_app(args: &mut Vec<OsString>, remove_args: bool) -> bool {
	// Convert args to C-style argv, argc
	let argv: Vec<*const u8>=args.iter().map(|s|s.as_bytes().as_ptr()).collect();
	let mut argc: c_int = argv.len() as i32;
	let argv_orig_p = argv.as_ptr() as *mut *mut c_char;
	let mut argv_p = argv_orig_p;

	let result=unsafe{native::Tao_InitializeApp(&mut argc, &mut argv_p, match remove_args { false => 0, true => 1 })};

	// Re-order and truncate args according to re-ordered and truncated argv
	let shifted=pointer_difference(argv_p,argv_orig_p);
	let keymap: HashMap<_,_>=argv[shifted as usize..(shifted+argc as isize) as usize].iter().enumerate().map(|(a,b)|(b,a)).collect();
	let mut i=0;
	let mut end=args.len();
	while i<end {
		match keymap.get(&args[i].as_bytes().as_ptr()) {
			Some(&idx) => {
				if i==idx {
					i+=1;
				} else {
					args.swap(i,idx);
				}
			}
			None => {
				args.swap_remove(i);
				end-=1;
			}
		}
	}

	match result { 0 => false, _ => true }
}

mod native {
	#![allow(non_upper_case_globals)]

	use libc::{c_void, c_int, c_char, size_t};

	pub const Tao_Tao_SealPolicyDefault: &'static str = "self";
	pub const Tao_Tao_SealPolicyConservative: &'static str = "few";
	pub const Tao_Tao_SealPolicyLiberal: &'static str = "any";
	pub const Tao_Tao_AttestationSigningContext: &'static str = "tao::Attestation Version 1";

	// tao::MessageChannel
	pub type TaoMessageChannel = c_void;

	// tao::FDMessageChannel
	pub type TaoFDMessageChannel = c_void;
	extern "C" {
		pub fn Tao_FDMessageChannel_new(readfd: c_int, writefd: c_int) -> *mut TaoFDMessageChannel;
		pub fn Tao_FDMessageChannel_as_MessageChannel(obj: *mut TaoFDMessageChannel) -> *mut TaoMessageChannel;
	}

	// tao::Tao
	pub type TaoTao = c_void;
	extern "C" {
		pub fn Tao_Tao_delete(obj: *mut TaoTao);
		pub fn Tao_Tao_GetTaoName(obj: *mut TaoTao, name: *mut *mut c_char, name_size: *mut size_t) -> c_int;
		pub fn Tao_Tao_GetRandomBytes(obj: *mut TaoTao, size: size_t, bytes: *mut *mut c_char, bytes_size: *mut size_t) -> c_int;
		pub fn Tao_Tao_Attest(obj: *mut TaoTao, message: *const c_char, message_size: size_t, attestation: *mut *mut c_char, attestation_size: *mut size_t) -> c_int;
		pub fn Tao_Tao_Seal(obj: *mut TaoTao, data: *const c_char, data_size: size_t, policy: *const c_char, policy_size: size_t, sealed: *mut *mut c_char, sealed_size: *mut size_t) -> c_int;
		pub fn Tao_Tao_Unseal(obj: *mut TaoTao, sealed: *const c_char, sealed_size: size_t, data: *mut *mut c_char, data_size: *mut size_t, policy: *mut *mut c_char, policy_size: *mut size_t) -> c_int;
	}

	// tao::TaoRPC
	pub type TaoTaoRPC = c_void;
	extern "C" {
		pub fn Tao_TaoRPC_new(channel: *mut TaoMessageChannel) -> *mut TaoTaoRPC;
		pub fn Tao_TaoRPC_as_Tao(obj: *mut TaoTaoRPC) -> *mut TaoTao;
	}

	// tao::InitializeApp
	extern "C" {
		pub fn Tao_InitializeApp(argc: *mut c_int, argv: *mut *mut *mut c_char, remove_args: c_int) -> c_int;
	}

	// tao::MarshalSpeaksfor
	extern "C" {
		pub fn Tao_MarshalSpeaksfor(key: *const c_char, key_size: size_t, binaryTaoName: *const c_char, binaryTaoName_size: size_t, out: *mut *mut c_char, out_size: *mut size_t) -> c_int;
	}
}
