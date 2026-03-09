APP_NAME := anbu
VERSION ?= dev-build
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CYAN := \033[0;36m
GREEN := \033[0;32m
NC := \033[0m
.PHONY: help clean build build-for build-all version assets
.DEFAULT_GOAL := help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-15s$(NC) %s\n", $$1, $$2}'

clean: ## Remove build artifacts
	rm -f $(APP_NAME)
	rm -rf dist/

build: ## Build for current platform
	@go build -ldflags="-s -w -X 'github.com/tanq16/anbu/cmd.AppVersion=$(VERSION)'" -o $(APP_NAME) .
	@echo "$(GREEN)Built: ./$(APP_NAME)$(NC)"

build-for: ## Build for specific platform (GOOS=linux GOARCH=amd64)
	@mkdir -p dist
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w -X 'github.com/tanq16/anbu/cmd.AppVersion=$(VERSION)'" -o dist/$(APP_NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe,) .

build-all: ## Build for all platforms
	@mkdir -p dist
	$(MAKE) build-for GOOS=linux GOARCH=amd64
	$(MAKE) build-for GOOS=linux GOARCH=arm64
	$(MAKE) build-for GOOS=darwin GOARCH=amd64
	$(MAKE) build-for GOOS=darwin GOARCH=arm64
	$(MAKE) build-for GOOS=windows GOARCH=amd64
	$(MAKE) build-for GOOS=windows GOARCH=arm64

STATIC_DIR := internal/generics/static

assets: ## Download frontend assets (JS, CSS, fonts)
	@mkdir -p $(STATIC_DIR)/js $(STATIC_DIR)/css $(STATIC_DIR)/fonts
	@echo "Downloading JS assets..."
	@curl -sL -o $(STATIC_DIR)/js/marked.min.js "https://cdn.jsdelivr.net/npm/marked@14/marked.min.js"
	@curl -sL -o $(STATIC_DIR)/js/mermaid.min.js "https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"
	@curl -sL -o $(STATIC_DIR)/js/highlight.min.js "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"
	@curl -sL -o $(STATIC_DIR)/js/tailwindcss.js "https://cdn.tailwindcss.com/3.4.16"
	@curl -sL -o $(STATIC_DIR)/js/lucide.min.js "https://unpkg.com/lucide@latest/dist/umd/lucide.min.js"
	@echo "Downloading CSS assets..."
	@curl -sL -o $(STATIC_DIR)/css/github-dark.min.css "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css"
	@echo "Downloading Inter font..."
	@curl -sL -A "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36" \
		-o /tmp/inter.css "https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap"
	@grep -o 'https://[^)]*' /tmp/inter.css | while read url; do \
		filename=$$(basename "$$url"); \
		curl -sL -o $(STATIC_DIR)/fonts/"$$filename" "$$url"; \
	done
	@sed -E 's|url\(https://[^)]*/([^/)]*)\)|url(../fonts/\1)|g' /tmp/inter.css > $(STATIC_DIR)/css/inter.css
	@echo "Downloading JetBrains Mono font..."
	@curl -sL -A "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36" \
		-o /tmp/jetbrains-mono.css "https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700&display=swap"
	@grep -o 'https://[^)]*' /tmp/jetbrains-mono.css | while read url; do \
		filename=$$(basename "$$url"); \
		curl -sL -o $(STATIC_DIR)/fonts/"$$filename" "$$url"; \
	done
	@sed -E 's|url\(https://[^)]*/([^/)]*)\)|url(../fonts/\1)|g' /tmp/jetbrains-mono.css > $(STATIC_DIR)/css/jetbrains-mono.css
	@rm -f /tmp/inter.css /tmp/jetbrains-mono.css
	@echo "$(GREEN)All assets downloaded to $(STATIC_DIR)/$(NC)"

version: ## Calculate next version from git tags and commit message
	@LATEST_TAG=$$(git tag --sort=-v:refname | head -n1 || echo "0.0.0"); \
	LATEST_TAG=$${LATEST_TAG#v}; \
	MAJOR=$$(echo "$$LATEST_TAG" | cut -d. -f1); \
	MINOR=$$(echo "$$LATEST_TAG" | cut -d. -f2); \
	PATCH=$$(echo "$$LATEST_TAG" | cut -d. -f3); \
	MAJOR=$${MAJOR:-0}; MINOR=$${MINOR:-0}; PATCH=$${PATCH:-0}; \
	COMMIT_MSG=$$(git log -1 --pretty=%B 2>/dev/null || echo ""); \
	if echo "$$COMMIT_MSG" | grep -q "\[major-release\]"; then \
		MAJOR=$$((MAJOR + 1)); MINOR=0; PATCH=0; \
	elif echo "$$COMMIT_MSG" | grep -q "\[minor-release\]"; then \
		MINOR=$$((MINOR + 1)); PATCH=0; \
	else \
		PATCH=$$((PATCH + 1)); \
	fi; \
	echo "v$$MAJOR.$$MINOR.$$PATCH"
