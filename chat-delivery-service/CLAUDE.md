# CLAUDE.md — chat-delivery-service

Quarkus 3 service that owns the real-time transport layer: every browser WebSocket connection lands here. See [README.md](README.md) for full env vars and JWT details.

## Commands

```bash
./mvnw quarkus:dev    # dev mode with hot reload (requires kafka + chat-message-service)
./mvnw package        # build JVM jar
./docker-build.sh     # build quarkus/chat-delivery-service JVM image
docker compose up     # run via compose (copy env-example first)
```

Native image: `./mvnw package -Pnative` (not required for local dev).

## Source layout

```
src/main/java/org/demo/chat/
  ChatSocket.java              # @WebSocket — WS lifecycle (onOpen/onMessage/onClose)
  ChatConsumer.java            # @Incoming("chat-in") — Kafka → fan-out to WS clients
  ChatProducer.java            # @Outgoing("chat-out") — WS message → Kafka
  ParticipantConsumer.java     # @Incoming("participant-created/removed") — updates in-memory groups
  Message.java / MessageType.java / WsKeys.java
  client/
    MessageService.java        # MicroProfile REST client interface → chat-message-service
    ServiceTokenFactory.java   # mints short-lived JWT for outbound calls
    ServiceAuthFilter.java     # injects "Authorization: Bearer" on MicroProfile client requests
    CreateConversationRequest/Response.java
  config/
    ServiceConfig.java         # @ConfigMapping for service.* properties
src/main/resources/
  application.properties       # all Quarkus/SmallRye/Kafka config
```

## Key patterns

**WebSocket endpoint (`ChatSocket.java`):** `@WebSocket(path = "/chat/{username}")` + `@Authenticated`. The `{username}` path param carries the user UUID (naming is a known REVISIT). On open, calls `MessageService.getConversationIDs(username)` and stores the `Set<Long>` in `connection.userData()` under `WsKeys.GROUPS`. If this REST call fails, connection closes with custom code `4503`.

**Close codes (RFC 6455 private-use range):** the browser store retries *every* close except an explicit fatal set, so a code's number is the contract — mirror the HTTP status whose retry semantics match.

| Code | Meaning | Client |
|---|---|---|
| `4503` | A dependency needed to open the session was unavailable | Retries with backoff |
| `4401` | Session rejected — reserved, not currently emitted | Fatal; prompts re-login |

`4401` is unused because `@Authenticated` rejects during the HTTP upgrade: the socket never opens, so there is no close frame and the browser only ever reports `1006`. Do not reuse `4400` — it reads as an echo of HTTP 400 (client's fault, don't retry), which is the opposite of what this failure means.

**Idle keepalive:** `quarkus.websockets-next.server.auto-ping-interval=30s` makes the server emit RFC 6455 ping frames; browsers answer automatically. An idle chat socket otherwise carries no traffic at all — `HeartbeatScheduler` pings presence over gRPC and never touches the socket — and edge proxies close what looks dead. nginx tolerates it via `proxy_read_timeout 3600`, but Cloudflare caps WebSocket idle time at 100s and does not make it configurable below Enterprise. Note these are control frames: Chrome DevTools does not show them in the WS Frames panel.

**Kafka fan-out (`ChatConsumer.java`):** Iterates all active `WebSocketConnection`s and sends to those whose `userData().get(WsKeys.GROUPS)` contains the message's `conversationId`. Each pod receives every message because `kafka.group.id=chat-delivery-service-${quarkus.uuid}` gives each instance a unique consumer group.

**Participant event handling (`ParticipantConsumer.java`):** On `chat.participants.created` / `chat.participants.removed`, finds the affected user's connection on this instance and adds/removes the `conversationId` from their in-memory group set. No reconnect required.

**Private conversation lazy creation (`ChatSocket.onMessage`):** If the inbound `Message.conversationId` is null, calls `MessageService.createConversation()` to create a new private conversation, adds the returned ID to the sender's group set, then publishes to Kafka.

**Service-to-service JWT (`ServiceTokenFactory.java`):** Extracts the user's `sub` from the inbound JWT (via `SecurityIdentity`), mints a new HS256 token with `iss: chat-delivery-service`, `aud: chat-message-service`, `exp: 30s`. `ServiceAuthFilter` injects this on every MicroProfile REST client request.

**WS JWT handshake (`application.properties`):**
```properties
quarkus.websockets-next.server.supported-subprotocols=bearer-token-carrier
quarkus.websockets-next.server.propagate-subprotocol-headers=true
```
The browser encodes the token as `quarkus-http-upgrade#Authorization#Bearer <token>` in the subprotocol list. Quarkus unpacks it into an `Authorization` header before SmallRye JWT validation runs.

**JWT verification config:**
```properties
smallrye.jwt.verify.secretkey=${JWT_JWK}   # base64url-encoded shared secret
mp.jwt.verify.publickey.algorithm=HS256
mp.jwt.verify.issuer=chat-web              # only tokens from the BFF are accepted
```
`JWT_JWK` and `JWT_SECRET` must both be set — different formats of the same secret.

## Kafka channel names (application.properties)

| Channel | Topic | Direction |
|---|---|---|
| `chat-out` | `chat.messages` | outgoing (producer) |
| `chat-in` | `chat.messages` | incoming (consumer) |
| `participant-created` | `chat.participants.created` | incoming |
| `participant-removed` | `chat.participants.removed` | incoming |
