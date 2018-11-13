HAS_DEP := $(shell command -v dep;)
DEP_VERSION := v0.5.0
pkgs = $(shell go list ./... | grep -v /vendor/)
IMAGE_NAME := emq_exporter
IMAGE_TAG ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))

all: bootstrap test build

bootstrap:
ifndef HAS_DEP
	wget -q -O $(GOPATH)/bin/dep https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-linux-amd64
	chmod +x $(GOPATH)/bin/dep
endif
	dep ensure

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
	go test $(pkgs) -coverprofile cover.out

docker: build
	@echo ">> building docker image"
	@docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

.PHONY: all fmt vet build docker bootstrap