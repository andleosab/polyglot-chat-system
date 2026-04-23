<!-- src/lib/components/ChatView.svelte -->
<script lang="ts">
  import { onMount, onDestroy, tick } from 'svelte';
  import { goto } from '$app/navigation';
  import SendIcon from '@lucide/svelte/icons/send';

  import { send, onceNewConversation, cancelNewConversationListener } from '$lib/store/ws';
  import { currentUser } from '$lib/store/user';
  import {
    messages,
    isLoading,
    hasMore,
    fetchError,
    seed,
    fetchOlderMessages,
    reset,
  } from '$lib/store/messages';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import type { ChatMessage, MessageResponse } from '$lib/api/types/message';

  // ─── Props ────────────────────────────────────────────────────────────────

  interface Props {
    conversationId?: number;
    chatName: string;
    initialMessages: MessageResponse[];
    targetUserUuid?: string | null;
    targetUsername?: string | null;
  }

  let {
    conversationId = $bindable<number>(),
    chatName,
    initialMessages,
    targetUserUuid = null,
    targetUsername = null,
  }: Props = $props();

  // ─── Local state ──────────────────────────────────────────────────────────

  let message = $state('');
  let isNewConvo = $derived(!conversationId && !!targetUserUuid);

  // ─── Send ─────────────────────────────────────────────────────────────────

  function handleSend(e: SubmitEvent) {
    e.preventDefault();
    if (!message.trim()) return;

    const body: ChatMessage = {
      id: crypto.randomUUID(),
      type: 'CHAT_MESSAGE',
      conversationId: conversationId,
      from: $currentUser!.userId,
      fromName: $currentUser!.username,
      to: isNewConvo ? (targetUserUuid ?? '') : '',
      toName: isNewConvo ? (targetUsername ?? '') : '',
      message: message,
      timestamp: Date.now(),
    };

    if (isNewConvo) {
      onceNewConversation((newId: number) => {
        conversationId = newId;
        goto(`/chats/${newId}/${encodeURIComponent(targetUsername ?? '')}`, {
          replaceState: true,
        });
      });
    }

    send(body);
    message = '';
    tick().then(() => {
      if (scrollContainer) scrollContainer.scrollTop = scrollContainer.scrollHeight;
    });
  }

  // ─── DOM refs ─────────────────────────────────────────────────────────────

  let scrollContainer = $state<HTMLElement | null>(null);
  let sentinel = $state<HTMLElement | null>(null);
  let observer: IntersectionObserver | null = null;

  // ─── Helpers ──────────────────────────────────────────────────────────────

  function isOwn(uuid: string) {
    return uuid === $currentUser?.userId;
  }

  function formatTime(iso: string) {
    return new Date(iso).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  // ─── Lifecycle ────────────────────────────────────────────────────────────

  onMount(() => {
    reset();

    (async () => {
      seed(conversationId!, [...initialMessages].reverse());
      await tick();

      if (scrollContainer) scrollContainer.scrollTop = scrollContainer.scrollHeight;

      if (!conversationId) return;

      let isFetching = false;

      observer = new IntersectionObserver(
        async (entries) => {
          if (!entries[0].isIntersecting || isFetching) return;

          isFetching = true;
          const prevScrollHeight = scrollContainer?.scrollHeight ?? 0;

          observer?.disconnect();
          const fetched = await fetchOlderMessages(conversationId!);

          if (fetched.length > 0) {
            await tick();
            if (scrollContainer) {
              scrollContainer.scrollTop = scrollContainer.scrollHeight - prevScrollHeight;
            }
          }

          isFetching = false;
          if (sentinel && $hasMore) observer?.observe(sentinel);
        },
        {
          root: scrollContainer,
          rootMargin: '120px',
          threshold: 0,
        }
      );

      if (sentinel && $hasMore) observer.observe(sentinel);
    })();
  });

  // Autoscroll on new WS messages if near bottom
  $effect(() => {
    $messages;
    if (!scrollContainer) return;
    const nearBottom =
      scrollContainer.scrollHeight - scrollContainer.scrollTop - scrollContainer.clientHeight < 80;
    if (nearBottom) {
      tick().then(() => {
        scrollContainer!.scrollTop = scrollContainer!.scrollHeight;
      });
    }
  });

  onDestroy(() => {
    reset();
    cancelNewConversationListener();
    observer?.disconnect();
  });
</script>

<div class="flex flex-col h-full min-h-0">

  <!-- Header -->
  <header class="bg-background border-b px-4 py-3 shrink-0 flex items-center gap-3">
    <h2 class="text-base font-semibold truncate">{chatName}</h2>
  </header>

  <!-- Message area -->
  <section
    bind:this={scrollContainer}
    class="flex-1 overflow-y-auto px-4 py-3 space-y-3 min-h-0"
  >
    <!-- Sentinel -->
    <div bind:this={sentinel}>
      {#if $isLoading}
        <div class="space-y-3 py-2">
          {#each [70, 50, 80] as w}
            <div class="flex items-center gap-2">
              <Skeleton class="h-8 rounded-xl" style="width:{w}%" />
            </div>
          {/each}
        </div>

      {:else if $fetchError}
        <div class="text-center py-2">
          <p class="text-destructive text-xs">{$fetchError}</p>
          <button
            onclick={() => conversationId && fetchOlderMessages(conversationId)}
            class="text-xs text-primary hover:underline mt-1"
          >Retry</button>
        </div>

      {:else if !$hasMore && $messages.length > 0}
        <p class="text-center text-xs text-muted-foreground py-2">
          Beginning of conversation
        </p>
      {/if}
    </div>

    <!-- Empty state for new private chat -->
    {#if isNewConvo && $messages.length === 0}
      <div class="flex flex-col items-center justify-center h-full text-center gap-2 py-12">
        <p class="text-sm text-muted-foreground">No messages yet.</p>
        <p class="text-xs text-muted-foreground">
          Say hello to <span class="font-medium text-foreground">{targetUsername}</span>!
        </p>
      </div>
    {/if}

    <!-- Messages -->
    {#each $messages as msg (msg.id)}
      {@const own = isOwn(msg.sender_uuid)}
      <div class="flex {own ? 'justify-end' : 'justify-start'} group">
        <div class="flex flex-col {own ? 'items-end' : 'items-start'} max-w-sm">

          {#if !own}
            <span class="text-xs text-muted-foreground mb-0.5 ml-1">
              {msg.sender_name}
            </span>
          {/if}

          <div class="px-3 py-2 rounded-2xl text-sm leading-relaxed {own
            ? 'bg-primary text-primary-foreground rounded-br-sm'
            : 'bg-muted text-foreground rounded-bl-sm'}">
            {msg.content}
          </div>

          <span class="text-[10px] text-muted-foreground mt-0.5 mx-1
                       opacity-0 group-hover:opacity-100 transition-opacity">
            {formatTime(msg.sent_at)}
          </span>

        </div>
      </div>
    {/each}
  </section>

  <!-- Input -->
  <footer class="border-t px-4 py-3 shrink-0 bg-background">
    <form onsubmit={handleSend} class="flex gap-2 items-center">
      <Input
        bind:value={message}
        type="text"
        placeholder="Type a message..."
        class="flex-1"
      />
      <Button type="submit" size="icon">
        <SendIcon />
      </Button>
    </form>
  </footer>

</div>