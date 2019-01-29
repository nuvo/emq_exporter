# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

IMAGE_NAME := emq_exporter
IMAGE_TAG ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))
IP = $(shell docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' emqx)

GO111MODULE := on
GO ?= GO111MODULE=$(GO111MODULE) go

all: test build ## Run tests and build the binary

init:
	@echo ">> running go mod download and tidy"
	$(GO) mod download
	$(GO) mod tidy

fmt: init ## Format code using go fmt
	@echo ">> formatting code"
	$(GO) fmt ./...

vet: init ## Vet code using go vet
	@echo ">> vetting code"
	$(GO) vet ./...

build: fmt vet test ## Build binaries
	@echo ">> building binaries"
	$(GO) build -o ./bin/emq_exporter emq_exporter.go

test: fmt vet ## Run tests using go test
	@echo ">> running tests"
	$(GO) test ./... -coverprofile cover.out

docker: build ## Build docker image
	@echo ">> building docker image"
	@docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

run-local: ## Start a local emq container for development
	@docker run --rm -d --name emqx -h emqx -p 18083:18083 -p 8080:8080 emqx/emqx:latest

run: build run-local ## Run the exporter locally using a local container
	./bin/emq_exporter --emq.uri="http://127.0.0.1:18083" --emq.node="emqx@$(IP)" --emq.api-version="v3" --log.level="debug"

help: ## Print this message and exit
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'

.PHONY: all fmt vet build docker bootstrap help