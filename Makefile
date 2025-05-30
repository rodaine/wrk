MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables
.SUFFIXES:
SHELL := /usr/bin/env bash -euo pipefail -c
.DELETE_ON_ERROR:
.DEFAULT_GOAL := all

GO ?= go

GO_MODS += .
GO_MODS += ./grpc

BIN = .bin
GOLANGCI_LINT_VERSION ?= v2.1.6
GOLANGCI_LINT = $(BIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

.PHONY: all
all: format lint test

.PHONY: format
format: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) fmt $(GO_MODS)
	for mod in $(foreach mod,$(GO_MODS),$(abspath $(mod))); do \
		cd $$mod; \
		$(GO) mod tidy; \
	done
	$(GO) work sync

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run $(GO_MODS)

.PHONY: test
test:
	$(GO) test -cover -race -vet=off $(foreach mod,$(GO_MODS),$(mod)/...)

$(GOLANGCI_LINT):
	GOBIN=$(abspath $(BIN)) $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	mv $(BIN)/golangci-lint $@
