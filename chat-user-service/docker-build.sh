#!/bin/sh

#docker buildx build --push -f Dockerfile -t spring/chat-user-service .
docker buildx build -f Dockerfile.builder -t spring/chat-user-service .
