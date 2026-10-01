.PHONY: help assets verify-assets font nerdfont fontawesome clean build build-for build-all docker-build docker-push version

APP_NAME    := anbu
DOCKER_USER := tanq16
MODULE      := github.com/tanq16/anbu

VERSION ?= dev-build
GOOS    ?= $(shell go env GOOS)
GOARCH  ?= $(shell go env GOARCH)

TAILWIND_VERSION    := 4.3.3
LUCIDE_VERSION      := 1.34.0
FONTAWESOME_VERSION := 7.3.1
XTERM_VERSION       := 6.0.0
XTERM_WEBGL_VERSION := 0.19.0
XTERM_FIT_VERSION   := 0.11.0
NERDFONT_VERSION    := 3.5.1

NERDFONT := 1

STATIC_DIR := internal/server/static
JS_DIR     := $(STATIC_DIR)/js
CSS_DIR    := $(STATIC_DIR)/css
FONTS_DIR  := $(STATIC_DIR)/fonts
FA_DIR     := $(STATIC_DIR)/fontawesome
STAMP      := $(STATIC_DIR)/.assets-stamp

MONO := $(if $(filter 1,$(NERDFONT)),nerdfont,font FAMILY="JetBrains+Mono" SLUG=jetbrains-mono WEIGHTS="400;700")

UA := Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36
UVX := uv tool run

