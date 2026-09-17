<script lang="ts">
  import { onMount } from 'svelte';
  import { me } from './lib/api.js';
  import { isAuthenticated, username } from './lib/stores.js';
  import Login from './routes/Login.svelte';
  import Dashboard from './routes/Dashboard.svelte';

  let ready = $state(false);

  onMount(async () => {
    const result = await me();
    if (result.username) {
      username.set(result.username);
      isAuthenticated.login();
    }
    ready = true;
  });
</script>

{#if ready}
  {#if $isAuthenticated}
    <Dashboard />
  {:else}
    <Login />
  {/if}
{/if}
