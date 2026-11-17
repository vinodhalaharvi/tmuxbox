GOROOT := $(shell go env GOROOT)

.PHONY: build web check demo clean

# the tmuxbox command, with the page and the solver inside it
build: web
	go build -ldflags="-s -w" -o tmuxbox ./cmd/tmuxbox

# what the page needs beside what is in the repository: the solver as
# WebAssembly, and Go's loader for it
web: web/lib/bean.wasm web/lib/wasm_exec.js

web/lib/bean.wasm: go.mod $(wildcard cmd/beansolve/*.go)
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $@ ./cmd/beansolve

web/lib/wasm_exec.js:
	cp "$(GOROOT)/lib/wasm/wasm_exec.js" $@

# the solver against brute force, and its timing on ten panes
check: web
	go vet ./...
	GOOS=js GOARCH=wasm go vet ./cmd/beansolve
	go run ./cmd/beansolve

# the recorded demo, with a Linux machine behind each pane, served without tmuxbox
demo: web
	scripts/fetch-third-party.sh
	python3 -m http.server 8080 --bind 127.0.0.1 --directory web

clean:
	rm -rf tmuxbox web/lib/bean.wasm web/lib/wasm_exec.js web/vm .cache
