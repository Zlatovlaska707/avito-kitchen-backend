.PHONY: up up-build down clean logs run-local migrate-up migrate-down migrate-status migrate-redo seed generate openapi-lint test-unit test-integration test-e2e test-all lint fmt tidy loadtest loadtest-spike loadtest-spike-default loadtest-catalog loadtest-order up-build-loadtest plantuml smoke cover-html cover-check ci-local

COMPOSE := docker compose
GO := go
GOLANGCI := golangci/golangci-lint:v2.5.0-alpine
GOFUMPT := mvdan.cc/gofumpt

up:
	$(COMPOSE) up -d

up-build:
	$(COMPOSE) up -d --build --wait

down:
	$(COMPOSE) down

clean:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f

run-local:
	$(GO) run ./cmd/api

migrate-up:
	$(GO) run ./cmd/migrator up

migrate-down:
	$(GO) run ./cmd/migrator down

migrate-status:
	$(GO) run ./cmd/migrator status

migrate-redo:
	$(GO) run ./cmd/migrator redo

seed:
	$(GO) run ./cmd/migrator seed

generate:
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 --config=api/oapi-codegen.server.yaml api/openapi.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 --config=api/oapi-codegen.client.yaml api/openapi.yaml
	$(GO) generate ./...

openapi-lint:
	docker run --rm -v "$(CURDIR)/api:/api" stoplight/spectral lint --ruleset /api/.spectral.yaml /api/openapi.yaml

test-unit:
	$(GO) test -race -cover ./internal/...
	cd venue-simulator && $(GO) test -race -cover ./...

test-integration:
	$(GO) test -race -tags=integration ./test/integration/... ./test/contract/...

test-e2e: up-build-e2e
	$(GO) test -tags=e2e ./test/api/...

up-build-e2e:
	$(COMPOSE) -f docker-compose.yml -f docker-compose.e2e.yml up -d --build --wait

test-all: test-unit test-integration test-e2e

lint:
	docker run --rm -v "$(CURDIR):/app" -w /app $(GOLANGCI) golangci-lint run ./...

fmt:
	$(GO) run $(GOFUMPT) -w .

tidy:
	$(GO) mod tidy
	cd venue-simulator && $(GO) mod tidy

LOADTEST_RESULTS := docs/loadtest

loadtest: up-build-loadtest
	@mkdir -p $(LOADTEST_RESULTS)
	$(COMPOSE) -f docker-compose.yml -f docker-compose.loadtest.yml --profile loadtest run --rm \
		-v "$(CURDIR)/$(LOADTEST_RESULTS):/results" \
		k6 run --summary-export=/results/mixed-summary.json /scripts/mixed.js \
		> $(LOADTEST_RESULTS)/mixed-console.log 2>&1; \
	cat $(LOADTEST_RESULTS)/mixed-console.log

loadtest-spike: up-build-loadtest
	@mkdir -p $(LOADTEST_RESULTS)
	$(COMPOSE) -f docker-compose.yml -f docker-compose.loadtest.yml --profile loadtest run --rm \
		-v "$(CURDIR)/$(LOADTEST_RESULTS):/results" \
		k6 run --summary-export=/results/spike-summary.json /scripts/spike.js \
		> $(LOADTEST_RESULTS)/spike-console.log 2>&1; \
	cat $(LOADTEST_RESULTS)/spike-console.log || true

loadtest-spike-default:
	docker compose up -d --wait
	@mkdir -p $(LOADTEST_RESULTS)
	$(COMPOSE) --profile loadtest run --rm \
		-v "$(CURDIR)/$(LOADTEST_RESULTS):/results" \
		k6 run --summary-export=/results/spike-summary.json /scripts/spike.js \
		> $(LOADTEST_RESULTS)/spike-console.log 2>&1; \
	cat $(LOADTEST_RESULTS)/spike-console.log || true

loadtest-catalog: up-build-loadtest
	$(COMPOSE) -f docker-compose.yml -f docker-compose.loadtest.yml --profile loadtest run --rm k6 run /scripts/catalog.js

loadtest-order: up-build-loadtest
	$(COMPOSE) -f docker-compose.yml -f docker-compose.loadtest.yml --profile loadtest run --rm k6 run /scripts/order.js

up-build-loadtest:
	$(COMPOSE) -f docker-compose.yml -f docker-compose.loadtest.yml up -d --build --wait

plantuml:
	docker run --rm -v "$(CURDIR)/docs:/docs" plantuml/plantuml:latest -tpng /docs/**/*.puml

smoke:
	@echo "Run smoke checks against http://localhost:8080 (see README)"

cover-html:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

cover-check:
	@packages=$$(go list ./internal/domain/... ./internal/usecase/... | grep -v '/uctest$$'); \
	$(GO) test -coverprofile=domain_usecase.out -covermode=atomic $$packages; \
	bash scripts/check_coverage.sh domain_usecase.out

ci-local: tidy fmt lint test-unit test-integration
