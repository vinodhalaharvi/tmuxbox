#!/bin/sh
# Downloads what the recorded demo needs to open a pane as a Linux machine in
# the page: the v86 emulator, a BIOS and a small Linux disc. Each is checked
# against a known checksum. None of it is kept in the repository, and tmuxbox
# itself does not use it: see the README for what each piece is and its licence.
set -eu

root=$(cd "$(dirname "$0")/.." && pwd)
cache="$root/.cache"
mkdir -p "$cache" "$root/web/vm"

V86_COMMIT=315b2b0b758703ef0e6cc7b4088d6068f83f3c20
IMAGES_COMMIT=db92d8fd4a53dbd73b8e0588bb4bd8fe37a1975a

fetch() { # url file
	[ -f "$cache/$2" ] || curl -fsSL -o "$cache/$2" "$1"
}
check() { # file sha256
	got=$(openssl dgst -sha256 "$1" | awk '{print $NF}')
	if [ "$got" != "$2" ]; then
		echo "checksum mismatch for $1: got $got, want $2" >&2
		exit 1
	fi
}

fetch https://registry.npmjs.org/v86/-/v86-0.5.473.tgz v86.tgz
fetch "https://raw.githubusercontent.com/copy/v86/$V86_COMMIT/bios/seabios.bin" seabios.bin
fetch "https://raw.githubusercontent.com/copy/v86/$V86_COMMIT/bios/vgabios.bin" vgabios.bin
fetch "https://raw.githubusercontent.com/copy/images/$IMAGES_COMMIT/linux.iso" linux.iso

tar -xzf "$cache/v86.tgz" -C "$cache" package/build/libv86.js package/build/v86.wasm
mv "$cache/package/build/libv86.js" "$cache/package/build/v86.wasm" "$cache/"
rm -rf "$cache/package"

check "$cache/libv86.js"    e257912c4a1129e36a86c1e2df69904ab969d87bffd24aa6676e84d6ddffc1c6
check "$cache/v86.wasm"     2f0c4150f266415b754a84a32e6f8e4c975a1d3abb94c78b047fa30027db1ef1
check "$cache/seabios.bin"  73e3f359102e3a9982c35fce98eb7cd08f18303ac7f1ba6ebfbe6cdc1c244d98
check "$cache/vgabios.bin"  a4bc0d80cc3ca028c73dafa8fee396b8d054ce87ebd8abfbd31b06b437607880
check "$cache/linux.iso"    b311d065b9acc17e5bf1dd4d12f06e70efaf2f0d8b2c7be3bbbb25076f494885

cp "$cache/libv86.js" "$cache/v86.wasm" "$root/web/vm/"

# The page reads the BIOS and the Linux disc as base64 text, so that it can be
# served from hosts that only serve text and WebAssembly.
b64() { openssl base64 -A -in "$1"; }
printf '{"seabios":"%s","vgabios":"%s"}' "$(b64 "$cache/seabios.bin")" "$(b64 "$cache/vgabios.bin")" > "$root/web/vm/bios.json"
b64 "$cache/linux.iso" > "$root/web/vm/linux.txt"

echo "the demo's files are in web/vm"
