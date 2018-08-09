GO    := go
GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))

pkgs         = $(shell $(GO) list ./... | grep -v /vendor/)


PREFIX                  ?= $(shell pwd)
BIN_DIR                 ?= $(shell pwd)
DOCKER_IMAGE_NAME       ?= jmx-json-exporter
#DOCKER_IMAGE_TAG        ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))
DOCKER_IMAGE_TAG        ?=latest

all: format test build docker

format:
	@echo ">> formatting code"
	@${GO} fmt .

build:
	@echo ">>building binary"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${GO} build .

test:
	@echo ">>testing"
	@${GO} test .

docker:
	@echo ">>building docker image"
	@docker build -t  "${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}" .

.PHONY: all