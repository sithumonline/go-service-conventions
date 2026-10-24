MOQ_VERSION := v0.8.1
OAPI_CODEGEN := go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1
UNIT_ENV := DB_HOST=localhost DB_USER=test DB_PASSWORD=test DB_NAME=test

.PHONY: spec mocks codegen build test-unit test-integration lint fmt check

spec: ## Regenerate spec/ from docs/openapi.yaml
	$(OAPI_CODEGEN) -config spec/oapi-codegen.yaml docs/openapi.yaml

mocks: ## Regenerate moq mocks with the pinned moq version
	@moq -version 2>/dev/null | grep -q "$(MOQ_VERSION)" || echo "warning: pin moq with: go install github.com/matryer/moq@$(MOQ_VERSION)"
	go generate ./...

codegen: spec mocks

build:
	go build ./...

test-unit:
	$(UNIT_ENV) go test ./...

test-integration: ## Needs TEST_DATABASE_DSN pointing at a disposable Postgres
	go test -tags=integration ./...

lint: ## The same command CI runs
	golangci-lint run --config .golangci.yaml ./...

fmt:
	golangci-lint fmt --config .golangci.yaml ./...

check: build test-unit lint
	go build -tags=integration ./...
