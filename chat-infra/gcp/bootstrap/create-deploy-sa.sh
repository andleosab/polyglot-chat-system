#!/bin/bash
# One-time setup: creates the GitHub Actions deploy service account and its key.
# Run this once before the first CI deploy. Requires an account with
# roles/iam.serviceAccountAdmin + roles/resourcemanager.projectIamAdmin (e.g. owner).
#
# Usage: ./create-deploy-sa.sh
# Requires: gcloud CLI authenticated with a project set (`gcloud config set project <PROJECT_ID>`)
#           gh CLI authenticated (optional — only for the automatic secret push)
set -euo pipefail

PROJECT_ID=$(gcloud config get-value project)
SA_NAME="github-actions"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
KEY_FILE="gha-key.json"

echo "Creating deploy service account: ${SA_EMAIL}"

gcloud iam service-accounts create "${SA_NAME}" \
  --project="${PROJECT_ID}" \
  --display-name="GitHub Actions deployer"

# Grant only the two roles the deploy workflows need
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.writer"   # push images to Artifact Registry

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/container.developer"        # helm upgrade against GKE

echo "Creating JSON key: ${KEY_FILE}"
gcloud iam service-accounts keys create "${KEY_FILE}" --iam-account="${SA_EMAIL}"

echo ""
echo "Done. Add these two GitHub Actions secrets:"
echo "  Option A (gh CLI):"
echo "    gh secret set GCP_PROJECT_ID --body \"${PROJECT_ID}\""
echo "    gh secret set GCP_SA_KEY < ${KEY_FILE}"
echo "  Option B (UI): repo Settings -> Secrets and variables -> Actions"
echo "    GCP_PROJECT_ID = ${PROJECT_ID}"
echo "    GCP_SA_KEY     = contents of ${KEY_FILE}"
echo ""
echo "After adding the secrets, delete the local key: rm ${KEY_FILE}"
