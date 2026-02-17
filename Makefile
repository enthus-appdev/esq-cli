BINARY_NAME := esq
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"

.PHONY: all build install clean lint test help

all: build ## Build the application (default)

build: ## Build to bin/esq
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/esq

install: build ## Build + copy to ~/bin/
	cp bin/$(BINARY_NAME) $(HOME)/bin/$(BINARY_NAME)

clean: ## Remove build artifacts
	rm -rf bin/ dist/ coverage.out coverage.html

lint: ## Run goimports + golangci-lint
	goimports -w . && golangci-lint run

test: ## Run tests
	go test -v ./...

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
