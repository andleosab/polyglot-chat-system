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
echo "Building chat-presence-service image..."
echo "======================================="
(cd chat-presence-service && ./docker-build.sh)

echo "======================================="
echo "Building chat-web image..."
echo "======================================="
(cd chat-web && ./build-docker.sh)

echo "Images built successfully."