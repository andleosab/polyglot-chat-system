terraform {
  required_version = ">= 1.6"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.0"
    }
  }

  backend "gcs" {
    # Same bucket as the cluster module, different prefix — the two states are
    # independent, which is the whole point of the split: `terraform destroy` in
    # ../cluster must not take the budget with it.
    #   terraform init -backend-config="bucket=<YOUR_PROJECT_ID>-tf-state"
    prefix = "chat-demo/budget"
  }
}

provider "google" {
  project = var.project_id
  region  = "us-west1"

  # billingbudgets.googleapis.com bills API quota to a project via the
  # X-Goog-User-Project header. Under user ADC that header is unset by default,
  # so the budget-create call falls back to a Google-owned project and fails with
  # a misleading SERVICE_DISABLED 403. These two settings make Terraform always
  # send your project. (Setting the quota project on the ADC via gcloud does NOT
  # help — Terraform ignores that field.)
  user_project_override = true
  billing_project       = var.project_id
}
