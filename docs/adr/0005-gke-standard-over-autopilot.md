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

## Update (2026-07-27): node resized to `e2-standard-2`

The decision above — GKE Standard over Autopilot — stands unchanged. The **machine
type** in it did not survive first contact.

`e2-medium` could not schedule the stack. It is a **shared-core burstable** type: 2 vCPU
of *burst* on a 1 vCPU *baseline*, and GKE derives allocatable from the baseline, so the
node offered **940m** rather than the ~1930m its "2 vCPU" label implies. GKE's own system
pods request ~753m of that (`kube-dns` alone is 270m), leaving ~187m for a stack
requesting 850m. `redpanda` (250m, the largest single request) sat `Pending` with
`0/1 nodes are available: 1 Insufficient cpu`.

Memory was never the constraint — it peaked at 43%. The original sizing was done against
a memory budget only, which is why this surfaced at deploy time rather than design time.

Now a single **`e2-standard-2`** (2 dedicated vCPU, 8GB): ~1930m allocatable, ~1180m free
after system overhead, ~330m of headroom. Still Spot, still one node, still one zonal
cluster, and boot disk unchanged at 20GB so the 30GB always-free disk layout holds.

Reason #4 above ("full control over node sizing") is what made this a one-line fix —
under Autopilot the per-pod minimums would have driven the same problem into the billing
model instead.

Superseded figures: reason #3's "~$12/mo" and "$300 new-account credit" were never
verified and do not match the billing account in use (CAD). See the cost model in
[`chat-infra/gcp/README.md`](../../chat-infra/gcp/README.md).
