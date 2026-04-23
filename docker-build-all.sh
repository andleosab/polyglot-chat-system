#!/usr/bin/env bash
set -e


echo "======================================="
echo "Building chat-user-service image..."
echo "======================================="
(cd chat-user-service && ./docker-build.sh)

echo "======================================="
echo "Building chat-delivery-service image..."
echo "======================================="
(cd chat-delivery-service && ./docker-build.sh)

echo "======================================="
echo "Building chat-message-service image..."
echo "======================================="
(cd chat-message-service && make docker-build)

echo "======================================="
echo "Building chat-web image..."
echo "======================================="
(cd chat-web && cp env-example .env.local.docker && ./build-docker.sh && rm .env.local.docker)

echo "Images built successfully."