<!-- AuthGuard.svelte - Protects routes that require authentication -->
<script lang="ts">
  // AuthGuard.svelte - Client-side guard for authenticated/pro users.
  // NOTE: This is a UX convenience only (redirects, hides UI).
  // Real access control must be enforced on the server/API with JWT/RLS checks.
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { currentUser } from '$lib/auth/store';
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
    const unsubscribe = currentUser.subscribe((user) => {
      console.log('AuthGuard: user state changed', user);
      
      // If we have a user, stop loading and mark as authenticated
      if (user) {
        loading = false;
        redirecting = false;
        authenticated = true;
        
        // Check if user meets pro requirement
        if (requirePro && user.subscriptionTier !== 'pro') {
          authenticated = false;
          redirecting = true;
          console.log('AuthGuard: redirecting to upgrade');
          goto('/upgrade');
          return;
        }
        return;
      }
      
      // No user - handle redirect if not already redirecting
      loading = false;
      authenticated = false;
      if (!redirecting) {
        redirecting = true;
        const returnUrl = encodeURIComponent($page.url.pathname + $page.url.search);
        console.log('AuthGuard: redirecting to login with returnUrl:', returnUrl);
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
