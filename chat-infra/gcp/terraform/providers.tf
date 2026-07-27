terraform {
  required_version = ">= 1.6"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 3.0"
    }
  }

  backend "gcs" {
    # Fill in after running bootstrap/create-state-bucket.sh
    bucket = "<PROJECT_ID>-tf-state"
    prefix = "chat-demo/state"
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

data "google_client_config" "default" {}

provider "kubernetes" {
  host                   = "https://${google_container_cluster.main.endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(google_container_cluster.main.master_auth[0].cluster_ca_certificate)
}

