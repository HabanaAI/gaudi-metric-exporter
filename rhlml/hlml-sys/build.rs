// Copyright (C) 2025 Intel Corporation

// This program is free software; you can redistribute it and/or modify it
// under the terms of the GNU General Public License version 2 or later, as published
// by the Free Software Foundation.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program; if not, see <http://www.gnu.org/licenses/>.

// SPDX-License-Identifier: GPL-2.0-or-later

use std::path::PathBuf;

fn main() {
    let libdir_path = PathBuf::from(".")
        // Canocilaclie the path as 'rustc-link-search' requires an absolut path.
        .canonicalize()
        .expect("cannot canonicalize path");

    // Load header file from system. This is the header file that we want to generate bindings for.
    let headers_path = PathBuf::from("/usr/include/habanalabs/hlml.h");
    let headers_path_str = headers_path.to_str().expect("Path is not a valid string");

    // Tell cargo to look for shared libraries in the specified directory
    println!("cargo:rustc-link-search={}", libdir_path.to_str().unwrap());
    println!("cargo:rustc-link-search=/usr/lib/habanalabs/");
    println!("cargo:rustc-link-lib=hlml");

    // Tell cargo to tell rustc to link our `hello` library. Cargo will
    // automatically know it must look for a `libhello.a` file.
    // println!("cargo:rustc-link-lib=hlml");

    // Tell cargo to invalidate the built crate whenever the header changes.
    println!("cargo:rerun-if-changed={}", headers_path_str);

    let bindings = bindgen::Builder::default()
        .header(headers_path_str)
        .parse_callbacks(Box::new(bindgen::CargoCallbacks::new()))
        .generate()
        .expect("Unable to generate binding");

    // Write the bindings to the $OUT_DIR/bindings.rs file.
    // let out_path = PathBuf::from(env::var("OUT_DIR").unwrap())
    //     .join("hlml/")
    //     .join("bindings.rs");
    let out_path = libdir_path.join("bindings.rs");
    bindings
        .write_to_file(out_path)
        .expect("Couldn't write bindings!");
}
