.PHONY: all build dist website clean

GENERATED = generated
DIST = dist
GO_SRC = image

all: dist

# Full build: generate bindings, copy runtime, compile WASM
build:
	go run ../../main.go --output $(GENERATED) $(GO_SRC)/main.go

# Build dist directory for serving
dist: build
	mkdir -p $(DIST)
	cp $(GENERATED)/worker.js $(GENERATED)/$(GO_SRC).wasm $(GENERATED)/wasm_exec.js $(DIST)/
	npx esbuild src/demo.ts --bundle --outfile=$(DIST)/demo.js

# Copy to website static directory
website: dist
	mkdir -p ../../website/static/examples/image-processing
	cp $(DIST)/* ../../website/static/examples/image-processing/

clean:
	rm -rf $(GENERATED) $(DIST)
	rm -f $(GO_SRC)/bindings_gen.go
