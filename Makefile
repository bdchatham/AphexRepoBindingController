# Get the currently used golang install path
GOPATH ?= $(shell go env GOPATH)
GOBIN ?= $(GOPATH)/bin

# Tool versions
GOLANGCI_LINT_VERSION ?= v1.64.8

GOLANGCI_LINT := $(GOBIN)/golangci-lint

# Image settings
IMG ?= ghcr.io/bdchatham/repobinding-controller
TAG ?= latest

.PHONY: all
all: fmt vet lint build

##@ Development

.PHONY: fmt
fmt: ## Run go fmt
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint
	$(GOLANGCI_LINT) run --timeout=5m

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint with auto-fix
	$(GOLANGCI_LINT) run --fix --timeout=5m

.PHONY: test
test: fmt vet ## Run tests
	go test ./... -v -race -coverprofile=coverage.out

.PHONY: test-short
test-short: ## Run tests without race detector
	go test ./... -v -short

.PHONY: coverage
coverage: test ## Generate coverage report
	go tool cover -html=coverage.out -o coverage.html

##@ Build

.PHONY: build
build: fmt vet ## Build the binary
	go build -o bin/controller ./main.go

.PHONY: run
run: fmt vet ## Run the controller locally
	go run ./main.go

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: verify-tidy
verify-tidy: tidy ## Verify go.mod is tidy
	@if [ -n "$$(git status --porcelain go.mod go.sum)" ]; then \
		echo "go.mod or go.sum is not tidy. Run 'go mod tidy' and commit the changes."; \
		git diff go.mod go.sum; \
		exit 1; \
	fi

##@ Container

.PHONY: docker-build
docker-build: ## Build docker image
	docker build -t $(IMG):$(TAG) .

.PHONY: docker-push
docker-push: ## Push docker image
	docker push $(IMG):$(TAG)

##@ Tools

.PHONY: golangci-lint
golangci-lint: ## Install golangci-lint
	@test -f $(GOLANGCI_LINT) || \
		(echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..." && \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION))

.PHONY: tools
tools: golangci-lint ## Install all tools

##@ CI

.PHONY: ci
ci: tidy fmt vet lint test ## Run all CI checks

##@ Help

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
