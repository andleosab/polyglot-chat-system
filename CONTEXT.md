# Context

Polyglot distributed chat system. Domain language for this codebase.

---

## Glossary

### Conversation
The canonical entity for a chat thread. Two types: `private` (exactly two participants, no name) and `group` (named, two or more participants). The UI surfaces these as "chats" (private) and "groups" (group) but both map to the same underlying `conversations` table. **"Chat" and "conversation" are interchangeable in the UI; `conversationId` is the canonical internal key.**

Avoid: "room", "channel", "thread".

### Participant
A user's membership in a conversation. Carries `user_uuid`, `username`, `is_admin`, and `joined_at`. A user can only be a participant once per conversation.

Avoid: "member" (use participant).

### User
A person using the system. Identified by `user_uuid` (UUID, the cross-service key) and `username` (human-readable). The full user profile (email, active status, timestamps) lives in `chat-user-service`.

Avoid: "userId" when referring to the UUID key — prefer `user_uuid`.

### Message
A piece of content sent within a conversation. Carries `conversation_id`, `sender_uuid`, `content`, and `sent_at`. In Kafka and WebSocket payloads the fields are named `from`, `fromName`, and `message` — these are transport aliases, not distinct domain concepts.

### Presence
Whether a user is currently online and when they were last seen. The presence service owns exactly two facts per user: **online/offline status** (volatile, TTL-based) and **last seen** timestamp (durable). It also handles **typing indicators**. It has no knowledge of conversations or group membership — those belong to `chat-message-service`. When a UI needs group members with their online status, the BFF composes that view by calling both services independently.

Online/offline state is driven by gRPC calls from `chat-delivery-service` — `Connected`, `Disconnected`, and `Heartbeat` RPCs defined in `proto/presence.proto` at the repo root. No Kafka topic is used for presence signals. Disconnect is a global user-level event — the user's WebSocket connection dropped entirely, not that they left a specific conversation.

Avoid: "room presence", "room members" (presence service does not own membership).

### Online Status
Stored in Redis as `user:{useruuid}:status` (string `"online"`, 30s TTL). Reset by each `Heartbeat` gRPC call. Expires automatically if heartbeats stop.

### Last Seen
The timestamp of a user's most recent `Disconnected` event. Redis (`user:{useruuid}:last_seen`) is a cache; Postgres (`user_last_seen` table) is the source of truth. On a cache miss, read from Postgres. On `Disconnected`, write both atomically.

### Heartbeat
A periodic signal (~every 20s) sent by `chat-delivery-service` to `chat-presence-service` over gRPC to indicate a user's WebSocket connection is still alive. Resets the 30s TTL on online status in Redis. If heartbeats stop (crash, network drop), the TTL expires and the user appears offline automatically.

### Typing Indicator
An ephemeral signal that a user is currently composing a message in a conversation. Flow: frontend sends a debounced `POST /presence/typing` → presence service publishes to Redis pub/sub `conversation:{conversationId}:typing` → delivery service fans out over WebSocket to connected participants. Not persisted. Receiving clients display "is typing..." and auto-clear after 3–4 seconds if no new signal arrives.

### Read Receipt
A signal that a user has read a message. Owned by `chat-delivery-service`, not the presence service — delivery already has the WebSocket fan-out facility needed to broadcast receipts to all conversation participants. Placing it in presence would require presence to know about conversations, violating its focused scope.

### Redis
Introduced with the presence service — the first Redis dependency in the system. The presence service is the sole writer (online status, last seen cache, typing pub/sub). The delivery service subscribes to the typing pub/sub channel to fan out typing indicators over WebSocket. No other service uses Redis.
