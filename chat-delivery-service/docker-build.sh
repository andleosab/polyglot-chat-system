#!/bin/sh

# docker buildx build -f src/main/docker/Dockerfile.native-micro.builder -t quarkus/chat-delivery-service .
docker build -f src/main/docker/Dockerfile.jvm.builder -t quarkus/chat-delivery-service .
