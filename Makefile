SHELL := /bin/bash

.PHONY: all check format vet lint build test generate tidy integration_test

-include Makefile.env

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  check               to do static check"
	@echo "  build               to create bin directory and build"
	@echo "  test                to run test"

check: vet

format:
	go fmt ./...

vet:
	go vet ./...

build: tidy format check
	go build -o bin/community ./cmd/community

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -v ./...
	go tool cover -html="coverage.txt" -o "coverage.html"

tidy:
	go mod tidy
	go mod verify
