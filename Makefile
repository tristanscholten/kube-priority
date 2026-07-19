SHELL := /usr/bin/env bash
IMG ?= ghcr.io/tristanscholten/kube-priority:latest
CONTAINER_TOOL ?= podman
KUSTOMIZE ?= kubectl kustomize
KUBECTL ?= kubectl

.PHONY: all fmt vet test test-unit test-envtest test-e2e build docker-build install uninstall deploy undeploy manifests lint govulncheck

all: fmt vet test

fmt:
	go fmt ./...

vet:
	go vet ./...

test: test-unit

test-unit:
	go test ./... -count=1

test-envtest:
	go test ./internal/controller ./internal/admission -count=1

test-e2e:
	go test ./test/e2e -count=1 -timeout=20m

build:
	go build -o bin/manager ./cmd/manager

docker-build:
	$(CONTAINER_TOOL) build -t $(IMG) .

install:
	$(KUBECTL) apply -k config/default

uninstall:
	$(KUBECTL) delete -k config/default --ignore-not-found

deploy: install

undeploy: uninstall

manifests:
	$(KUSTOMIZE) config/default

lint:
	golangci-lint run --timeout=5m

govulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
