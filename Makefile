.PHONY: build test lint format check clean

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
