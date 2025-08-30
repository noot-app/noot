<!-- AuthGuard.svelte - DEPRECATED: No longer used with SSR auth -->
<script lang="ts">
  // AuthGuard.svelte - DEPRECATED: Client-side guard replaced by SSR auth in hooks.server.ts
  // This component is kept for potential future client-only auth scenarios but is not used
  // in the current SSR auth implementation. Protected routes are now server-side protected.
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { user } from '$lib/auth/store';
  import { onMount } from 'svelte';

  interface Props {
    children?: import("svelte").Snippet;
    requirePro?: boolean;
  }

  let { children, requirePro = false }: Props = $props();

  let loading = $state(true);
  let redirecting = $state(false);
  let authenticated = $state(false);

  onMount(() => {
    const unsubscribe = user.subscribe((currentUser) => {
      
      // If we have a user, stop loading and mark as authenticated
      if (currentUser) {
        loading = false;
        redirecting = false;
        authenticated = true;
        
        // Note: Pro requirement checking would need to be updated to fetch from user profile
        // since the new session system doesn't include subscription tier in the user object
        if (requirePro) {
          console.warn('Pro requirement checking not implemented in new auth system');
        }
        return;
      }
      
      // No user - handle redirect if not already redirecting
      loading = false;
      authenticated = false;
      if (!redirecting) {
        redirecting = true;
        const returnUrl = encodeURIComponent($page.url.pathname + $page.url.search);
        goto(`/login?returnUrl=${returnUrl}`);
      }
    });

    return unsubscribe;
  });
</script>

{#if loading}
  <div class="flex justify-center items-center min-h-screen">
    <div class="loading loading-spinner loading-lg"></div>
    <span class="ml-4 text-lg">Loading...</span>
  </div>
{:else if authenticated}
  {@render children?.()}
{:else}
  <!-- This shouldn't render as redirect should happen, but just in case -->
  <div class="flex justify-center items-center min-h-screen">
    <div class="text-center">
      <h2 class="text-xl font-bold mb-4">Authentication Required</h2>
      <p class="mb-4">Please sign in to access this page.</p>
      <a href="/login" class="btn btn-primary">Sign In</a>
    </div>
  </div>
{/if}
