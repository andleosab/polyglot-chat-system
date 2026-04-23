APP_NAME=chat-message-service
BIN_DIR=bin
BIN=$(BIN_DIR)/server
MAIN=cmd/server/main.go

DOCKER_FILE=./docker/Dockerfile.scratch
DOCKER_IMAGE=$(APP_NAME)
COMPOSE_FILE=docker/docker-compose.yml
COMPOSE=docker compose
SQLC=sqlc
# MIGRATE=migrate
MIGRATE=./scripts/migrate.sh
# MIGRATIONS_DIR=internal/db/migrations

# Load .env if present
-include .env
export

.PHONY: all build run sqlc migrate-up migrate-down docker-build docker-push compose-up compose-down restart clean

## -------------------------------
## Default target
## -------------------------------
all: build

## -------------------------------
## Go build
## -------------------------------
build:
	@echo "==> Building Go binary..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN) $(MAIN)
	@echo "Binary created at $(BIN)"

run:
	@$(BIN)

clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)

## -------------------------------
## SQLC generation (on demand)
## -------------------------------
sqlc:
	@echo "==> Running SQLC generation..."
	@$(SQLC) generate
	@echo "SQLC complete."

## -------------------------------
## Database migrations (on demand)
## -------------------------------
migrate-up:
	@echo "==> Migrating database UP..."
# 	@$(MIGRATE) -path "$(MIGRATIONS_DIR)" -database "$(LOCALHOST_DB_URL)" up
	@$(MIGRATE) up

migrate-down:
	@echo "==> Migrating database DOWN..."
# 	@$(MIGRATE) -path "$(MIGRATIONS_DIR)" -database "$(LOCALHOST_DB_URL)" down
	@$(MIGRATE) down -all

## -------------------------------
## Docker image build & push
## -------------------------------
docker-build:
	@echo "==> Building Docker image using $(DOCKER_FILE)..."
	docker build -t $(DOCKER_IMAGE) -f $(DOCKER_FILE) .

docker-push:
	@echo "==> Pushing Docker image to registry..."
	docker push $(DOCKER_IMAGE)

## -------------------------------
## Docker Compose
## -------------------------------
compose-up:
	@echo "==> Bringing up Docker Compose..."
	DOCKER_IMAGE=$(DOCKER_IMAGE) $(COMPOSE) -f $(COMPOSE_FILE) --env-file .env up -d

compose-down:
	@echo "==> Shutting down Docker Compose..."
	$(COMPOSE) -f $(COMPOSE_FILE) down

restart: compose-down compose-up
	@echo "==> Restarted Docker Compose."
