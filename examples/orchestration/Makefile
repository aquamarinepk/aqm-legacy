GO ?= go
COMPOSE_FILE ?= deploy/local/docker-compose.yml
COMPOSE_ENV ?= deploy/local/env.sample

TASKS_MONGO_URI ?= mongodb://admin:password@localhost:27017/tasks?authSource=admin
TASKS_MONGO_DATABASE ?= tasks
TASKS_MONGO_COLLECTION ?= tasks
TASKS_HTTP_PORT ?= :8082
SHARED_LOG_LEVEL ?= debug

export TASKS_MONGO_URI
export TASKS_MONGO_DATABASE
export TASKS_MONGO_COLLECTION
export TASKS_HTTP_PORT
export SHARED_LOG_LEVEL

SERVICES := tasks

.PHONY: help dev dev-tasks build build-tasks test lint compose-up compose-down

help:
	@echo "Available targets:"
	@echo "  dev            - Run all services (currently Tasks)"
	@echo "  dev-tasks      - Run only the Tasks service"
	@echo "  build          - Build every service binary"
	@echo "  build-tasks    - Build the Tasks service binary"
	@echo "  compose-up     - Launch docker compose stack from deploy/local"
	@echo "  compose-down   - Stop docker compose stack"
	@echo "  test           - go test ./..."
	@echo "  lint           - placeholder for future linting"

dev: dev-tasks

dev-tasks:
	cd services/tasks && $(GO) run .

build: $(addprefix build-,$(SERVICES))

build-tasks:
	@mkdir -p bin
	cd services/tasks && $(GO) build -o ../../bin/tasks .

test:
	$(GO) test ./...

lint:
	@echo "lint target pending (add golangci-lint when ready)"


compose-up:
	@if [ ! -f $(COMPOSE_FILE) ]; then \
		echo "docker-compose file not found"; \
		exit 1; \
	fi
	@if [ ! -f $(COMPOSE_ENV) ]; then \
		echo "env file not found (expected $(COMPOSE_ENV))"; \
		exit 1; \
	fi
	docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) up --build

compose-down:
	@if [ ! -f $(COMPOSE_FILE) ]; then \
		echo "docker-compose file not found"; \
		exit 1; \
	fi
	docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) down
