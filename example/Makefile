.PHONY: all build generate verify serve clean setup

# Build WASM, generate types, and verify
all: setup build generate verify

# Copy wasm_exec.js from TinyGo installation
setup:
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" .

# Build the WASM binary with TinyGo (optimized for size)
# Flags:
#   -target wasm    Target WebAssembly
#   -opt=z          Optimize aggressively for size (like gcc -Oz)
#   -no-debug       Strip debug symbols and DWARF info
#   -panic=trap     Use wasm trap instruction for panics instead of full panic
#                   handler (smaller but no stack traces on panic)
build:
	tinygo build -o example.wasm \
		-target wasm \
		-opt=z \
		-no-debug \
		-panic=trap \
		./wasm/

# Debug build with TinyGo (larger but has debug info and panic messages)
build-debug:
	tinygo build -o example.wasm -target wasm ./wasm/

# Alternative: Build with standard Go (produces ~2.4MB binary)
build-go:
	GOOS=js GOARCH=wasm go build -o example.wasm ./wasm/

setup-go:
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" .

# Generate TypeScript declarations using go-wasm-ts-gen
generate:
	go run ../cmd/go-wasm-ts-gen/main.go --tests "wasm/*_test.go" --output types.d.ts

# Verify generated types work with Deno
verify:
	deno test --allow-read verify_test.ts

# Start a local web server to test the demo
serve: all
	@echo "Starting server at http://localhost:8080"
	@echo "Open http://localhost:8080/web/ in your browser"
	python3 -m http.server 8080

# Clean generated files
clean:
	rm -f example.wasm types.d.ts wasm_exec.js
