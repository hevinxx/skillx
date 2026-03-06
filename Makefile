BINARY_NAME ?= skillx
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
MODULE := github.com/hevinxx/skillx
LDFLAGS := -X main.version=$(VERSION) -X main.binaryName=$(BINARY_NAME)

.PHONY: build clean test lint

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) .

clean:
	rm -rf bin/

test:
	go test ./...

lint:
	go vet ./...
