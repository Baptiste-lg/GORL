WASM_OUT = web/main.wasm

.PHONY: build serve dev clean

build:
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $(WASM_OUT) .
	@# Go >=1.24 uses lib/wasm, older versions use misc/wasm
	@if [ -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" web/; \
	elif [ -f "$$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/; \
	else \
		echo "ERROR: wasm_exec.js not found in GOROOT"; exit 1; \
	fi

serve:
	cd web && python3 -m http.server 8080

dev: build serve

clean:
	rm -f $(WASM_OUT) web/wasm_exec.js
