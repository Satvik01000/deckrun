BINARY_NAME := deckrun
BUILD_DIR := bin
CMD_DIR := ./cmd/deckrun
VERSION ?= 2.0.0
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build install test clean vet

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

install:
	go install -ldflags="$(LDFLAGS)" $(CMD_DIR)
	@echo "Installed $(BINARY_NAME) to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

test:
	go test -v ./...

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)

