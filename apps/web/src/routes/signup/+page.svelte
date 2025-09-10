<script lang="ts">
  import { goto } from "$app/navigation"
  import { page } from "$app/stores"
  import { dev } from "$app/environment"
  import {
    signUp,
    signInWithGitHub,
    signInWithGoogle,
    user,
  } from "$lib/auth/store"
  import { getAuthErrorMessage } from "$lib/auth/error-messages"
  import { onMount } from "svelte"
  import { DEFAULT_REDIRECT_PATH, getRedirectParam } from "$lib/utils/redirect"
  import Alert from "$lib/components/Alert.svelte"
  import DevModeAlert from "$lib/components/DevModeAlert.svelte"

  let email = ""
  let password = ""
  let confirmPassword = ""
  let fullName = ""
  let username = ""
  let loading = false
  let githubLoading = false
  let googleLoading = false
  let error: string | null = null
  let success = false

  // Get return URL from query params
  const redirect = getRedirectParam($page.url, DEFAULT_REDIRECT_PATH)

  // Redirect if already authenticated
  onMount(() => {
    const errorParam = $page.url.searchParams.get("error")
    if (errorParam) {
      error = getAuthErrorMessage(
        errorParam,
        "An authentication error occurred. Please try again.",
      )
    }
    const hash = $page.url.hash || ""
    if (hash.includes("error=")) {
      const params = new URLSearchParams(hash.replace(/^#/, ""))
      const oauthError = params.get("error_code") || params.get("error")
      if (oauthError) {
        error = getAuthErrorMessage(
          oauthError,
          "An authentication error occurred. Please try again.",
        )
      }
    }
    const unsubscribe = user.subscribe((currentUser) => {
      if (currentUser) {
        goto(redirect)
      }
    })
    return unsubscribe
  })

  async function handleSignUp() {
    if (!email || !password || !confirmPassword || !username) {
      error = "Please fill in all required fields"
      return
    }

    if (password !== confirmPassword) {
      error = "Passwords do not match"
      return
    }

    if (password.length < 8) {
      error = "Password must be at least 8 characters long"
      return
    }

    // Username validation
    if (username.length < 3) {
      error = "Username must be at least 3 characters long"
      return
    }

    if (!/^[a-zA-Z0-9_]+$/.test(username)) {
      error = "Username can only contain letters, numbers, and underscores"
      return
    }

    loading = true
    error = null
    success = false

    try {
      const result = await signUp(email, password, {
        fullName: fullName || undefined, // Optional: only include if provided
      })

      if (result.error) {
        error = result.error.message
      } else {
        success = true
        // Note: With Supabase, user may need to verify email before login
      }
    } catch (err) {
      error =
        err instanceof Error ? err.message : "An unexpected error occurred"
    } finally {
      loading = false
    }
  }

  function handleLoginRedirect() {
    goto(`/login?redirect=${encodeURIComponent(redirect)}`)
  }

  function handleResetPassword() {
    goto("/reset-password")
  }

  async function handleGitHubSignup() {
    if (githubLoading) return
    githubLoading = true
    error = null

    try {
      const result = await signInWithGitHub(redirect)
      if (result.error) {
        error = getAuthErrorMessage(result.error.name, result.error.message)
        githubLoading = false
      }
      // Success will redirect to GitHub and back via /auth/callback
    } catch (err) {
      console.error("✌️ GitHub signup error:", err)
      error = "An error occurred during authentication. Please try again."
      githubLoading = false
    }
  }

  async function handleGoogleSignup() {
    if (googleLoading) return
    googleLoading = true
    error = null

    try {
      const result = await signInWithGoogle(redirect)
      if (result.error) {
        error = getAuthErrorMessage(result.error.name, result.error.message)
        googleLoading = false
      }
      // Success will redirect to Google and back via /auth/callback
    } catch (err) {
      console.error("✌️ Google signup error:", err)
      error = "An error occurred during authentication. Please try again."
      googleLoading = false
    }
  }
</script>

<svelte:head>
  <title>Sign Up - Noot</title>
  <meta
    name="description"
    content="Create your Noot account to start tracking your nutrition."
  />
</svelte:head>

<div
  class="min-h-screen flex items-center justify-center bg-base-200 py-12 px-4 sm:px-6 lg:px-8"
>
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

    <DevModeAlert />

    {#if success}
      <Alert type="success">
        <strong>Account created!</strong> Please check your email for a verification
        link before signing in.
      </Alert>
    {/if}

    <!-- OAuth sign up options -->
    <div class="space-y-4">
      <button
        type="button"
        disabled={githubLoading || loading}
        class="btn btn-outline w-full relative"
        on:click={handleGitHubSignup}
        aria-describedby="github-button-status"
      >
        <span
          class="inline-flex items-center justify-center gap-2 whitespace-nowrap"
          class:opacity-0={githubLoading}
        >
          <svg
            class="w-5 h-5"
            fill="currentColor"
            viewBox="0 0 20 20"
            aria-hidden="true"
          >
            <path
              fill-rule="evenodd"
              d="M10 0C4.477 0 0 4.484 0 10.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0110 4.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.203 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.942.359.31.678.921.678 1.856 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0020 10.017C20 4.484 15.522 0 10 0z"
              clip-rule="evenodd"
            ></path>
          </svg>
          Continue with GitHub
        </span>
        {#if githubLoading}
          <span
            class="loading loading-spinner loading-sm absolute"
            aria-hidden="true"
          ></span>
          <span class="sr-only" id="github-button-status"
            >Connecting to GitHub, please wait</span
          >
        {/if}
      </button>

      <!-- Placeholder for Google (coming soon) -->
      <button
        type="button"
        disabled={googleLoading || loading}
        class="btn btn-outline w-full relative"
        on:click={handleGoogleSignup}
        aria-describedby="google-button-status"
      >
        <span
          class="inline-flex items-center justify-center gap-2 whitespace-nowrap"
          class:opacity-0={googleLoading}
        >
          <svg class="w-5 h-5" viewBox="0 0 48 48" aria-hidden="true">
            <path
              fill="#FFC107"
              d="M43.611 20.083H42V20H24v8h11.303C33.983 32.91 29.369 36 24 36c-6.627 0-12-5.373-12-12s5.373-12 12-12c3.059 0 5.842 1.156 7.961 3.039l5.657-5.657C34.871 6.053 29.718 4 24 4 12.955 4 4 12.955 4 24s8.955 20 20 20 20-8.955 20-20c0-1.341-.138-2.651-.389-3.917z"
            />
            <path
              fill="#FF3D00"
              d="M6.306 14.691l6.571 4.819C14.239 16.15 18.793 12 24 12c3.059 0 5.842 1.156 7.961 3.039l5.657-5.657C34.871 6.053 29.718 4 24 4 16.318 4 9.716 8.337 6.306 14.691z"
            />
            <path
              fill="#4CAF50"
              d="M24 44c5.304 0 10.165-2.033 13.828-5.343l-6.383-5.396C29.435 34.203 26.863 35.2 24 35.2c-5.334 0-9.845-3.417-11.469-8.147l-6.56 5.056C8.35 38.614 15.627 44 24 44z"
            />
            <path
              fill="#1976D2"
              d="M43.611 20.083H42V20H24v8h11.303c-1.688 4.91-6.302 8-11.303 8-5.334 0-9.845-3.417-11.469-8.147l-6.56 5.056C8.35 38.614 15.627 44 24 44c11.045 0 20-8.955 20-20 0-1.341-.138-2.651-.389-3.917z"
            />
          </svg>
          Continue with Google
        </span>
        {#if googleLoading}
          <span
            class="loading loading-spinner loading-sm absolute"
            aria-hidden="true"
          ></span>
          <span class="sr-only" id="google-button-status"
            >Connecting to Google, please wait</span
          >
        {/if}
      </button>
    </div>

    <!-- Divider -->
    <div class="relative my-6">
      <div class="absolute inset-0 flex items-center">
        <div class="w-full border-t border-base-300"></div>
      </div>
      <div class="relative flex justify-center text-sm">
        <span class="px-2 bg-base-200 text-base-content/70"
          >Or sign up with email</span
        >
      </div>
    </div>

    <!-- Email sign up form -->
    <form on:submit|preventDefault={handleSignUp} class="space-y-6">
      <div>
        <label
          for="username"
          class="block text-sm font-medium text-base-content"
        >
          Username *
        </label>
        <input
          id="username"
          name="username"
          type="text"
          bind:value={username}
          required
          autocomplete="username"
          class="input input-bordered w-full"
          placeholder="Enter your username"
          disabled={loading}
        />
      </div>

      <div>
        <label
          for="fullName"
          class="block text-sm font-medium text-base-content"
        >
          Full Name
        </label>
        <input
          id="fullName"
          name="fullName"
          type="text"
          bind:value={fullName}
          autocomplete="name"
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
        <label
          for="password"
          class="block text-sm font-medium text-base-content"
        >
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
        <label
          for="confirmPassword"
          class="block text-sm font-medium text-base-content"
        >
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
        <Alert type="error">
          {error}
        </Alert>
      {/if}

      <div>
        <button
          type="submit"
          disabled={loading || success}
          class="btn btn-primary w-full"
          class:loading
        >
          {loading ? "Creating account..." : "Create account"}
        </button>
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

      <div class="text-xs text-base-content/60 text-center">
        By creating an account, you agree to our terms of service and privacy
        policy.
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
