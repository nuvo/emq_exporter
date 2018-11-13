pkgs = $(shell go list ./... | grep -v /vendor/)
IMAGE_NAME := emq_exporter
IMAGE_TAG ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))

all: test build

fmt:
	@echo ">> formatting code"
	go fmt $(pkgs)

vet:
	@echo ">> vetting code"
	go vet $(pkgs)

build: fmt vet
	@echo ">> building binaries"
	go build -o ./bin/emq_exporter $(pkgs)

test: fmt vet
	@echo ">> running tests"
	go test $(pkgs)

docker: build
	@echo ">> building docker image"
	@docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

.PHONY: all fmt vet build docker