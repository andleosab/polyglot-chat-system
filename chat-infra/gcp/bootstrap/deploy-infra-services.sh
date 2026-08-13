#!/bin/bash
# Deploy infra services (Postgres, Redis, Redpanda, Nginx, cloudflared) to the GKE cluster.
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

# cloudflared is the cluster's only external ingress, and it needs a tunnel token to
# start at all — with an empty one it crashloops and hangs `helm --wait` until the
# timeout. The token comes from chat-secrets (set via tunnel_token in the cluster
# module's tfvars), so an empty value means no tunnel has been created yet: skip the
# release rather than fail the script.
TUNNEL_TOKEN=$(kubectl get secret chat-secrets -n "$NAMESPACE" \
  -o jsonpath='{.data.tunnel-token}' 2>/dev/null | base64 -d 2>/dev/null || true)

if [ -n "$TUNNEL_TOKEN" ]; then
  helm upgrade --install cloudflared "$CHARTS_DIR/cloudflared" \
    --namespace "$NAMESPACE" \
    --wait --timeout 2m
else
  echo "Skipping cloudflared: tunnel-token is empty in chat-secrets."
  echo "  Set tunnel_token in terraform/cluster/terraform.tfvars, re-apply, and re-run this script."
  echo "  Until then, reach the app with: kubectl port-forward svc/nginx 8080:80 -n $NAMESPACE"
fi

echo "All infra services deployed."
