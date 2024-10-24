CONTAINER := ghcr.io/KillianMeersman/chaperone
TAG ?= latest

.ONESHELL:
SHELL = /usr/bin/bash
.SHELLFLAGS = -ce

tidy:
	go mod tidy

vendor: tidy
	go mod vendor
.PHONY: vendor

vet:
	go vet ./...

test: vendor
	go test -v -race ./...

fuzz: vendor
	go test -fuzztime 30s -fuzz FuzzInvalidJWT ./pkg/auth/
	go test -fuzztime 30s -fuzz FuzzPassword ./pkg/auth/

build: vendor vet test
	mkdir dist || true
	go build -o dist/$(TARGET) ./cmd/$(TARGET)/main.go

container: vendor vet test
	docker build -t $(CONTAINER):$(TAG) .

run: container
	docker run \
		--rm \
		-p "127.0.0.1:8080:8080" \
		$(CONTAINER):$(TAG) proxy

debug:
	go run cmd/$(TARGET)/main.go proxy

clean:
	rm -rf dist/
	rm -rf coverage.*
	go clean ./...