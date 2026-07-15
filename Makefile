WASM_OUT = web/main.wasm
GOROOT_WASM = $(shell go env GOROOT)/lib/wasm/wasm_exec.js

.PHONY: build serve dev clean

build:
	GOOS=js GOARCH=wasm go build -o $(WASM_OUT) .
	cp $(GOROOT_WASM) web/wasm_exec.js

serve:
	cd web && python3 -m http.server 8080

dev: build serve

clean:
	rm -f $(WASM_OUT) web/wasm_exec.js
