#!/bin/sh
ENV_FILE=.env.local.docker \
docker compose -f docker-compose.yml up -d