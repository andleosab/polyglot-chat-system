<script lang="ts">
  import { enhance } from '$app/forms';
  import GroupForm from '$lib/components/GroupForm.svelte';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import { Button } from '$lib/components/ui/button/index.js';

  let { data, params } = $props();
  let isSubmitting = $state(false);
</script>

<div class="flex flex-col h-full min-h-0">
  <header class="border-b px-4 py-3 shrink-0 flex items-center gap-3">
    <Button variant="ghost" size="icon-sm" href="/groups">
      <ArrowLeftIcon />
    </Button>
    <h2 class="text-base font-semibold">Edit Group</h2>
  </header>

  <section class="flex-1 overflow-y-auto min-h-0 p-4">
    <form
      method="POST"
      action="?/update"
      use:enhance={() => {
        isSubmitting = true;
        return async ({ update }) => {
          try {
            await update();
          } finally {
            isSubmitting = false;
          }          
        };
      }}
    >
      <GroupForm
        currentUserId={data.currentUser!.userId}
        createdBy={data.createdBy}
        groupName={data.groupName}
        originalParticipants={data.participants}
        existingParticipants={data.participants}
        allUsers={data.users}
        submitLabel="Save Changes"
        {isSubmitting}
      />
    </form>
  </section>
</div>