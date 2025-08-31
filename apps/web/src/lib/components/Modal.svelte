<script lang="ts">
  export let show: boolean = false
  export let title: string = ""
  export let size: "sm" | "md" | "lg" | "xl" = "md"
  export let closable: boolean = true

  function closeModal() {
    if (closable) {
      show = false
    }
  }

  function handleOutsideClick(event: MouseEvent) {
    if (event.target === event.currentTarget && closable) {
      closeModal()
    }
  }

  function handleEscapeKey(event: KeyboardEvent) {
    if (event.key === "Escape" && closable) {
      closeModal()
    }
  }

  $: sizeClass = {
    sm: "max-w-sm",
    md: "max-w-md",
    lg: "max-w-2xl",
    xl: "max-w-4xl",
  }[size]
</script>

<svelte:window on:keydown={handleEscapeKey} />

{#if show}
  <div
    class="modal modal-open"
    on:click={handleOutsideClick}
    on:keydown={handleEscapeKey}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
  >
    <div class="modal-box {sizeClass}">
      {#if title}
        <h3 class="font-bold text-lg mb-4">{title}</h3>
      {/if}

      <slot />

      {#if $$slots.actions}
        <div class="modal-action">
          <slot name="actions" />
        </div>
      {/if}

      {#if closable && !$$slots.actions}
        <div class="modal-action">
          <button class="btn" on:click={closeModal}>Close</button>
        </div>
      {/if}
    </div>
  </div>
{/if}
