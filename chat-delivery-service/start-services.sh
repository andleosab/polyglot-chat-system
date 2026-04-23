#!/bin/bash

source .env.docker

# convert standard Base64 to Base64URL first, then embed in JWK
JWT_KEY_URL=$(echo -n "$JWT_SECRET" | tr '+/' '-_' | tr -d '=')
JWT_JWK=$(echo -n "{\"kty\":\"oct\",\"alg\":\"HS256\",\"k\":\"$JWT_KEY_URL\",\"use\":\"sig\"}" | base64 | tr '+/' '-_' | tr -d '=')

env | grep JWT

JWT_JWK=$JWT_JWK ENV_FILE=.env.docker \
docker compose -f docker-compose.yml up -d

