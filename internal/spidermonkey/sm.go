package spidermonkey

import _ "embed"

//go:embed js.wasm
var WasmBlob []byte
