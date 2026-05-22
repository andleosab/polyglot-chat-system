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

  # Allow destroy without extra confirmation step
  deletion_protection = false
}

resource "google_container_node_pool" "main" {
  name     = "main-pool"
  cluster  = google_container_cluster.main.name
  location = "us-west1-a"

  node_count = 1

  node_config {
    machine_type = "e2-medium"
    disk_size_gb = 20
    disk_type    = "pd-standard"
    image_type   = "COS_CONTAINERD"

    # Spot pricing reduces node cost to ~$12/mo
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
