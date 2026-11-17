// Package web holds the page tmuxbox serves.
package web

import "embed"

// Files is the page and everything it loads. lib/bean.wasm and
// lib/wasm_exec.js are built by `make web` before this package compiles.
//
//go:embed index.html lib
var Files embed.FS
