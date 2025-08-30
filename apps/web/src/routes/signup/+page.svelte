<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { dev } from '$app/environment';
  import { signUp, user } from '$lib/auth/store';
  import { onMount } from 'svelte';
  
  let email = '';
  let password = '';
  let confirmPassword = '';
  let fullName = '';
  let username = '';
  let loading = false;
  let error: string | null = null;
  let success = false;

  // Get return URL from query params
  const returnUrl = $page.url.searchParams.get('returnUrl') || '/';

  // Redirect if already authenticated
  onMount(() => {
    const unsubscribe = user.subscribe((currentUser) => {
      if (currentUser) {
        goto(returnUrl);
      }
    });
    return unsubscribe;
  });

  async function handleSignUp() {
    if (!email || !password || !confirmPassword || !username) {
      error = 'Please fill in all required fields';
      return;
    }

    if (password !== confirmPassword) {
      error = 'Passwords do not match';
      return;
    }

    if (password.length < 8) {
      error = 'Password must be at least 8 characters long';
      return;
    }

    // Username validation
    if (username.length < 3) {
      error = 'Username must be at least 3 characters long';
      return;
    }

    if (!/^[a-zA-Z0-9_]+$/.test(username)) {
      error = 'Username can only contain letters, numbers, and underscores';
      return;
    }

    loading = true;
    error = null;
    success = false;

    try {
      const result = await signUp(email, password, { 
        fullName: fullName || undefined // Optional: only include if provided
      });
      
      if (result.error) {
        error = result.error.message;
      } else {
        success = true;
        // Note: With Supabase, user may need to verify email before login
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'An unexpected error occurred';
    } finally {
      loading = false;
    }
  }

  function handleLoginRedirect() {
    goto(`/login?returnUrl=${encodeURIComponent(returnUrl)}`);
  }
</script>

<svelte:head>
  <title>Sign Up - Noot</title>
  <meta name="description" content="Create your Noot account to start tracking your nutrition." />
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-base-200 py-12 px-4 sm:px-6 lg:px-8">
  <div class="max-w-md w-full space-y-8">
    <div>
      <h2 class="mt-6 text-center text-3xl font-extrabold text-base-content">
        Create your account
      </h2>
      <p class="mt-2 text-center text-sm text-base-content/70">
        Or 
        <button 
          type="button"
          class="link link-primary"
          on:click={handleLoginRedirect}
        >
          sign in to your existing account
        </button>
      </p>
    </div>

    {#if dev}
      <div class="alert alert-info">
        <div class="flex-1">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <span>
            <strong>Development Mode:</strong> Authentication may be using development settings.
            Check your environment variables for production deployment.
          </span>
        </div>
      </div>
    {/if}

    {#if success}
      <div class="alert alert-success">
        <div class="flex-1">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <span>
            <strong>Account created!</strong> Please check your email for a verification link before signing in.
          </span>
        </div>
      </div>
    {/if}

    <form on:submit|preventDefault={handleSignUp} class="space-y-6">
      <div>
        <label for="username" class="block text-sm font-medium text-base-content">
          Username *
        </label>
        <input
          id="username"
          name="username"
          type="text"
          bind:value={username}
          required
          class="input input-bordered w-full"
          placeholder="Enter your username"
          disabled={loading}
        />
      </div>

      <div>
        <label for="fullName" class="block text-sm font-medium text-base-content">
          Full Name
        </label>
        <input
          id="fullName"
          name="fullName"
          type="text"
          bind:value={fullName}
          class="input input-bordered w-full"
          placeholder="Enter your full name (optional)"
          disabled={loading}
        />
      </div>

      <div>
        <label for="email" class="block text-sm font-medium text-base-content">
          Email Address *
        </label>
        <input
          id="email"
          name="email"
          type="email"
          autocomplete="email"
          required
          bind:value={email}
          class="input input-bordered w-full"
          placeholder="your.email@example.com"
          disabled={loading}
        />
      </div>

      <div>
        <label for="password" class="block text-sm font-medium text-base-content">
          Password *
        </label>
        <input
          id="password"
          name="password"
          type="password"
          autocomplete="new-password"
          required
          bind:value={password}
          class="input input-bordered w-full"
          placeholder="At least 8 characters"
          disabled={loading}
        />
      </div>

      <div>
        <label for="confirmPassword" class="block text-sm font-medium text-base-content">
          Confirm Password *
        </label>
        <input
          id="confirmPassword"
          name="confirmPassword"
          type="password"
          autocomplete="new-password"
          required
          bind:value={confirmPassword}
          class="input input-bordered w-full"
          placeholder="Confirm your password"
          disabled={loading}
        />
      </div>

      {#if error}
        <div class="alert alert-error">
          <div class="flex-1">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <span>{error}</span>
          </div>
        </div>
      {/if}

      <div>
        <button
          type="submit"
          disabled={loading || success}
          class="btn btn-primary w-full"
          class:loading
        >
          {loading ? 'Creating account...' : 'Create account'}
        </button>
      </div>

      <div class="text-xs text-base-content/60 text-center">
        By creating an account, you agree to our terms of service and privacy policy.
      </div>
    </form>

    {#if success}
      <div class="text-center">
        <button
          type="button"
          class="btn btn-outline"
          on:click={handleLoginRedirect}
        >
          Go to Sign In
        </button>
      </div>
    {/if}
  </div>
</div>
