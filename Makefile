.PHONY: all build test lint vet format check clean test-e2e test-example-browser

# Default target: check code quality and build
all: check build

# Build the CLI binary
build:
	go build -o bin/gowasm-bindgen ./cmd/gowasm-bindgen

# Run all tests
test:
	go test -v ./...

# Run golangci-lint (excludes example which needs WASM build tags)
lint:
	golangci-lint run ./cmd/... ./internal/...
	GOOS=js GOARCH=wasm golangci-lint run ./example/...

# Run go vet (excludes example which needs WASM build tags)
vet:
	go vet ./cmd/... ./internal/...
	GOOS=js GOARCH=wasm go vet ./example/...

# Format code with gofmt
format:
	gofmt -w .

# Run format, lint, and test
check: format lint test

# Remove build artifacts and generated test files
clean:
	rm -rf bin/
	rm -f testdata/e2e/test.wasm testdata/e2e/test.d.ts testdata/e2e/wasm_exec.js
	go clean -cache -testcache

# End-to-end test: build WASM, generate .d.ts, run TypeScript tests
test-e2e: build
	# Copy wasm_exec.js from Go installation
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" testdata/e2e/
	# Build WASM binary
	GOOS=js GOARCH=wasm go build -o testdata/e2e/test.wasm ./testdata/e2e/wasm/
	# Generate TypeScript declarations
	./bin/gowasm-bindgen --tests "testdata/e2e/wasm/*_test.go" --output testdata/e2e/test.d.ts
	# Run TypeScript tests
	npx tsx --test testdata/e2e/verify_test.ts

# Browser test: build example and run Playwright tests
test-example-browser:
	$(MAKE) -C example dist
	cd example && npx playwright test
