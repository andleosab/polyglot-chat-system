# chat-infra

Docker Compose infrastructure for local development: PostgreSQL, Redpanda (Kafka-compatible broker), and Nginx as reverse proxy. All services share a Docker bridge network named `chat-demo-net`.

## Components

| Service | Image | Purpose |
|---|---|---|
| `chat-postgres` | `postgres:17-alpine` | Shared database host (three separate databases) |
| `chat-kafka` | `redpandadata/redpanda:v24.1.2` | Kafka-compatible broker |
| `nginx` | `nginx:1.27-alpine` | Reverse proxy / WebSocket gateway |

## Two compose modes

**Persistent** (`docker-compose.yml`): Postgres data and Kafka data are mounted to `~/docker-volumes/`. Use this for day-to-day development — data survives container restarts.

**Ephemeral** (`docker-compose-ephemeral.yml`): No volume mounts. Postgres initialises from `init-db/` scripts on first start. Kafka state is lost on stop. Use this for CI, demos, or clean-slate testing. This is what `chat-demo/compose-up.sh` uses.

## Database initialisation

`init-db/01-create-db.sql` creates three databases on the single Postgres instance:

```sql
CREATE DATABASE "chat-user";    -- chat-user-service
CREATE DATABASE "chat-message"; -- chat-message-service  
CREATE DATABASE "chat-auth";    -- Better Auth (chat-web)
```

`init-db/02-better-auth.sql` connects to `chat-auth` and creates Better Auth's schema: `user`, `session`, `account`, and `verification` tables, with all required indexes. The `user` table includes the two custom columns that Better Auth in `chat-web` expects: `username` and `useruuid`.

These scripts only run on first start (Postgres's `docker-entrypoint-initdb.d` mechanism). Individual service schemas are managed separately: `chat-message-service` uses `golang-migrate`, Spring Boot uses `ddl-auto`.

## Nginx

`nginx.conf` routes all traffic through a single entry point on port 80:

```nginx
location / {
    proxy_pass http://chat-web:3000;       # SvelteKit BFF
}

location /chat/ {
    proxy_pass http://chat-delivery-service:8080;  # WebSocket gateway
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;
    proxy_read_timeout  3600;              # 1-hour keepalive for WS
    proxy_send_timeout  3600;
    proxy_buffering off;
}
```

The `map $http_upgrade $connection_upgrade` block handles the WebSocket upgrade correctly. The 1-hour timeouts are set to prevent Nginx from closing idle WebSocket connections.

## Kafka (Redpanda)

Redpanda runs with two advertised listeners:

| Listener | Address | Use |
|---|---|---|
| `PLAINTEXT` | `kafka:9092` | Internal Docker network (service-to-service) |
| `OUTSIDE` | `localhost:19092` | Host access for local tooling (`rpk`, `kafka-console-consumer`, etc.) |

`auto_create_topics_enabled=true` is set for convenience — topics are created on first produce. Disable this in production.

## Running

```bash
# copy env file
cp env-example .env

# persistent mode
./start-services.sh
./stop-services.sh

# ephemeral mode (used by chat-demo compose-up.sh)
./start-services-ephemeral.sh
./stop-services-ephemeral.sh
```

Or via the umbrella project's `./compose-up.sh`, which also polls Postgres and Kafka health before starting dependent services.

## Network

All service `docker-compose.yml` files attach to `chat-demo-net` as an **external** network. This network is created by `chat-infra` and must exist before other services start. The umbrella `compose-up.sh` script handles startup order automatically.
