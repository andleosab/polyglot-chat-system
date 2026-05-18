APP_NAME=chat-presence-service
DOCKER_FILE=Dockerfile
DOCKER_IMAGE=$(APP_NAME)
COMPOSE_FILE=docker-compose.yml
COMPOSE=docker compose

# Load .env if present
-include .env
export

.PHONY: gen-proto dev test lint docker-build docker-push compose-up compose-down restart

## -------------------------------
## Local development
## -------------------------------
gen-proto:
	uv run python -m grpc_tools.protoc \
	  -I proto \
	  --python_out=app/proto \
	  --grpc_python_out=app/proto \
	  proto/presence.proto

dev:
	uv run fastapi dev main.py

test:
	uv run pytest

lint:
	uv run ruff check .
	uv run ruff format --check .

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
	DOCKER_IMAGE=$(DOCKER_IMAGE) $(COMPOSE) -f $(COMPOSE_FILE) --env-file .env.docker up -d

compose-down:
	@echo "==> Shutting down Docker Compose..."
	$(COMPOSE) -f $(COMPOSE_FILE) down

restart: compose-down compose-up
	@echo "==> Restarted Docker Compose."
