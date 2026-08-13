# GKE Demo Cluster

GCP infrastructure for the polyglot chat demo. Designed to stay within the GCP always-free tier.

## Deployment model

| Layer | What | How |
|---|---|---|
| Billing budget | Budget alert | `terraform apply` in `terraform/budget/` — **applied once, never destroyed** |
| GCP resources | Cluster, VPC, disks, namespace, K8s secrets | `terraform apply` in `terraform/cluster/` |
| Infra services | Postgres, Redis, Redpanda, Nginx, cloudflared | `bootstrap/deploy-infra-services.sh` (run once after apply) |
| Microservices | 5 app services | GitHub Actions on push to `main` |

### Two Terraform root modules

`terraform/cluster/` and `terraform/budget/` are independent root modules sharing one
state bucket under different prefixes (`chat-demo/state` and `chat-demo/budget`). The
split exists so that `terraform destroy` on the cluster — which you will run often, since
the node is the only real cost — does not take the budget alert with it.

The boundary is **lifetime, not subject matter**. Everything else the cluster module owns
is deliberately destroyed with it, including the Postgres and Redpanda disks and the
Artifact Registry repository. Recreating means empty databases and a full image rebuild
(dispatch the five deploy workflows). That is the intended behaviour while the schema is
still in flux — a clean slate on recreate, not a cost saving. The two disks and the
registry would cost well under a dollar a month to keep.

## What's here

