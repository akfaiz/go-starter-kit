SHELL := /bin/bash
APP := go-starter-kit
BIN := bin/$(APP)
DOCKER_IMAGE := $(APP)

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

GINKGO := ginkgo

# NOTE: SKIP_PACKAGES and TEST_EXCLUDE must be kept in sync manually.
# SKIP_PACKAGES is used by ginkgo --skip-package (comma-separated)
# TEST_EXCLUDE is used by grep -Ev (regex) to build UNIT_TEST_PKGS / COVER_PKGS
SKIP_PACKAGES := test/mocks,db/migrations,cmd,test/e2e
TEST_EXCLUDE := /test/mocks$$|/db/migrations$$|/cmd($$|/)|/test/e2e$$

UNIT_TEST_PKGS = $(shell go list ./... | grep -Ev '$(TEST_EXCLUDE)')
COVER_PKGS = $(shell go list ./... | grep -Ev '$(TEST_EXCLUDE)' | tr '\n' ',' | sed 's/,$$//')

COVDATA_DIR := .covdata
COVDATA_ABS_DIR := $(CURDIR)/$(COVDATA_DIR)

# ─── Dev ──────────────────────────────────────────────────────────────────────

.PHONY: tidy
tidy:
	go mod tidy

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

# ─── Build ────────────────────────────────────────────────────────────────────

.PHONY: build
build:
	go build -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)" -o $(BIN) .

.PHONY: clean
clean:
	rm -rf $(BIN) $(COVDATA_DIR) coverage.out coverage.html

# ─── Test ─────────────────────────────────────────────────────────────────────

.PHONY: test
test:
	$(GINKGO) run -r --skip-package="$(SKIP_PACKAGES)" ./...

.PHONY: test-e2e
test-e2e:
	$(GINKGO) run -r ./test/e2e

.PHONY: coverage
coverage:
	$(GINKGO) run -r --skip-package="$(SKIP_PACKAGES)" --cover --coverpkg="$(COVER_PKGS)" --coverprofile=coverage.out --output-dir=. ./...
	go tool cover -func=coverage.out

.PHONY: coverage-all
coverage-all:
	rm -rf "$(COVDATA_ABS_DIR)"
	mkdir -p "$(COVDATA_ABS_DIR)/unit" "$(COVDATA_ABS_DIR)/e2e" "$(COVDATA_ABS_DIR)/merged"
	go test $(UNIT_TEST_PKGS) \
		-cover \
		-coverpkg="$(COVER_PKGS)" \
		-run . \
		-args -test.gocoverdir="$(COVDATA_ABS_DIR)/unit"
	go test ./test/e2e \
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
	@command -v xdg-open &>/dev/null && xdg-open coverage.html || \
	 command -v open &>/dev/null && open coverage.html || \
	 echo "Coverage report written to coverage.html"

# ─── Database ─────────────────────────────────────────────────────────────────

.PHONY: migrate-up
migrate-up:
	go run . migrate up

.PHONY: migrate-down
migrate-down:
	go run . migrate down

# ─── Docker ───────────────────────────────────────────────────────────────────

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run:
	docker run --rm -p 8080:8080 --env-file .env $(DOCKER_IMAGE) serve