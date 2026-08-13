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

# No ingress firewall rule: nothing reaches this cluster from the outside. cloudflared
# dials Cloudflare's edge outbound and traffic arrives back down that connection, which
# the default egress-allow rule already covers.
