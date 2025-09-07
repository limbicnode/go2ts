# Makefile for go2ts
GO ?= go
GOFMT ?= gofmt "-s"
APP_NAME := go2ts
GOFILES := $(shell find . -name "*.go")
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_DIR := bin
CMD_DIR := ./cmd/go2ts
PKG_DIR := ./pkg/go2ts
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(shell git rev-parse --short HEAD) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

.PHONY: help fmt vet lint build clean test coverage benchmark release info misspell misspell-check

## help          
help:
	@echo "Usage: make [target]"; echo
	@echo "Available targets:"
	@awk '/^##/ { printf "  %-15s %s\n", $$2, substr($$0, index($$0,$$3)) }' $(MAKEFILE_LIST)

## misspell      
misspell:
	@echo "Running misspell to fix common typos..."
	@if ! command -v misspell >/dev/null 2>&1; then \
		echo "⚠️  misspell not found, installing..."; \
		$(GO) install github.com/client9/misspell/cmd/misspell@latest; \
	fi
	misspell -w $(GOFILES)
	@echo "✅ misspell corrections applied"

## misspell-check
misspell-check:
	@echo "Checking code for common typos..."
	@if ! command -v misspell >/dev/null 2>&1; then \
		echo "⚠️  misspell not found, installing..."; \
		$(GO) install github.com/client9/misspell/cmd/misspell@latest; \
	fi
	misspell -error $(GOFILES)
	@echo "✅ misspell check completed"

## vet         
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✅ Go vet completed"

fmt:
	@echo "Formatting code..."
	@if command -v gofumpt >/dev/null 2>&1; then \
		gofumpt -w .; \
	else \
		gofmt -s -w .; \
	fi
	@echo "✅ Code formatting completed"

## lint         
lint:
	@echo "Running golangci-lint..."
	@$(shell go env GOPATH)/bin/golangci-lint run --verbose --timeout=10m
	@echo "✅ Linting completed"

## build        
build: fmt
	@echo "Building $(APP_NAME) CLI version $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo "✅ Build completed: $(BUILD_DIR)/$(APP_NAME)"

## clean          
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.txt coverage.html
	@rm -f benchmark-results.txt
	@go clean -cache -testcache
	@echo "✅ Clean completed"

## test          
test: fmt
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.txt ./...
	@echo "✅ Tests completed"

## test-short    
test-short: fmt
	@echo "Running tests (short mode)..."
	@go test -v -short ./...
	@echo "✅ Tests completed"

## coverage      
coverage: test
	@echo "Generating coverage report..."
	@go tool cover -func=coverage.txt
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

## benchmark     
benchmark: fmt
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./... > benchmark-results.txt
	@echo "✅ Benchmarks completed: benchmark-results.txt"

## ci            
ci: fmt vet lint misspell-check test coverage
	@echo "✅  All CI checks passed!"

## release        
release: clean build-all
	@echo "Creating release for version $(VERSION)..."
	@tar -czf $(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-linux-amd64
	@tar -czf $(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-darwin-amd64
	@zip $(APP_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe
	@echo "✅ Release packages created:"
	@ls -la $(APP_NAME)-$(VERSION)-*.tar.gz $(APP_NAME)-$(VERSION)-*.zip

## install       
install: build
	@echo "Installing $(APP_NAME)..."
	@cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/ || sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	@echo "✅ $(APP_NAME) installed to /usr/local/bin/"

## uninstall     
uninstall:
	@echo "Uninstalling $(APP_NAME)..."
	@rm -f /usr/local/bin/$(APP_NAME) || sudo rm -f /usr/local/bin/$(APP_NAME)
	@echo "✅ $(APP_NAME) uninstalled"

## info          
info:
	@echo "App: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Go version: $(shell go version)"
	@echo "OS/ARCH: $(shell go env GOOS)/$(shell go env GOARCH)"
	@echo "Build time: $(shell date -u +%Y-%m-%dT%H:%M:%SZ)"
	@echo "Git commit: $(shell git rev-parse --short HEAD)"
	@echo "Git status: $(shell git status --porcelain | wc -l) files changed"