#!/bin/bash

set -e
COMPOSE_FILE=../docker/docker-compose.yml
ENV_FILE="../.env"
CHAT_SERVICE_NAME="chat-message-service" # Name of the chat service in docker-compose.yml

# Tail the logs of the chat service
# need to start with the same params as docker compose up!
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" -p chat-demo logs -f "$CHAT_SERVICE_NAME"