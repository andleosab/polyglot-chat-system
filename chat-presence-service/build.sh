#!/bin/bash
set -e
cd "$(dirname "$0")"
uv lock
make gen-proto
echo "Build complete. Run with: uv run uvicorn main:app --host 0.0.0.0 --port 8000"
