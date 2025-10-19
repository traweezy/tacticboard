GO ?= go
BINARY := bin/tacticboard

.PHONY: all dev build test lint clean migrate

all: build

dev:
	APP_ENV=development $(GO) run ./cmd/server

build:
	mkdir -p $(dir $(BINARY))
	$(GO) build -trimpath -o $(BINARY) ./cmd/server

test:
	$(GO) test ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed"; \
		exit 1; \
	fi

migrate:
	@echo "No database migrations implemented yet. See migrations/ for starter SQL."

clean:
	rm -rf $(BINARY)
