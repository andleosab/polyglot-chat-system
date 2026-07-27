#!/bin/bash
# One-time setup: creates the private GCS bucket for Terraform remote state.
# Run this BEFORE `terraform init`. Never manage this bucket with Terraform itself.
#
# Usage: ./create-state-bucket.sh
# Requires: gcloud CLI authenticated with a project set (`gcloud config set project <PROJECT_ID>`)
set -euo pipefail

PROJECT_ID=$(gcloud config get-value project)
BUCKET_NAME="${PROJECT_ID}-tf-state"
REGION="us-west1"

echo "Creating Terraform state bucket: gs://${BUCKET_NAME}"

gcloud storage buckets create "gs://${BUCKET_NAME}" \
  --project="${PROJECT_ID}" \
  --location="${REGION}" \
  --uniform-bucket-level-access \
  --public-access-prevention

gcloud storage buckets update "gs://${BUCKET_NAME}" --versioning

echo ""
echo "Done. Initialize Terraform with this bucket (do not hardcode it in providers.tf):"
echo "  terraform init -backend-config=\"bucket=${BUCKET_NAME}\""
