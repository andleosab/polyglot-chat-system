<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { connect, isConnected, disconnect } from '$lib/store/ws';
  import { currentUser } from '$lib/store/user';
  import * as Sidebar from '$lib/components/ui/sidebar/index.js';
  import AppSidebar from '$lib/components/app-sidebar.svelte';

  import { userSession } from '$lib/state/user-session.svelte';

  let { data, children } = $props();

  onMount(() => {
    currentUser.set(data.currentUser);
    connect($currentUser!.userId);
    userSession.currentUser = data.currentUser;
    // connect(userSession.currentUser!.userId);
    // Then next step is updating other files that 
    // import $lib/store/user (ChatView, etc.) to use userSession too.
  });

   onDestroy(() => {
    disconnect();
  }); 

</script>

<Sidebar.Provider class="h-svh overflow-hidden">
  <AppSidebar user={data.currentUser!} isConnected={$isConnected} />
  <Sidebar.Inset class="overflow-hidden">
    {@render children()}
  </Sidebar.Inset>
</Sidebar.Provider>