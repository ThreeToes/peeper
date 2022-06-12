.PHONY: build test

build: vendor
	go build -o bin/peeper ./cmd/main.go

vendor:
	go mod vendor

build-docker:
	docker build -t peeper:latest .