```
bootstrap/
  create-state-bucket.sh      # one-time GCS state bucket setup — run before terraform init
  create-deploy-sa.sh         # one-time: GitHub Actions deploy SA + roles + JSON key
  deploy-infra-services.sh    # deploy Postgres, Redis, Redpanda, Nginx, cloudflared via Helm

terraform/
  cluster/                    # created and destroyed freely — state prefix chat-demo/state
    providers.tf              # GCS backend, google/kubernetes providers
    variables.tf              # input variables (secrets via terraform.tfvars)
    gke.tf                    # GKE Standard cluster + e2-standard-2 Spot node pool + Artifact Registry
    vpc.tf                    # VPC + subnet (no ingress firewall rule — cloudflared dials out)
    disks.tf                  # pre-provisioned pd-standard disks: Postgres (5GB) + Redpanda (5GB)
    namespaces.tf             # chat namespace + shared K8s Secret (chat-secrets)
    helm_releases.tf          # intentionally empty — see file for explanation
    terraform.tfvars.example  # template — copy to terraform.tfvars and fill in secrets
  budget/                     # applied once, outlives cluster teardown — prefix chat-demo/budget
    providers.tf              # GCS backend, google provider (with the billing quota-project override)
    variables.tf              # project_id + billing_account
    budget.tf                 # 1-unit budget alert in the billing account's currency (fires while trial credit drains)
    terraform.tfvars.example  # template

helm/
  microservice/               # generic chart reused by all 5 microservices
  nginx/                      # Nginx reverse proxy, ClusterIP (reuses existing nginx.conf routing rules)
  cloudflared/                # Cloudflare Tunnel connector — the only external ingress
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
| e2-standard-2 Spot node | ~$15/mo (approximate — check the billing console) — absorbed by the trial credit |
| 30GB pd-standard disks (20GB boot + 5GB Postgres + 5GB Redpanda) | $0 — always-free cap |
| GCS state bucket | $0 — state file is a few KB, well under 5GB always-free cap |
| Artifact Registry | $0 — under 0.5GB/mo free tier (five images, two of them JVM, may exceed it; overage is ~$0.10/GB/mo) |
| Cloudflare Tunnel | $0 — the tunnel itself is free; only the domain costs anything (Cloudflare Registrar sells at cost) |
| Budget alert | 1-unit threshold in the billing account's currency, `EXCLUDE_ALL_CREDITS=true` so it fires during trial period |

## Resource budget

Every workload shares one 8GB `e2-standard-2`. **Budget both CPU and memory** — the
scheduler refuses a pod if *either* request can't be met, and CPU is the tighter of the
two here.

| Pod | CPU request | Memory request | Memory limit |
|---|---|---|---|
| chat-web | 100m | 128Mi | 450Mi |
| chat-user-service | 100m | 256Mi | 350Mi |
| chat-message-service | 50m | 64Mi | 128Mi |
| chat-delivery-service | 100m | 256Mi | 350Mi |
| chat-presence-service | 50m | 128Mi | 256Mi |
| postgres | 100m | 128Mi | 300Mi |
| redpanda | 250m | 700Mi | 700Mi |
| redis | 50m | 32Mi | 64Mi |
| nginx | 50m | 32Mi | 64Mi |
| cloudflared (×2) | 40m | 64Mi | 256Mi |
| **Total** | **890m** | **~1.8GB** | **~2.9GB** |

### Budget against allocatable minus system overhead

Never size against the machine's advertised specs. Two deductions come first:

| | CPU | Memory |
|---|---|---|
| Machine | 2000m | 8GB |
| Node allocatable (after GKE's reserve) | ~1930m | ~6.1GB |
| GKE system pods (measured) | ~753m | ~1.04GB |
| **Left for this stack** | **~1180m** | **~5GB** |
| Stack requests | 890m | ~1.8GB |

The ~753m of system overhead is roughly **fixed regardless of workload** and dominated
by `kube-dns` (270m), which no application can do without. Logging and monitoring
agents (`fluentbit` 105m, `kube-state-metrics` 105m, `metrics-server` 43m,
`gke-metrics-agent` 21m) account for most of the rest. Verify on a live node with:

```bash
kubectl describe node | grep -A30 "Non-terminated Pods"
```

> ### ⚠️ Do not switch to `e2-medium` to save money
>
> It looks like a 2 vCPU machine and is advertised as one, but it is **shared-core
> burstable**: 2 vCPU of *burst* on a 1 vCPU *baseline*, and GKE derives allocatable
> from the baseline. A GKE `e2-medium` offers **940m**, not ~1930m — and after the
> ~753m of system overhead, only ~187m remains for a stack that requests 850m.
>
> This is not theoretical. The cluster ran on `e2-medium` initially and `redpanda`
> (250m, the largest single request) sat `Pending` with `0/1 nodes are available:
> 1 Insufficient cpu` while every pod already placed looked perfectly healthy. Memory
> was never the constraint — it sat at 43%.

CPU requests gate scheduling and set relative shares under contention; they do not cap
usage (limits do), so the values above are reservations rather than ceilings. JVM
services (`chat-user-service`, `chat-delivery-service`) additionally bound heap via
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

On a brand-new GCP project, `terraform apply` will fail unless these are done first.
Both affect the **budget** module; the cluster module applies without them.

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

> Since the split, a missing Billing Budgets API no longer produces a *partial* apply
> of the cluster — the two modules fail independently. That is the main day-to-day
> benefit of the split beyond surviving `destroy`.

> **The budget also needs a quota project.** `billingbudgets.googleapis.com` bills
> API quota to a project via the `X-Goog-User-Project` header. Under user ADC that
> header is unset by default, so the call falls back to a Google-owned project and
> fails with a misleading `SERVICE_DISABLED` 403 (the error's `consumer` field
> points at a project number that isn't yours). `budget/providers.tf` sets
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

**3. Create the Cloudflare tunnel — optional, and skippable on a first pass.** External
access runs through Cloudflare Tunnel; the token has to be in `terraform.tfvars` before
the cluster apply, so the tunnel is created *before* the cluster exists. That is fine —
Cloudflare does not check the origin until traffic flows.

In the Cloudflare dashboard: **Zero Trust → Networks → Tunnels → Create a tunnel →
Cloudflared**. Name it, copy the token, and add a public hostname:

| Field | Value |
|---|---|
| Subdomain / domain | e.g. `chat` / `alslabs.dev` |
| Service type | HTTP |
| URL | `nginx.chat.svc.cluster.local:80` |

One catch-all hostname is all that is needed — nginx owns the `/` vs `/chat/` split, and
Cloudflare upgrades the `/chat/` WebSocket natively. Cloudflare creates the DNS record for
you. Put the token in `tunnel_token` in the cluster module's `terraform.tfvars`.

**Skip this entirely if you have no Cloudflare account yet.** Leave `tunnel_token` empty:
`deploy-infra-services.sh` skips the cloudflared release, the cluster comes up with no
external ingress, and `kubectl port-forward` is the way in (step 9 below). Everything
except the tunnel leg is testable in that state.

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

# 3. Apply the budget — separate root module, own state prefix. Run this ONCE; it is
#    deliberately not touched by cluster teardown, so you never repeat it.
cd ../terraform/budget
terraform init -backend-config="bucket=<YOUR_PROJECT_ID>-tf-state"
cat > terraform.tfvars <<EOF
project_id      = "<YOUR_PROJECT_ID>"
billing_account = "<XXXXXX-XXXXXX-XXXXXX>"
EOF
terraform apply
#    If a budget named chat-demo-budget already exists (e.g. hand-created in the
#    console after an earlier destroy), delete it there first. GCP does not enforce
#    unique display names, so applying over it produces two budgets, not one.

# 4. Init the cluster module and create its terraform.tfvars (gitignored)
cd ../cluster
terraform init -backend-config="bucket=<YOUR_PROJECT_ID>-tf-state"
cat > terraform.tfvars <<EOF
project_id         = "<YOUR_PROJECT_ID>"
jwt_secret         = "<shared HMAC secret>"
jwt_jwk            = "<base64url-encoded JWT secret for Quarkus>"
postgres_password  = "<postgres superuser password>"
better_auth_secret = "<Better Auth session signing secret>"
tunnel_token       = "<Cloudflare tunnel token, or leave empty>"
EOF
#    No billing_account here — that belongs to the budget module.

# 5. Apply in two steps — kubernetes provider can't connect until the cluster exists
#    Step 5a: create cluster (Terraform auto-includes VPC + subnet as dependencies)
terraform apply -target=google_container_cluster.main

#    Step 5b: full apply — cluster now exists, kubernetes provider can connect
#             Creates: node pool, disks, Artifact Registry, chat namespace,
#                      chat-secrets K8s Secret. No firewall rule — cloudflared dials out.
terraform apply

# 6. Configure kubectl for the new cluster
gcloud container clusters get-credentials chat-demo --zone us-west1-a

# 7. Deploy infra services (Postgres, Redis, Redpanda, Nginx, cloudflared)
cd ../../bootstrap
POSTGRES_PASSWORD=<same as terraform.tfvars> ./deploy-infra-services.sh
#    cloudflared is skipped, with a message, if tunnel_token was left empty.

# 8. Create the GitHub Actions deploy service account + Workload Identity Federation
#    No key file is created — workflows authenticate with a short-lived OIDC token.
#    Safe to re-run; every step skips work that already exists.
./create-deploy-sa.sh
#    The script prints three ready-to-paste commands — run them to store the values
#    as GitHub Actions repo secrets (or add them by hand at:
#    repo Settings -> Secrets and variables -> Actions -> New repository secret):
#      gh secret set GCP_PROJECT_ID   --body "<printed value>"
#      gh secret set GCP_SA_EMAIL     --body "<printed value>"
#      gh secret set GCP_WIF_PROVIDER --body "<printed value>"
#    These live on the GitHub repo, not in GCP.
#    Do NOT skip ahead: the secrets must exist before STEP 9, not step 10. Step 9 edits
#    helm/values/chat-web.yaml, which is in the Deploy chat-web path filter, so pushing
#    it fires a deploy immediately. Without the secrets that run dies in seconds at
#    "Authenticate to GCP" with PROJECT_ID empty and:
#      the GitHub Action workflow must specify exactly one of
#      "workload_identity_provider" or "credentials_json"
#    which reads like a workflow bug but only means the secrets aren't set yet.

# 9. Point chat-web at the origin your browser will actually use
#    Two ways in — pick ONE and use it for both env vars below.
#
#    (a) port-forward — the only way in with no tunnel configured:
#            kubectl port-forward svc/nginx 8080:80 -n chat
#        origin = http://localhost:8080
#        Forward nginx, NOT chat-web: the browser needs / and the /chat/ WebSocket
#        on a single origin, and nginx is what joins them. Works against a ClusterIP
#        Service and needs no firewall rule — it tunnels via the API server.
#        This is the default already committed in helm/values/chat-web.yaml, so with
#        no tunnel there is nothing to edit here.
#
#    (b) Cloudflare tunnel — once cloudflared is deployed:
#        origin = https://<your tunnel hostname>, e.g. https://chat.alslabs.dev
#        Stable across Spot node replacement. https only — Cloudflare terminates TLS,
#        and a .dev domain is HSTS-preloaded, so browsers refuse plain http to it.
#
# Edit helm/values/chat-web.yaml — set BOTH BETTER_AUTH_URL and ORIGIN to the origin
# you picked. They are runtime values and must match the browser's origin; a mismatch
# shows up as auth misbehaving rather than a clean error.
# The WebSocket needs no configuration — in production its URL comes from
# location.host, so it follows whichever origin you pick.
# Commit and push the change to main — this triggers Deploy chat-web automatically.
# That push is also the FIRST REAL TEST of the WIF handshake from step 8; nothing
# verifies it statically. If it fails at "Authenticate to GCP" with a 403 rather than
# the "must specify exactly one of" message, the secrets are set but the trust is
# wrong: check the provider's --attribute-condition (repository_owner) and the
# principalSet:// binding (exact owner/repo). All five workflows already declare
# `permissions: id-token: write`, so that is not the cause. Nor is it a missing IAM
# role — those govern what the SA may do once impersonation has already succeeded.

# 10. First-time microservice bootstrap (path filters won't fire on a fresh cluster)
#    Go to Actions → select each "Deploy <service>" workflow → Run workflow
#    After first run, path-based triggers handle redeployment automatically
#    Note: chat-web already ran from step 9's push (its path filter covers
#    helm/values/chat-web.yaml), so dispatching it here runs it a second time.
#    Harmless — don't read the duplicate run as a failure.
#
#    If a run builds and pushes fine but then times out at `helm --wait`, check:
#      kubectl describe pod -n chat -l app=<service> | tail -20
#    "Failed to pull image ...: 403 Forbidden" means the NODE cannot read Artifact
#    Registry. Pushing and pulling are different identities: CI pushes as
#    github-actions@ (granted by create-deploy-sa.sh), the kubelet pulls as the node
#    pool's default Compute Engine SA. terraform grants that read access
#    (google_artifact_registry_repository_iam_member.node_pull in cluster/gke.tf), so this
#    only appears on clusters applied before that resource existed. Fix in place:
#      PROJECT_NUMBER=$(gcloud projects describe "$(gcloud config get-value project)" \
#        --format='value(projectNumber)')
#      gcloud artifacts repositories add-iam-policy-binding chat-demo \
#        --location=us-west1 \
#        --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
#        --role="roles/artifactregistry.reader"
#    Pods recover at the next pull backoff (~5m), or force it with
#      kubectl rollout restart deployment/<service> -n chat

# 11. Access the app — at the same origin you configured in step 9
kubectl port-forward svc/nginx 8080:80 -n chat && open http://localhost:8080   # if (a)
# open https://<your tunnel hostname>                                          # if (b)

# 12. Tear down when done. Run this from terraform/cluster — NOT from terraform/budget,
#     which is applied once and left alone.
#     This DELETES the Postgres and Redpanda disks and the Artifact Registry repository
#     along with everything else. Databases come back empty and all five images must be
#     rebuilt (step 10) on the next apply. The budget alert survives.
cd ../terraform/cluster
terraform destroy
```

