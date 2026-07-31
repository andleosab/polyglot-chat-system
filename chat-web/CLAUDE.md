# CLAUDE.md — chat-web

SvelteKit 5 app acting as both the UI and BFF. See [README.md](README.md) for full env vars and auth flow details.

## Commands

```bash
pnpm install
pnpm dev          # dev server on :3000
pnpm build        # production build (@sveltejs/adapter-node)
pnpm check        # svelte-check type checking
```

Env: copy `env-example` → `.env.local.docker`

## Source layout

```
src/
  hooks.server.ts               # BFF: session resolution + JWT injection + /(app)/ guard
  lib/
    auth.ts                     # Better Auth server config; databaseHooks provisions chat-user-service
    auth-client.ts              # Better Auth browser client
    server/
      jwt.ts                    # issueServiceToken() / issueWsToken()
      config.ts                 # env var imports
    store/
      ws.ts                     # WebSocket lifecycle (connect/disconnect/send/reconnect)
      messages.ts               # cursor-paginated message store (seed/append/fetchOlderMessages/reset)
      user.ts                   # CurrentUser store
    api/types/                  # shared TS types (conversation, message, user)
    state/user-session.svelte.ts
  routes/
    (app)/                      # auth-gated routes
      chats/                    # conversation list + [id]/[name] chat view
      groups/                   # group list + [id]/[name] + new
      users/                    # user directory
    api/
      ws-token/+server.ts       # POST: issues short-lived WS JWT
      chats/[id]/+server.ts     # GET: cursor-paginated messages proxy
    sign-in/ sign-up/
```

## Key patterns

**JWT injection (`hooks.server.ts`):** Overrides `event.fetch` with a version that matches the URL against `SERVICE_MAP` (keyed by `MESSAGE_API_BASE` / `USER_API_BASE`) and injects a signed Bearer token. All `+page.server.ts` files call `fetch()` normally — they have no knowledge of auth. One token per service base is cached per request via `tokenCache`.

**Two token types (`src/lib/server/jwt.ts`):**
- `issueServiceToken(audience, user)` — signs with the raw `JWT_SECRET` string (`getJwtSecret()`), `exp: 10m`
- `issueWsToken(audience, user)` — signs with `JWT_SECRET` decoded from base64url to bytes (`getJwtSecretBytes()`), `exp: 60s`. Required because Quarkus SmallRye JWT expects raw bytes for JWK key material.

Both read `JWT_SECRET` through `$env/dynamic/private` on call, never at module load — `pnpm build` must not require it.

**WebSocket handshake (`store/ws.ts`):** Fetches a WS token from `/api/ws-token`, then connects with two subprotocols: `"bearer-token-carrier"` and the URL-encoded `"quarkus-http-upgrade#Authorization#Bearer <token>"` string. Quarkus unpacks the second protocol into an `Authorization` header before JWT validation.

**Reconnect logic (`store/ws.ts`):** Up to 5 attempts with 3s delay. Custom close code `4400` (server-rejected session) skips reconnect entirely.

**Message store (`store/messages.ts`):** Hybrid server-loaded + live feed.
- `seed(conversationId, initial)` — populates from server-loaded history, sets cursor to oldest message id
- `append(msg)` — called by ws.ts on incoming WS message; dedupes by id, ignores other conversations
- `fetchOlderMessages(conversationId)` — fetches backwards via `/api/chats/[id]?before=<cursor>&limit=30`; backend returns DESC, stored reversed to ASC
- `reset()` — must be called in `onDestroy` to clear state between conversations

**One-shot new conversation listener (`store/ws.ts`):** `onceNewConversation(cb)` registers a callback that fires once when the server echoes back a message with a populated `conversationId` (new private chat). The callback fires then clears itself; used by ChatView to navigate to the permanent conversation URL.

**User provisioning (`lib/auth.ts`):** `databaseHooks.user.create.after` calls `chat-user-service POST /api/users` with the Better Auth UUID as `userid`. If this call fails, the error propagates and registration rolls back.

## Svelte 5 note

This project uses Svelte 5 runes (`$state`, `$derived`, etc.) and the new snippet/component API. Avoid Svelte 4 patterns (`$:`, `<slot>`, stores as the sole reactive primitive for component-local state).
This project also uses taiwindcss for styling and svelte port of shadcn UI components.
See https://www.shadcn-svelte.com/

See details on shadcn at https://www.shadcn-svelte.com/llms.txt

## Browser Automation

Use `agent-browser` for web automation. Run `agent-browser --help` for all commands.

Core workflow:

1. `agent-browser open <url>` - Navigate to page
2. `agent-browser snapshot -i` - Get interactive elements with refs (@e1, @e2)
3. `agent-browser click @e1` / `fill @e2 "text"` - Interact using refs
4. Re-snapshot after page changes

### UI testing checklist

After any reactive event fires (WS message received, store updated, presence change, etc.):

1. **Check the console for JS errors** — a screenshot proving the UI rendered is not enough; Svelte reactive crashes (`effect_update_depth_exceeded`, etc.) only show up in the console:
   ```bash
   agent-browser eval "document.querySelectorAll('*').length"  # page still responsive?
   ```
2. **Verify interactivity** — click a nav link or button after the event to confirm the page hasn't locked up. A crashed Svelte effect silently freezes the UI while the DOM still looks correct in a screenshot.
3. **Confirm you're testing the new image** — after a rebuild, verify the running container is using the new image before running tests. Stale containers can make broken code appear to pass.
