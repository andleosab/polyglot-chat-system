resource "google_compute_network" "main" {
  name                    = "chat-demo-vpc"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "main" {
  name          = "chat-demo-subnet"
  ip_cidr_range = "10.0.0.0/16"
  region        = "us-west1"
  network       = google_compute_network.main.id

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.1.0.0/16"
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.2.0.0/16"
  }
}

# Allows external traffic to reach Nginx on NodePort 30080 only.
resource "google_compute_firewall" "nodeport_http" {
  name    = "chat-demo-nodeport-http"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["30080"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["gke-node"]
}
