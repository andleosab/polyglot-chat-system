# GKE Demo Cluster

GCP infrastructure for the polyglot chat demo. Designed to stay within the GCP always-free tier.

## Deployment model

| Layer | What | How |
|---|---|---|
| GCP resources | Cluster, VPC, disks, namespace, K8s secrets | `terraform apply` |
| Infra services | Postgres, Redis, Redpanda, Nginx | `bootstrap/deploy-infra-services.sh` (run once after apply) |
| Microservices | 5 app services | GitHub Actions on push to `main` |

## What's here

```
bootstrap/
  create-state-bucket.sh      # one-time GCS state bucket setup — run before terraform init
  deploy-infra-services.sh    # deploy Postgres, Redis, Redpanda, Nginx via Helm

terraform/
  providers.tf                # GCS backend, google/kubernetes providers
  variables.tf                # input variables (secrets via terraform.tfvars)
  gke.tf                      # GKE Standard cluster + e2-medium Spot node pool + Artifact Registry
  vpc.tf                      # VPC, subnet, firewall (NodePort 30080 only)
  disks.tf                    # pre-provisioned pd-standard disks: Postgres (5GB) + Redpanda (5GB)
  namespaces.tf               # chat namespace + shared K8s Secret (chat-secrets)
  helm_releases.tf            # intentionally empty — see file for explanation
  budget.tf                   # $1 budget alert (fires even while $300 trial credit drains)

helm/
  microservice/               # generic chart reused by all 5 microservices
  nginx/                      # Nginx reverse proxy (reuses existing nginx.conf routing rules)
  redis/                      # ephemeral Redis — no PVC, data is TTL-based
  postgres/                   # StatefulSet on pre-created GCE disk, init SQL as ConfigMap
  redpanda/                   # single-broker Redpanda, memory-capped at 600MB
  values/
    chat-web.yaml             # per-service Helm values (env vars, ports, resource limits)
    chat-user-service.yaml
    chat-message-service.yaml
    chat-delivery-service.yaml
    chat-presence-service.yaml
```

## Cost model

| Resource | Cost |
|---|---|
| GKE cluster management fee | $0 — covered by GKE $74.40/mo credit |
| e2-medium Spot node | ~$12/mo — absorbed by $300 new-account credit |
| 30GB pd-standard disks (20GB boot + 5GB Postgres + 5GB Redpanda) | $0 — always-free cap |
| GCS state bucket | $0 — state file is a few KB, well under 5GB always-free cap |
| Artifact Registry | $0 — under 0.5GB/mo free tier |
| Budget alert | $1.00 threshold, `EXCLUDE_ALL_CREDITS=true` so it fires during trial period |

## First-run sequence

```bash
# 1. Authenticate and set project
gcloud auth login
gcloud config set project <YOUR_PROJECT_ID>

# 2. Create the Terraform state bucket (one time only — never re-run)
cd chat-infra/gcp/bootstrap
./create-state-bucket.sh
# Then update the bucket name in terraform/providers.tf

# 3. Init Terraform
cd ../terraform
terraform init

# 4. Create terraform.tfvars (gitignored — never commit this)
cat > terraform.tfvars <<EOF
project_id         = "<YOUR_PROJECT_ID>"
billing_account    = "<XXXXXX-XXXXXX-XXXXXX>"
jwt_secret         = "<shared HMAC secret>"
jwt_jwk            = "<base64url-encoded JWT secret for Quarkus>"
postgres_password  = "<postgres superuser password>"
better_auth_secret = "<Better Auth session signing secret>"
EOF

# 5. Apply (provisions GCP resources + namespace + K8s secrets only)
terraform apply

# 6. Configure kubectl for the new cluster
gcloud container clusters get-credentials chat-demo --zone us-west1-a

# 7. Deploy infra services (Postgres, Redis, Redpanda, Nginx)
cd ../bootstrap
POSTGRES_PASSWORD=<same as terraform.tfvars> ./deploy-infra-services.sh

# 8. Get the node's external IP and update BETTER_AUTH_URL
kubectl get nodes -o wide
# Edit helm/values/chat-web.yaml — replace REPLACE_WITH_NODE_EXTERNAL_IP
# Then trigger a deploy via: git commit --allow-empty -m "trigger deploy" && git push

# 9. Microservices deploy automatically via GitHub Actions on push to main
#    Or trigger manually: Actions → Build and Deploy → Run workflow

# 10. Access the app
open http://<EXTERNAL-IP>:30080

# 11. Tear down when done (disks are retained — data survives)
terraform destroy
```

## Debugging without external exposure

All services except Nginx are ClusterIP-only. Use `kubectl port-forward` for direct access:

```bash
# Postgres
kubectl port-forward svc/postgres 5432:5432 -n chat

# Redpanda (Kafka)
kubectl port-forward svc/redpanda 9092:9092 -n chat

# Presence service HTTP
kubectl port-forward svc/chat-presence-service 8000:8000 -n chat
```

## GitHub Actions setup

Add these secrets in repo Settings → Secrets and variables → Actions:

| Secret | Value |
|---|---|
| `GCP_PROJECT_ID` | Your GCP project ID |
| `GCP_SA_KEY` | JSON key for a service account with **Artifact Registry Writer** + **GKE Developer** roles |

The workflow (`.github/workflows/build-images.yml`) builds each service image and deploys it via `helm upgrade --install` on every push to `main`. Each service deploys independently in parallel.

## Architecture decisions

See [`docs/adr/0005-gke-standard-over-autopilot.md`](../../docs/adr/0005-gke-standard-over-autopilot.md) for why GKE Standard was chosen over Autopilot.
