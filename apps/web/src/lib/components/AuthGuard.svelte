<!-- AuthGuard.svelte - Protects routes that require authentication -->
<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { currentUser } from '$lib/auth/store';
  import { onMount } from 'svelte';

  interface Props {
    children?: import("svelte").Snippet;
    requirePro?: boolean;
  }

  let { children, requirePro = false }: Props = $props();

  let loading = true;
  let redirecting = false;

  onMount(() => {
    const unsubscribe = currentUser.subscribe((user) => {
      console.log('AuthGuard: user state changed', user);
      loading = false;

      if (!user && !redirecting) {
        // Not authenticated - redirect to login with return URL
        redirecting = true;
        const returnUrl = encodeURIComponent($page.url.pathname + $page.url.search);
        console.log('AuthGuard: redirecting to login with returnUrl:', returnUrl);
        goto(`/login?returnUrl=${returnUrl}`);
        return;
      }

      if (user && requirePro && user.subscriptionTier !== 'pro' && !redirecting) {
        // User needs pro subscription
        redirecting = true;
        console.log('AuthGuard: redirecting to upgrade');
        goto('/upgrade');
        return;
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
{:else if $currentUser && (!requirePro || $currentUser.subscriptionTier === 'pro')}
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