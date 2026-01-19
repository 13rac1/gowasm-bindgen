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
	rm -f testdata/e2e/test.wasm testdata/e2e/client.ts testdata/e2e/wasm_exec.js testdata/e2e/wasm/bindings_gen.go
	go clean -cache -testcache

# End-to-end test: build WASM, generate bindings, run TypeScript tests
test-e2e: build
	# Copy wasm_exec.js from Go installation
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" testdata/e2e/
	# Generate TypeScript client and Go bindings
	./bin/gowasm-bindgen --output testdata/e2e/client.ts --go-output testdata/e2e/wasm/bindings_gen.go --sync testdata/e2e/wasm/main.go
	# Build WASM binary (includes generated bindings)
	GOOS=js GOARCH=wasm go build -o testdata/e2e/test.wasm ./testdata/e2e/wasm/
	# Run TypeScript tests
	npx tsx --test testdata/e2e/verify_test.ts

# Browser test: build example and run Playwright tests
test-example-browser:
	$(MAKE) -C example dist
	cd example && npx playwright test
