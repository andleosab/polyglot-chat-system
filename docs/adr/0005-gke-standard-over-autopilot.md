# ADR-0005: GKE Standard over Autopilot for demo cluster

## Status
Accepted

## Context
The polyglot chat system needs a GKE cluster for demo purposes. GKE offers two cluster modes:
- **Autopilot** — Google manages nodes automatically; billing is per-pod based on requested CPU/memory.
- **Standard** — operator manages node pools; billing is per-node VM.

## Decision
Use **GKE Standard** (zonal, `us-west1-a`) with a single `e2-medium` Spot node.

## Reasons
1. **Autopilot per-pod minimums cost $30–50/mo at baseline**, before any real workload. Each pod has a minimum resource request enforced by GKE, making it expensive for a demo with 9 pods (5 services + Redis + Redpanda + Postgres + Nginx).
2. **GKE Standard's cluster management fee ($74.40/mo) is fully covered** by Google's GKE cluster credit — one free zonal Standard cluster per billing account. The only real cost is the node VM.
3. **A single `e2-medium` Spot instance costs ~$12/mo**, absorbed by the $300 new-account credit. This is the entire node cost.
4. **Full control over node sizing** — we can pack all workloads onto one node and cap memory via `JAVA_TOOL_OPTIONS` and Redpanda startup flags in a way that Autopilot's minimum enforcement would prevent.

## Consequences
- The cluster is intentionally ephemeral: `terraform apply` to start, `terraform destroy` when idle.
- Spot instances can be preempted by GCP. Pods restart automatically; persistent data is on separate `pd-standard` disks that survive node replacement.
- Autopilot should be reconsidered if the system ever needs multi-node scale or production SLAs, as it eliminates node management overhead at the cost of higher per-workload pricing.
