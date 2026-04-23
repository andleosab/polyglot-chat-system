#!/usr/bin/env bash
set -e

echo "======================================="
echo "Stopping infra..."
echo "======================================="
(cd chat-infra/docker && cp env-example .env && ./stop-services-ephemeral.sh && rm .env)

echo "======================================="
echo "Starting service chat-user-service..."
echo "======================================="
(cd chat-user-service && cp env-example .env.docker && ./stop-services.sh && rm .env.docker)

echo "======================================="
echo "Starting service chat-delivery-service..."
echo "======================================="
(cd chat-delivery-service && cp env-example .env.docker && ./stop-services.sh && rm .env.docker)   

echo "======================================="
echo "Starting service chat-message-service..."
echo "======================================="
(cd chat-message-service && cp env-example .env && make compose-down && rm .env)

echo "======================================="
echo "Starting chat-web..."
echo "======================================="
(cd chat-web && cp env-example .env.local.docker && ./compose-down.sh && rm .env.local.docker)

echo "Services stopped successfully."