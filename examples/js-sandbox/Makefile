.PHONY: all build dist website clean

GENERATED = generated
DIST = dist
GO_SRC = goja

all: dist

# Full build: generate bindings, copy runtime, compile WASM (uses standard Go for Goja)
build:
	go run ../../main.go --output $(GENERATED) --compiler go $(GO_SRC)/main.go

# Build dist directory for serving
dist: build
	mkdir -p $(DIST)
	cp $(GENERATED)/worker.js $(GENERATED)/$(GO_SRC).wasm $(GENERATED)/wasm_exec.js $(DIST)/
	npx esbuild src/demo.ts --bundle --outfile=$(DIST)/demo.js

# Copy to website static directory
website: dist
	mkdir -p ../../website/static/examples/js-sandbox
	cp $(DIST)/* ../../website/static/examples/js-sandbox/

clean:
	rm -rf $(GENERATED) $(DIST)
	rm -f $(GO_SRC)/bindings_gen.go
