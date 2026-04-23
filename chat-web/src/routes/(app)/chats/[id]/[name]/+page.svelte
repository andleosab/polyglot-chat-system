<!-- src/routes/(app)/chats/[id]/[name]/+page.svelte -->
<script lang="ts">
  import ChatView from '$lib/components/ChatView.svelte';

  let { data, params } = $props();
</script>

<ChatView
  conversationId={data.conversationId}
  chatName={decodeURIComponent(params.name)}
  initialMessages={data.messages}
/>