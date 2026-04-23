<script lang="ts">
  import MessageSquareIcon from '@lucide/svelte/icons/message-square';
  import UsersIcon from '@lucide/svelte/icons/users';
  import UsersRoundIcon from '@lucide/svelte/icons/users-round';
  import CircleIcon from '@lucide/svelte/icons/circle';
  import * as Sidebar from '$lib/components/ui/sidebar/index.js';
  import * as Avatar from '$lib/components/ui/avatar/index.js';
  import type { CurrentUser } from '$lib/store/user';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import LogOutIcon from '@lucide/svelte/icons/log-out';

  import { client } from "$lib/auth-client";
  import { goto } from "$app/navigation";
  import { env } from '$env/dynamic/public';

  let { user, isConnected }: { user: CurrentUser; isConnected: boolean } = $props();

  const navItems = [
    { title: 'Chats',  url: '/chats',  icon: MessageSquareIcon },
    { title: 'Users',  url: '/users',  icon: UsersIcon },
    { title: 'Groups', url: '/groups', icon: UsersRoundIcon },
  ];

  // Generate initials from username
  function initials(name: string) {
    return name.slice(0, 2).toUpperCase();
  }
</script>

<Sidebar.Root collapsible="icon">

  <Sidebar.Header>
    <Sidebar.Menu>
      <Sidebar.MenuItem>
        <Sidebar.MenuButton size="lg">
          <MessageSquareIcon class="size-5" />
          <span class="font-semibold text-base">Chat App</span>
        </Sidebar.MenuButton>
      </Sidebar.MenuItem>
    </Sidebar.Menu>
  </Sidebar.Header>

  <Sidebar.Content>
    <Sidebar.Group>
      <Sidebar.GroupContent>
        <Sidebar.Menu>
          {#each navItems as item (item.title)}
            <Sidebar.MenuItem>
              <Sidebar.MenuButton tooltipContent={item.title}>
                {#snippet child({ props })}
                  <a href={item.url} {...props}>
                    <item.icon />
                    <span>{item.title}</span>
                  </a>
                {/snippet}
              </Sidebar.MenuButton>
            </Sidebar.MenuItem>
          {/each}
        </Sidebar.Menu>
      </Sidebar.GroupContent>
    </Sidebar.Group>
  </Sidebar.Content>

<Sidebar.Footer>
  <Sidebar.Menu>
    <Sidebar.MenuItem>
      <DropdownMenu.Root>
        <DropdownMenu.Trigger>
          {#snippet child({ props })}
            <Sidebar.MenuButton size="lg" tooltipContent={user.username} {...props}>
              <Avatar.Root class="size-7 rounded-lg">
                <Avatar.Fallback class="rounded-lg text-xs">
                  {initials(user.username)}
                </Avatar.Fallback>
              </Avatar.Root>
              <div class="flex flex-col flex-1 text-left text-sm leading-tight">
                <span class="truncate font-medium">{user.username}</span>
                <!--
                <span class="truncate text-xs text-muted-foreground">{user.email}</span>
              -->
              </div>
              <span class="flex items-center gap-1 text-xs text-muted-foreground">
                <CircleIcon class="size-2 {isConnected
                  ? 'fill-green-500 text-green-500'
                  : 'fill-red-500 text-red-500'}"
                />
                {isConnected ? 'Online' : 'Offline'}
              </span>
            </Sidebar.MenuButton>
          {/snippet}
        </DropdownMenu.Trigger>
        <DropdownMenu.Content side="top" class="w-(--bits-dropdown-menu-anchor-width)">
          <DropdownMenu.Item>
            {#snippet child({ props })}
              <button
                {...props}
                class="flex items-center gap-2 {props.class ?? ''}"
                onclick={async () => {
                  try {
                    await client.signOut();
                    goto('/sign-in'); // optional redirect after signout
                  } catch (err) {
                    console.error('Sign out failed', err);
                  }
                }}
              >
                <LogOutIcon class="size-4" />
                Sign out
              </button>     
            {/snippet}
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    </Sidebar.MenuItem>
  </Sidebar.Menu>
</Sidebar.Footer>

  <Sidebar.Rail />

</Sidebar.Root>