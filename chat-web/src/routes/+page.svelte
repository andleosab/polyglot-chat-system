<script>
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';

  console.log('==> in root +page.svelte');

  onMount(() => {
    console.log('==> in root +page.svelte onMount: Redirecting to /chats');
    goto('/chats');
  });
</script>