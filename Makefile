APP=reflex
PKG=./cmd/reflex

.PHONY: build run test lint fmt clean

build:
	go build -o bin/$(APP) $(PKG)

run:
	go run $(PKG) --help

test:
	go test ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin

