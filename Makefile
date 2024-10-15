.DEFAULT_GOAL := help

PROJECT_NAME := tmff-discord-app

.PHONY: help
help:
	@echo "------------------------------------------------------------------------"
	@echo "${PROJECT_NAME}"
	@echo "------------------------------------------------------------------------"
	@grep -E '^[a-zA-Z0-9_/%\-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONMY: build_binary
build_binary: ## Build the application
	GOOS=linux GOARCH=amd64 go build -o tmff cmd/main.go

.PHONY: tools
make tools: ## Install required tools
	go install github.com/go-bindata/go-bindata/go-bindata@latest

.PHONY: run
run: ## Runs the application
	go run cmd/main.go

.PHONY: build
build: check test ## Checks and tests

.PHONY: check
check: ## Runs code checks
	docker run -t --rm -v $(PWD):/app -v ~/.cache/golangci-lint/v1.61.0:/root/.cache -w /app golangci/golangci-lint:v1.61.0 golangci-lint run

.PHONY: fix
fix: ## Fix trivial linting issues
	docker run -t --rm -v $(PWD):/app -v ~/.cache/golangci-lint/v1.61.0:/root/.cache -w /app golangci/golangci-lint:v1.61.0 golangci-lint run --fix

.PHONY: mock
mock: ## Generate mocks
	docker run -v $(PWD):/src -w /src vektra/mockery --all

.PHONY: test
test: ## Runs unit tests
	CGO_ENABLED=1 go run gotest.tools/gotestsum@latest -- -race ./...

.PHONY: package_migration
package_migration: ## Package the migration
	go-bindata  -prefix "db/migrations/" -o db/bindata.go -pkg db db/migrations/...

.PHONY: compile_for_pi
compile_for_pi: ## Compiles the application for Raspberry Pi
	./scripts/build-for-pi.sh