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
  create-deploy-sa.sh         # one-time: GitHub Actions deploy SA + roles + JSON key
  deploy-infra-services.sh    # deploy Postgres, Redis, Redpanda, Nginx via Helm

terraform/
  providers.tf                # GCS backend, google/kubernetes providers
  variables.tf                # input variables (secrets via terraform.tfvars)
  gke.tf                      # GKE Standard cluster + e2-medium Spot node pool + Artifact Registry
  vpc.tf                      # VPC, subnet, firewall (NodePort 30080 only)
  disks.tf                    # pre-provisioned pd-standard disks: Postgres (5GB) + Redpanda (5GB)
  namespaces.tf               # chat namespace + shared K8s Secret (chat-secrets)
  helm_releases.tf            # intentionally empty — see file for explanation
  budget.tf                   # 1-unit budget alert in the billing account's currency (fires while trial credit drains)
  terraform.tfvars.example    # template — copy to terraform.tfvars and fill in secrets

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
| Budget alert | 1-unit threshold in the billing account's currency, `EXCLUDE_ALL_CREDITS=true` so it fires during trial period |

## Resource budget

Every workload shares one 4GB `e2-medium`. After GKE system overhead, ~2.9GB is
schedulable. Memory **request** (what the scheduler must place) and **limit**
(burst ceiling) per pod:

| Pod | Request | Limit |
|---|---|---|
| chat-web | 128Mi | 450Mi |
| chat-user-service | 256Mi | 350Mi |
| chat-message-service | 64Mi | 128Mi |
| chat-delivery-service | 256Mi | 350Mi |
| chat-presence-service | 128Mi | 256Mi |
| postgres | 128Mi | 300Mi |
| redpanda | 700Mi | 700Mi |
| redis | 32Mi | 64Mi |
| nginx | 32Mi | 64Mi |
| **Total** | **~1.7GB** | **~2.6GB** |

Requests (~1.7GB) fit the node's ~2.9GB allocatable with headroom for `kube-system`.
JVM services (`chat-user-service`, `chat-delivery-service`) additionally bound heap via
`JAVA_TOOL_OPTIONS=-Xms64m -Xmx256m`. Redis is ephemeral (no disk); Redpanda's process
is capped at 600MB via `--memory=600M`.

## Prerequisites

Install these before starting. The first-run sequence assumes all of them are on `PATH`.

| Tool | Needed for | Check |
|---|---|---|
| `gcloud` | Everything — auth, cluster, billing | `gcloud version` |
| `gke-gcloud-auth-plugin` | Every `kubectl`/`helm` call against GKE | `gke-gcloud-auth-plugin --version` |
| `terraform` | Steps 3–5 | `terraform version` |
| `kubectl` | Steps 6, 9 | `kubectl version --client` |
| `helm` | Step 7 and the deploy workflows | `helm version --short` |
| `gh` | Step 8, pushing repo secrets (optional — the UI works too) | `gh auth status` |
| `jq` | The billing-IAM check in pre-apply setup | `jq --version` |

> **`gke-gcloud-auth-plugin` is the one that bites.** Step 6 writes a kubeconfig that
> references it, so if it's missing every later `kubectl` and `helm` command fails with
> an opaque credential error rather than a "not installed" message. Install it with
> `gcloud components install gke-gcloud-auth-plugin`.

## Pre-apply setup (fresh project only)

On a brand-new GCP project, `terraform apply` will fail partway through unless
these are done first. Both cause a *partial* apply — cluster/disks/networks
succeed, then the budget resource errors out.

**1. Enable the APIs the apply touches.** `billingbudgets.googleapis.com` and
`cloudresourcemanager.googleapis.com` are **not** guaranteed enabled on a fresh
project, and there is no `google_project_service` resource to enable them — so the
apply hard-fails without this. (The budget needs Billing Budgets; the budget's
`google_project` lookup that resolves the project number needs Resource Manager.)

```bash
gcloud services enable \
  cloudresourcemanager.googleapis.com \
  container.googleapis.com \
  compute.googleapis.com \
  artifactregistry.googleapis.com \
  billingbudgets.googleapis.com \
  --project=<YOUR_PROJECT_ID>
```

`compute`/`container` self-enable on first use, but enabling all four up-front is
faster and avoids one-time race warnings.

