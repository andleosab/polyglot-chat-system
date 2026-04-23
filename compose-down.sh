#!/usr/bin/env bash
set -e

echo "======================================="
echo "Stopping infra..."
echo "======================================="
(cd chat-infra/docker && ./stop-services-ephemeral.sh)

echo "======================================="
echo "Starting service chat-user-service..."
echo "======================================="
(cd chat-user-service && ./stop-services.sh)

echo "======================================="
echo "Starting service chat-delivery-service..."
echo "======================================="
(cd chat-delivery-service && ./stop-services.sh)   

echo "======================================="
echo "Starting service chat-message-service..."
echo "======================================="
(cd chat-message-service && make compose-down )

echo "======================================="
echo "Starting chat-web..."
echo "======================================="
(cd chat-web && ./compose-down.sh)

echo "Services stopped successfully."