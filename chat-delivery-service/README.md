# chat-delivery-service

Quarkus service that owns the real-time transport layer. Every browser WebSocket connection lands here. Inbound messages go to Kafka; Kafka messages come back and are fanned out to connected clients based on their conversation membership.

## Responsibilities

- Accept authenticated WebSocket connections at `/chat/{userUUID}`
- On connect: fetch the user's conversation IDs from `chat-message-service` and store them in connection-local state
- On message: publish to Kafka `chat.messages` (and lazily create a private conversation if `conversationId` is absent)
- Consume `chat.messages` from Kafka and deliver to every connection on **this instance** that belongs to the target conversation
- Consume participant events from `chat-message-service` and update connected users' in-memory group membership without requiring a reconnect

## Stack

| | |
|---|---|
| Runtime | Java 21 |
| Framework | Quarkus 3.23 |
| WebSockets | `quarkus-websockets-next` |
| Messaging | SmallRye Reactive Messaging → Kafka |
| REST client | MicroProfile REST Client |
| JWT (inbound) | SmallRye JWT, HS256 |
| JWT (outbound) | SmallRye JWT builder (mints service tokens) |

## WebSocket endpoint

**Path:** `/chat/{username}` — `{username}` carries the user UUID (path param naming is a pending cleanup).

The endpoint is `@Authenticated` — SmallRye JWT validates the token before the handshake completes.

**On open:**
1. Call `chat-message-service GET /messaging/conversations/ids?id={uuid}` to fetch conversation IDs
2. Store `Set<Long> conversationIds` and `username` in `WebSocketConnection.userData()`
3. Send `USER_JOINED` message to the connecting client
4. If the REST call fails, close the connection with custom code `4400` — the browser WS store watches for this code and does not attempt to reconnect

**On message:**
- If `conversationId` is null, call `chat-message-service POST /messaging/conversations` to create a new private conversation, then add the returned ID to this connection's group set
- Publish the (resolved) message to Kafka `chat-out` channel

**On close:** Broadcast `USER_LEFT` to connections on this instance.

## Kafka consumer fan-out

```java
@Incoming("chat-in")
public void consumeConversation(Message message) {
    connections.stream()
        .filter(conn -> conn.userData().get(WsKeys.GROUPS).contains(message.conversationId()))
        .forEach(conn -> conn.sendText(message));
}
```

Each instance only delivers to connections it hosts. For horizontal scaling, every pod gets a unique consumer group ID:

```properties
kafka.group.id=chat-delivery-service-${quarkus.uuid}
```

This ensures every pod receives every Kafka message and can fan-out to its own connected clients. No shared session store is required.

**Participant events** (`chat.participants.created` / `chat.participants.removed`): If the affected user is connected to this instance, the service immediately adds or removes the conversation ID from their in-memory group set.

## Service-to-service JWT (identity propagation)

When `chat-delivery-service` calls `chat-message-service` it cannot forward the user's inbound JWT (wrong audience). Instead, `ServiceTokenFactory` mints a new short-lived JWT (30s TTL) propagating the user's `sub` claim with correct `iss` / `aud`:

```
inbound JWT (iss:chat-web, aud:chat-delivery-service)
    → ServiceTokenFactory
        → new JWT (iss:chat-delivery-service, aud:chat-message-service, sub:<user UUID>, exp:30s)
    → ServiceAuthFilter injects "Authorization: Bearer <token>"
    → MicroProfile REST Client call
```

## WebSocket JWT handshake

Browser WebSocket upgrades don't reliably support custom headers. The subprotocol workaround encodes the token as a protocol string:

```js
// client
ws = new WebSocket(url, [
    "bearer-token-carrier",
    encodeURIComponent("quarkus-http-upgrade#Authorization#Bearer " + token)
]);
```

```properties
# server
quarkus.websockets-next.server.supported-subprotocols=bearer-token-carrier
quarkus.websockets-next.server.propagate-subprotocol-headers=true
```

Quarkus unpacks the encoded subprotocol into the `Authorization` header before JWT validation.

## JWT verification

```properties
smallrye.jwt.verify.secretkey=${JWT_JWK}     # base64url-encoded shared secret
mp.jwt.verify.publickey.algorithm=HS256
mp.jwt.verify.issuer=chat-web                # only BFF-issued tokens accepted
```

`JWT_JWK` is the shared secret in base64url format (required by SmallRye's JWK key parsing). It is derived from the same `JWT_SECRET` used across all services — the BFF encodes it to raw bytes when signing WS tokens.

## Environment variables

| Variable | Notes |
|---|---|
| `KAFKA_BOOTSTRAP` | e.g. `kafka:9092` |
| `MESSAGE_SERVICE_URL` | e.g. `http://chat-message-service:8080/messaging/conversations` |
| `JWT_JWK` | Base64url-encoded shared secret for inbound JWT verification |
| `JWT_SECRET` | Raw secret for outbound service token signing |
| `HTTP_PORT` | Default `8080` |

## Running

```bash
# JVM mode
./mvnw quarkus:dev

# docker (JVM image)
./docker-build.sh   # builds quarkus/chat-delivery-service image
docker compose up
```

Native image build is supported (`-Pnative`) but not required for local dev. Multiple Dockerfiles are provided: `Dockerfile.jvm` (default), `Dockerfile.native`, `Dockerfile.native-micro`.
