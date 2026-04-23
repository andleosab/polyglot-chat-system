#!/bin/sh

docker compose -f docker-compose-ephemeral.yml --env-file .env -p chat-demo-infra down
