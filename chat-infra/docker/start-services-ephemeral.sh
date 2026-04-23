#!/bin/sh

docker compose -f docker-compose-ephemeral.yml --env-file .env up -d
