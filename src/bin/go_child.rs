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

extern crate tao;

use tao::*;

use std::env::args_os;
use std::fmt::Write as FmtWrite;
use std::io::{Write,stderr};

fn hexencode(input: &[u8]) -> String {
	let mut output = String::with_capacity(input.len()*2);
	for byte in input {
		write!(output,"{:02x}",byte).unwrap();
	}
	output
}

fn main() {
	initialize_app(&mut args_os().collect(),false);

	// This code expects fd 3 and 4 to be the pipes from and to the Tao, so it
	// doesn't need to take any parameters. It will establish a Tao Child Channel
	// directly with these fds.
	let mut msg=FDMessageChannel::new(3, 4);
	let mut tao=TaoRPC::new(&mut msg);

	let bytes=
	match tao.get_random_bytes(10) {
		Some(v) => v,
		None => panic!("FATAL: Couldn't get 10 bytes from the Tao RPC channel")
	};

	if bytes.len() == 10 {
		writeln!(stderr(),"INFO: Got 10 bytes from the Tao RPC channel").unwrap();
	} else {
		panic!("FATAL: Got {} bytes from the channel, but expected 10", bytes.len());
	}

	let encoded_bytes=hexencode(&bytes);
	writeln!(stderr(),"INFO: Encoded bytes: {}",encoded_bytes).unwrap();

	let sealed=
	match tao.seal(&bytes,TAO_SEAL_POLICY_DEFAULT) {
		Some(v) => v,
		None => panic!("FATAL: Couldn't seal bytes across the channel")
	};

	let encoded_sealed=hexencode(&sealed);
	writeln!(stderr(),"INFO: Encoded sealed bytes: {}",encoded_sealed).unwrap();

	let (unsealed,policy)=
	match tao.unseal(&sealed) {
		Some(v) => v,
		None => panic!("FATAL: Couldn't unseal the tao-sealed data")
	};
	writeln!(stderr(),"INFO: Got a seal policy '{}'",policy).unwrap();

	if policy != TAO_SEAL_POLICY_DEFAULT {
		panic!("FATAL: The policy returned by Unseal didn't match the Seal policy");
	}

	if unsealed != bytes {
		panic!("FATAL: The unsealed data didn't match the sealed data");
	}

	let encoded_unsealed=hexencode(&unsealed);
	writeln!(stderr(),"INFO: Encoded unsealed bytes: {}",encoded_unsealed).unwrap();

	// Set up a fake attestation using a fake key.
	let tao_name=
	match tao.get_tao_name() {
		Some(v) => v,
		None => panic!("FATAL: Couldn't get the name of the Tao")
	};

	let msf=
	match marshal_speaksfor("This is a fake key".as_bytes(), &tao_name) {
		Some(v) => v,
		None => panic!("FATAL: Couldn't marshal a speaksfor statement")
	};

	let attest=
	match tao.attest(&msf) {
		Some(v) => v,
		None => panic!("FATAL: Couldn't attest to a fake key delegation")
	};

	let encoded_attest=hexencode(&attest);
	writeln!(stderr(),"INFO: Got attestation {}",encoded_attest).unwrap();

	writeln!(stderr(),"INFO: All Go Tao tests pass").unwrap();
}
