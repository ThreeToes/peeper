.PHONY: build test

build: vendor test
	go build -o bin/peeper ./cmd/main.go

vendor:
	go mod vendor

test: vendor generate
	go test -cover ./...

build-docker:
	docker build -t peeper:latest .

generate:
	go generate ./...