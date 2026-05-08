# ============================================================
# netvisor Makefile
# ============================================================
# Variables
BINARY        := netvisor
CMD_PATH      := ./cmd/netvisor
VERSION       ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS       := -ldflags "-X main.version=$(VERSION) -s -w"
# -s -w strips debug info and DWARF — makes the binary smaller for release.

GO            := go
GOFLAGS       :=
LINT          := golangci-lint

# Directories
OUT_DIR       := bin

.PHONY: all build test lint clean help tidy

## all: build + test (default target)
all: build test

## build: compile the binary into ./bin/netvisor
build:
	@mkdir -p $(OUT_DIR)
	$(GO) build $(LDFLAGS) -o $(OUT_DIR)/$(BINARY) $(CMD_PATH)
	@echo "✓ built $(OUT_DIR)/$(BINARY) (version=$(VERSION))"

## test: run all tests with race detector enabled
# The race detector instruments memory accesses at runtime to catch
# data races — concurrent reads/writes to shared memory without locking.
# It adds ~5x overhead but is essential to run in CI.
test:
	$(GO) test -race -count=1 ./...
	@echo "✓ all tests passed"

## lint: run golangci-lint (must be installed: https://golangci-lint.run/usage/install/)
lint:
	$(LINT) run ./...

## tidy: clean up go.mod and go.sum
tidy:
	$(GO) mod tidy

## clean: remove build artifacts
clean:
	rm -rf $(OUT_DIR)

## help: print this message
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^##' Makefile | sed 's/## /  /'
