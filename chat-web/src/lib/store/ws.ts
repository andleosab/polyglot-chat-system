import { writable, type Writable } from 'svelte/store';
import { append } from '$lib/store/messages';         // ← shared store
import type { ChatMessage, UserCreatedMessage, TypingSignal } from '$lib/api/types/message';
import { env } from '$env/dynamic/public';

export const userCreated: Writable<UserCreatedMessage | null> = writable(null);
export const isConnected: Writable<boolean> = writable(false);
export const presenceUpdate: Writable<ChatMessage | null> = writable(null);
export const typingSignal: Writable<TypingSignal | null> = writable(null);
let _typingClearTimer: ReturnType<typeof setTimeout> | null = null;

// Set when the BFF refuses to mint a socket token. Components can watch this to prompt
// a fresh sign-in; nothing reconnects while it is true.
export const sessionExpired: Writable<boolean> = writable(false);

// Close codes that must NOT be retried. Every other close is treated as transient — a
// preempted node, a restarting pod, an idle socket reaped by an edge proxy — so retrying
// is the default and this set is the exception.
//   1000  normal closure, including our own disconnect()
//   4401  session rejected by the delivery service (reserved; see ChatSocket)
// Note 4503 is deliberately absent: it means a dependency was unavailable, which is
// exactly the case worth retrying.
const FATAL_CLOSE_CODES = new Set([1000, 4401]);

const RECONNECT_BASE_DELAY = 1000;
const MAX_RECONNECT_DELAY = 30000;
const MAX_MESSAGES = 300;

let reconnectAttempts = 0;
let ws: WebSocket | null = null;
let connecting = false;
let intentionalClose = false;
let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;

function scheduleReconnect(userUuid: string): void {
    if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
    }

    reconnectAttempts++;
    const delay = Math.min(
        RECONNECT_BASE_DELAY * Math.pow(1.5, reconnectAttempts),
        MAX_RECONNECT_DELAY
    );

    console.log(`Reconnect attempt ${reconnectAttempts} in ${(delay / 1000).toFixed(1)}s`);
    reconnectTimeout = setTimeout(() => {
        reconnectTimeout = null;
        void connect(userUuid);
    }, delay);
}

// ─── One-shot new conversation listener ───────────────────────────────────────
// ChatView registers this before sending the very first private message.
// Fires once when the server echoes back a message with a populated conversationId,
// then clears itself automatically.

let newConvoListener: ((id: number) => void) | null = null;

export function onceNewConversation(cb: (id: number) => void): void {
  newConvoListener = cb;
}

export function cancelNewConversationListener(): void {
  newConvoListener = null;
}


export async function connect(userUuid: string) {
    // `connecting` guards the await below: ws is still null while the token is in flight,
    // so the ws check alone would let a second call open a duplicate socket.
    if (!userUuid || ws || connecting) return;

    connecting = true;
    intentionalClose = false;

    try {
        const isDev = import.meta.env.DEV;

        console.log("env: ", isDev);

        const proto = location.protocol === 'https:' ? 'wss' : 'ws';

        const wsUrl = isDev ? `${env.PUBLIC_DELIVERY_API_BASE}/chat/${userUuid}`
            : `${proto}://${location.host}/chat/${userUuid}`;

        console.log("==> ", wsUrl);

        const res = await fetch("/api/ws-token", { method: "POST" });

        // The only failure the browser can positively identify as fatal. A rejected token
        // fails here rather than at the upgrade, where @Authenticated returns a 401
        // handshake and the socket reports 1006 — indistinguishable from a dead pod.
        if (res.status === 401) {
            console.error('WS token refused: session is gone. Not reconnecting.');
            sessionExpired.set(true);
            return;
        }

        if (!res.ok) throw new Error(`ws-token request failed: ${res.status}`);

        const { token } = await res.json();
        const quarkusHeaderProtocol = encodeURIComponent("quarkus-http-upgrade#Authorization#Bearer " + token);

        console.log('Connecting to WebSocket with token:', token);

        ws = new WebSocket(wsUrl,
            ["bearer-token-carrier", quarkusHeaderProtocol]);

        ws.onopen = () => {
            isConnected.set(true);
            reconnectAttempts = 0;
            sessionExpired.set(false);
            console.log('WebSocket connected');
        };

        ws.onmessage = (event) => {
            const rawMsg = JSON.parse(event.data);

            // User created broadcast — update the users store
            if (rawMsg.type === 'USER_CREATED') {
                console.log('User created event:', rawMsg);
                userCreated.set(rawMsg as UserCreatedMessage);
                return;
            }

            const msg: ChatMessage = rawMsg;

            // Live presence updates — set store, never append to message history
            if (msg.type === 'USER_JOINED' || msg.type === 'USER_LEFT') {
                console.log('Presence update:', msg.type, msg.from);
                presenceUpdate.set(msg);
                return;
            }

            // Typing indicator — set store then auto-clear after 4s
            if (msg.type === 'TYPING') {
                if (_typingClearTimer) clearTimeout(_typingClearTimer);
                typingSignal.set({ conversationId: msg.conversationId!, from: msg.from, fromName: msg.fromName });
                _typingClearTimer = setTimeout(() => typingSignal.set(null), 4000);
                return;
            }

            console.log('Received message:', msg);

            append(msg);   // ← write into shared messages store
            // Fire the one-shot listener if a new conversationId comes back
            if (msg.conversationId && newConvoListener) {
                newConvoListener(msg.conversationId);
                newConvoListener = null;
            }
        };

        ws.onclose = (event) => {
            isConnected.set(false);
            ws = null;
            console.log('WebSocket disconnected. code:', event.code, 'reason:', event.reason);

            if (intentionalClose || FATAL_CLOSE_CODES.has(event.code)) {
                if (event.code === 4401) sessionExpired.set(true);
                console.log('Not reconnecting.');
                return;
            }

            scheduleReconnect(userUuid);
        };
    } catch (err) {
        // A failed token fetch or a rejected handshake must not end the retry chain —
        // without this the promise rejects unhandled and the socket stays dead.
        console.error('WebSocket connect failed:', err);
        scheduleReconnect(userUuid);
    } finally {
        connecting = false;
    }
}

export function disconnect(): void {
  // With no attempt cap left, the flag is what tells onclose this close was ours.
  intentionalClose = true;
  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
    reconnectTimeout = null;
  }
  reconnectAttempts = 0;
  ws?.close();
  ws = null;
  isConnected.set(false);
}

export function send(body: ChatMessage) : void {
    if (ws?.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(body));
    } else {
        console.warn('WebSocket not connected. Message not sent.');
    }
}

export function sendTyping(conversationId: number): void {
    fetch('/api/presence/presence/typing', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ conversationId: String(conversationId) }),
    }).catch(() => {});
}