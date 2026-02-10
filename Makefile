BINARY_NAME := esq
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

.PHONY: build install clean lint test

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/esq

install: build
	cp bin/$(BINARY_NAME) $(HOME)/bin/$(BINARY_NAME)

clean:
	rm -rf bin/

lint:
	goimports -w . && golangci-lint run

test:
	go test -v ./...
