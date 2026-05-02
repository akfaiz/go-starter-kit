SHELL := /bin/bash
APP := go-starter-kit
BIN := bin/$(APP)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: run
run:
	go run . serve

.PHONY: migrate-up
migrate-up:
	go run . migrate up

.PHONY: migrate-down
migrate-down:
	go run . migrate down

.PHONY: build
build:
	go build -o $(BIN) .

.PHONY: docker-build
docker-build:
	docker build -t go-api-starter-kit .

.PHONY: docker-run
docker-run:
	docker run --rm -p 8080:8080 --env-file .env go-api-starter-kit serve
