resource "google_compute_disk" "postgres" {
  name  = "chat-demo-postgres"
  type  = "pd-standard"
  zone  = "us-west1-a"
  size  = 5
}

resource "google_compute_disk" "redpanda" {
  name  = "chat-demo-redpanda"
  type  = "pd-standard"
  zone  = "us-west1-a"
  size  = 5
}
