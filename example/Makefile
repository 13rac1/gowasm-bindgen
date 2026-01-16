.PHONY: all build generate generate-worker typecheck verify serve clean setup format lint check web dist

# Build directory for all artifacts
DIST := dist

# Build WASM, generate types, and verify
all: setup build generate typecheck verify dist

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

# Generate TypeScript client with Web Worker (default - async, non-blocking)
generate:
	go run ../cmd/gowasm-bindgen/main.go --tests "wasm/*_test.go" --output client.ts

# Generate synchronous API (blocks main thread)
generate-sync:
	go run ../cmd/gowasm-bindgen/main.go --tests "wasm/*_test.go" --output client.ts --sync

# Type-check TypeScript (strict verification that generated types work)
typecheck:
	npx tsc --noEmit --project tsconfig.json
	npx tsc --noEmit --project web/tsconfig.json

# Run TypeScript tests (also exercises the WASM functions at runtime)
verify:
	npx tsx --test verify_test.ts

# Build dist directory with runtime artifacts for serving
dist: setup build generate
	@mkdir -p $(DIST)
	cp example.wasm $(DIST)/
	cp wasm_exec.js $(DIST)/
	cp web/index.html $(DIST)/
	npx esbuild web/app.ts --bundle --outfile=$(DIST)/app.js

# Start a local web server to test the demo
serve: all
	@echo "Starting server at http://localhost:8080"
	npx serve $(DIST) -l 8080

# Compile TypeScript web demo (uses generated types.d.ts) - legacy target
web: generate
	npx esbuild web/app.ts --bundle --outfile=web/app.js

# Format code
format:
	gofmt -w .

# Lint code (requires WASM build tags for syscall/js)
lint:
	GOOS=js GOARCH=wasm golangci-lint run ./...

# Run format and lint
check: format lint

# Clean generated files
clean:
	rm -rf $(DIST)
	rm -f example.wasm wasm_exec.js web/app.js
	rm -f client.ts worker.js types.d.ts wasm_exec.d.ts
