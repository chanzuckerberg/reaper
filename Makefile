SHA=$(shell git rev-parse --short HEAD)
VERSION=$(shell cat VERSION)
DIRTY=$(shell if `git diff-index --quiet HEAD --`; then echo false; else echo true; fi)
# TODO add release flag
LDFLAGS=-ldflags "-w -s -X github.com/chanzuckerberg/reaper/cmd.GitSha=${SHA} -X github.com/chanzuckerberg/reaper/cmd.Version=${VERSION} -X github.com/chanzuckerberg/reaper/cmd.Dirty=${DIRTY}"
export GOFLAGS=-mod=vendor
export GO111MODULE=on

all: test install
.PHONY:all

setup: ## setup development endencies
	curl -L https://raw.githubusercontent.com/chanzuckerberg/bff/master/download.sh | sh
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	curl -sfL https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh| sh
.PHONY: setup

lint: ## run the fast go linters
	gometalinter --vendor --fast ./...
.PHONY: lint

lint-slow: ## run all linters, even the slow ones
	gometalinter --vendor --deadline 120s ./...
.PHONY:lint-slow

lint-ci: ## run the fast go linters
	./bin/reviewdog -conf .reviewdog.yml  -reporter=github-pr-review
.PHONY: lint-ci

release: ## run a release
	./bin/bff bump
	git push
	goreleaser release
.PHONY: release

release-prerelease: build ## release to github as a 'pre-release'
	version=`./reaper version`; \
	git tag v"$$version"; \
	git push
	git push --tags
	goreleaser release -f .goreleaser.prerelease.yml --debug
.PHONY: release-prerelease

release-snapshot: ## run a release
	goreleaser release --snapshot
.PHONY: release-snapshot

build: deps ## build the binary
	go build ${LDFLAGS} .
.PHONY: build

coverage: ## run the go coverage tool, reading file coverage.out
	go tool cover -html=coverage.out
.PHONY: coverage

test: deps ## run the tests
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test

test-ci: ## run the tests
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test

deps:
	go mod tidy
	go mod vendor
.PHONY: deps

install: deps ## install the reaper binary in $GOPATH/bin
	go install ${LDFLAGS} .
.PHONY: install

help: ## display help for this makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

clean: ## clean the repo
	rm reaper 2>/dev/null || true
	go clean
	rm -rf dist
.PHONY: clean

docker: ## build docker image
	docker build .
.PHONY: docker
