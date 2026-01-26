.PHONY: all build test lint vet format check clean test-e2e test-example-browser test-website

# Default target: check code quality and build
all: check build

# Build the CLI binary
build:
	go build -o bin/gowasm-bindgen .

# Run all tests
test:
	go test -v ./...

# Run golangci-lint (excludes examples which needs WASM build tags)
lint:
	golangci-lint run . ./internal/...
	GOOS=js GOARCH=wasm golangci-lint run ./examples/...

# Run go vet (excludes examples which needs WASM build tags)
vet:
	go vet . ./internal/...
	GOOS=js GOARCH=wasm go vet ./examples/...

# Format code with gofmt
format:
	gofmt -w .

# Run format, lint, and test
check: format lint test

# Remove build artifacts and generated test files
clean:
	rm -rf bin/
	rm -rf test/e2e/generated/
	rm -f test/e2e/wasm/bindings_gen.go
	go clean -cache -testcache

# End-to-end test: build WASM, generate bindings, run TypeScript tests
test-e2e: build
	# Generate bindings and build WASM
	./bin/gowasm-bindgen test/e2e/wasm/main.go --output test/e2e/generated --mode sync --compiler go
	# Run TypeScript tests
	npx tsx --test test/e2e/verify_test.ts

# Browser test: build example and run Playwright tests
test-example-browser:
	$(MAKE) -C examples/simple dist
	cd examples/simple && npx playwright test

# Build website with all examples (requires hugo, tinygo, node)
test-website: build
	cd examples/image-processing && npm install && $(MAKE) website
	cd examples/js-sandbox && npm install && $(MAKE) website
	cd website && hugo --minify
