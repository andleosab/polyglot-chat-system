# chat-user-service

Spring Boot microservice that maintains chat-specific user profiles. It is a **pure profile store** — it holds no credentials and issues no tokens. Authentication is owned entirely by the SvelteKit BFF via Better Auth.

## Responsibilities

- Provision a user profile record when Better Auth registers a new account (called by the BFF `databaseHooks.user.create.after` hook)
- Serve user profiles to other services (BFF, admin tooling)
- Soft-deactivate users (`isActive = false`)

## Stack

| | |
|---|---|
| Runtime | Java 21 |
| Framework | Spring Boot 3.5 |
| Persistence | Spring Data JPA → PostgreSQL |
| Security | Spring Security OAuth2 Resource Server (HMAC-SHA256 JWT) |
| API docs | SpringDoc OpenAPI / Swagger UI |
| Mapping | MapStruct + Lombok |

## API

Base path: `/user` (all routes require a valid Bearer JWT except as noted)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/users` | Provision a new user profile. `userid` (UUID) is optional — the BFF passes the Better Auth UUID to keep identity consistent across services |
| `GET` | `/api/users/{id}` | Get profile by UUID |
| `GET` | `/api/users` | List all users |
| `PUT` | `/api/users/{id}/deactivate` | Soft-deactivate (`isActive = false`) |

## Data model

```sql
users (
  id         BIGINT       PK (sequence, internal only)
  user_id    UUID         NOT NULL UNIQUE  -- public identifier
  username   VARCHAR      NOT NULL UNIQUE
  email      VARCHAR      NOT NULL UNIQUE
  created_at TIMESTAMPTZ  NOT NULL
  updated_at TIMESTAMPTZ
  is_active  BOOLEAN      NOT NULL DEFAULT true
)
```

The dual-key pattern is intentional: `id` is an efficient internal surrogate; `user_id` (UUID) is what every other service references. The `user_id` is generated via `@PrePersist` if not supplied in the request, but in practice the BFF always passes the Better Auth UUID so identity is consistent across `chat-auth`, `chat-user`, and JWT `sub` claims.

## JWT verification

Validates HMAC-SHA256 tokens signed by the BFF (`chat-web`). Configured in `SecurityConfig`:

```java
NimbusJwtDecoder.withSecretKey(
    new SecretKeySpec(secret.getBytes(), "HmacSHA256")
).build()
```

`JWT_SECRET` must match the value used by `chat-web` to sign service tokens.

## Environment variables

| Variable | Default | Notes |
|---|---|---|
| `DB_URL` | — | `jdbc:postgresql://host:5432/chat-user` |
| `DB_USERNAME` | — | |
| `DB_PASSWORD` | — | |
| `JWT_SECRET` | — | Shared HMAC secret (base64) |
| `SERVER_PORT` | `8080` | |

Copy `env-example` → `.env.docker` before running via Docker Compose.

## Running

```bash
# local (requires Postgres on localhost:5432/chat-user)
./mvnw spring-boot:run

# docker
docker build -t spring/chat-user-service .
docker compose up
```

Schema is managed by Hibernate `ddl-auto: update` in dev. In staging/prod profiles (`ST`, `PR`, `ND`) DDL is disabled — manage schema externally.

Seed data (`data.sql`) inserts three test users on startup using `ON CONFLICT DO NOTHING`.

## Notes

- `spring-boot-starter-oauth2-resource-server` is used in place of the full security starter — only the JWT resource server filter chain is configured, no form login or basic auth.
- `spring-boot-starter-integration` and related dependencies are declared but not active — scaffolding for a planned Kafka-based `UserCreated` event (would allow `chat-message-service` to be notified of new users without a REST call).
- `PaginatedResponse<T>` and `PaginationMetadata` are present but not yet wired into any endpoint — future-proofing for paginated user lists.
