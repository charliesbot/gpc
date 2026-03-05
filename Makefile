BINARY := gpc
MODULE := github.com/charliesbot/gpc
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build install test lint build-all clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY) .

install:
	go install $(LDFLAGS) .

test:
	go test ./... -v

lint:
	golangci-lint run ./...

build-all:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe .

clean:
	rm -rf bin/
