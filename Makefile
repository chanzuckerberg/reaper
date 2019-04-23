SHA=$(shell git rev-parse --short HEAD)
VERSION=$(shell cat VERSION)
DIRTY=$(shell if `git diff-index --quiet HEAD --`; then echo false; else echo true;  fi)
# TODO add release flag
LDFLAGS=-ldflags "-w -s -X github.com/chanzuckerberg/reaper/util.GitSha=${SHA} -X github.com/chanzuckerberg/reaper/util.Version=${VERSION} -X github.com/chanzuckerberg/reaper/util.Dirty=${DIRTY}"

all: test install
.PHONY:all

setup: ## setup development dependencies
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	go get github.com/rakyll/gotest
	go install github.com/rakyll/gotest
	curl -L https://raw.githubusercontent.com/chanzuckerberg/bff/master/download.sh | sh
.PHONY: setup

dep: ## ensure dependencies are up to date
	dep ensure
.PHONEY: dep

lint: dep ## run the fast go linters
	gometalinter --vendor --fast ./...
.PHONY: lint

lint-slow: dep ## run all linters, even the slow ones
	gometalinter --vendor --deadline 120s ./...
.PHONY:lint-slow

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

build: dep ## build the binary
	go build ${LDFLAGS} .
.PHONY: build

coverage: ## run the go coverage tool, reading file coverage.out
	go tool cover -html=coverage.out
.PHONY: coverage

test: dep ## run the tests
	gotest -race -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test

install: dep ## install the reaper binary in $GOPATH/bin
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

docker: dep ## build docker image
	docker build .
.PHONY: docker
