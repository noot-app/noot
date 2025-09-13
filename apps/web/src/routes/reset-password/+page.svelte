<script lang="ts">
  import { goto } from "$app/navigation"
  import { resetPassword } from "$lib/auth/store"
  import Alert from "$lib/components/Alert.svelte"
  import DevModeAlert from "$lib/components/DevModeAlert.svelte"

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
  class="min-h-full flex flex-1 items-center justify-center bg-base-200 py-12 px-4 sm:px-6 lg:px-8"
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

    <DevModeAlert />

    {#if success}
      <Alert type="success">
        <strong>Password reset email sent!</strong> Check your inbox for a link
        to reset your password.
      </Alert>
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
