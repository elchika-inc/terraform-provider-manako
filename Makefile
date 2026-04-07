BINARY=terraform-provider-manako
HOSTNAME=registry.terraform.io
NAMESPACE=elchika-inc
NAME=manako
VERSION=0.1.0
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

default: build

generate:
	go generate ./...

build:
	go build -o $(BINARY)

install: build
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)
	mv $(BINARY) ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)

test:
	go test ./... -v -count=1

testacc:
	TF_ACC=1 go test ./... -v -count=1 -timeout 120m

lint:
	golangci-lint run ./...

.PHONY: default generate build install test testacc lint
