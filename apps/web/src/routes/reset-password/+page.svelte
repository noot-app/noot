<script lang="ts">
  import { goto } from "$app/navigation"
  import { dev } from "$app/environment"
  import { resetPassword } from "$lib/auth/store"

  let email = ""
  let loading = false
  let error: string | null = null
  let success = false

  async function handleResetPassword() {
    if (!email) {
      error = "Please enter your email address"
      return
    }

    loading = true
    error = null
    success = false

    try {
      console.debug("🔑 Password reset initiated for email:", email.substring(0, 3) + "***@" + email.split("@")[1])
      
      const result = await resetPassword(email)

      if (result.error) {
        console.warn("❌ Password reset failed:", result.error.message)
        error = result.error.message
      } else {
        console.debug("✅ Password reset email sent successfully")
        success = true
      }
    } catch (err) {
      console.error("💥 Password reset exception:", err)
      error =
        err instanceof Error ? err.message : "An unexpected error occurred"
    } finally {
      loading = false
    }
  }

  function handleBackToLogin() {
    goto("/login")
  }
</script>

<svelte:head>
  <title>Reset Password - Noot</title>
  <meta name="description" content="Reset your Noot account password." />
</svelte:head>

<div
  class="min-h-screen flex items-center justify-center bg-base-200 py-12 px-4 sm:px-6 lg:px-8"
>
  <div class="max-w-md w-full space-y-8">
    <div>
      <h2 class="mt-6 text-center text-3xl font-extrabold text-base-content">
        Reset your password
      </h2>
      <p class="mt-2 text-center text-sm text-base-content/70">
        Enter your email address and we'll send you a link to reset your
        password.
      </p>
    </div>

    {#if dev}
      <div class="alert alert-info">
        <svg
          class="w-6 h-6"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          ></path>
        </svg>
        <div>
          <strong>Development Mode:</strong> Authentication may be using development
          settings. Check your environment variables for production deployment.
        </div>
      </div>
    {/if}

    {#if success}
      <div class="alert alert-success">
        <svg
          class="w-6 h-6"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          ></path>
        </svg>
        <div>
          <strong>Password reset email sent!</strong> Check your inbox for a link
          to reset your password.
        </div>
      </div>
    {/if}

    <form class="mt-8 space-y-6" on:submit|preventDefault={handleResetPassword}>
      <div>
        <label for="email" class="block text-sm font-medium text-base-content">
          Email Address
        </label>
        <input
          id="email"
          name="email"
          type="email"
          autocomplete="email"
          required
          bind:value={email}
          class="input input-bordered w-full mt-1"
          placeholder="your.email@example.com"
          disabled={loading || success}
        />
      </div>

      {#if error}
        <div class="alert alert-error">
          <svg
            class="w-6 h-6"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            ></path>
          </svg>
          <div>{error}</div>
        </div>
      {/if}

      <div>
        <button
          type="submit"
          disabled={loading || success}
          class="btn btn-primary w-full"
          class:loading
        >
          {loading ? "Sending..." : "Send reset email"}
        </button>
      </div>
    </form>

    <div class="text-center">
      <button
        type="button"
        class="link link-primary text-sm"
        on:click={handleBackToLogin}
      >
        Back to sign in
      </button>
    </div>

    {#if success}
      <div class="text-center">
        <p class="text-sm text-base-content/70">
          Didn't receive the email? Check your spam folder or try again.
        </p>
      </div>
    {/if}
  </div>
</div>
