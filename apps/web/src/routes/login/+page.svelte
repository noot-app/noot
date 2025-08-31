<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { dev } from '$app/environment';
  import { signIn, signInWithGitHub, user } from '$lib/auth/store';
  import { getAuthErrorMessage } from '$lib/auth/error-messages';
  import { onMount } from 'svelte';
  
  let email = '';
  let password = '';
  let loading = false;
  let githubLoading = false;
  let error: string | null = null;
  let redirecting = false; // Prevent multiple simultaneous redirects

  // Get return URL from query params
  const redirect = $page.url.searchParams.get('redirect') || '/';

  // Check for OAuth errors in URL
  onMount(() => {
    const errorParam = $page.url.searchParams.get('error');
    if (errorParam) {
      error = getAuthErrorMessage(errorParam, 'An authentication error occurred. Please try again.');
    }

    const unsubscribe = user.subscribe((currentUser) => {
      if (currentUser && !redirecting) {
        redirecting = true;
        loading = false; // Reset loading state on successful auth
        goto(redirect).catch((err) => {
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
      } else {
        // Successful sign-in. Perform explicit client redirect to avoid race conditions
        // with invalidateAll/onAuthStateChange timing.
        redirecting = true;
        goto(redirect).catch((err) => {
          console.error('❌ Navigation failed:', err);
          redirecting = false;
          loading = false;
        });
      }
    } catch (err) {
      console.error('❌ Exception in handleLogin:', err);
      error = err instanceof Error ? err.message : 'An unexpected error occurred';
      loading = false;
    }
  }

  function handleSignUpRedirect() {
    goto(`/signup?redirect=${encodeURIComponent(redirect)}`);
  }

  function handleResetPassword() {
    goto('/reset-password');
  }

  async function handleGitHubLogin() {
    if (githubLoading) return;
    
    githubLoading = true;
    error = null;

    try {
      const result = await signInWithGitHub(redirect);
      
      if (result.error) {
        error = getAuthErrorMessage(result.error.name, result.error.message);
        githubLoading = false;
      }
      // If successful, the OAuth flow will redirect to GitHub
      // and then back to our callback, so we don't reset loading here
    } catch (err) {
      console.error('❌ GitHub login error:', err);
      error = 'An error occurred during GitHub authentication. Please try again.';
      githubLoading = false;
    }
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

      <!-- GitHub OAuth Sign In -->
      <div class="mt-6">
        <div class="relative">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-base-300"></div>
          </div>
          <div class="relative flex justify-center text-sm">
            <span class="px-2 bg-base-200 text-base-content/70">Or continue with</span>
          </div>
        </div>

        <div class="mt-6">
          <button
            type="button"
            disabled={githubLoading || loading}
            class="btn btn-outline w-full relative"
            on:click={handleGitHubLogin}
            aria-describedby="github-button-status"
          >
            <span class:opacity-0={githubLoading}>
              <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                <path fill-rule="evenodd" d="M10 0C4.477 0 0 4.484 0 10.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0110 4.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.203 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.942.359.31.678.921.678 1.856 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0020 10.017C20 4.484 15.522 0 10 0z" clip-rule="evenodd"></path>
              </svg>
              Continue with GitHub
            </span>
            {#if githubLoading}
              <span class="loading loading-spinner loading-sm absolute" aria-hidden="true"></span>
              <span class="sr-only" id="github-button-status">Connecting to GitHub, please wait</span>
            {/if}
          </button>
        </div>
      </div>
    </form>
  </div>
</div>
