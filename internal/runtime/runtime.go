// Package runtime provides embedded TypeScript declarations for the Go WASM runtime.
package runtime

import _ "embed"

//go:embed wasm_exec.d.ts
var WasmExecDTS string
