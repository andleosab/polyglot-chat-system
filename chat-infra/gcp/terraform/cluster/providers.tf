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
    # Bucket is supplied at init time, not hardcoded — keeps the project-specific
    # name out of version control. After running bootstrap/create-state-bucket.sh,
    # init with your real bucket name:
    #   terraform init -backend-config="bucket=<YOUR_PROJECT_ID>-tf-state"
    #
    # The budget module shares this bucket under a different prefix, which keeps
    # the two states independent. Do not point both at the same prefix.
    prefix = "chat-demo/state"
  }
}

provider "google" {
  project = var.project_id
  region  = "us-west1"
}

data "google_client_config" "default" {}

provider "kubernetes" {
  host                   = "https://${google_container_cluster.main.endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(google_container_cluster.main.master_auth[0].cluster_ca_certificate)
}

