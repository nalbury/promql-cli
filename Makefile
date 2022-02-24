OS ?= darwin ## OS we're building for (e.g. linux)
ARCH ?= amd64 ## Arch we're building for  (e.g. amd64)
VERSION ?= master ## Version we're releasing
INSTALL_PATH ?= /usr/local/bin ## Install path for make install

GOOS = $(strip $(OS))
GOARCH = $(strip $(ARCH))
V = $(strip $(VERSION))
P = $(strip $(INSTALL_PATH))

BUILD_PATH = ./build/bin
ARTIFACT_PATH = ./build/artifacts

export GO_BUILD=GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_PATH)/promql_$(GOOS)_$(GOARCH) ./
export TAR=tar -czvf $(ARTIFACT_PATH)/promql-$(V)-$(GOOS)-$(GOARCH).tar.gz -C $(BUILD_PATH) promql

build: setup ## Build promql binary
		$(GO_BUILD)

setup: ## Setup build/artifact paths.
		mkdir -p $(BUILD_PATH)
		mkdir -p $(ARTIFACT_PATH)

clean: ## Cleanup build dir
		rm -rf ./build/*

build-all: ## Build binaries for linux and macOS
		OS="darwin" ARCH="amd64" make build
		OS="linux" ARCH="amd64" make build

build-artifact: setup ## Build binary and create release artifact
		$(GO_BUILD)

release: ## Build binaries and create release artifacts for both linux and macOS
		OS="darwin" ARCH="amd64" make build-artifact
		OS="darwin" ARCH="arm64" make build-artifact
		OS="linux" ARCH="amd64" make build-artifact
		OS="linux" ARCH="arm64" make build-artifact

install: ## Build binary and install to the specified install path (default /usr/local/bin)
		$(GO_BUILD)
		cp $(BUILD_PATH)/promql $(P)/promql

help: ## Print Makefile help
	@echo "Makefile for promql"
	@echo "#### Examples ####"
	@echo "Build a linux binary:"
	@echo "   OS=linux make build"
	@echo "Build both linux and macOS binaries and create release artifacts:"
	@echo "   VERSION=v0.2.1 make release"
	@echo "#### Environment Variables ####"
	@awk '$$4 == "##" {gsub(/\?=./, "", $$0); $$2="(default: "$$2")"; printf "-- %s \n", $$0}' Makefile
	@echo "#### Targets ####"
	@awk '$$1 ~ /^.*:$$/ {gsub(":", "", $$1);printf "-- %s \n", $$0}' Makefile
