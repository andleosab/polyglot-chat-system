<!-- src/routes/(app)/users/+page.svelte -->
<script lang="ts">
  import type { PageProps } from './$types';
  import type { CreateUserResponse } from '$lib/api/types/user';
  import type { PresenceInfo } from '$lib/api/types/presence';
  import * as Avatar from '$lib/components/ui/avatar/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import MessageSquareIcon from '@lucide/svelte/icons/message-square';
  import UsersIcon from '@lucide/svelte/icons/users';
  import { userCreated, presenceUpdate } from '$lib/store/ws';

  let { data } = $props() as PageProps;
  let wsUsers = $state<CreateUserResponse[]>([]);
  let users = $derived([...(data.users ?? []), ...wsUsers]);
  let ssrPresenceMap = $derived(
    (data.presenceMap ?? {}) as Record<string, { online: boolean; last_seen: string | null }>
  );
  let livePresence = $state<Record<string, { online: boolean; last_seen: string | null }>>({});
  let presenceMap = $derived({ ...ssrPresenceMap, ...livePresence });

  $effect(() => {
    const newUser = $userCreated;
    if (newUser && !users.some(u => u.userId === newUser.userId)) {
      wsUsers = [...wsUsers, {
        userId: newUser.userId,
        username: newUser.username,
        email: newUser.email,
        isActive: true,
        createdAt: new Date(newUser.timestamp).toISOString(),
        updatedAt: ''
      }];
    }
  });

  $effect(() => {
    const update = $presenceUpdate;
    if (!update) return;
    const online = update.type === 'USER_JOINED';
    livePresence = {
      ...livePresence,
      [update.from]: {
        online,
        last_seen: online ? null : new Date(update.timestamp).toISOString()
      }
    };
  });

  function initials(name: string) {
    return name.slice(0, 2).toUpperCase();
  }

  function relativeTime(iso: string | null): string {
    if (!iso) return 'a while ago';
    const diffMs = Date.now() - new Date(iso).getTime();
    const diffMins = Math.floor(diffMs / 60_000);
    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;
    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d ago`;
  }
</script>

<div class="flex flex-col h-full min-h-0">
  <header class="border-b px-4 py-3.5 shrink-0">
    <h1 class="text-base font-semibold">Users</h1>
  </header>

  <section class="flex-1 overflow-y-auto min-h-0">
    {#if users.length === 0}
      <div class="flex flex-col items-center justify-center h-full gap-3 text-muted-foreground">
        <div class="size-14 rounded-2xl bg-muted flex items-center justify-center">
          <UsersIcon class="size-6 opacity-50" />
        </div>
        <div class="text-center">
          <p class="text-sm font-medium text-foreground">No users yet</p>
          <p class="text-xs mt-0.5">Users will appear here when they sign up</p>
        </div>
      </div>
    {:else}
      <ul class="divide-y divide-border/60">
        {#each users as user (user.userId)}
          {@const presence = presenceMap[user.userId]}
          {@const isOnline = presence?.online === true}
          {@const lastSeen = presence?.last_seen ?? null}
          <li class="flex items-center gap-3 px-4 py-3 hover:bg-muted/50 transition-colors">
            <div class="relative shrink-0">
              <Avatar.Root class="size-10 rounded-xl">
                <Avatar.Fallback class="rounded-xl text-sm font-medium bg-primary/10 text-primary">
                  {initials(user.username)}
                </Avatar.Fallback>
              </Avatar.Root>
              {#if isOnline}
                <span class="absolute bottom-0 right-0 size-2.5 rounded-full bg-green-500 ring-2 ring-background"></span>
              {/if}
            </div>

            <div class="flex-1 min-w-0">
              <p class="font-medium text-sm truncate">{user.username}</p>
              {#if isOnline || (!isOnline && !lastSeen)}
                <p class="text-xs text-muted-foreground truncate">{user.email}</p>
              {:else}
                <p class="text-xs text-muted-foreground truncate">Last seen {relativeTime(lastSeen)}</p>
              {/if}
            </div>

            <Button
              href="/chats/new/{user.userId}/{user.username}"
              variant="ghost"
              size="icon-sm"
              class="shrink-0 rounded-lg text-muted-foreground hover:text-foreground"
            >
              <MessageSquareIcon class="size-4" />
            </Button>
          </li>
        {/each}
      </ul>
    {/if}
  </section>
</div>
