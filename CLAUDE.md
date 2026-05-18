# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Polyglot distributed chat system: SvelteKit BFF + Spring Boot (user profiles) + Go (message persistence) + Quarkus (WebSocket/Kafka fan-out), connected by Kafka and a shared HMAC-SHA256 JWT secret. See [README.md](README.md) and [docs/architecture.svg](docs/architecture.svg) for the full picture.

## Stack per service

| Directory | Stack | Role |
|---|---|---|
| `chat-web` | SvelteKit 5, Better Auth | UI + BFF; issues JWTs, proxies all API calls |
| `chat-user-service` | Spring Boot 3, JPA | Chat user profile store |
| `chat-message-service` | Go, Chi, SQLC | Conversations, participants, message history |
| `chat-delivery-service` | Quarkus 3, SmallRye | WebSocket gateway + Kafka fan-out |
| `chat-infra` | Docker Compose, Nginx, Redpanda | Local infrastructure |

Each service has its own `CLAUDE.md` with stack-specific guidance.

## Full-stack commands

```bash
./docker-build-all.sh   # build all Docker images
./compose-up.sh         # start infra + services (polls pg/kafka health first)
./compose-down.sh       # tear down everything
```

Access at `http://localhost` (Nginx on :80).

## Cross-cutting architecture

**Send path:** Browser → WS → `chat-delivery` → Kafka `chat.messages` → (fan-out to WS clients) + (persist via `chat-message`)

**Participant events:** `chat-message` publishes to `chat.participants.created` / `chat.participants.removed`; any `chat-delivery` pod with the affected user connected updates their in-memory group set immediately.

**JWT chain:**
1. BFF (`chat-web`) signs all tokens with `JWT_SECRET` (HMAC-SHA256, `iss: chat-web`)
2. REST service tokens: `exp: 10m`, `aud: <target-service>`
3. WS tokens: `exp: 60s`, signed with `JWT_SECRET` decoded to bytes (SmallRye JWK requirement)
4. When `chat-delivery` calls `chat-message`, it mints a new 30s token (`iss: chat-delivery-service`) via `ServiceTokenFactory` — the original user token has the wrong `aud`

**Identity thread:** A single `useruuid` (UUID generated browser-side at signup) flows through Better Auth's user table → JWT `sub` → all Kafka message fields → Postgres foreign keys across every service.

**Kafka fan-out:** Each `chat-delivery` pod uses `kafka.group.id=chat-delivery-service-${quarkus.uuid}` so every pod receives every `chat.messages` event and fans out to its own connected clients. No shared session store needed.

## Shared environment variables

| Variable | Notes |
|---|---|
| `JWT_SECRET` | Raw HMAC secret. `chat-delivery` also needs it as `JWT_JWK` (base64url-encoded). |
| `KAFKA_BOOTSTRAP` | e.g. `kafka:9092` |
| `DATABASE_URL` | Per-service; three separate DBs on one Postgres instance |

## Infrastructure (chat-infra)

See [chat-infra/README.md](chat-infra/README.md). Two modes: persistent (`~/docker-volumes/`) and ephemeral (no mounts, used by `compose-up.sh`). Nginx routes `/chat/` to `chat-delivery` with WS upgrade headers; everything else goes to `chat-web`.

Three databases on shared Postgres: `chat-auth` (Better Auth), `chat-user` (Spring Boot), `chat-message` (Go migrations).

## Known scaffolding (not yet wired)

- `chat-message-service/internal/grpc/` — gRPC stubs; `GET /conversations/ids` REST endpoint is the current substitute
- `chat-user-service` Spring Integration deps — planned `UserCreated` Kafka event
- `chat-user-service` `PaginatedResponse<T>` — present but not wired to any endpoint
- `TODO` / `REVISIT` comments in the code are intentional evolution markers

## Workflow
- Always present a numbered plan before making any file changes
- Wait for explicit approval before executing
- Do not proceed if the plan involves more than 3 file changes without confirmation

## Development practices

**Adding features:** Implement as clean additions where the architecture allows — new files, new endpoints, new consumers — without touching existing logic. When a change to existing code is genuinely required, keep it surgical. The current system is fully tested and working.

**Nginx routing:** Nginx only handles protocol-level concerns (the `/chat/` WebSocket upgrade route). New REST services do not require Nginx route additions — the BFF proxies them internally.

**Verification (three layers):**
1. UI smoke test — build and start the full stack (`./docker-build-all.sh`, `./compose-up.sh`), verify the golden path in the browser. Human-executed; the browser is the correct tool for WebSocket flows.
2. API smoke test — targeted `curl` commands against the running stack for new REST endpoints.
3. Unit tests — service layer tests with mocked dependencies, run per-component.

## Agent skills

### Issue tracker

Issues live in GitHub Issues (`github.com/andleosab/polyglot-chat-system`). See `docs/agents/issue-tracker.md`.

### Triage labels

Default five-role label vocabulary (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.
