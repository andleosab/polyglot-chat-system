# ADR-0001: Read receipts belong in chat-delivery-service, not chat-presence-service

## Status
Accepted

## Context
When adding the presence service, read receipts were initially considered as a presence feature (a user "has seen" a message is related to their activity state). The presence service tracks online/offline status and last seen.

## Decision
Read receipts are implemented in `chat-delivery-service` (Quarkus), not in `chat-presence-service`.

## Reasons
1. Read receipts require fan-out — when user A reads a message, all other participants in the conversation need to be notified. `chat-delivery-service` already owns the WebSocket fan-out facility.
2. Placing read receipts in presence would require the presence service to know which users are in a conversation, violating its focused scope (it has no knowledge of conversations or membership).

## Consequences
- `chat-delivery-service` handles the full lifecycle: receive read event → fan out over WebSocket to conversation participants.
- `chat-presence-service` stays conversation-agnostic.
