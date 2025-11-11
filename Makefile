# Root Makefile to coordinate aqm orchestration workflows.

ROOT_DIR := $(abspath .)
CACHE_ENV := GOCACHE=$(ROOT_DIR)/.gocache GOMODCACHE=$(ROOT_DIR)/.gomodcache

PROJECT_NAME ?= aqm-orchestration
ORCH_DIR ?= examples/orchestration
SERVICES ?= tasks # accounts activity (pending)
#PKG_LIBS ?= core telemetry (future shared libs)

LOG_DIR ?= logs
LOG_DIR_ABS := $(ROOT_DIR)/$(LOG_DIR)
TAIL_LINES ?= 200

COMPOSE_FILE ?= $(ORCH_DIR)/deploy/local/docker-compose.yml
COMPOSE_ENV ?= $(ORCH_DIR)/deploy/local/env.sample
COMPOSE_LOG_FILTER ?= aqm-orchestration-mongodb
COMPOSE_MONGO_USER ?= admin
COMPOSE_MONGO_PASS ?= password

NOMAD_ADDR ?= http://127.0.0.1:4646
NOMAD_JOBS_DIR ?= tmp/deployments/nomad/jobs
NOMAD_JOBS ?= $(NOMAD_JOBS_DIR)/mongodb.nomad $(NOMAD_JOBS_DIR)/pulap-services.nomad # placeholders from reference repo

MONGO_URL ?= mongodb://admin:password@localhost:27017/tasks?authSource=admin
TASKS_DB ?= tasks

GOFMT ?= gofmt
GOLANGCI_LINT ?= golangci-lint
GO_TEST ?= go test

.PHONY: help build build-services dev dev-% run-all stop-all stop-tasks test lint fmt \\
	log-stream log-clean logs log-clear \\
	compose-log-stream compose-log-clean compose-logs \\
	run-compose run-compose-neat stop-compose reset-compose-data \\
	nomad-run nomad-stop nomad-status

help:
	@echo "$(PROJECT_NAME) orchestration helper"
	@echo "Targets:"
	@echo "  build               - Build orchestration services via sub-make"
	@echo "  dev                 - Run servicios de ejemplo dentro de $(ORCH_DIR)"
	@echo "  run-all             - Arranca binarios locales en background (logs en $(LOG_DIR))"
	@echo "  stop-all            - Detiene lo iniciado por run-all"
	@echo "  dev-<service>       - Run a specific service (e.g., make dev-tasks)"
	@echo "  test                - Run go test ./..."
	@echo "  lint                - Run golangci-lint if available (placeholder)"
	@echo "  fmt                 - gofmt relevant modules"
	@echo "  run-compose         - docker compose up using $(COMPOSE_FILE)"
	@echo "  run-compose-neat    - compose up while filtering $(COMPOSE_LOG_FILTER)"
	@echo "  stop-compose        - docker compose down"
	@echo "  reset-compose-data  - Drop Mongo DBs inside the compose stack"
	@echo "  log-stream/log-clean - Tail de logs locales"
	@echo "  compose-log-*       - Tail de logs de docker compose"
	@echo "  nomad-run|stop|status - Placeholder pass-through to tmp jobs"

build: build-services

build-services:
	@$(CACHE_ENV) $(MAKE) -C $(ORCH_DIR) build

dev:
	@$(MAKE) -C $(ORCH_DIR) dev

run-all: build-services
	@mkdir -p $(LOG_DIR_ABS)
	@$(MAKE) stop-tasks >/dev/null
	@echo "🚀 Starting Tasks service (logs -> $(LOG_DIR)/tasks.log)"
	@cd $(ORCH_DIR)/services/tasks && nohup ../../bin/tasks > $(LOG_DIR_ABS)/tasks.log 2>&1 & echo $$! > $(LOG_DIR_ABS)/tasks.pid
	@echo "✅ Tasks is running (PID: $$(cat $(LOG_DIR_ABS)/tasks.pid))"

stop-all: stop-tasks

stop-tasks:
	@if [ -f $(LOG_DIR_ABS)/tasks.pid ]; then \
		PID=$$(cat $(LOG_DIR_ABS)/tasks.pid); \
		if kill $$PID >/dev/null 2>&1; then \
			echo "🛑 Stopped Tasks (PID $$PID)"; \
		else \
			echo "ℹ️  Tasks already stopped"; \
		fi; \
		rm -f $(LOG_DIR_ABS)/tasks.pid; \
	else \
		echo "ℹ️  No Tasks PID file found"; \
	fi

# Allow make dev-tasks to proxy into the orchestration Makefile.
dev-%:
	@$(MAKE) -C $(ORCH_DIR) dev-$*

test:
	@$(GO_TEST) ./...
	@$(MAKE) -C $(ORCH_DIR) test

lint:
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run ./...; \
	else \
		echo "golangci-lint not installed - skipping"; \
	fi

fmt:
	@find . -name '*.go' \
		-not -path './.gomodcache/*' \
		-not -path './.gocache/*' \
		-exec $(GOFMT) -w {} +

