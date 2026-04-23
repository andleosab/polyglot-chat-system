import { writable, get } from 'svelte/store';
import { type MessageResponse, type ChatMessage, toMessageResponse } from '$lib/api/types/message';

export const LIMIT = 30;

// ─── Exported stores ──────────────────────────────────────────────────────────

export const messages = writable<MessageResponse[]>([]);
export const isLoading = writable(false);
export const hasMore = writable(true);
export const fetchError = writable<string | null>(null);

// ─── Internal cursor (not a store — no UI needs to react to it) ───────────────

let cursor: number | null = null;

let activeConversationId: number | null = null;

// ─── seed ─────────────────────────────────────────────────────────────────────
// Called once in onMount with the server-loaded initial page (already ASC order).
// Sets the cursor to the oldest message id in that page.

export function seed(conversationId: number, initial: MessageResponse[]): void {
    messages.set(initial);
    cursor = initial.length > 0 ? initial[0].id : null;
    hasMore.set(initial.length >= LIMIT);
    activeConversationId = conversationId;

    console.log('seeded ids:', initial.map(m => m.id));
    console.log('cursor set to:', cursor);    

    console.log('==> active conversation set to:', activeConversationId);
}

// ─── append ───────────────────────────────────────────────────────────────────
// Called by ws.ts when a live WS message arrives — appends to the same array.

export function append(msg: ChatMessage): void {

  const msgResponse: MessageResponse = toMessageResponse(msg);

  // ignore messages for other conversations
  if (msgResponse.conversation_id !== activeConversationId) {
    return;
  }  

  messages.update(current => {

    // If the message already has a DB id, dedupe it
    if (msgResponse.id != null && current.some(m => m.id === msgResponse.id)) {
      // Deduped — return a shallow copy to trigger reactivity
      return [...current];
    }

    const next = [...current, msgResponse];

    // Soft cap — trim oldest if store grows too large
    return next.length > 300 ? next.slice(-300) : next;
  });
}

// ─── fetchOlderMessages ───────────────────────────────────────────────────────
// Fetches the next older page using the current cursor.
// Returns the fetched page so the component can use it for scroll anchoring.
// Returns [] if already loading, no more pages, or cursor is unset.

export async function fetchOlderMessages(conversationId: number): Promise<MessageResponse[]> {
    
    if (get(isLoading) || !get(hasMore) || cursor === null) return [];

    isLoading.set(true);
    fetchError.set(null);

    try {
        const res = await fetch(
            `/api/chats/${conversationId}?before=${cursor}&limit=${LIMIT}`
        );
        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const page: MessageResponse[] = await res.json();
console.log('cursor was:', cursor);
console.log('fetched ids:', page.map(m => m.id));        

        if (page.length === 0) {
        hasMore.set(false);
        return [];
        }        

        if (page.length < LIMIT) hasMore.set(false);

        if (page.length > 0) {
            const newCursor = page[page.length - 1].id;
            if (newCursor === cursor) {
                // cursor didn't advance — truly at the beginning
                hasMore.set(false);
                return [];
            }            
            // Update cursor to oldest message in this batch (last item — backend returns DESC)
            cursor = newCursor
            // Reverse DESC → ASC then prepend
            messages.update(current => [...page.toReversed(), ...current]);
        }

        return page;
    } catch (e) {
        fetchError.set(e instanceof Error ? e.message : 'Failed to load messages');
        return [];
    } finally {
        isLoading.set(false);
    }
}

// ─── reset ────────────────────────────────────────────────────────────────────
// Called in onDestroy — clears everything so stale state doesn't bleed into
// the next conversation.

export function reset(): void {
    messages.set([]);
    isLoading.set(false);
    hasMore.set(true);
    fetchError.set(null);
    cursor = null;
    activeConversationId = null;
}