> **The budget also needs a quota project.** `billingbudgets.googleapis.com` bills
> API quota to a project via the `X-Goog-User-Project` header. Under user ADC that
> header is unset by default, so the call falls back to a Google-owned project and
> fails with a misleading `SERVICE_DISABLED` 403 (the error's `consumer` field
> points at a project number that isn't yours). `providers.tf` sets
> `user_project_override = true` + `billing_project = var.project_id` so Terraform
> always sends your project — no manual step needed. Note that `gcloud auth
> application-default set-quota-project` fixes this for the gcloud CLI but **not**
> for Terraform, which ignores the ADC quota-project field.

**2. Confirm your identity can create the budget.** `google_billing_budget`
needs the `billing.budgets.create` permission on the **billing account** (not the
project) — granted by `roles/billing.costsManager` (least-privilege) or
`roles/billing.admin`. Note that `roles/billing.user` is **not** enough (it only
covers project↔billing-account association). Fresh service-account setups usually
lack all of these:

```bash
gcloud beta billing accounts get-iam-policy <BILLING_ACCOUNT_ID> --format=json \
  | jq '.bindings[] | select(.members[] | contains("<your-identity>"))'
```

If missing, grant `roles/billing.costsManager` on the billing account before
applying (the grantor needs `roles/billing.admin`, e.g. the account owner):

```bash
gcloud billing accounts add-iam-policy-binding <BILLING_ACCOUNT_ID> \
  --member='user:you@example.com' \
  --role='roles/billing.costsManager'
```

Or via the Console: **Billing → <account> → Account management → Add principal →
role "Billing Account Costs Manager"**.

> Budget alerts have no notification channel configured, so on threshold breach
> GCP emails the Billing Account Administrators and Users by default (no
> Slack/SMS). This is expected.

## First-run sequence

```bash
# 1. Authenticate — the gcloud CLI AND Application Default Credentials.
#    Terraform reads ADC, not the gcloud CLI login, so BOTH are required.
gcloud auth login                        # authorizes gcloud/gsutil commands
gcloud auth application-default login    # sets ADC — Terraform authenticates with this
gcloud config set project <YOUR_PROJECT_ID>

# 2. Create the Terraform state bucket (one time only — never re-run)
cd chat-infra/gcp/bootstrap
./create-state-bucket.sh
# The script prints the bucket name (<YOUR_PROJECT_ID>-tf-state) for the next step

# 3. Init Terraform — the bucket is passed here, not hardcoded in providers.tf
cd ../terraform
terraform init -backend-config="bucket=<YOUR_PROJECT_ID>-tf-state"

# 4. Create terraform.tfvars (gitignored — never commit this)
cat > terraform.tfvars <<EOF
project_id         = "<YOUR_PROJECT_ID>"
billing_account    = "<XXXXXX-XXXXXX-XXXXXX>"
jwt_secret         = "<shared HMAC secret>"
jwt_jwk            = "<base64url-encoded JWT secret for Quarkus>"
postgres_password  = "<postgres superuser password>"
better_auth_secret = "<Better Auth session signing secret>"
EOF

# 5. Apply in two steps — kubernetes provider can't connect until the cluster exists
#    Step 5a: create cluster (Terraform auto-includes VPC + subnet as dependencies)
terraform apply -target=google_container_cluster.main

#    Step 5b: full apply — cluster now exists, kubernetes provider can connect
#             Creates: node pool, disks, firewall, Artifact Registry, budget,
#                      chat namespace, chat-secrets K8s Secret
terraform apply

# 6. Configure kubectl for the new cluster
gcloud container clusters get-credentials chat-demo --zone us-west1-a

# 7. Deploy infra services (Postgres, Redis, Redpanda, Nginx)
cd ../bootstrap
POSTGRES_PASSWORD=<same as terraform.tfvars> ./deploy-infra-services.sh

# 8. Create the GitHub Actions deploy service account + Workload Identity Federation
#    No key file is created — workflows authenticate with a short-lived OIDC token.
#    Safe to re-run; every step skips work that already exists.
./create-deploy-sa.sh
#    Then add the three repo secrets the script prints (GCP_PROJECT_ID,
#    GCP_SA_EMAIL, GCP_WIF_PROVIDER). CI auth must exist before the first run.

# 9. Get the node's external IP and point chat-web at it
kubectl get nodes -o wide
# Edit helm/values/chat-web.yaml — replace REPLACE_WITH_NODE_EXTERNAL_IP in ALL
# THREE env vars: BETTER_AUTH_URL, ORIGIN, and PUBLIC_DELIVERY_API_BASE.
# Missing the third is the easy mistake — the page loads fine and only the
# WebSocket connect fails.
# Commit and push the change to main — this triggers Deploy chat-web automatically

# 10. First-time microservice bootstrap (path filters won't fire on a fresh cluster)
#    Go to Actions → select each "Deploy <service>" workflow → Run workflow
#    After first run, path-based triggers handle redeployment automatically
#    Note: chat-web already ran from step 9's push (its path filter covers
#    helm/values/chat-web.yaml), so dispatching it here runs it a second time.
#    Harmless — don't read the duplicate run as a failure.

# 11. Access the app
open http://<EXTERNAL-IP>:30080

# 12. Tear down when done (disks are retained — data survives)
terraform destroy
```

## The node IP is ephemeral — expect it to change

The node pool runs a **single Spot `e2-medium`** (`gke.tf`), chosen deliberately for
cost (see ADR-0005). Spot nodes are reclaimable: GCP can preempt the node at any
time, and GKE replaces it. The replacement comes back with a **new external IP**.

Because step 9 bakes that IP into three env vars in `helm/values/chat-web.yaml`, a
preemption leaves `chat-web` pointing at an address that no longer exists — logins
fail against the stale `BETTER_AUTH_URL` and the WebSocket never connects. Nothing
in the cluster reports unhealthy; the pods are fine, the URL is wrong.

**Recovery:** re-run step 9 with the new IP from `kubectl get nodes -o wide`, commit,
and let the workflow redeploy.

If the churn becomes annoying, reserve a static external IP and assign it to the node
(`gcloud compute addresses create`), which pins the value across replacements — at the
cost of a small charge for the reserved address when it is not attached to a running
instance.

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

The deploy workflows authenticate to GCP with **Workload Identity Federation** — no
service-account key is created, downloaded, or stored. Each workflow run mints a
short-lived OIDC token from GitHub and exchanges it for GCP credentials, so there is
no long-lived secret to leak, rotate, or accidentally commit.

Create the service account, bind its roles, and wire up the identity pool with the
one-time bootstrap script (requires an account with `roles/iam.serviceAccountAdmin` +
`roles/resourcemanager.projectIamAdmin`, e.g. the project owner):

```bash
cd chat-infra/gcp/bootstrap
./create-deploy-sa.sh
```

The script is idempotent — re-running it skips anything that already exists. It prints
the exact `gh secret set` commands (or UI values) for the three secrets it produces:

| Secret | Value |
|---|---|
| `GCP_PROJECT_ID` | Your GCP project ID |
| `GCP_SA_EMAIL` | `github-actions@<PROJECT_ID>.iam.gserviceaccount.com` — holds **Artifact Registry Writer** + **GKE Developer** |
| `GCP_WIF_PROVIDER` | Full provider resource path, `projects/<NUMBER>/locations/global/workloadIdentityPools/github/providers/github` |

**How access is restricted.** The pool trusts GitHub's OIDC issuer, which by itself
would let *any* repo on GitHub request a token. Two things narrow it to yours: the
provider's `--attribute-condition` pins `assertion.repository_owner`, and the
`roles/iam.workloadIdentityUser` binding uses a `principalSet://` scoped to the exact
`owner/repo`. Both must match before a token is honoured. The script derives the repo
from your `origin` remote — override with `GITHUB_REPO=owner/name ./create-deploy-sa.sh`
if that's wrong (e.g. on a fork).

Each workflow declares `permissions: id-token: write` so the runner may request the
OIDC token, plus `contents: read` for `actions/checkout` — declaring `permissions`
explicitly drops everything not listed, so both are required.

There is one workflow per service under `.github/workflows/deploy-<service>.yml`. Each workflow triggers on push to `main` when its own service directory, its Helm values file, or the shared microservice chart changes. Use `workflow_dispatch` (Actions → select workflow → Run workflow) for the first-time bootstrap or to force a redeploy without a code change.

## Architecture decisions

See [`docs/adr/0005-gke-standard-over-autopilot.md`](../../docs/adr/0005-gke-standard-over-autopilot.md) for why GKE Standard was chosen over Autopilot.
