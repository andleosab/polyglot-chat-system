# ADR-0004: All delivery→presence signals go via gRPC — no Kafka topic

## Status
Accepted

## Context
The original presence design proposed a `chat.connections` Kafka topic for `chat-delivery-service` to signal `connected`, `disconnected`, and `heartbeat` events to `chat-presence-service`. Three signals need to flow from delivery to presence:

- `connected` — user opened a WebSocket connection
- `disconnected` — user closed the browser or connection dropped
- `heartbeat` — user's connection is still alive (~every 20s), resets the 30s online TTL

## Decision
All three signals are sent via gRPC direct calls from `chat-delivery-service` to `chat-presence-service`. The `chat.connections` Kafka topic is not created.

## Reasons
Kafka durability for these events provides no meaningful benefit:
- Missed `connected`: the next heartbeat (≤20s) catches it.
- Missed `disconnected`: the 30s TTL on `user:{useruuid}:status` expires and the user appears offline automatically.

gRPC is a better fit: lower latency, single transport, single `.proto` definition, no new Kafka topic or consumer to operate.

Auth/TLS is not applied for now — these are plain internal cluster calls. The cluster network boundary provides sufficient isolation for the first iteration.

## Consequences
- `chat.connections` Kafka topic is never created.
- `chat-presence-service` exposes a gRPC server with a single service covering `Connected`, `Disconnected`, and `Heartbeat` RPCs.
- `chat-delivery-service` gains a gRPC client stub for presence.
- Presence service resilience relies on the Redis TTL + heartbeat cycle rather than Kafka durability.
