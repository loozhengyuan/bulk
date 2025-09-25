-include .env
export

REPOSITORY_OWNER := loozhengyuan
REPOSITORY_NAME := bulk

PROJECT_NAME := $(REPOSITORY_OWNER)/$(REPOSITORY_NAME)
MODULE_NAME := github.com/$(PROJECT_NAME)

BUILD_TIMESTAMP := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_COMMIT := $(shell git rev-parse --verify HEAD)
BUILD_VERSION := v0.0.0

EXECUTABLE_NAME := bulk

COVERAGE_FILE := coverage.out

.PHONY: install
install:
	go install \
		-ldflags=" \
			-X $(MODULE_NAME)/internal/build.Version=$(BUILD_VERSION) \
			-X $(MODULE_NAME)/internal/build.CommitHash=$(BUILD_COMMIT) \
			-X $(MODULE_NAME)/internal/build.Timestamp=$(BUILD_TIMESTAMP)" \
		-v \
		./cmd/$(EXECUTABLE_NAME)/

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: check
check:
	go mod verify
	go vet ./...

.PHONY: lint
lint:
	test -z $$(gofmt -l .)

.PHONY: test
test:
	go test \
		-race \
		-cover \
		-coverprofile=$(COVERAGE_FILE) \
		-covermode=atomic \
		./...
	go tool cover \
		-func=$(COVERAGE_FILE)

.PHONY: test-all
test-all: install
	go test \
		-race \
		-cover \
		-coverprofile=$(COVERAGE_FILE) \
		-covermode=atomic \
		-tags=unit,integration,e2e \
		./...
	go tool cover \
		-func=$(COVERAGE_FILE)

.PHONY: bench
bench:
	go test \
		-benchmem \
		-bench=. \
		./...

.PHONY: build
build:
	CGO_ENABLED=0 \
	go build \
		-ldflags=" \
			-X $(MODULE_NAME)/internal/build.Version=$(BUILD_VERSION) \
			-X $(MODULE_NAME)/internal/build.CommitHash=$(BUILD_COMMIT) \
			-X $(MODULE_NAME)/internal/build.Timestamp=$(BUILD_TIMESTAMP)" \
		-v \
		-o $(EXECUTABLE_NAME) \
		./cmd/$(EXECUTABLE_NAME)/main.go

.PHONY: image
image:
	docker build \
		--build-arg VERSION=$(BUILD_VERSION) \
		--build-arg REVISION=$(BUILD_COMMIT) \
		--build-arg CREATED=$(BUILD_TIMESTAMP) \
		--tag ${PROJECT_NAME}:latest \
		.

.PHONY: clean
clean: clean-build clean-mod clean-test clean-cov

.PHONY: clean-build
clean-build:
	go clean \
		-cache

.PHONY: clean-mod
clean-mod:
	go clean \
		-modcache

.PHONY: clean-test
clean-test:
	go clean \
		-testcache

.PHONY: clean-cov
clean-cov:
	find . \
		-type f \
		-name $(COVERAGE_FILE) \
		-exec rm -rf {} \;
