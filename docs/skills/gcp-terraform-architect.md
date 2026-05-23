---
name: gcp-terraform-architect
description: Use when creating, editing, or reviewing any .tf or .tfvars file targeting Google Cloud Platform. Triggers on resource definitions, IAM bindings, GCS backends, GKE configs, service accounts, and variable blocks.
type: skill-profile
target: claude-code-workspace
engine: terraform-gcp
version: 1.0.0
---

# GCP Terraform Guardrails

## State & Layout

Each layer gets an isolated GCS backend. Cross-layer references use `terraform_remote_state` only — no hardcoded IDs.

Layers: `vpc/` → `iam/` → `compute/` → `data/`

## Mandatory Patterns

**Labels** (every resource):
```hcl
locals {
  common_labels = {
    environment = var.environment
    managed_by  = "terraform"
    layer       = basename(path.cwd)
  }
}
```

**Naming**: `${project_id}-${environment}-${component}` → `myapp-dev-message-svc`

**Variable validation**:
```hcl
validation {
  condition     = contains(["us-central1", "us-west1", "northamerica-northeast1"], var.region)
  error_message = "Region must be one of the approved GCP regions."
}

validation {
  condition     = contains(["dev", "staging", "prod"], var.environment)
  error_message = "Environment must be dev, staging, or prod."
}
```

**Lifecycle** (compute resources): prefer `create_before_destroy = true`

**Outputs**: root modules must export resource IDs and self-links.

## IAM Rules

- ❌ Never: `roles/owner`, `roles/editor`, `roles/viewer`
- ❌ Never: `google_service_account_key`
- ❌ Never: `google_project_iam_binding` (overwrites existing members)
- ✓ Always: `google_project_iam_member` (additive)
- ✓ Always: dedicated service account per component
- ✓ Always: GKE Workload Identity — no key files

## Status

Parked — written 2026-05-22. Not yet installed as an active skill.
Apply if/when `chat-infra/gcp/terraform/` is refactored into a layered layout.
Current flat layout (`gke.tf`, `namespaces.tf`, etc.) is intentional for a single-node demo cluster.
