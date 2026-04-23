<script lang="ts">
  import ChatView from '$lib/components/ChatView.svelte';
  let { data } = $props();
</script>

<ChatView
  chatName={data.targetUsername}
  initialMessages={[]}
  targetUserUuid={data.targetUserUuid}
  targetUsername={data.targetUsername}
/>