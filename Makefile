SHELL := /bin/bash
APP := go-starter-kit
BIN := bin/$(APP)

GOCACHE_DIR := $(CURDIR)/.gocache
GINKGO := ginkgo
SKIP_PACKAGES := internal/mocks,db/migrations,cmd,test/e2e
TEST_EXCLUDE := /internal/mocks$$|/db/migrations$$|/cmd($$|/)|/test/e2e$$
UNIT_TEST_PKGS := $(shell GOCACHE="$(GOCACHE_DIR)" go list ./... | grep -Ev '$(TEST_EXCLUDE)')
COVER_PKGS := $(shell GOCACHE="$(GOCACHE_DIR)" go list ./... | grep -Ev '$(TEST_EXCLUDE)' | tr '\n' ',' | sed 's/,$$//')
COVDATA_DIR := .covdata
COVDATA_ABS_DIR := $(CURDIR)/$(COVDATA_DIR)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	GOCACHE="$(GOCACHE_DIR)" $(GINKGO) run -r --skip-package="$(SKIP_PACKAGES)" ./...

.PHONY: test-e2e
test-e2e:
	GOCACHE="$(GOCACHE_DIR)" $(GINKGO) run -r ./test/e2e

.PHONY: coverage
coverage:
	GOCACHE="$(GOCACHE_DIR)" $(GINKGO) run -r --skip-package="$(SKIP_PACKAGES)" --cover --coverpkg="$(COVER_PKGS)" --coverprofile=coverage.out --output-dir=. ./...
	go tool cover -func=coverage.out

.PHONY: coverage-all
coverage-all:
	rm -rf "$(COVDATA_ABS_DIR)"
	mkdir -p "$(COVDATA_ABS_DIR)/unit" "$(COVDATA_ABS_DIR)/e2e" "$(COVDATA_ABS_DIR)/merged"
	GOCACHE="$(GOCACHE_DIR)" go test $(UNIT_TEST_PKGS) \
		-cover \
		-coverpkg="$(COVER_PKGS)" \
		-run . \
		-args -test.gocoverdir="$(COVDATA_ABS_DIR)/unit"
	GOCACHE="$(GOCACHE_DIR)" go test ./test/e2e \
		-cover \
		-coverpkg="$(COVER_PKGS)" \
		-run . \
		-args -test.gocoverdir="$(COVDATA_ABS_DIR)/e2e"
	go tool covdata merge -i="$(COVDATA_ABS_DIR)/unit,$(COVDATA_ABS_DIR)/e2e" -o="$(COVDATA_ABS_DIR)/merged"
	go tool covdata textfmt -i="$(COVDATA_ABS_DIR)/merged" -o=coverage.out
	go tool cover -func=coverage.out

.PHONY: coverage-html
coverage-html:
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run --timeout 5m

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix --timeout 5m

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
