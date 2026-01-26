.PHONY: all build test generate generate-sync typecheck verify serve clean format lint check dist

# Directories
GO_SRC := wasm
SRC := src
PUBLIC := public
GENERATED := generated
DIST := dist

# Build everything: test, generate, compile WASM, verify types, bundle for serving
all: test build typecheck verify dist

# Run Go unit tests
test:
	go test -v ./$(GO_SRC)/...

# Full build: generate bindings, copy runtime, compile WASM
build:
	go run ../../main.go --output $(GENERATED) $(GO_SRC)/main.go

# Generate only (skip WASM compilation)
generate:
	go run ../../main.go --output $(GENERATED) --no-build $(GO_SRC)/main.go

# Generate synchronous API (blocks main thread)
generate-sync:
	go run ../../main.go --output $(GENERATED) --no-build --mode sync $(GO_SRC)/main.go

# Build with standard Go (larger binary, ~2.4MB)
build-go:
	go run ../../main.go --output $(GENERATED) --compiler go $(GO_SRC)/main.go

# Build without optimizations (larger but has debug info)
build-debug:
	go run ../../main.go --output $(GENERATED) --optimize=false $(GO_SRC)/main.go

# Type-check TypeScript (strict verification that generated types work)
typecheck:
	npx tsc --noEmit --project tsconfig.json

# Run TypeScript tests (also exercises the WASM functions at runtime)
verify:
	npx tsx --test $(SRC)/verify_test.ts

# Build dist directory for serving
dist: build
	mkdir -p $(DIST)
	cp $(GENERATED)/worker.js $(GENERATED)/$(GO_SRC).wasm $(GENERATED)/wasm_exec.js $(DIST)/
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
lint: build
	GOOS=js GOARCH=wasm golangci-lint run ./$(GO_SRC)/...
	npx tsc --noEmit --project tsconfig.json

# Run format and lint
check: format lint

# Clean generated files
clean:
	rm -rf $(GENERATED) $(DIST)
	rm -f $(GO_SRC)/bindings_gen.go
