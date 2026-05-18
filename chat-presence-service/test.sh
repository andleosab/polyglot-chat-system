#!/bin/bash
set -e
cd "$(dirname "$0")"
uv lock
make gen-proto
uv run pytest -v
