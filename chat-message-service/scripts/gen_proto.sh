#!/bin/bash
protoc --go_out=internal/proto --go-grpc_out=internal/proto proto/**/*.proto


#!/bin/bash
set -e

PROTO_DIR=proto
OUT_DIR=internal/grpc

mkdir -p $OUT_DIR

protoc \
  --go_out=$OUT_DIR \
  --go-grpc_out=$OUT_DIR \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  $PROTO_DIR/*.proto


#!/bin/bash
set -e

# Root directory of the project
ROOT_DIR="$(dirname "$0")/.."

# Paths
PROTO_DIR="$ROOT_DIR/internal/grpc/proto"
OUT_DIR="$ROOT_DIR/internal/grpc/server"

# Make sure output directory exists
mkdir -p "$OUT_DIR"

echo "Generating Go code from proto files..."

# Generate Go code (messages + gRPC)
protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$OUT_DIR" \
  --go-grpc_out="$OUT_DIR" \
  "$PROTO_DIR"/*.proto

echo "Done."
