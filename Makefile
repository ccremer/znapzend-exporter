.PHONY: build fmt dist clean test
SHELL := /usr/bin/env bash

build:
	@go build ./...

fmt:
	@[[ -z $$(go fmt ./...) ]]

dist:
	@goreleaser release --snapshot --rm-dist --skip-sign

clean:
	@rm znapzend-exporter c.out

test:
	@go test -coverprofile c.out ./...
