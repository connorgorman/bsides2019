TESTFLAGS=-race -p 4
BASE_DIR=$(CURDIR)
GO_VERSION="1.10"

.PHONY: all
all: deps style test image

###########
## Style ##
###########
.PHONY: style
style: fmt imports lint vet

.PHONY: fmt
fmt:
	@echo "+ $@"
	@find . -name vendor -prune -o -name '*.go' -print | xargs gofmt -s -l -w

.PHONY: imports
imports:
	@echo "+ $@"
	@find . -name vendor -prune -o -name '*.go' -print | xargs goimports -w

.PHONY: lint
lint:
	@echo "+ $@"
	@set -e; for pkg in $(shell go list -e ./... | grep -v vendor); do golint -set_exit_status $$pkg; done

.PHONY: vet
vet:
	@echo "+ $@"
	@go vet $(shell go list -e ./... | grep -v vendor)

deps: Gopkg.toml Gopkg.lock
	@echo "+ $@"
# `dep status` exits with a nonzero code if there is a toml->lock mismatch.
	dep status
	dep ensure
	@touch deps

.PHONY: clean-deps
clean-deps:
	@echo "+ $@"
	@rm -f deps

###########
## Build ##
###########
.PHONY: build
build:
	@echo "+ $@"
	@mkdir -p build/bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./container/bin/fsmonitor ./fsmonitor
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./container/bin/capable ./capable

.PHONY: test
test:
	@go test -cover $(TESTFLAGS) -v $(shell go list -e ./... | grep -v vendor) 2>&1 | tee test.log

###########
## Image ##
###########
image: build
	@echo "+ $@"
	@mkdir -p container/bin
	docker build -t connorgorman/bsides2019:latest container/

###########
## Clean ##
###########
.PHONY: clean
clean: clean-image
	@echo "+ $@"

.PHONY: clean-image
clean-image:
	@echo "+ $@"