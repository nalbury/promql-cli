GOOS ?= darwin ## OS we're building for (e.g. linux)
GOARCH ?= amd64 ## Arch we're building for  (e.g. amd64)
VERSION ?= master ## Version we're releasing

BUILD_PATH = ./build/bin/$(GOOS)/$(GOARCH)
ARTIFACT_PATH = ./build/artifacts

export GO_BUILD=GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_PATH)/promql ./
export TAR=tar -czvf $(ARTIFACT_PATH)/promql-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz -C $(BUILD_PATH) promql

help: ## Print Makefile help
	@echo "Makefile for promql"
	@echo "#### Examples ####"
	@echo "Build a linux binary:"
	@echo "   GOOS=linux make build"
	@echo "Build both linux and macOS binaries and create release artifacts:"
	@echo "   VERSION=v0.2.1 make release"
	@echo "#### Environment Variables ####"
	@awk '$$4 == "##" {gsub(/\?=./, "", $$0); $$2="(default: "$$2")"; printf "-- %s \n", $$0}' Makefile
	@echo "#### Targets ####"
	@awk '$$1 ~ /^.*:$$/ {gsub(":", "", $$1);printf "-- %s \n", $$0}' Makefile

setup: ## Setup build/artifact paths.
	mkdir -p $(BUILD_PATH)
	mkdir -p $(ARTIFACT_PATH)

clean: ## Cleanup build dir
	 rm -rf ./build/*

build: setup ## Build promql binary
	$(GO_BUILD)

build-all: ## Build binaries for linux and macOS
	GOOS="darwin" GOARCH="amd64" make build
	GOOS="linux" GOARCH="amd64" make build

build-artifact: setup ## Build binary and create release artifact
	$(GO_BUILD)
	$(TAR)

release: ## Build binaries and create release artifactors for both linux and macOS
	GOOS="darwin" GOARCH="amd64" make build-artifact
	GOOS="linux" GOARCH="amd64" make build-artifact
