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

use std::process::Command;
use std::io::ErrorKind;
use std::env;
use std::path::Path;
use std::fs;

fn find_cloudproxy_src() -> String {
	match env::var("CLOUDPROXY_SRC") {
		Err(e) => panic!("Could not obtain CLOUDPROXY_SRC: {}",e),
		Ok(s) => s
	}
}

fn debug_or_release<'a>() -> &'a str {
	match env::var("PROFILE") {
		Ok(ref p) if p=="release" => "Release",
		_ => "Debug"
	}
}

fn need_bootstrap(src_path: &str) -> bool {
	!fs::metadata(Path::new(src_path).join("third_party/ninja/ninja")).is_ok()
}

const STATIC_LIBS: [&'static str;6]=["apps/libtao-c.a", "tao/libtao.a", "third_party/chromium/libchromium.a", "third_party/google-glog/libglog.a", "third_party/gflags/libgflags.a", "third_party/modp/libmodp.a"];
const SYSTEM_LIBS: [&'static str;7]=["pthread", "protobuf", "crypto", "ssl", "virt", "tspi", "stdc++"];

fn libpath_split(libpath: &str) -> (&str,&str) {
	let slash_pos=libpath.rfind('/').map_or(0,|v|v+1);
	let fname=&libpath[slash_pos..];
	let extstart=fname.len()-2;
	if &fname[0..3]!="lib" || &fname[extstart..]!=".a" {
		panic!("Invalid static library specification {}", libpath);
	}
	(&libpath[0..slash_pos],&fname[3..extstart])
}

fn main() {
	let src_path=find_cloudproxy_src();

	if need_bootstrap(&src_path) {
		run(Command::new("build/bootstrap.sh").current_dir(&src_path),"bootstrap.sh");
	} else {
		run(Command::new("build/config.sh").current_dir(&src_path),"config.sh");
	}
	run(Command::new("build/compile.sh").arg(debug_or_release()).current_dir(&src_path),"compile.sh");

	for lib in STATIC_LIBS.iter() {
		let (path,name)=libpath_split(lib);
		println!("cargo:rustc-link-search=native={}/out/{}/{}",src_path,debug_or_release(),path);
		println!("cargo:rustc-link-lib=static={}",name);
	}

	for lib in SYSTEM_LIBS.iter() {
		println!("cargo:rustc-link-lib={}",lib);
	}
}

// From git2-rs/build.rs
fn run(cmd: &mut Command, program: &str) {
    println!("running: {:?}", cmd);
    let status = match cmd.status() {
        Ok(status) => status,
        Err(ref e) if e.kind() == ErrorKind::NotFound => {
            fail(&format!("failed to execute command: {}\nis `{}` not installed?",
                          e, program));
        }
        Err(e) => fail(&format!("failed to execute command: {}", e)),
    };
    if !status.success() {
        fail(&format!("command did not execute successfully, got: {}", status));
    }
}

fn fail(s: &str) -> ! {
    panic!("\n{}\n\nbuild script failed, must exit now", s)
}
