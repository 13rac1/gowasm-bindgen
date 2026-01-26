.PHONY: all build generate clean

DIST = ../../website/static/examples/image-processing
GO_SRC = go

all: build

generate:
	../../bin/gowasm-bindgen \
		--ts-output $(DIST)/client.ts \
		--go-output $(GO_SRC)/bindings_gen.go \
		--mode sync \
		$(GO_SRC)/main.go

build: generate
	mkdir -p $(DIST)
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" $(DIST)/
	tinygo build -o $(DIST)/image.wasm -target wasm -no-debug ./$(GO_SRC)/

clean:
	rm -rf $(DIST)
	rm -f $(GO_SRC)/bindings_gen.go
