# chat-web

SvelteKit 5 application that serves as both the user interface and the **Backend-For-Frontend (BFF)**. The browser communicates with backend services only through this layer — the BFF resolves sessions, signs JWTs, and injects them transparently into every outbound fetch.

## Responsibilities

- Sign-in / sign-up via email+password and Google OAuth (Better Auth)
- Session management backed by PostgreSQL (`chat-auth` database)
- JWT issuance: signs service tokens for REST API calls and short-lived WS tokens for the WebSocket handshake
- Transparent JWT injection: all `+page.server.ts` fetch calls to backend services get a `Bearer` token injected via a `hooks.server.ts` fetch override — callers don't handle auth
- UI: conversation list, group management, user directory, real-time chat
- WebSocket client lifecycle: connect on app mount, reconnect on drop, disconnect on navigate away

## Stack

| | |
|---|---|
| Framework | SvelteKit 2 + Svelte 5 |
| Auth | [Better Auth](https://www.better-auth.com/) 1.6 |
| Auth DB | PostgreSQL (`pg`) → `chat-auth` database |
| Styling | Tailwind CSS 4 + `bits-ui` |
| JWT signing | `jsonwebtoken` (server-side only) |
| Adapter | `@sveltejs/adapter-node` |

## Route structure

```
/                    → redirect to /chats
/sign-in             → email + Google login
/sign-up             → registration
/api/ws-token        → POST: issues a short-lived WebSocket JWT
/(app)/chats         → conversation list (server-loaded)
/(app)/chats/[id]/[name]  → chat view (server-loaded history + live WS)
/(app)/groups        → group list
/(app)/groups/new    → create group
/(app)/groups/[id]/[name] → group settings / participant management
/(app)/users         → user directory
/(app)/chats/new/[userId]/[username] → new private chat
```

All `/(app)/` routes redirect to `/sign-in` if no session is present (enforced in `hooks.server.ts`).

## Auth flow

Better Auth is configured with two custom user fields: `username` and `useruuid`. The `useruuid` is generated client-side at registration and stored in Better Auth's `user` table. On account creation, a `databaseHooks.user.create.after` hook fires a server-side fetch to `chat-user-service` to provision the chat profile:

```typescript
after: async (user) => {
    const token = issueServiceToken('chat-user-service', { userId: user.useruuid, ... });
    await fetch(`${USER_API_BASE}/api/users`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: JSON.stringify({ username, email, userid: user.useruuid })
    });
}
```

If this call fails the error propagates and registration is rolled back.

## BFF JWT injection (`hooks.server.ts`)

The hook overrides `event.fetch` to intercept calls to known backend URLs and inject a signed JWT:

```typescript
const SERVICE_MAP = {
    [MESSAGE_API_BASE]: { issueToken: (user) => issueServiceToken('chat-message-service', user) },
    [USER_API_BASE]:    { issueToken: (user) => issueServiceToken('chat-user-service', user) }
};

event.fetch = async (input, init) => {
    const match = Object.entries(SERVICE_MAP).find(([base]) => url.startsWith(base));
    if (match && event.locals.currentUser) {
        // inject Authorization header
    }
    return originalFetch(input, init);
};
```

`+page.server.ts` files call `fetch()` normally — they have no knowledge of JWT handling.

## JWT signing (`src/lib/server/jwt.ts`)

Two token types, both HMAC-SHA256 with `JWT_SECRET`:

**Service tokens** (REST API calls):
- `iss: chat-web`, `aud: <target>`, `exp: 10m`
- Signed with the raw `JWT_SECRET` string

**WS tokens** (WebSocket handshake):
- Adds `scope: "ws"` claim
- Signed with `JWT_SECRET` decoded from base64url to raw bytes — required by Quarkus SmallRye JWT
- `exp: 60s` (one-time use for handshake only)

The WS token is fetched from `/api/ws-token` in the browser just before `new WebSocket(...)` is called.

## WebSocket client (`src/lib/store/ws.ts`)

- Fetches a WS token from `/api/ws-token` then connects with the Quarkus subprotocol workaround
- Reconnects up to 5 times with a 3s delay on unexpected close; does **not** reconnect on custom code `4400` (server-rejected session)
- Routes presence events (`USER_JOINED` / `USER_LEFT`) separately from chat messages
- **One-shot `onceNewConversation` listener**: when the server echoes back a message with a resolved `conversationId` for a new private chat, the listener fires once and navigates the client to the permanent conversation URL

## Message store (`src/lib/store/messages.ts`)

Implements a hybrid server-loaded + live feed with keyset cursor pagination:

- `seed(conversationId, initialMessages)`: populates from server-loaded history, sets cursor to oldest message ID
- `fetchOlderMessages(conversationId)`: fetches backwards using `id < cursor`, prepends pages, advances cursor
- `append(msg)`: adds live WS messages, deduplicates by ID, ignores messages for other conversations
- `reset()`: clears all state on conversation change

## Environment variables

| Variable | Notes |
|---|---|
| `DATABASE_URL` | PostgreSQL connection for Better Auth (`chat-auth` DB) |
| `BETTER_AUTH_SECRET` | Better Auth session secret |
| `BETTER_AUTH_URL` | Public base URL (used for OAuth callbacks) |
| `JWT_SECRET` | Shared HMAC secret for signing service + WS tokens |
| `USER_API_BASE` | e.g. `http://chat-user-service:8080/user` |
| `MESSAGE_API_BASE` | e.g. `http://chat-message-service:8080/messaging` |
| `PUBLIC_DELIVERY_API_BASE` | WebSocket URL, public (in browser), e.g. `ws://localhost/chat` |
| `GOOGLE_CLIENT_ID` | Optional — enables Google OAuth |
| `GOOGLE_CLIENT_SECRET` | Optional |

Copy `env-example` → `.env.local.docker` for Docker Compose.

## Running

```bash
# install
pnpm install

# dev (requires backend services running)
pnpm dev

# build
pnpm build

# docker
docker build -t chat-web:dist .
docker compose up
```

A `Dockerfile.distroless` variant is also provided for smaller production images.
