.PHONY: build test lint format check clean test-e2e

# Build the CLI binary
build:
	go build -o bin/go-wasm-ts-gen ./cmd/go-wasm-ts-gen

# Run all tests
test:
	go test -v ./...

# Run golangci-lint
lint:
	golangci-lint run

# Format code with gofmt
format:
	gofmt -w .

# Run format, lint, and test
check: format lint test

# Remove build artifacts
clean:
	rm -rf bin/
	go clean -cache -testcache

# End-to-end test: build WASM, generate .d.ts, run Deno tests
test-e2e: build
	# Copy wasm_exec.js from Go installation
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" testdata/e2e/
	# Build WASM binary
	GOOS=js GOARCH=wasm go build -o testdata/e2e/test.wasm ./testdata/e2e/wasm/
	# Generate TypeScript declarations
	./bin/go-wasm-ts-gen --tests "testdata/e2e/wasm/*_test.go" --output testdata/e2e/test.d.ts
	# Run Deno tests
	deno test --allow-read testdata/e2e/verify_test.ts
