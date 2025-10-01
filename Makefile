.PHONY: test test-quick test-position test-settlement test-balance test-service test-cache test-integration test-all test-coverage check-env build clean

# Load .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

check-env:
	@if [ ! -f .env ]; then \
		echo "Warning: .env not found. Copy .env.example to .env"; \
		exit 1; \
	fi

test-quick:
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi && \
	go test -v ./tests -run TestPositionBehavior -timeout=2m

test-position: check-env
	@set -a && . ./.env && set +a && \
	go test -v ./tests -run TestPositionBehaviorSuite -timeout=5m

test-settlement: check-env
	@set -a && . ./.env && set +a && \
	go test -v ./tests -run TestSettlementBehaviorSuite -timeout=5m

test-balance: check-env
	@set -a && . ./.env && set +a && \
	go test -v ./tests -run TestBalanceBehaviorSuite -timeout=5m

test-service: check-env
	@set -a && . ./.env && set +a && \
	go test -v ./tests -run TestServiceDiscoveryBehaviorSuite -timeout=5m

test-cache: check-env
	@set -a && . ./.env && set +a && \
	go test -v ./tests -run TestCacheBehaviorSuite -timeout=5m

test-integration: check-env
	@set -a && . ./.env && set +a && \
	go test -v ./tests -run TestIntegrationBehaviorSuite -timeout=10m

test-all: check-env
	@set -a && . ./.env && set +a && \
	go test -v ./tests -timeout=15m

test-coverage: check-env
	@set -a && . ./.env && set +a && \
	go test -v ./tests -coverprofile=coverage.out -timeout=15m
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

build:
	go build -v ./...

clean:
	rm -f coverage.out coverage.html
	go clean -testcache
