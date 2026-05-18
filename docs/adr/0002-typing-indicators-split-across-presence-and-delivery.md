# ADR-0002: Typing indicator publish in presence, fan-out in delivery

## Status
Accepted

## Context
Typing indicators require two steps: accepting the signal from the user and broadcasting it to other conversation participants. The presence service handles presence signals; the delivery service owns all WebSocket fan-out.

## Decision
Split the typing indicator across two services:
- `chat-presence-service` accepts `POST /presence/typing` and publishes to Redis pub/sub channel `conversation:{conversationId}:typing`.
- `chat-delivery-service` subscribes to that Redis channel and fans out the signal over WebSocket to connected participants.

## Reasons
Adding WebSocket infrastructure to the presence service would duplicate what Quarkus already owns. Presence is HTTP-only; delivery is the single WebSocket gateway. The Redis pub/sub channel is the natural handoff point between them.

## Consequences
- Presence service remains HTTP-only with no WebSocket dependency.
- Delivery gains a Redis pub/sub subscription in addition to its Kafka consumer.
- The fan-out path for typing indicators mirrors the intended fan-out path for read receipts (ADR-0001).
