#!/usr/bin/env bash
set -e

echo "======================================="
echo "Starting infra..."
echo "======================================="
(cd chat-infra/docker && cp env-example .env && ./start-services-ephemeral.sh && rm .env)

# ------------------ wait for Postgres ------------------
POSTGRES_CONTAINER=chat-postgres-ephemeral   # adjust if container_name is different
POSTGRES_USER=postgres                      # adjust to your DB user

echo "Waiting for Postgres to be ready..."
until docker exec "$POSTGRES_CONTAINER" pg_isready -U "$POSTGRES_USER" > /dev/null 2>&1; do
    echo "Postgres not ready yet... sleeping 2s"
    sleep 2
done
echo "Postgres is ready!"
# -------------------------------------------------------

# ------------------ wait for Kafka ------------------
KAFKA_CONTAINER=chat-kafka-ephemeral  # adjust if container_name is different
echo "Waiting for Kafka broker port..."

echo "Waiting for Kafka to be ready..."
until docker exec "$KAFKA_CONTAINER" rpk cluster health > /dev/null 2>&1; do
    echo "Kafka not ready yet... sleeping 2s"
    sleep 2
done

echo "Kafka port is reachable!"
# -------------------------------------------------------

echo "======================================="
echo "Starting service chat-user-service..."
echo "======================================="
(cd chat-user-service && cp env-example .env.docker && ./start-services.sh && rm .env.docker)

echo "======================================="
echo "Starting service chat-delivery-service..."
echo "======================================="
(cd chat-delivery-service && cp env-example .env.docker && ./start-services.sh && rm .env.docker)   

echo "======================================="
echo "Starting service chat-message-service..."
echo "======================================="
(cd chat-message-service && cp env-example .env && make migrate-up compose-up && rm .env)

echo "======================================="
echo "Starting chat-web..."
echo "======================================="
# (cd chat-web && cp env-example .env.local.docker && ./compose-up.sh && rm .env.local.docker)
(cd chat-web && \
  CREATED=0; \
  if [ ! -f .env.local.docker ]; then cp env-example .env.local.docker && CREATED=1; fi; \
  ./compose-up.sh; \
  [ "$CREATED" -eq 1 ] && rm .env.local.docker)

echo "Services started successfully."