<script lang="ts">
  export let show: boolean = false
  export let title: string
  export let message: string = ""
  export let confirmText: string = "Confirm"
  export let cancelText: string = "Cancel"
  export let confirmVariant:
    | "primary"
    | "secondary"
    | "accent"
    | "error"
    | "warning"
    | "info"
    | "success" = "error"
  export let loading: boolean = false

  // Event handlers
  export let onConfirm: () => void | Promise<void>
  export let onCancel: (() => void) | undefined = undefined

  async function handleConfirm() {
    await onConfirm()
  }

  function handleCancel() {
    if (onCancel) {
      onCancel()
    } else {
      show = false
    }
  }

  function handleBackdropClick() {
    handleCancel()
  }
</script>

{#if show}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg">{title}</h3>

      {#if message}
        <p class="py-4">{message}</p>
      {/if}

      <!-- Allow custom content via slot -->
      <slot />

      <div class="modal-action">
        <button
          class="btn btn-{confirmVariant} {loading ? 'loading' : ''}"
          disabled={loading}
          on:click={handleConfirm}
        >
          {loading ? "" : confirmText}
        </button>
        <button class="btn" disabled={loading} on:click={handleCancel}>
          {cancelText}
        </button>
      </div>
    </div>

    <!-- Backdrop -->
    <button
      class="modal-backdrop"
      on:click={handleBackdropClick}
      disabled={loading}
      aria-label="Close modal"
    ></button>
  </div>
{/if}