run-compose:
	@if [ ! -f "$(COMPOSE_FILE)" ]; then \
		echo "❌ docker compose file '$(COMPOSE_FILE)' not found"; \
		exit 1; \
	fi
	@if [ ! -f "$(COMPOSE_ENV)" ]; then \
		echo "❌ env file '$(COMPOSE_ENV)' not found"; \
		exit 1; \
	fi
	@echo "Starting docker compose using $(COMPOSE_FILE)..."
	@docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) up --build

run-compose-neat:
	@if [ ! -f "$(COMPOSE_FILE)" ]; then \
		echo "❌ docker compose file '$(COMPOSE_FILE)' not found"; \
		exit 1; \
	fi
	@if [ ! -f "$(COMPOSE_ENV)" ]; then \
		echo "❌ env file '$(COMPOSE_ENV)' not found"; \
		exit 1; \
	fi
	@echo "Starting docker compose using $(COMPOSE_FILE) (filter: $(COMPOSE_LOG_FILTER))..."
	@if [ -z "$(COMPOSE_LOG_FILTER)" ]; then \
		docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) up --build; \
	else \
		docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) up --build 2>&1 | grep -v '^$(COMPOSE_LOG_FILTER)'; \
	fi

stop-compose:
	@if [ ! -f "$(COMPOSE_FILE)" ]; then \
		echo "❌ docker compose file '$(COMPOSE_FILE)' not found"; \
		exit 1; \
	fi
	@docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) down

reset-compose-data:
	@if [ ! -f "$(COMPOSE_FILE)" ]; then \
		echo "❌ docker compose file '$(COMPOSE_FILE)' not found"; \
		exit 1; \
	fi
	@if ! docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) ps --status running mongodb >/dev/null 2>&1; then \
		echo "❌ compose MongoDB service is not running. Start it first (make run-compose)."; \
		exit 1; \
	fi
	@echo "🧹 Clearing MongoDB database $(TASKS_DB) inside compose..."
	@docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) exec mongodb mongosh --quiet --username $(COMPOSE_MONGO_USER) --password $(COMPOSE_MONGO_PASS) --authenticationDatabase admin --eval 'db.getSiblingDB("$(TASKS_DB)").dropDatabase()'
	@echo "✅ Compose MongoDB database cleared."

log-stream:
	@if [ ! -f $(LOG_DIR)/tasks.log ]; then \
		echo "ℹ️  $(LOG_DIR)/tasks.log not found. Run 'make run-all' first."; \
		exit 1; \
	fi
	@echo "📜 Tailing $(LOG_DIR)/tasks.log (Ctrl+C to exit)"
	@tail -n $(TAIL_LINES) -F $(LOG_DIR)/tasks.log

log-clean:
	@if [ ! -f $(LOG_DIR)/tasks.log ]; then \
		echo "ℹ️  $(LOG_DIR)/tasks.log not found. Run 'make run-all' first."; \
		exit 1; \
	fi
	@echo "📜 Tailing condensed logs (hora | mensaje)"
	@tail -n $(TAIL_LINES) -F $(LOG_DIR)/tasks.log | \
	awk '{ printf("[%s] %s\n", strftime("%H:%M:%S"), $$0); }'

log-clear:
	@rm -f $(LOG_DIR)/*.log >/dev/null 2>&1 || true
	@rm -f $(LOG_DIR)/*.pid >/dev/null 2>&1 || true
	@echo "🧹 Logs/PIDs eliminados de $(LOG_DIR)"

logs: log-stream

compose-log-stream:
	@if [ ! -f "$(COMPOSE_FILE)" ]; then \
		echo "❌ docker compose file '$(COMPOSE_FILE)' not found"; \
		exit 1; \
	fi
	@if [ ! -f "$(COMPOSE_ENV)" ]; then \
		echo "❌ env file '$(COMPOSE_ENV)' not found"; \
		exit 1; \
	fi
	@echo "📜 Streaming docker compose logs (Ctrl+C to stop)..."
	@docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) logs -f

compose-log-clean:
	@if [ ! -f "$(COMPOSE_FILE)" ]; then \
		echo "❌ docker compose file '$(COMPOSE_FILE)' not found"; \
		exit 1; \
	fi
	@if [ ! -f "$(COMPOSE_ENV)" ]; then \
		echo "❌ env file '$(COMPOSE_ENV)' not found"; \
		exit 1; \
	fi
	@echo "📜 Streaming condensed docker compose logs"
	@docker compose --env-file $(COMPOSE_ENV) -f $(COMPOSE_FILE) logs -f | \
	awk -v filter="$(COMPOSE_LOG_FILTER)" '{ \
		svc=$$1; gsub(/ +/,"",svc); \
		if (filter != "" && svc == filter) next; \
		sep=index($$0,"|"); msg=$$0; if (sep>0) {msg=substr($$0,sep+1)}; gsub(/^ +/ ,"", msg); \
		printf("[%s] %-18s %s\n", strftime("%H:%M:%S"), svc, msg); \
	}'

compose-logs: compose-log-stream

nomad-run:
	@echo "⚠️  Nomad jobs not yet adapted. Refer to $(NOMAD_JOBS_DIR) for reference manifests."

nomad-stop:
	@echo "⚠️  Nomad stop placeholder"

nomad-status:
	@echo "⚠️  Nomad status placeholder"
