.PHONY: all build generate clean

DIST = ../../website/static/examples/js-sandbox
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
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(DIST)/
	cd $(GO_SRC) && GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o ../$(DIST)/goja.wasm .

clean:
	rm -rf $(DIST)
	rm -f $(GO_SRC)/bindings_gen.go
