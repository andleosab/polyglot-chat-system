# Helm releases are no longer managed by Terraform.
#
# Infra services (Postgres, Redis, Redpanda, Nginx) are deployed manually:
#   ./bootstrap/deploy-infra-services.sh
#
# Microservices are deployed by GitHub Actions on every push to main:
#   .github/workflows/build-images.yml (build + helm upgrade per service)
#
# Per-service Helm values live in: helm/values/<service-name>.yaml
