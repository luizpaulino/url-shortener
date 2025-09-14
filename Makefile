SHELL := /bin/bash

APP_NAME := url-shortener
PORT ?= 8080
URLS_TABLE ?= urls
DYNAMO_ENDPOINT ?= http://localhost:8000

AWS_ENV = AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local AWS_DEFAULT_REGION=us-east-1
ENV_NOPROXY = NO_PROXY=localhost,127.0.0.1,::1 no_proxy=localhost,127.0.0.1,::1

.PHONY: deps
deps:
	@go mod tidy
	@go mod download

.PHONY: run
run:
	@$(ENV_NOPROXY) APP_ENV=local PORT=$(PORT) URLS_TABLE=$(URLS_TABLE) DYNAMO_ENDPOINT=$(DYNAMO_ENDPOINT) \
	OTEL_EXPORTER_OTLP_ENDPOINT= \
	http_proxy= HTTPS_PROXY= \
	go run ./cmd/api

.PHONY: test
test:
	@go test ./... -count=1 -race

.PHONY: test-integration
test-integration:
	@RUN_INTEGRATION=1 DYNAMO_ENDPOINT=$(DYNAMO_ENDPOINT) URLS_TABLE=$(URLS_TABLE) \
	go test ./... -tags=Integration -count=1

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME) ./cmd/api

.PHONY: docker-build
docker-build:
	@docker build -t $(APP_NAME):dev -f Dockerfile .

.PHONY: docker-run
docker-run:
	@docker run --rm -p $(PORT):8080 \
		-e PORT=8080 -e URLS_TABLE=$(URLS_TABLE) -e DYNAMO_ENDPOINT=$(DYNAMO_ENDPOINT) \
		$(APP_NAME):dev

# ---- Dynamo Local ----
.PHONY: dynamo-up
dynamo-up:
	@docker compose up -d dynamodb
	@sleep 2
	@$(AWS_ENV) aws --no-cli-pager dynamodb list-tables --endpoint-url $(DYNAMO_ENDPOINT) || true

.PHONY: dynamo-down
dynamo-down:
	@docker compose down -v

.PHONY: dynamo-init-cli
dynamo-init-cli:
	@$(AWS_ENV) $(ENV_NOPROXY) aws --no-cli-pager dynamodb list-tables --endpoint-url $(DYNAMO_ENDPOINT) || true
	@$(AWS_ENV) $(ENV_NOPROXY) aws --no-cli-pager dynamodb create-table \
	  --endpoint-url $(DYNAMO_ENDPOINT) \
	  --table-name $(URLS_TABLE) \
	  --attribute-definitions \
	    AttributeName=PK,AttributeType=S \
	    AttributeName=SK,AttributeType=S \
	    AttributeName=GSI1PK,AttributeType=S \
	    AttributeName=GSI1SK,AttributeType=S \
	  --key-schema \
	    AttributeName=PK,KeyType=HASH \
	    AttributeName=SK,KeyType=RANGE \
	  --global-secondary-indexes \
	    "IndexName=GSI1,KeySchema=[{AttributeName=GSI1PK,KeyType=HASH},{AttributeName=GSI1SK,KeyType=RANGE}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5}" \
	  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 || true

# Inicializador em Go (se preferir evitar o CLI)
.PHONY: dynamo-init
dynamo-init:
	@echo "Initializing table via Go helper..."
	@$(ENV_NOPROXY) DYNAMO_ENDPOINT=$(DYNAMO_ENDPOINT) URLS_TABLE=$(URLS_TABLE) go run ./cmd/dyninit

.PHONY: dynamo-logs
dynamo-logs:
	@docker logs -f dynamodb-local

.PHONY: dynamo-ps
dynamo-ps:
	@docker compose ps dynamodb
