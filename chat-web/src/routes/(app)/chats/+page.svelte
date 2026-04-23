<script lang="ts">
  import type { PageProps } from './$types';
  import MessageSquareIcon from '@lucide/svelte/icons/message-square';
  import type { ConversationResponse } from '$lib/api/types/conversation';

  let { data } = $props() as PageProps;
  let conversations = $derived(data.conversations as ConversationResponse[]);

  function formatDate(date: Date | string | null | undefined) {
    if (!date) return '';
    return new Date(date).toLocaleDateString([], {
      month: 'short',
      day: 'numeric',
    });
  }
</script>

<div class="flex flex-col h-full min-h-0">
  <header class="border-b px-4 py-3 shrink-0">
    <h2 class="text-base font-semibold">Chats</h2>
  </header>

  <section class="flex-1 overflow-y-auto min-h-0">
    {#if conversations.length === 0}
      <div class="flex flex-col items-center justify-center h-full gap-2 text-muted-foreground">
        <MessageSquareIcon class="size-8 opacity-40" />
        <p class="text-sm">No conversations yet.</p>
      </div>

    {:else}
      <ul>
        {#each conversations as conversation (conversation.conversation_id)}
          <li>
            <a
              href="/chats/{conversation.conversation_id}/{encodeURIComponent(conversation.name ?? '')}"
              class="flex items-center gap-3 px-4 py-3 hover:bg-muted transition-colors border-b"
            >
              <div class="flex-1 min-w-0">
                <p class="font-medium text-sm truncate">{conversation.name ?? 'Unnamed'}</p>
                <p class="text-xs text-muted-foreground truncate">
                  {conversation.type === 'group' ? 'Group' : 'Private'}
                </p>
              </div>
              {#if conversation.last_message_at}
                <span class="text-xs text-muted-foreground shrink-0">
                  {formatDate(conversation.last_message_at)}
                </span>
              {/if}
            </a>
            
          </li>
        {/each}
      </ul>
    {/if}
  </section>
</div>