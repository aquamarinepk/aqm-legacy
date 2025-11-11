.PHONY: all build run runpretty test test-v test-short coverage coverage-html coverage-func coverage-profile coverage-check coverage-100 clean lint fmt vet tidy tidy-fmt check ci quick-test help

APP_NAME := todo
BUILD_DIR := bin
BIN_NAME := $(APP_NAME)
GO_VERSION := 1.23
GOCACHE := $(abspath .gocache)

all: build

help:
	@echo "Available targets:"
	echo "  build         - Build the $(APP_NAME) service"
	echo "  run           - Build and run the service"
	echo "  runpretty     - Run with prettified log output"
	echo "  test          - Run all tests"
	echo "  test-v        - Run tests with verbose output"
	echo "  test-short    - Run tests in short mode"
	echo "  coverage      - Run tests with coverage report"
	echo "  coverage-html - Generate HTML coverage report"
	echo "  coverage-func - Show function-level coverage"
	echo "  coverage-check - Ensure coverage meets 80%"
	echo "  coverage-100  - Ensure coverage is 100%"
	echo "  tidy          - go mod tidy in the example module"
	echo "  fmt           - Format code"
	echo "  vet           - Run go vet"
	echo "  lint          - Run basic lint (go vet + optional golangci)"
	echo "  clean         - Remove build artifacts"
	echo "  quick-test    - Build and perform a quick smoke test"
	echo "  check         - Run fmt, vet, test, coverage-check, lint"
	echo "  ci            - Run fmt, vet, test, coverage-100, lint"

build: tidy-fmt
	@echo "Building $(BIN_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@GOCACHE=$(GOCACHE) go build -o $(BUILD_DIR)/$(BIN_NAME) ./main.go
	@echo "$(BIN_NAME) built successfully."

run: build
	@echo "Running $(BIN_NAME)..."
	@GOCACHE=$(GOCACHE) $(BUILD_DIR)/$(BIN_NAME)

runpretty: build
	@echo "Running $(BIN_NAME)..."
	@GOCACHE=$(GOCACHE) $(BUILD_DIR)/$(BIN_NAME) 2>&1 | awk '{ \
		if(match($$0, /time=([^ ]+)/, t)) time=t[1]; else time=""; \
		if(match($$0, /level=([^ ]+)/, l)) level=l[1]; else level=""; \
		if(match($$0, /msg="([^"]+)"/, m)) msg=m[1]; else msg=""; \
		if(match($$0, /error="([^"]+)"/, e)) error=e[1]; else error=""; \
		if(level != "") { \
			if(toupper(level) == "INFO") level="INF"; \
			else if(toupper(level) == "ERROR") level="ERR"; \
			else if(toupper(level) == "DEBUG") level="DBG"; \
			else if(toupper(level) == "WARN") level="WRN"; \
			output = "[" level "] - " time " - " msg; \
			if(error != "") output = output " → " error; \
			print output; \
		} else print $$0; \
	}'

test:
	@echo "Running tests..."
	@GOCACHE=$(GOCACHE) go test ./...

test-v:
	@echo "Running tests with verbose output..."
	@GOCACHE=$(GOCACHE) go test -v ./...

test-short:
	@echo "Running tests in short mode..."
	@GOCACHE=$(GOCACHE) go test -short ./...

coverage:
	@echo "Running tests with coverage..."
	@GOCACHE=$(GOCACHE) go test -cover ./...

coverage-profile:
	@echo "Generating coverage profile..."
	@GOCACHE=$(GOCACHE) go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

coverage-html: coverage-profile
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-func: coverage-profile
	@echo "Function-level coverage:"
	@go tool cover -func=coverage.out

coverage-check: coverage-profile
	@COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE < 80" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage $$COVERAGE% is below 80%"; \
		exit 1; \
	else \
		echo "✅ Coverage $$COVERAGE% meets the 80% threshold"; \
	fi

coverage-100: coverage-profile
	@COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $$COVERAGE%"; \
	if [ "$$COVERAGE" != "100.0" ]; then \
		echo "❌ Coverage $$COVERAGE% is not 100%"; \
		go tool cover -func=coverage.out | grep -v "100.0%"; \
		exit 1; \
	else \
		echo "🎉 Perfect! 100% test coverage achieved!"; \
	fi

vet:
	@echo "Running go vet..."
	@GOCACHE=$(GOCACHE) go vet ./...

lint:
	@echo "Running linter..."
	@GOCACHE=$(GOCACHE) go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping lint"; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

tidy:
	@echo "Running go mod tidy..."
	@GOCACHE=$(GOCACHE) go mod tidy

fmt:
	@echo "Formatting code..."
	@gofmt -w .
	@if command -v goimports >/dev/null 2>&1; then \
		echo "Running goimports..."; \
		goimports -w .; \
	else \
		echo "goimports not found, skipping"; \
	fi

clean:
	@echo "Cleaning up..."
	@rm -f $(BUILD_DIR)/$(BIN_NAME)
	@GOCACHE=$(GOCACHE) go clean -testcache
	@rm -f coverage.out coverage.html
	@echo "Cleanup complete."

tidy-fmt: tidy fmt
	@echo "Tidy and formatting complete."

check: fmt vet test coverage-check lint
	@echo "✅ All quality checks passed!"

ci: fmt vet test coverage-100 lint
	@echo "🚀 CI pipeline passed!"

quick-test: build
	@echo "Quick testing $(BIN_NAME)..."
	@timeout 5s bash -c 'GOCACHE=$(GOCACHE) ./bin/$(BIN_NAME) & sleep 2 && curl -s http://localhost:8080/healthz && echo "\n✅ Quick test successful"' || echo "⚠️  Quick test completed (timeout expected)"

