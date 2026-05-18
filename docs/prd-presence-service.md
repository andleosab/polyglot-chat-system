# PRD: chat-presence-service — online status, last seen, and typing indicators

## Problem Statement

Users of the chat system have no visibility into whether other users are currently online, when they were last active, or whether someone is currently typing a reply. This makes conversations feel unresponsive and leaves users uncertain about whether their messages have been seen or whether a reply is coming.

## Solution

Introduce a new `chat-presence-service` (FastAPI) that tracks each user's online/offline status and last seen timestamp. Typing indicators are also handled by this service. The delivery service (`chat-delivery-service`) drives presence state via gRPC and fans out typing signals to connected clients over WebSocket. The BFF (`chat-web`) proxies presence queries from the frontend and enriches group member lists with online status by composing responses from both the presence and message services.

## User Stories

1. As a chat user, I want to see a green indicator next to a user's name when they are online, so that I know they are available to chat.
2. As a chat user, I want to see when a user was last seen when they are offline, so that I have a sense of when they might reply.
3. As a chat user, I want the online indicator to disappear automatically when a user closes their browser or loses their connection, so that I am not misled about their availability.
4. As a chat user, I want the online indicator to reflect the current state in near real-time, not a stale snapshot, so that the information is trustworthy.
5. As a chat user, I want to see a typing bubble (three animated dots, as seen in iMessage) when another user is composing a reply in our conversation, so that I know a response is coming without distracting text.
6. As a chat user, I want the typing indicator to disappear automatically after a few seconds if the other user stops typing, so that it does not linger incorrectly.
7. As a chat user, I want the typing indicator to only appear for the conversation I am currently viewing, so that I am not distracted by activity in other conversations.
8. As a chat user, I want my online status to be updated immediately when I connect, so that other users see me as online as soon as I open the app.
9. As a chat user, I want my last seen timestamp to be recorded when I disconnect, so that other users can see when I was last active.
10. As a chat user, I want my online status to recover automatically if my connection drops and reconnects, so that presence reflects my actual state without manual intervention.
11. As a chat user, I want presence checks to be fast, so that online indicators load without noticeably delaying the UI.
12. As a developer, I want the presence service to be conversation-agnostic, so that it remains simple and focused and does not need to be updated when conversation features change.
13. As a developer, I want the presence service to degrade gracefully if it is temporarily unavailable, so that the rest of the chat system continues to function.

## Implementation Decisions

### New service: `chat-presence-service` (FastAPI + Python)

Standalone FastAPI microservice. Exposes HTTP endpoints for the BFF and a gRPC server for internal calls from the delivery service.

**Presence service module** — deep module encapsulating all online/offline state:
- `set_online(user_uuid)` — writes `user:{useruuid}:status = "online"` to Redis with 30s TTL
- `set_offline(user_uuid)` — deletes the status key, writes `last_seen` to Redis (cache) and Postgres (source of truth) atomically
- `get_presence(user_uuid)` — returns `{ online: bool, last_seen: ISO | null }`; on Redis cache miss for `last_seen`, reads from Postgres

**Typing service module** — deep module for typing indicators:
- `publish_typing(user_uuid, conversation_id)` — publishes to Redis pub/sub channel `conversation:{conversationId}:typing`

**gRPC server** — implements `PresenceService` with three unary RPCs: `Connected`, `Disconnected`, `Heartbeat`. All accept `ConnectionEvent { user_uuid }` and return `Ack { ok }`. No auth on gRPC — plain internal cluster calls.

**HTTP endpoints** (JWT required, `aud: chat-presence-service`):
- `GET /presence/{userUuid}` — returns `{ online: bool, last_seen: ISO | null }`
- `POST /presence/typing` — body `{ conversationId }`, user UUID extracted from JWT `sub`

**Proto contract** — defined in `proto/presence.proto` at the repo root. Consumed by both the FastAPI server (`grpcio`) and the Quarkus client (SmallRye gRPC).

**Persistence:**
- Redis: `user:{useruuid}:status` (string, 30s TTL), `user:{useruuid}:last_seen` (string, no TTL)
- Postgres: `user_last_seen(useruuid UUID PRIMARY KEY, last_seen TIMESTAMPTZ NOT NULL)`
- Redis is the cache; Postgres is the source of truth for `last_seen`

### Modifications to `chat-delivery-service` (Quarkus) — surgical

