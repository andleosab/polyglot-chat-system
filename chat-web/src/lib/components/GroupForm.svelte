<script lang="ts">
  import { Input } from '$lib/components/ui/input/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import SearchIcon from '@lucide/svelte/icons/search';
  import XIcon from '@lucide/svelte/icons/x';
  import * as Avatar from '$lib/components/ui/avatar/index.js';
  import type { CreateUserResponse } from '$lib/api/types/user';
  import type { Participant } from '$lib/api/types/conversation';
    import type { CurrentUser } from '$lib/store/user';

  // ─── Internal selected type ───────────────────────────────────────────────
  // Unified shape used internally — mapped from either source

  interface SelectedUser {
    userId: string;
    username: string;
    email?: string;
    isAdmin?: boolean;
  }

  // ─── Props ────────────────────────────────────────────────────────────────

  interface Props {
    currentUserId: string;
    createdBy: string;                 
    groupName?: string;
    originalParticipants?: Participant[];
    existingParticipants?: Participant[];  // from BE /groups/:id/participants
    allUsers: CreateUserResponse[];        // from BE /api/users — required
    submitLabel?: string;
    isSubmitting?: boolean;
  }

  let {
    currentUserId,
    createdBy,
    groupName = '',
    originalParticipants = [],
    existingParticipants = [],
    allUsers,
    submitLabel = 'Create Group',
    isSubmitting = false,
  }: Props = $props();

  // ─── State ────────────────────────────────────────────────────────────────

  let name = $state(groupName);
  let searchQuery = $state('');

  // Map existing Participant[] → SelectedUser[] on init
  let selected = $state<SelectedUser[]>(
    existingParticipants.map(p => ({
      userId: p.user_uuid,
      username: p.username,
      isAdmin: p.isAdmin,
    }))
  );

  // console.log('Initial selected:', selected);

  // ─── Local filtering ──────────────────────────────────────────────────────

  let searchResults = $derived(
    searchQuery.trim().length > 0
      ? allUsers.filter(u =>
          u.username.toLowerCase().includes(searchQuery.trim().toLowerCase()) &&
          !isSelected(u.userId)
        )
      : []
  );

  // ─── Helpers ──────────────────────────────────────────────────────────────

  function initials(username: string) {
    return username.slice(0, 2).toUpperCase();
  }

  function isSelected(userId: string) {
    return selected.some(p => p.userId === userId);
  }

  function addParticipant(u: CreateUserResponse) {
    if (!isSelected(u.userId)) {
      selected = [...selected, {
        userId: u.userId,
        username: u.username,
        email: u.email,
      }];
    }
    searchQuery = '';
  }

  function removeParticipant(userId: string) {
    selected = selected.filter(p => p.userId !== userId);
  }
</script>

<div class="flex flex-col gap-5">

  <!-- Group name -->
  <div class="flex flex-col gap-1.5">
    <label class="text-sm font-medium" for="group-name">Group name</label>
    <Input
      id="group-name"
      name="name"
      bind:value={name}
      placeholder="e.g. Design Team"
      required
    />
  </div>

  <!-- Participant search -->
  <div class="flex flex-col gap-1.5">
    <label class="text-sm font-medium" for="participant-search">Add participants</label>
    <div class="relative">
      <SearchIcon class="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
      <Input
        id="participant-search"
        bind:value={searchQuery}
        placeholder="Search by name..."
        class="pl-8"
      />
    </div>

    {#if searchResults.length > 0}
      <ul class="border rounded-lg divide-y mt-1">
        {#each searchResults as result (result.userId)}
          <li>
            <button
              type="button"
              onclick={() => addParticipant(result)}
              class="flex items-center gap-2 w-full px-3 py-2 text-left hover:bg-muted transition-colors"
            >
              <Avatar.Root class="size-7 rounded-lg shrink-0">
                <Avatar.Fallback class="rounded-lg text-xs">
                  {initials(result.username)}
                </Avatar.Fallback>
              </Avatar.Root>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium truncate">{result.username}</p>
                <p class="text-xs text-muted-foreground truncate">{result.email}</p>
              </div>
            </button>
          </li>
        {/each}
      </ul>
    {:else if searchQuery.trim()}
      <p class="text-xs text-muted-foreground mt-1 px-1">No users found.</p>
    {/if}
  </div>

  <!-- Selected participants -->
  {#if selected.length > 0}
    <div class="flex flex-col gap-1.5">
      <p class="text-sm font-medium">Participants ({selected.length})</p>
      <ul class="flex flex-col gap-1">
        {#each selected as participant (participant.userId)}
          <li class="flex items-center gap-2 px-2 py-1.5 rounded-lg bg-muted">
            <Avatar.Root class="size-6 rounded-md shrink-0">
              <Avatar.Fallback class="rounded-md text-xs">
                {initials(participant.username)}
              </Avatar.Fallback>
            </Avatar.Root>
            <span class="flex-1 text-sm truncate">{participant.username}</span>
            {#if participant.userId === createdBy}
              <span class="text-xs text-muted-foreground">owner</span>
            {:else}   
              {#if currentUserId === createdBy}
                <button
                  type="button"
                  onclick={() => removeParticipant(participant.userId)}
                  class="text-muted-foreground hover:text-foreground transition-colors"
                >
                  <XIcon class="size-3.5" />
                </button>              
              {/if}         

            {/if}            
          </li>
        {/each}
      </ul>
    </div>
  {/if}

  <!-- what exists on the server now. preserve for updates -->
  <!-- rename this to existingParticipants  and existingParticipants to modifiedParticipants later ! -->  
  <input
    type="hidden"
    name="originalParticipants"
    value={JSON.stringify(originalParticipants)}
  />  

  <!-- Hidden input carries selected participants to server as Participant[] shape -->
  <input
    type="hidden"
    name="participants"
    value={JSON.stringify(selected.map(p => ({
      user_uuid: p.userId,
      username: p.username,
      is_admin: p.isAdmin ?? false,
    })))}
  />

  <Button type="submit" disabled={isSubmitting || !name.trim()}>
    {isSubmitting ? 'Saving...' : submitLabel}
  </Button>

</div>