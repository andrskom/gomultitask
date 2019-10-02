SRV = $(notdir $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))))
PROJECT = github.com/andrskom/${SRV}

all: vendor lint test
.PHONY: all

vendor:
	@echo "+ $@"
	@GO111MODULE=on go mod vendor
.PHONY: vendor

test:
	@echo "+ $@"
	@go test -count=5 -cover ./...
.PHONY: test

lint:
	@echo "+ $@"
	@docker run --rm -i  \
		-v ${PWD}:/go/src/${PROJECT} \
		-w /go/src/${PROJECT} golangci/golangci-lint:v1.18 golangci-lint run --enable-all --skip-dirs vendor,version,pkg/gen ./...
.PHONY: lint
