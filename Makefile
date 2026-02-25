APP_NAME := anbu
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev-build")
LDFLAGS := -s -w -X github.com/tanq16/anbu/cmd.AnbuVersion=$(VERSION)

.PHONY: build build-all clean version help

build: ## Build for current platform
	go build -ldflags="$(LDFLAGS)" -o $(APP_NAME) .

build-all: ## Build for all platforms
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(APP_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/$(APP_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(APP_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/$(APP_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(APP_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/$(APP_NAME)-windows-arm64.exe .

clean: ## Remove build artifacts
	rm -f $(APP_NAME)
	rm -rf dist/

version: ## Print current version
	@echo $(VERSION)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
