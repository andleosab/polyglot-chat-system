# ADR-0006: nginx NodePort vs Cloudflare Tunnel for GKE demo ingress

## Status
Accepted. Revised 2026-07-30 — the original decision (keep NodePort) is reversed, and the
migration path it sketched is superseded by the one below. See "What changed on revision".

## Context
The GKE demo cluster exposed the application via nginx on NodePort 30080, reachable at the node's external IP. The node runs as a Spot instance. When GCP preempts the node and replaces it, the external IP changes — `BETTER_AUTH_URL` in `chat-web` becomes stale and auth breaks until the value is updated and redeployed.

Two alternatives to that setup were evaluated:

**Option A — GCP LoadBalancer + static IP:** Change the nginx Service to `type: LoadBalancer` and reserve a GCP static IP. The IP survives node replacement. Cost: ~$18/month for the forwarding rule alone, roughly doubling the demo cluster cost. Rejected.

**Option B — Cloudflare Tunnel:** Run a `cloudflared` Deployment that dials out to Cloudflare's edge. The tunnel provides a stable public URL — no IP involved, and no inbound firewall rule. Requires a domain managed on Cloudflare. Cost: zero beyond domain registration ([Cloudflare Registrar](https://developers.cloudflare.com/registrar/) sells at-cost).

## Decision
Adopt Cloudflare Tunnel (option B). **nginx is retained**, as a ClusterIP Service, with
cloudflared pointing a single catch-all rule at it.

Retaining nginx is the part worth recording, because the obvious reading of option B is
that cloudflared replaces it — cloudflared can do path-based routing itself. Two reasons
not to:

1. **nginx is the origin-unifier, not just the NodePort holder.** The browser needs `/`
   and the `/chat/` WebSocket on a single origin: `BETTER_AUTH_URL` is single-valued, and
   `ws.ts` derives the WebSocket URL from `location.host` in production. nginx is what
   joins those two upstreams. Delete it and `kubectl port-forward` stops being a usable
   way into the cluster, because there is no longer any single port that serves both.
2. **The routing rules already exist and are shared with local compose.** Moving them into
   cloudflared's config creates a second, independently-maintained copy of the same
   routing for GKE only.

So cloudflared is a dumb pipe to one upstream, and nginx keeps owning the `/` vs `/chat/`
split for both GKE and compose.

The tunnel is **remotely-managed** (token-based): it is created in the Cloudflare
dashboard, and cloudflared pulls its ingress rules from there. A locally-managed tunnel
with a `config.yaml` ConfigMap was rejected because, with nginx retained, that ConfigMap
would encode exactly one hardcoded upstream — the real routing is in nginx's ConfigMap,
which is already in version control.

Cloudflare handles the WebSocket upgrade for `/chat/` natively; no `Upgrade` /
`Connection` headers need to be set at the edge.

## Consequences
- `BETTER_AUTH_URL` and `ORIGIN` are set once to the tunnel hostname and survive Spot node
  replacement. The recovery step that re-baked the node IP into `chat-web` is gone.
- The cluster has **no inbound firewall rule at all**. cloudflared dials outbound; the
  NodePort rule in `vpc.tf` is deleted rather than replaced.
- Google OAuth becomes viable — it needs a redirect URI registered in advance, which a
  stable https origin now provides. Still opt-in via `google_client_id` /
  `google_client_secret`.
- The hostname → `nginx:80` mapping lives in the Cloudflare dashboard, not in this repo.
  That is one manual step in the first-run sequence that `terraform apply` cannot cover.
- A Cloudflare account and domain are now a hard dependency for external access. Until one
  exists, `tunnel_token` stays empty, `deploy-infra-services.sh` skips the cloudflared
  release, and `kubectl port-forward svc/nginx 8080:80 -n chat` is the only way in — which
  is a supported state, not a broken one.

## What changed on revision

The original decision was to keep NodePort and accept the IP fragility, on the grounds
that the recovery path was cheap. The sketched migration path assumed nginx would be
deleted and cloudflared would take over path routing via a ConfigMap.

Both parts changed:

- **Deferred → adopted.** The deferral was conditional on having no domain. That condition
  is being resolved.
- **nginx deleted → nginx retained.** The original scope treated nginx as a NodePort
  holder. It is also the origin-unifier (see Decision), which the original did not account
  for.
- **ConfigMap → dashboard-managed token.** Follows from retaining nginx: with one upstream
  there is nothing left for a local ingress config to express.

## Verification status

The nginx-to-ClusterIP change, the firewall removal, and the port-forward path are
exercised. **The cloudflared leg is untested** — it has not run against a real tunnel,
because no Cloudflare account exists yet. Treat the chart and the dashboard steps as
unverified until a tunnel is live.
