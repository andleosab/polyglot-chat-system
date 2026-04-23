#!/bin/sh

docker run --rm \
  --name chat-web \
  --network chat-demo-net \
  -p 3000:3000 \
  --env-file .env.local.docker \
  chat-web