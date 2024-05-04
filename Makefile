.PHONY: dev

dev:
	@air --build.cmd="go build -o bin/ultraphx-core cmd/core/main.go" --build.bin="./bin/ultraphx-core"

.PHONY: build

build:
	@go build -o bin/ultraphx-core -ldflags "-s -w" cmd/core/main.go