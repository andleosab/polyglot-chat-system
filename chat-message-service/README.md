# chat-message-service

Go service that owns all persistent chat data: conversations (private and group), participants, and message history. Also acts as a Kafka consumer (persists inbound messages from the delivery layer) and producer (publishes participant lifecycle events).

## Responsibilities

- CRUD for conversations, participants, and messages
- Cursor-based message pagination with `joined_at` visibility guard (users only see messages sent after they joined)
- Private conversation deduplication — creating a private conversation between two users is idempotent
- Kafka consumer: persists `chat.messages` events from `chat-delivery-service`
- Kafka producer: publishes `chat.participants.created` / `chat.participants.removed` events so connected users' group membership is updated in real time

## Stack

| | |
|---|---|
| Runtime | Go 1.25 |
| HTTP router | `go-chi/chi` v5 |
| Database driver | `jackc/pgx` v5 (native, no `database/sql` wrapper) |
| Query layer | [SQLC](https://sqlc.dev/) — compile-time type-safe SQL → Go |
| Migrations | `golang-migrate` (Makefile targets) |
| Kafka client | `twmb/franz-go` v1 |
| JWT | `golang-jwt/jwt` v5 |

## API

Base path: `/messaging` (all routes require Bearer JWT)

### Conversations

| Method | Path | Notes |
|---|---|---|
| `POST` | `/conversations/` | Create private or group conversation. Private creation is transactional and idempotent — returns existing conversation if one already exists between the two participants |
| `GET` | `/conversations/` | List conversations for a user (`?id=<uuid>`). Uses a lateral join to resolve display names and last-message timestamps |
| `GET` | `/conversations/ids` | Return conversation IDs only (`?id=<uuid>`). Optimised endpoint called by `chat-delivery` on WebSocket connect |
| `GET` | `/conversations/{id}` | Get by ID |
| `PUT` | `/conversations/{id}` | Update name (group only) |
| `GET` | `/conversations/private` | Lookup private conversation ID between two users (`?user1=&user2=`) |

### Groups / Participants

| Method | Path | Notes |
|---|---|---|
| `GET` | `/groups/` | List groups for a user (`?id=<uuid>`) |
| `POST` | `/groups/{id}/participants` | Add participant. Publishes `chat.participants.created` event |
| `GET` | `/groups/{id}/participants` | List members |
| `DELETE` | `/groups/{id}/participants/{userId}` | Remove participant. Publishes `chat.participants.removed` event |

### Messages

| Method | Path | Notes |
|---|---|---|
| `POST` | `/conversations/{id}/messages/` | Persist a message directly via HTTP |
| `GET` | `/conversations/{id}/messages/` | Offset-based, limit 100 |
| `GET` | `/conversations/{id}/messages/cursor` | Keyset pagination (`?mid=<last_id>&limit=20`). `mid` defaults to `MaxInt64` to return the latest page |

## Schema

```sql
conversations (id, name, type ENUM('private','group'), created_by UUID, last_message_at)
  -- CHECK: name IS NULL for private, NOT NULL for group

participants  (id, conversation_id FK, user_uuid UUID, username, is_admin, joined_at)
  -- UNIQUE(conversation_id, user_uuid)

messages      (id, conversation_id FK, sender_user_uuid UUID, content, sent_at, edited_at, is_deleted, client_ts BIGINT)
```

`client_ts` (client epoch ms) threads through Kafka so message ordering is preserved across delivery latency. Messages are soft-deleted (`is_deleted = true`).

Key indexes: `(conversation_id, sent_at DESC)` for message retrieval, `(conversation_id, id DESC)` for lateral join optimisation, `(conversation_id, user_uuid)` composite for the most-used join path.

## Kafka integration

**Consumer** (`chat.messages` topic, manual commit, batch up to 100):

```
chat-delivery → Kafka [chat.messages] → Consumer → PostgreSQL
```

Poison-pill handling: unmarshalable records are committed and skipped rather than blocking the partition.

**Producer** (`AllISRAcks`, 10s delivery timeout):

```
POST /groups/:id/participants → AddParticipant() → [chat.participants.created]
DELETE /groups/:id/participants/:userId → RemoveParticipant() → [chat.participants.removed]
```

Participant events are published synchronously after a successful DB write. The Kafka record key is `conversationId` (string), ensuring ordering per conversation.

## JWT verification

Validates HMAC-SHA256 tokens. Accepts multiple issuers (comma-separated `JWT_ISSUER`) so both `chat-web` and `chat-delivery-service` (which mints short-lived identity-propagation tokens) can call this service.

## Environment variables

| Variable | Default | Notes |
|---|---|---|
| `DATABASE_URL` | `postgres://...` | Full connection string |
| `HTTP_PORT` | `:8080` | |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated |
| `KAFKA_CONSUMER_TOPIC` | `chat.messages` | |
| `KAFKA_CONSUMER_GROUP` | `chat-message-service` | |
| `KAFKA_PRODUCER_TOPIC_PARTICIPANT_CREATED` | `chat.participants.created` | |
| `KAFKA_PRODUCER_TOPIC_PARTICIPANT_REMOVED` | `chat.participants.removed` | |
| `JWT_SECRET` | — | |
| `JWT_ISSUER` | — | Comma-separated, e.g. `chat-delivery-service,chat-web` |
| `JWT_AUDIENCE` | `chat-message-service` | |

## Running

```bash
# prerequisites: postgres + kafka running

# run migrations
make migrate-up

# build and run
make build && make run

# generate SQLC (after changing SQL queries)
make sqlc

# docker
make docker-build
make compose-up
```

The `Makefile` uses `Dockerfile.scratch` by default (minimal image, binary only). Alternative Dockerfiles for JVM-style (`Dockerfile.jvm`-style via distroless) are in `docker/`.

## gRPC

`internal/grpc/` contains scaffolding and a `user.proto` definition. Not yet implemented. The `GET /conversations/ids` endpoint is the current substitute for intra-cluster calls from `chat-delivery-service` — the plan is to replace it with a gRPC call.