- `ChatSocket.java` — add gRPC `Connected` call in `@OnOpen`, `Disconnected` call in `@OnClose`. User UUID taken from JWT `sub` claim; fall back to `{username}` path param if unavailable.
- New `PresenceGrpcClient` — gRPC client stub wrapping the presence service connection.
- New `HeartbeatScheduler` — `@Scheduled` task (~every 20s) that iterates all active `WebSocketConnection` objects and fires a `Heartbeat` gRPC call per connected user.
- New Redis pub/sub subscriber — subscribes to `conversation:{conversationId}:typing` channels and fans out typing signals over WebSocket to connected participants in that conversation. Delivery service is subscriber-only; presence service is the sole Redis writer.

### Modifications to `chat-infra`

- docker-compose: add Redis container, `chat-presence` Postgres database (fourth DB on shared Postgres), `chat-presence-service` container.
- No Nginx changes — the `/chat/` Nginx route exists solely for WebSocket protocol upgrade. Presence is REST-only and is proxied by the BFF like all other services.

### Modifications to `chat-web` (BFF + frontend)

- New proxy route: `/presence/*` → `chat-presence-service`. Mint short-lived JWTs with `aud: chat-presence-service` following the existing service token pattern.
- Frontend: online status indicators on user and member list views. Typing indicator in conversation view (debounced `POST /presence/typing` on keystroke; auto-clear display after 3–4s).
- Group member list: BFF composes the enriched view — fetch members from `chat-message-service`, fetch presence for each `user_uuid` from `chat-presence-service`, merge and return.

### Read receipts — out of scope, tracked separately

Read receipts were considered for this service but belong in `chat-delivery-service` — that service already owns WebSocket fan-out, and placing receipts in presence would require presence to know about conversations. See ADR-0001.

## Testing Decisions

A good test verifies external behaviour — what the module returns or what state it produces — not how it achieves it internally. Tests should not assert on Redis key names, internal method calls, or Postgres query structure.

**Modules with unit tests:**

- `presence.py` service — test `get_presence` (online user, offline user, unknown user with no last_seen), `set_online` (status key written with correct TTL), `set_offline` (status deleted, last_seen written to both Redis and Postgres atomically). Use mocked Redis and asyncpg.
- `typing.py` service — test that `publish_typing` publishes to the correct Redis pub/sub channel with the correct payload. Use mocked Redis.

**Verification (UI smoke test — primary, human-executed):**

Build and start the full stack (`./docker-build-all.sh`, `./compose-up.sh`). Walk through:
1. Open the app — verify online indicator appears for the logged-in user as seen by another session.
2. Close the browser tab — verify the indicator disappears and last seen timestamp updates.
3. Open a group conversation — verify member list shows online/offline status correctly.
4. Start typing in a conversation — verify the other participant sees a typing bubble (three animated dots) that auto-clears after ~4s.

**API smoke test:**
- `GET /presence/{userUuid}` on a connected user returns `{ online: true, last_seen: null }`.
- `GET /presence/{userUuid}` on a disconnected user returns `{ online: false, last_seen: "<ISO timestamp>" }`.
- `POST /presence/typing` with a valid JWT and `{ conversationId }` returns 200.

## Out of Scope

- **Read receipts** — implemented in `chat-delivery-service` in a separate piece of work (ADR-0001).
- **Presence privacy / per-conversation visibility** — any authenticated user can query any other user's presence. Membership-scoped visibility deferred.
- **Automated UI tests (Playwright)** — manual browser testing is the approach for this iteration.
- **gRPC auth / mTLS** — plain internal cluster calls for now; auth can be layered on later.
- **Cross-service integration test harness** — out of scope; UI smoke test + unit tests are sufficient at this stage.
- **`chat-delivery-service` read receipt fan-out** — tracked separately.

## Further Notes

- ADR-0001: Read receipts in delivery, not presence.
- ADR-0002: Typing indicator publish in presence, fan-out in delivery.
- ADR-0003: Presence service uses a fixed Kafka consumer group.
- ADR-0004: All delivery→presence signals go via gRPC — no Kafka topic.
- Domain glossary: `CONTEXT.md` at the repo root defines canonical terms (Conversation, Participant, Presence, Heartbeat, Typing Indicator, etc.). Use those terms throughout implementation.
- Implementation principle: add new code where possible; modify existing code only when genuinely required, and keep changes surgical. The current system is fully tested and working.
