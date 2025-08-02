#!/bin/bash
git clone https://github.com/libbpf/blazesym.git
cd blazesym
cargo build -p blazesym-c
sudo cp target/debug/libblazesym_c.a /usr/local/lib/
sudo cp capi/include/blazesym.h /usr/local/include/
