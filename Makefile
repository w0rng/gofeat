BIN_DIR := $(CURDIR)/bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
GOLANGCI_LINT_VERSION := v2.8.0

export PATH := $(BIN_DIR):$(PATH)

.PHONY: test bench lint lint-install

test:
	go test -v -race ./...

bench:
	go test -run=^$$ -bench=. ./...

lint: $(GOLANGCI_LINT)
	golangci-lint run

lint-install: $(GOLANGCI_LINT)

$(GOLANGCI_LINT):
	mkdir -p $(BIN_DIR)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b $(BIN_DIR) $(GOLANGCI_LINT_VERSION)
