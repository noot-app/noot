<script lang="ts">
  import { goto } from '$app/navigation';
  import { dev } from '$app/environment';
  import { signIn, currentUser } from '$lib/auth/store';
  import { onMount } from 'svelte';
  
  let email = '';
  let password = '';
  let loading = false;
  let error: string | null = null;

  // Redirect if already authenticated
  onMount(() => {
    const unsubscribe = currentUser.subscribe((user) => {
      if (user) {
        goto('/');
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
        error = result.error.message;
      } else if (result.user) {
        // Successful login - redirect handled by onMount subscription
        console.log('Login successful');
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'An unexpected error occurred';
    } finally {
      loading = false;
    }
  }

  function handleSignUpRedirect() {
    goto('/signup');
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
          <label>
            <strong>Development Mode:</strong> Authentication may be using development settings.
            Check your environment variables for production deployment.
          </label>
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
            placeholder="Email address"
            disabled={loading}
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
            placeholder="Password"
            disabled={loading}
          />
        </div>
      </div>

      {#if error}
        <div class="alert alert-error">
          <div class="flex-1">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <label>{error}</label>
          </div>
        </div>
      {/if}

      <div class="flex items-center justify-between">
        <button
          type="button"
          class="link link-primary text-sm"
          on:click={handleResetPassword}
        >
          Forgot your password?
        </button>
      </div>

      <div>
        <button
          type="submit"
          disabled={loading}
          class="btn btn-primary w-full"
          class:loading
        >
          {loading ? 'Signing in...' : 'Sign in'}
        </button>
      </div>
    </form>
  </div>
</div>