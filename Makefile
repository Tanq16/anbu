APP_NAME := anbu
VERSION ?= dev-build
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CYAN := \033[0;36m
GREEN := \033[0;32m
NC := \033[0m
.PHONY: help clean build build-for build-all version
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
