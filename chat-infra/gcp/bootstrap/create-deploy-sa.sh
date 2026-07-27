#!/bin/bash
# One-time setup: creates the GitHub Actions deploy service account and wires it to
# GitHub via Workload Identity Federation (WIF). No service-account key is created —
# workflows authenticate with a short-lived OIDC token instead.
#
# Run this once before the first CI deploy. Requires an account with
# roles/iam.serviceAccountAdmin + roles/resourcemanager.projectIamAdmin (e.g. owner).
#
# Usage: ./create-deploy-sa.sh
# Requires: gcloud CLI authenticated with a project set (`gcloud config set project <PROJECT_ID>`)
#
# Safe to re-run: every step is idempotent and skips work that already exists.
set -euo pipefail

PROJECT_ID=$(gcloud config get-value project)
PROJECT_NUMBER=$(gcloud projects describe "${PROJECT_ID}" --format='value(projectNumber)')
SA_NAME="github-actions"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
POOL_ID="github"
PROVIDER_ID="github"

# The repo allowed to impersonate the service account. Derived from the git remote so
# a fork doesn't silently grant access to the upstream repo; override by exporting
# GITHUB_REPO=owner/name before running.
GITHUB_REPO="${GITHUB_REPO:-$(git config --get remote.origin.url \
  | sed -E 's#(git@github.com:|https://github.com/)##; s#\.git$##')}"

if [[ ! "${GITHUB_REPO}" =~ ^[^/]+/[^/]+$ ]]; then
  echo "Could not determine the GitHub repo (got: '${GITHUB_REPO}')." >&2
  echo "Set it explicitly: GITHUB_REPO=owner/name ./create-deploy-sa.sh" >&2
  exit 1
fi
REPO_OWNER="${GITHUB_REPO%%/*}"

echo "Project: ${PROJECT_ID} (${PROJECT_NUMBER})"
echo "Repo:    ${GITHUB_REPO}"
echo ""

# WIF needs the STS and IAM Credentials APIs; neither is on by default everywhere.
echo "Enabling STS + IAM Credentials APIs..."
gcloud services enable sts.googleapis.com iamcredentials.googleapis.com \
  --project="${PROJECT_ID}"

echo "Creating deploy service account: ${SA_EMAIL}"
if gcloud iam service-accounts describe "${SA_EMAIL}" --project="${PROJECT_ID}" &>/dev/null; then
  echo "  already exists, skipping"
else
  gcloud iam service-accounts create "${SA_NAME}" \
    --project="${PROJECT_ID}" \
    --display-name="GitHub Actions deployer"
fi

# Grant only the two roles the deploy workflows need.
# add-iam-policy-binding is idempotent — re-adding an existing binding is a no-op.
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.writer" \
  --condition=None --quiet >/dev/null   # push images to Artifact Registry

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/container.developer" \
  --condition=None --quiet >/dev/null   # helm upgrade against GKE

echo "Creating workload identity pool: ${POOL_ID}"
if gcloud iam workload-identity-pools describe "${POOL_ID}" \
     --project="${PROJECT_ID}" --location="global" &>/dev/null; then
  echo "  already exists, skipping"
else
  gcloud iam workload-identity-pools create "${POOL_ID}" \
    --project="${PROJECT_ID}" \
    --location="global" \
    --display-name="GitHub Actions"
fi

# The attribute condition is the security boundary. Without it ANY GitHub repo's OIDC
# token could impersonate this service account — the pool trusts GitHub's issuer, not
# your repo. Restricting to the owner here, and to the exact repo in the IAM binding
# below, means both must match before a token is honoured.
echo "Creating OIDC provider: ${PROVIDER_ID}"
if gcloud iam workload-identity-pools providers describe "${PROVIDER_ID}" \
     --project="${PROJECT_ID}" --location="global" \
     --workload-identity-pool="${POOL_ID}" &>/dev/null; then
  echo "  already exists, skipping"
else
  gcloud iam workload-identity-pools providers create-oidc "${PROVIDER_ID}" \
    --project="${PROJECT_ID}" \
    --location="global" \
    --workload-identity-pool="${POOL_ID}" \
    --display-name="GitHub Actions OIDC" \
    --issuer-uri="https://token.actions.githubusercontent.com" \
    --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository,attribute.repository_owner=assertion.repository_owner" \
    --attribute-condition="assertion.repository_owner == '${REPO_OWNER}'"
fi

# Allow only this repo to impersonate the SA.
POOL_NAME="projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_ID}"
echo "Binding ${GITHUB_REPO} to ${SA_EMAIL}"
gcloud iam service-accounts add-iam-policy-binding "${SA_EMAIL}" \
  --project="${PROJECT_ID}" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/${POOL_NAME}/attribute.repository/${GITHUB_REPO}" \
  --quiet >/dev/null

PROVIDER_RESOURCE="${POOL_NAME}/providers/${PROVIDER_ID}"

echo ""
echo "Done. No key file was created — nothing to delete or rotate."
echo ""
echo "Add these three GitHub Actions secrets:"
echo "  Option A (gh CLI):"
echo "    gh secret set GCP_PROJECT_ID  --body \"${PROJECT_ID}\""
echo "    gh secret set GCP_SA_EMAIL    --body \"${SA_EMAIL}\""
echo "    gh secret set GCP_WIF_PROVIDER --body \"${PROVIDER_RESOURCE}\""
echo "  Option B (UI): repo Settings -> Secrets and variables -> Actions"
echo "    GCP_PROJECT_ID   = ${PROJECT_ID}"
echo "    GCP_SA_EMAIL     = ${SA_EMAIL}"
echo "    GCP_WIF_PROVIDER = ${PROVIDER_RESOURCE}"
