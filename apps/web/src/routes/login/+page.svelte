<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { dev } from '$app/environment';
  import { signIn, user } from '$lib/auth/store';
  import { getAuthErrorMessage } from '$lib/auth/error-messages';
  import { onMount } from 'svelte';
  
  let email = '';
  let password = '';
  let loading = false;
  let error: string | null = null;
  let redirecting = false; // Prevent multiple simultaneous redirects

  // Get return URL from query params
  const returnUrl = $page.url.searchParams.get('returnUrl') || '/';

  // Redirect if already authenticated
  onMount(() => {
    const unsubscribe = user.subscribe((currentUser) => {
      if (currentUser && !redirecting) {
        redirecting = true;
        loading = false; // Reset loading state on successful auth
        goto(returnUrl).catch((err) => {
          console.error('❌ Navigation failed:', err);
          redirecting = false; // Reset on failure
          loading = false;
        });
      }
    });
    return unsubscribe;
  });

  async function handleLogin() {
    if (!email || !password) {
      error = 'Please fill in all fields';
      return;
    }

    loading = true;
    error = null;

    try {
      const result = await signIn(email, password);
      
      if (result.error) {
        // Get the error message, handling both string and Error object types
        const errorMessage = result.error.message;

        // log the error
        console.warn('❌ Login failure:', errorMessage);
        
        // Try to parse error as JSON to get error code, otherwise use message
        let errorCode = null;
        try {
          const errorData = JSON.parse(errorMessage);
          errorCode = errorData.code;
        } catch {
          // If not JSON, we'll handle it in the default case
        }
        
        // Get user-friendly error message
        error = getAuthErrorMessage(errorCode, errorMessage);
        loading = false;
      }
      // If no error, sign-in was successful - auth state change listener will handle redirect
      // Don't set loading = false here, let the redirect happen
    } catch (err) {
      console.error('❌ Exception in handleLogin:', err);
      error = err instanceof Error ? err.message : 'An unexpected error occurred';
      loading = false;
    }
  }

  function handleSignUpRedirect() {
    goto(`/signup?returnUrl=${encodeURIComponent(returnUrl)}`);
  }

  function handleResetPassword() {
    goto('/reset-password');
  }
</script>

<svelte:head>
  <title>Login - Noot</title>
  <meta name="description" content="Sign in to your Noot account to track your nutrition." />
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-base-200 py-12 px-4 sm:px-6 lg:px-8">
  <div class="max-w-md w-full space-y-8">
    <div>
      <h2 class="mt-6 text-center text-3xl font-extrabold text-base-content">
        Sign in to your account
      </h2>
      <p class="mt-2 text-center text-sm text-base-content/70">
        Or 
        <button 
          type="button"
          class="link link-primary"
          on:click={handleSignUpRedirect}
        >
          create a new account
        </button>
      </p>
    </div>

    {#if dev}
      <div class="alert alert-info">
        <div class="flex-1">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <div>
            <strong>Development Mode:</strong> Authentication may be using development settings.
            Check your environment variables for production deployment.
          </div>
        </div>
      </div>
    {/if}

    <form class="mt-8 space-y-6" on:submit|preventDefault={handleLogin}>
      <div class="rounded-md shadow-sm -space-y-px">
        <div>
          <label for="email" class="sr-only">Email address</label>
          <input
            id="email"
            name="email"
            type="email"
            autocomplete="email"
            required
            bind:value={email}
            class="input input-bordered w-full rounded-t-md rounded-b-none"
            class:input-error={error}
            placeholder="Email address"
            disabled={loading}
            aria-describedby={error ? "error-message" : undefined}
          />
        </div>
        <div>
          <label for="password" class="sr-only">Password</label>
          <input
            id="password"
            name="password"
            type="password"
            autocomplete="current-password"
            required
            bind:value={password}
            class="input input-bordered w-full rounded-t-none rounded-b-md"
            class:input-error={error}
            placeholder="Password"
            disabled={loading}
            aria-describedby={error ? "error-message" : undefined}
          />
        </div>
      </div>

      <!-- Fixed height error container to prevent layout shift -->
      <div class="min-h-[4rem] flex items-start">
        {#if error}
          <div 
            id="error-message"
            class="alert alert-error w-full"
            role="alert"
            aria-live="polite"
          >
            <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <span>{error}</span>
          </div>
        {/if}
      </div>

      <div class="flex items-center justify-between">
        <button
          type="button"
          class="link link-primary text-sm"
          on:click={handleResetPassword}
          disabled={loading}
        >
          Forgot your password?
        </button>
      </div>

      <div>
        <button
          type="submit"
          disabled={loading}
          class="btn btn-primary w-full relative"
          aria-describedby="button-status"
        >
          <span class:opacity-0={loading}>Sign in</span>
          {#if loading}
            <span class="loading loading-spinner loading-sm absolute" aria-hidden="true"></span>
            <span class="sr-only" id="button-status">Signing in, please wait</span>
          {/if}
        </button>
      </div>
    </form>
  </div>
</div>
