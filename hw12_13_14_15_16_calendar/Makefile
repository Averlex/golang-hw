BIN := "./bin/calendar"
TOOLS_DIR := $(PWD)/tools/bin

DOCKER_IMG="calendar:develop"

JQ_MARKER_FILE := .jq_installed

HOST := "localhost"
DBPORT := "5432"
DBNAME := "calendar"
CALENDAR_STORAGE_SQL_USER := "calendar_user"
CALENDAR_STORAGE_SQL_PASSWORD := "calendar_pass"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

# --- Build and run ---

build:
	go build -tags=viper_bind_struct -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	CALENDAR_STORAGE_SQL_USER=$(CALENDAR_STORAGE_SQL_USER) \
	CALENDAR_STORAGE_SQL_PASSWORD=$(CALENDAR_STORAGE_SQL_PASSWORD) \
	$(BIN) --config ./configs/calendar/config.toml

run-json: build
	CALENDAR_STORAGE_SQL_USER=$(CALENDAR_STORAGE_SQL_USER) \
	CALENDAR_STORAGE_SQL_PASSWORD=$(CALENDAR_STORAGE_SQL_PASSWORD) \
	$(BIN) --config ./configs/calendar/config.toml | jq -R 'fromjson?' 2>/dev/null

# --- Database ---

# WARNING: docker-compose.yaml must be in sync with this Makefile
run-db:
	docker compose -f ./deployments/docker-compose.yaml up -d

stop-db:
	docker compose -f ./deployments/docker-compose.yaml down

# --- App CLI flags ---

version: build
	$(BIN) --version

help: build
	$(BIN) --help

# --- Testing ---

test:
	go test -v -count=1 -race -timeout=1m ./internal/... ./pkg/...

test-fast:
	go test -v -count=1 -timeout=1m ./internal/... ./pkg/...

test-cover:
	go test -v -count=1 -race -timeout=1m -coverprofile=coverage.out ./internal/... ./pkg/...
	go tool cover -html=coverage.out -o coverage.html

# --- Linting ---

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.57.2

lint: install-lint-deps
	golangci-lint run ./...

# --- Migrations ---

# Apply all pending migrations (default migration target)
migrate: migrate-up

# Apply all pending migrations
migrate-up: run-db
	goose -dir ./migrations postgres "host=$(HOST) port=$(DBPORT) user=$(CALENDAR_STORAGE_SQL_USER) password=$(CALENDAR_STORAGE_SQL_PASSWORD) dbname=$(DBNAME) sslmode=disable" up

# Apply one next migration
migrate-up-1: run-db
	goose -dir ./migrations postgres "host=$(HOST) port=$(DBPORT) user=$(CALENDAR_STORAGE_SQL_USER) password=$(CALENDAR_STORAGE_SQL_PASSWORD) dbname=$(DBNAME) sslmode=disable" up 1

# Rollback one last migration
migrate-down-1: run-db
	goose -dir ./migrations postgres "host=$(HOST) port=$(DBPORT) user=$(CALENDAR_STORAGE_SQL_USER) password=$(CALENDAR_STORAGE_SQL_PASSWORD) dbname=$(DBNAME) sslmode=disable" down 1

# --- Generate ---

generate: setup-mockery
	go generate ./...

# --- Docker ---

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

# --- Setup additional tools: mocks, migration and pretty shell json logs ---

setup-tools: setup-mockery setup-goose setup-jq

setup-mockery:
    GOBIN=$(TOOLS_DIR) go install github.com/vektra/mockery/v2@latest

setup-goose:
    GOBIN=$(TOOLS_DIR) go install github.com/pressly/goose/v3@v3.24.3

setup-jq:
	@command -v jq >/dev/null 2>&1 || { \
        echo "Installing jq..."; \
        sudo apt-get update && sudo apt-get install -y jq; \
        touch $(JQ_MARKER_FILE); \
    }

# --- Cleanup additional tools ---

clean-tools: clean-mockery clean-goose clean-jq

clean-mockery:
	rm -f $(TOOLS_DIR)/mockery

clean-goose:
	rm -f $(TOOLS_DIR)/goose

clean-jq:
	@if [ -f $(JQ_MARKER_FILE) ]; then \
		echo "Removing jq (installed via Makefile)..."; \
		sudo apt-get remove --purge -y jq; \
		rm -f $(JQ_MARKER_FILE); \
		echo "jq successfully removed."; \
	else \
		echo "jq was not installed via Makefile."; \
		echo "To remove it manually: sudo apt-get remove --purge jq"; \
	fi

.PHONY: build run run-json version help test test-fast test-cover lint install-lint-deps migrate-up migrate-up-1 migrate-down-1 generate build-img run-img run-db stop-db clean-tools clean-mockery clean-goose clean-jq setup-tools setup-mockery setup-goose setup-jq
