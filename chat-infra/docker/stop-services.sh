#!/bin/sh

docker compose -f docker-compose.yml --env-file .env -p chat-demo-infra down
