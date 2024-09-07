# Makefile
.DEFAULT_GOAL := build

FLAGS = -ldflags="-s -w" -trimpath

fmt:
	go fmt
.PHONY: fmt

lint: fmt
	staticcheck
.PHONY: lint

vet: fmt
	go vet
.PHONY: vet

build: vet
	go mod tidy
	go build $(FLAGS)
.PHONY: build
