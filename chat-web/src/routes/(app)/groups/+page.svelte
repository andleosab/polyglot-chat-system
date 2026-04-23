<script lang="ts">
  import type { PageProps } from './$types';
  import { Button } from '$lib/components/ui/button/index.js';
  import UsersRoundIcon from '@lucide/svelte/icons/users-round';
  import PlusIcon from '@lucide/svelte/icons/plus';
  import MessageSquareIcon from '@lucide/svelte/icons/message-square';

  let { data } = $props() as PageProps;
  let groups = $derived(data.groups);
</script>

<div class="flex flex-col h-full min-h-0">
  <header class="border-b px-4 py-3 shrink-0 flex items-center justify-between">
    <h2 class="text-base font-semibold">Groups</h2>
    <Button size="sm" href="/groups/new">
      <PlusIcon />
      New Group
    </Button>
  </header>

  <section class="flex-1 overflow-y-auto min-h-0">
    {#if groups.length === 0}
      <div class="flex flex-col items-center justify-center h-full gap-2 text-muted-foreground">
        <UsersRoundIcon class="size-8 opacity-40" />
        <p class="text-sm">No groups yet.</p>
      </div>
    {:else}
      <ul>
        {#each groups as group (group.conversation_id)}
          <li class="flex items-center border-b hover:bg-muted transition-colors">
            <a
              href="/groups/{group.conversation_id}/{encodeURIComponent(group.name ?? '')}"
              class="flex-1 px-4 py-3 min-w-0"
            >
              <p class="font-medium text-sm truncate">{group.name ?? 'Unnamed'}</p>
              <p class="text-xs text-muted-foreground">Group</p>
            </a>
            <Button
              href="/chats/{group.conversation_id}/{encodeURIComponent(group.name ?? '')}"
              variant="ghost"
              size="icon-sm"
              class="mr-3 shrink-0"
            >
              <MessageSquareIcon />
            </Button>
          </li>
        {/each}
      </ul>
    {/if}
  </section>
</div>