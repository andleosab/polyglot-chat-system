# ADR-0006: nginx NodePort vs Cloudflare Tunnel for GKE demo ingress

## Status
Accepted

## Context
The GKE demo cluster exposes the application via nginx on NodePort 30080, reachable at the node's external IP. The node runs as a Spot instance. When GCP preempts the node and replaces it, the external IP changes — `BETTER_AUTH_URL` in `chat-web` becomes stale and auth breaks until the value is updated and redeployed (README first-run step 8).

Two alternatives to the current setup were evaluated:

**Option A — GCP LoadBalancer + static IP:** Change the nginx Service to `type: LoadBalancer` and reserve a GCP static IP. The IP survives node replacement. Cost: ~$18/month for the forwarding rule alone, roughly doubling the demo cluster cost. Rejected.

**Option B — Cloudflare Tunnel:** Replace nginx entirely with a `cloudflared` Deployment. The tunnel provides a stable public URL backed by Cloudflare's edge — no IP involved. Requires a domain managed on Cloudflare. Cost: zero beyond domain registration. Domain can be registered directly at [Cloudflare Registrar](https://developers.cloudflare.com/registrar/) at-cost (400+ TLDs, auto-configured on Cloudflare nameservers). Deferred — no domain currently in place.

## Decision
Keep nginx + NodePort 30080. Accept the IP fragility at demo scale. The recovery path (README step 8 + automatic CD) is fast enough that the cost of fixing it is low.

## Migration path to Cloudflare Tunnel

When a Cloudflare account and domain are available, nginx can be removed entirely. `cloudflared` does path-based routing via its ingress config, matching what nginx currently does:

```yaml
ingress:
  - hostname: <domain>
    path: /chat/
    service: http://chat-delivery-service.chat.svc.cluster.local:8080
  - hostname: <domain>
    service: http://chat-web.chat.svc.cluster.local:3000
  - service: http_status:404
```

Cloudflare handles the WebSocket upgrade for `/chat/` natively — no `Upgrade` / `Connection` headers need to be set manually.

Implementation scope (~9 file changes):
- Remove `chat-infra/gcp/helm/nginx/`
- Add `chat-infra/gcp/helm/cloudflared/` (Deployment + ConfigMap with ingress rules above)
- Update `chat-infra/gcp/bootstrap/deploy-infra-services.sh` (deploy cloudflared, not nginx)
- Add `tunnel_token` to `chat-infra/gcp/terraform/secrets.tf` (K8s Secret `chat-secrets`)
- Add `tunnel_token` variable to `terraform/variables.tf` and `terraform.tfvars.example`
- Remove the NodePort firewall rule from `terraform/vpc.tf`
- Replace `BETTER_AUTH_URL` placeholder in `helm/values/chat-web.yaml` with the stable tunnel URL
- Update first-run steps in `chat-infra/gcp/README.md`

Once done, `BETTER_AUTH_URL` is set once and never needs updating on Spot node replacement.

## Consequences
- NodePort fragility is intentional and accepted at demo scale.
- Revisit this ADR if Spot preemption becomes frequent or a stable public URL is required (e.g., for a recorded demo or external reviewer access).
