import { writable, type Writable } from 'svelte/store';
import { append } from '$lib/store/messages';         // ← shared store
import type { ChatMessage, UserCreatedMessage, TypingSignal } from '$lib/api/types/message';
import { env } from '$env/dynamic/public';

export const userCreated: Writable<UserCreatedMessage | null> = writable(null);
export const isConnected: Writable<boolean> = writable(false);
export const presenceUpdate: Writable<ChatMessage | null> = writable(null);
export const typingSignal: Writable<TypingSignal | null> = writable(null);
let _typingClearTimer: ReturnType<typeof setTimeout> | null = null;

let reconnectAttempts = 0;
const MAX_RECONNECT = 5;
const RECONNECT_DELAY = 3000;
const MAX_MESSAGES = 300;

let ws: WebSocket | null = null;
let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;

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
    if (!userUuid || ws) return;

    const isDev = import.meta.env.DEV;

    console.log("env: ", isDev);

    const proto = location.protocol === 'https:' ? 'wss' : 'ws';

    const wsUrl = isDev ? `${env.PUBLIC_DELIVERY_API_BASE}/chat/${userUuid}`
        : `${proto}://${location.host}/chat/${userUuid}`;    

    console.log("==> ", wsUrl);

    const { token } = await fetch("/api/ws-token", { method: "POST" }).then(r => r.json());
    const quarkusHeaderProtocol = encodeURIComponent("quarkus-http-upgrade#Authorization#Bearer " + token);

    console.log('Connecting to WebSocket with token:', token);

    ws = new WebSocket(wsUrl,
        ["bearer-token-carrier", quarkusHeaderProtocol]);

    ws.onopen = () => {
        isConnected.set(true);
        reconnectAttempts = 0;
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
        console.log('close code:', event.code, 'reason:', event.reason);

        console.log('WebSocket disconnected');

        if (reconnectTimeout) {
            clearTimeout(reconnectTimeout);
            reconnectTimeout = null;
        }

        //REVISIT: we should not reconnect if the server rejected session. 
        // in this case we need to ask user to login again. 
        // but how to do it? maybe we can set some flag in a store and then react to it in a component?
        // custom code 4400 is not working
        if (event.code === 4400) {
            console.error('Server rejected session, not reconnecting. Error: ', event);
            return;
        }

        if (reconnectAttempts < MAX_RECONNECT) {
            reconnectAttempts++;
            console.log(`Reconnect attempt ${reconnectAttempts} in ${RECONNECT_DELAY}ms`);
            reconnectTimeout = setTimeout(() => connect(userUuid), RECONNECT_DELAY);
        } else {
            console.warn('Max reconnect attempts reached. Not retrying.');
        }
    };
}

export function disconnect(): void {
  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
    reconnectTimeout = null;
  }
  reconnectAttempts = MAX_RECONNECT; // prevent auto-reconnect
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