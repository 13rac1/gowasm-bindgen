.PHONY: all build test generate generate-sync typecheck verify serve clean setup format lint check dist

# Directories
GO_SRC := go
SRC := src
GENERATED := generated
PUBLIC := public
DIST := dist

# Build WASM, generate types, and verify
all: test setup generate build typecheck verify dist

# Run Go unit tests
test:
	go test -v ./$(GO_SRC)/...

# Copy wasm_exec.js from TinyGo installation
setup:
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" .

# Build the WASM binary with TinyGo (optimized for size)
build:
	tinygo build -o $(DIST)/example.wasm \
		-target wasm \
		-opt=z \
		-no-debug \
		-panic=trap \
		./$(GO_SRC)/

# Debug build with TinyGo (larger but has debug info and panic messages)
build-debug:
	tinygo build -o $(DIST)/example.wasm -target wasm ./$(GO_SRC)/

# Alternative: Build with standard Go (produces ~2.4MB binary)
build-go:
	GOOS=js GOARCH=wasm go build -o $(DIST)/example.wasm ./$(GO_SRC)/

setup-go:
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" .

# Generate TypeScript client with Web Worker (default - async, non-blocking)
generate:
	@mkdir -p $(GENERATED)
	go run ../../main.go \
		--output $(GENERATED)/client.ts \
		--go-output $(GO_SRC)/bindings_gen.go \
		--wasm-path example.wasm \
		$(GO_SRC)/main.go

# Generate synchronous API (blocks main thread)
generate-sync:
	@mkdir -p $(GENERATED)
	go run ../../main.go \
		--output $(GENERATED)/client.ts \
		--go-output $(GO_SRC)/bindings_gen.go \
		--sync \
		$(GO_SRC)/main.go

# Type-check TypeScript (strict verification that generated types work)
typecheck:
	npx tsc --noEmit --project tsconfig.json

# Run TypeScript tests (also exercises the WASM functions at runtime)
verify:
	npx tsx --test $(SRC)/verify_test.ts

# Build dist directory with runtime artifacts for serving
dist: setup generate build
	@mkdir -p $(DIST)
	cp wasm_exec.js $(DIST)/
	cp $(GENERATED)/worker.js $(DIST)/
	cp $(PUBLIC)/index.html $(DIST)/
	npx esbuild $(SRC)/app.ts --bundle --outfile=$(DIST)/app.js

# Start a local web server to test the demo
serve: all
	@echo "Starting server at http://localhost:8080"
	npx serve $(DIST) -l 8080

# Format code
format:
	gofmt -w $(GO_SRC)/

# Lint Go and TypeScript code
lint: generate
	GOOS=js GOARCH=wasm golangci-lint run ./$(GO_SRC)/...
	npx tsc --noEmit --project tsconfig.json

# Run format and lint
check: format lint

# Clean generated files
clean:
	rm -rf $(DIST)
	rm -rf $(GENERATED)
	rm -f wasm_exec.js
	rm -f $(GO_SRC)/bindings_gen.go
