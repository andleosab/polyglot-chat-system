#!/bin/bash
set -e

echo "Generating SQLC code..."
sqlc generate -f ../sqlc.yaml
echo "Done."
