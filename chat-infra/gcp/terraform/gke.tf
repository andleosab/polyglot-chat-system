resource "google_container_cluster" "main" {
  name     = "chat-demo"
  location = "us-west1-a"

  remove_default_node_pool = true
  initial_node_count       = 1

  network    = google_compute_network.main.name
  subnetwork = google_compute_subnetwork.main.name

  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }

  # Pinned explicitly so future GKE default shifts surface in `terraform plan`.
  release_channel {
    channel = "REGULAR"
  }

  # Allow destroy without extra confirmation step
  deletion_protection = false
}

resource "google_container_node_pool" "main" {
  name     = "main-pool"
  cluster  = google_container_cluster.main.name
  location = "us-west1-a"

  node_count = 1

  node_config {
    # e2-standard-2 has 2 DEDICATED vCPU (~1930m allocatable). Do not "downgrade" to
    # e2-medium to save cost: it is a shared-core burstable type — 2 vCPU of burst on a
    # 1 vCPU baseline — and GKE derives allocatable from the baseline, giving only 940m.
    # GKE's own system pods request ~753m of that, leaving ~187m for a stack that needs
    # 850m. See the resource budget in ../README.md.
    machine_type = "e2-standard-2"
    disk_size_gb = 20
    disk_type    = "pd-standard"
    image_type   = "COS_CONTAINERD"

    # Spot pricing keeps the node to roughly a third of on-demand (~$15/mo at the time
    # of writing — approximate; check the billing console for the current rate)
    spot = true

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]

    metadata = {
      disable-legacy-endpoints = "true"
    }

    tags = ["gke-node"]

    kubelet_config {
      # Cap container logs at 10Mi per file, 3 files max (30MB total per container)
      # Protects the 20GB boot disk from log overrun
      container_log_max_size  = "10Mi"
      container_log_max_files = 3
    }
  }
}

resource "google_artifact_registry_repository" "images" {
  location      = "us-west1"
  repository_id = "chat-demo"
  format        = "DOCKER"
  description   = "Docker images for chat-demo services"
}

# Pushing and pulling are done by two different identities, and only the pusher was
# ever granted anything. CI pushes as github-actions@ (create-deploy-sa.sh grants it
# roles/artifactregistry.writer); the kubelet pulls as the node pool's service
# account. The node_config above sets no service_account, so that is the default
# Compute Engine SA — which, on projects created under the current
# automaticIamGrantsForDefaultServiceAccounts policy, holds no roles at all.
#
# Without this binding every deploy workflow builds and pushes successfully, then
# times out at `helm --wait` while the pods sit in ImagePullBackOff:
#   Failed to pull image ".../chat-demo/<service>:<sha>": 403 Forbidden
# The image is present and correctly tagged; only the read permission is missing,
# which makes it read like a CI or tagging fault rather than an IAM one.
#
# Scoped to this repository rather than the whole project: pulling these images is
# the only registry access a node needs.
resource "google_artifact_registry_repository_iam_member" "node_pull" {
  project    = var.project_id
  location   = google_artifact_registry_repository.images.location
  repository = google_artifact_registry_repository.images.repository_id
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${data.google_project.current.number}-compute@developer.gserviceaccount.com"
}
