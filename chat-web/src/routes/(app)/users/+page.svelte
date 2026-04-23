<!-- src/routes/(app)/users/+page.svelte -->
<script lang="ts">
  import type { PageProps } from './$types';
  import * as Avatar from '$lib/components/ui/avatar/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import MessageSquareIcon from '@lucide/svelte/icons/message-square';

  let { data } = $props() as PageProps;
  let users = $derived(data.users);

  function initials(name: string) {
    return name.slice(0, 2).toUpperCase();
  }
</script>

<div class="flex flex-col h-full min-h-0">
  <header class="border-b px-4 py-3 shrink-0">
    <h2 class="text-base font-semibold">Users</h2>
  </header>

  <section class="flex-1 overflow-y-auto min-h-0">
    {#if users.length === 0}
      <div class="flex flex-col items-center justify-center h-full gap-2 text-muted-foreground">
        <p class="text-sm">No users found.</p>
      </div>
    {:else}
      <ul>
        {#each users as user (user.userId)}
          <li class="flex items-center gap-3 px-4 py-3 border-b hover:bg-muted transition-colors">
            <Avatar.Root class="size-8 rounded-lg shrink-0">
              <Avatar.Fallback class="rounded-lg text-xs">
                {initials(user.username)}
              </Avatar.Fallback>
            </Avatar.Root>

            <div class="flex-1 min-w-0">
              <p class="font-medium text-sm truncate">{user.username}</p>
              <p class="text-xs text-muted-foreground truncate">{user.email}</p>
            </div>

            <Button
              href="/chats/new/{user.userId}/{user.username}"
              variant="ghost"
              size="icon-sm"
            >
              <MessageSquareIcon />
            </Button>
          </li>
        {/each}
      </ul>
    {/if}
  </section>
</div>