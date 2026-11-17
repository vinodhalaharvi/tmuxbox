GOROOT := $(shell go env GOROOT)

.PHONY: web check serve clean

# everything the page needs: the solver as WebAssembly, Go's loader for it,
# and the third-party files
web: web/lib/bean.wasm web/lib/wasm_exec.js
	scripts/fetch-third-party.sh

web/lib/bean.wasm: go.mod $(wildcard cmd/beansolve/*.go)
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $@ ./cmd/beansolve

web/lib/wasm_exec.js:
	cp "$(GOROOT)/lib/wasm/wasm_exec.js" $@

# the solver against brute force, and its timing on ten panes
check:
	go vet ./...
	GOOS=js GOARCH=wasm go vet ./...
	go run ./cmd/beansolve

serve: web
	python3 -m http.server 8080 --bind 127.0.0.1 --directory web

clean:
	rm -rf web/lib/bean.wasm web/lib/wasm_exec.js web/lib/three.min.js web/vm .cache
