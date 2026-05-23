#!/bin/bash
# Deploy infra services (Postgres, Redis, Redpanda, Nginx) to the GKE cluster.
# Run AFTER terraform apply. Re-run any time an infra service config changes.
#
# Prerequisites:
#   gcloud container clusters get-credentials chat-demo --zone us-west1-a
#   POSTGRES_PASSWORD env var set (must match the value in terraform.tfvars)
set -euo pipefail

NAMESPACE="chat"
CHARTS_DIR="$(cd "$(dirname "$0")/../helm" && pwd)"
PROJECT_ID=$(gcloud config get-value project)
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:?POSTGRES_PASSWORD env var is required}"

echo "Deploying infra services to namespace: $NAMESPACE (project: $PROJECT_ID)"

helm upgrade --install postgres "$CHARTS_DIR/postgres" \
  --namespace "$NAMESPACE" \
  --set gcpProjectId="$PROJECT_ID" \
  --set postgresPassword="$POSTGRES_PASSWORD" \
  --wait --timeout 3m

helm upgrade --install redis "$CHARTS_DIR/redis" \
  --namespace "$NAMESPACE" \
  --wait --timeout 2m

helm upgrade --install redpanda "$CHARTS_DIR/redpanda" \
  --namespace "$NAMESPACE" \
  --set gcpProjectId="$PROJECT_ID" \
  --wait --timeout 3m

helm upgrade --install nginx "$CHARTS_DIR/nginx" \
  --namespace "$NAMESPACE" \
  --wait --timeout 2m

echo "All infra services deployed."