CYAN  := \033[0;36m
GREEN := \033[0;32m
NC    := \033[0m

help: ## Show this help
	@echo "$(CYAN)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

assets: $(STAMP) ## Download pinned frontend assets (never committed)
	@:

$(STAMP): $(MAKEFILE_LIST)
	@mkdir -p $(JS_DIR) $(CSS_DIR) $(FONTS_DIR) $(FA_DIR)/css $(FA_DIR)/webfonts
	@curl -sfL "https://cdn.jsdelivr.net/npm/@tailwindcss/browser@$(TAILWIND_VERSION)" -o "$(JS_DIR)/tailwind.js"
	@curl -sfL "https://cdn.jsdelivr.net/npm/lucide@$(LUCIDE_VERSION)/dist/umd/lucide.min.js" -o "$(JS_DIR)/lucide.min.js"
	@curl -sfL "https://cdn.jsdelivr.net/npm/@xterm/xterm@$(XTERM_VERSION)/lib/xterm.js" -o "$(JS_DIR)/xterm.js"
	@curl -sfL "https://cdn.jsdelivr.net/npm/@xterm/xterm@$(XTERM_VERSION)/css/xterm.css" -o "$(CSS_DIR)/xterm.css"
	@curl -sfL "https://cdn.jsdelivr.net/npm/@xterm/addon-webgl@$(XTERM_WEBGL_VERSION)/lib/addon-webgl.js" -o "$(JS_DIR)/addon-webgl.js"
	@curl -sfL "https://cdn.jsdelivr.net/npm/@xterm/addon-fit@$(XTERM_FIT_VERSION)/lib/addon-fit.js" -o "$(JS_DIR)/addon-fit.js"
	@$(MAKE) --no-print-directory fontawesome
	@$(MAKE) --no-print-directory font FAMILY="Inter" SLUG=inter WEIGHTS="400;500;600;700"
	@$(MAKE) --no-print-directory font FAMILY="Google+Sans" SLUG=google-sans WEIGHTS="400;500;700"
	@$(MAKE) --no-print-directory $(MONO)
	@touch $(STAMP)
	@echo "$(GREEN)Assets downloaded$(NC)"

font:
	@curl -sfL -H "User-Agent: $(UA)" \
	  "https://fonts.googleapis.com/css2?family=$(FAMILY):wght@$(WEIGHTS)&display=swap" \
	  -o "$(CSS_DIR)/$(SLUG).raw"
	@awk '/^\/\* /{keep = ($$0 ~ /^\/\* latin(-ext)? \*\/$$/)} keep' \
	  "$(CSS_DIR)/$(SLUG).raw" > "$(CSS_DIR)/$(SLUG).css"
	@rm -f "$(CSS_DIR)/$(SLUG).raw"
	@grep -o 'https://fonts.gstatic.com/[^)]*' "$(CSS_DIR)/$(SLUG).css" | sort -u \
	  | xargs -P 8 -I{} sh -c 'curl -sfL "$$1" -o "$(FONTS_DIR)/$$(basename "$$1")"' _ {}
	@sed -i.bak -E 's|https://fonts\.gstatic\.com/[^)]*/([^/)]+)|/static/fonts/\1|g' "$(CSS_DIR)/$(SLUG).css"
	@rm -f "$(CSS_DIR)/$(SLUG).css.bak"

nerdfont:
	@set -e; tmp="$$(mktemp -d)"; trap 'rm -rf "$$tmp"' EXIT; \
	curl -sfL -o "$$tmp/JetBrainsMono.tar.xz" \
	  "https://github.com/ryanoasis/nerd-fonts/releases/download/v$(NERDFONT_VERSION)/JetBrainsMono.tar.xz"; \
	tar -xJf "$$tmp/JetBrainsMono.tar.xz" -C "$$tmp" \
	  JetBrainsMonoNerdFontMono-Regular.ttf JetBrainsMonoNerdFontMono-Bold.ttf; \
	for w in Regular Bold; do \
	  $(UVX) -q --from "fonttools[woff]" fonttools ttLib.woff2 compress \
	    -o "$(FONTS_DIR)/JetBrainsMonoNerdFontMono-$$w.woff2" \
	    "$$tmp/JetBrainsMonoNerdFontMono-$$w.ttf" >/dev/null 2>&1; \
	done
	@{ \
	  for pair in 400:Regular 700:Bold; do \
	    printf '@font-face{font-family:"JetBrains Mono";font-style:normal;font-weight:%s;font-display:swap;src:url("/static/fonts/JetBrainsMonoNerdFontMono-%s.woff2") format("woff2");}\n' \
	      "$${pair%%:*}" "$${pair##*:}"; \
	  done; \
	} > "$(CSS_DIR)/jetbrains-mono.css"

fontawesome:
	@curl -sfL "https://cdn.jsdelivr.net/npm/@fortawesome/fontawesome-free@$(FONTAWESOME_VERSION)/css/all.min.css" -o "$(FA_DIR)/css/all.min.css"
	@for f in fa-brands-400 fa-regular-400 fa-solid-900; do \
	  curl -sfL "https://cdn.jsdelivr.net/npm/@fortawesome/fontawesome-free@$(FONTAWESOME_VERSION)/webfonts/$$f.woff2" -o "$(FA_DIR)/webfonts/$$f.woff2"; \
	done
	@sed -i.bak 's|../webfonts/|/static/fontawesome/webfonts/|g' "$(FA_DIR)/css/all.min.css"
	@rm -f "$(FA_DIR)/css/all.min.css.bak"

verify-assets: ## Fail early if the embedded tree is missing an asset
	@test -s $(JS_DIR)/tailwind.js || (echo "tailwind.js missing, run 'make assets'" && exit 1)
	@test -s $(JS_DIR)/lucide.min.js || (echo "lucide.min.js missing, run 'make assets'" && exit 1)
	@test -s $(JS_DIR)/xterm.js || (echo "xterm.js missing, run 'make assets'" && exit 1)
	@test -s $(JS_DIR)/addon-webgl.js || (echo "addon-webgl.js missing, run 'make assets'" && exit 1)
	@test -s $(JS_DIR)/addon-fit.js || (echo "addon-fit.js missing, run 'make assets'" && exit 1)
	@test -s $(CSS_DIR)/xterm.css || (echo "xterm.css missing, run 'make assets'" && exit 1)
	@test -s $(CSS_DIR)/inter.css || (echo "inter.css missing, run 'make assets'" && exit 1)
	@test -s $(CSS_DIR)/google-sans.css || (echo "google-sans.css missing, run 'make assets'" && exit 1)
	@test -s $(CSS_DIR)/jetbrains-mono.css || (echo "jetbrains-mono.css missing, run 'make assets'" && exit 1)
	@test -s $(FA_DIR)/css/all.min.css || (echo "fontawesome missing, run 'make assets'" && exit 1)

clean: ## Remove built binaries and downloaded assets
	@rm -f $(APP_NAME) $(APP_NAME)-* $(STAMP)
	@rm -rf $(JS_DIR) $(CSS_DIR) $(FONTS_DIR) $(FA_DIR)
	@echo "$(GREEN)Cleaned$(NC)"

build: assets verify-assets ## Build for the current platform
	@go build -ldflags="-s -w -X '$(MODULE)/cmd.AppVersion=$(VERSION)'" -o $(APP_NAME) .
	@echo "$(GREEN)Built: ./$(APP_NAME)$(NC)"

build-for: verify-assets ## Build for a specific GOOS/GOARCH
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	  -ldflags="-s -w -X '$(MODULE)/cmd.AppVersion=$(VERSION)'" \
	  -o $(APP_NAME)-$(GOOS)-$(GOARCH) .
	@echo "$(GREEN)Built: ./$(APP_NAME)-$(GOOS)-$(GOARCH)$(NC)"

build-all: assets verify-assets ## Build every platform binary
	@$(MAKE) build-for GOOS=linux  GOARCH=amd64
	@$(MAKE) build-for GOOS=linux  GOARCH=arm64
	@$(MAKE) build-for GOOS=darwin GOARCH=amd64
	@$(MAKE) build-for GOOS=darwin GOARCH=arm64

docker-build: ## Build the container image for this machine
	@docker build --build-arg VERSION=$(VERSION) -t $(DOCKER_USER)/$(APP_NAME):$(VERSION) .
	@docker tag $(DOCKER_USER)/$(APP_NAME):$(VERSION) $(DOCKER_USER)/$(APP_NAME):latest

docker-push: ## Build linux/amd64 and linux/arm64 and push one manifest
	@docker buildx build --platform linux/amd64,linux/arm64 \
	  --build-arg VERSION=$(VERSION) \
	  -t $(DOCKER_USER)/$(APP_NAME):$(VERSION) \
	  -t $(DOCKER_USER)/$(APP_NAME):latest \
	  --push .

version: ## Print the next version, derived from the last commit message
	@LATEST_TAG=$$(git tag --sort=-v:refname | head -n1 || echo "0.0.0"); \
	LATEST_TAG=$${LATEST_TAG#v}; \
	MAJOR=$$(echo "$$LATEST_TAG" | cut -d. -f1); \
	MINOR=$$(echo "$$LATEST_TAG" | cut -d. -f2); \
	PATCH=$$(echo "$$LATEST_TAG" | cut -d. -f3); \
	MAJOR=$${MAJOR:-0}; MINOR=$${MINOR:-0}; PATCH=$${PATCH:-0}; \
	COMMIT_MSG="$$(git log -1 --pretty=%B)"; \
	if echo "$$COMMIT_MSG" | grep -q "\[major-release\]"; then \
		MAJOR=$$((MAJOR + 1)); MINOR=0; PATCH=0; \
	elif echo "$$COMMIT_MSG" | grep -q "\[minor-release\]"; then \
		MINOR=$$((MINOR + 1)); PATCH=0; \
	else \
		PATCH=$$((PATCH + 1)); \
	fi; \
	echo "v$${MAJOR}.$${MINOR}.$${PATCH}"