## The node IP no longer matters

The node pool runs a **single Spot `e2-standard-2`** (`cluster/gke.tf`), chosen
deliberately for cost (see ADR-0005). Spot nodes are reclaimable: GCP can preempt the node
at any time, GKE replaces it, and the replacement comes back with a new external IP.

That used to break the app. The origin was `http://<node IP>:30080`, baked into
`BETTER_AUTH_URL` and `ORIGIN` in `helm/values/chat-web.yaml`, so a preemption left
`chat-web` pointing at an address that no longer existed — logins failed and the WebSocket
never connected, while every pod reported healthy.

**Nothing in the ingress path references the node IP any more.** cloudflared dials out to
Cloudflare's edge and the hostname is stable, so a preemption costs a pod restart and
nothing else. `kubectl port-forward` is likewise unaffected — it goes through the API
server, not the node's external address. See ADR-0006.

## Debugging without external exposure

**No Service in this cluster is externally reachable** — everything is ClusterIP,
including nginx, and there is no ingress firewall rule. External traffic arrives only
through the cloudflared tunnel. Use `kubectl port-forward` for direct access:

```bash
# Postgres
kubectl port-forward svc/postgres 5432:5432 -n chat

# Redpanda (Kafka)
kubectl port-forward svc/redpanda 9092:9092 -n chat

# Presence service HTTP
kubectl port-forward svc/chat-presence-service 8000:8000 -n chat

# The whole app — forward nginx, not chat-web (it joins / and the /chat/ WebSocket)
kubectl port-forward svc/nginx 8080:80 -n chat
```

Check the tunnel's own health with `kubectl logs -n chat -l app=cloudflared`. Its
readiness probe hits cloudflared's `/ready`, which returns 200 only while a connection to
Cloudflare's edge is registered — so `0/1 READY` means the tunnel leg is genuinely down,
not just that the process is slow to start.

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

- [`docs/adr/0005-gke-standard-over-autopilot.md`](../../docs/adr/0005-gke-standard-over-autopilot.md) — why GKE Standard over Autopilot.
- [`docs/adr/0006-nginx-nodeport-vs-cloudflare-tunnel.md`](../../docs/adr/0006-nginx-nodeport-vs-cloudflare-tunnel.md) — why Cloudflare Tunnel replaced NodePort, and why nginx was kept rather than folded into cloudflared's own routing.
