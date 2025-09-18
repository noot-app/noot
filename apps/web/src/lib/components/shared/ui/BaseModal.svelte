<script lang="ts">
  import { browser } from "$app/environment"

  export let show: boolean = false
  export let title: string = ""
  export let size: "sm" | "md" | "lg" | "xl" = "md"
  export let closable: boolean = true
  export let type: "default" | "confirm" = "default"
  
  // For confirm modal type
  export let message: string = ""
  export let confirmText: string = "Confirm"
  export let cancelText: string = "Cancel"
  export let confirmVariant: "primary" | "secondary" | "accent" | "error" | "warning" | "info" | "success" = "error"
  export let loading: boolean = false
  export let onConfirm: (() => void | Promise<void>) | undefined = undefined
  export let onCancel: (() => void) | undefined = undefined

  let modalElement: HTMLDivElement
  let modalBoxElement: HTMLDivElement
  let previouslyFocused: HTMLElement | null = null

  function closeModal() {
    if (closable && !loading) {
      show = false
      restoreFocus()
    }
  }

  function handleOutsideClick(event: MouseEvent) {
    if (event.target === event.currentTarget && closable && !loading) {
      if (type === "confirm" && onCancel) {
        onCancel()
      } else {
        closeModal()
      }
    }
  }

  function handleEscapeKey(event: KeyboardEvent) {
    if (event.key === "Escape" && closable && !loading) {
      if (type === "confirm" && onCancel) {
        onCancel()
      } else {
        closeModal()
      }
    }
  }

  async function handleConfirm() {
    if (onConfirm && type === "confirm") {
      await onConfirm()
    }
  }

  function handleCancel() {
    if (onCancel) {
      onCancel()
    } else {
      show = false
    }
  }

  // Focus trap functionality
  function trapFocus(event: KeyboardEvent) {
    if (!modalBoxElement || event.key !== "Tab") return

    const focusableElements = modalBoxElement.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    )
    const firstFocusable = focusableElements[0] as HTMLElement
    const lastFocusable = focusableElements[focusableElements.length - 1] as HTMLElement

    if (event.shiftKey) {
      if (document.activeElement === firstFocusable) {
        lastFocusable?.focus()
        event.preventDefault()
      }
    } else {
      if (document.activeElement === lastFocusable) {
        firstFocusable?.focus()
        event.preventDefault()
      }
    }
  }

  // Touch handling for mobile - prevent overscroll and excessive dragging
  function handleTouchStart(event: TouchEvent) {
    if (!browser || event.target !== modalElement) return
    event.preventDefault()
  }

  function handleTouchMove(event: TouchEvent) {
    if (!browser || !modalBoxElement) return
    
    const touch = event.touches[0]
    const elementAtTouch = document.elementFromPoint(touch.clientX, touch.clientY)
    
    if (!modalBoxElement.contains(elementAtTouch as Node)) {
      event.preventDefault()
    }
  }

  function manageFocus() {
    if (!browser) return

    if (show) {
      previouslyFocused = document.activeElement as HTMLElement
      
      setTimeout(() => {
        if (modalBoxElement) {
          const firstFocusable = modalBoxElement.querySelector(
            'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
          ) as HTMLElement
          if (firstFocusable) {
            firstFocusable.focus()
          } else {
            modalBoxElement.focus()
          }
        }
      }, 100)
      
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
  }

  function restoreFocus() {
    if (!browser) return
    
    if (previouslyFocused) {
      previouslyFocused.focus()
      previouslyFocused = null
    }
  }

  $: sizeClass = {
    sm: "max-w-sm",
    md: "max-w-md",
    lg: "max-w-2xl",
    xl: "max-w-4xl",
  }[size]

  $: if (browser) {
    manageFocus()
  }
</script>

<svelte:window on:keydown={handleEscapeKey} />

{#if show}
  <div
    bind:this={modalElement}
    class="modal modal-open modal-mobile-optimized"
    on:click={handleOutsideClick}
    on:keydown={trapFocus}
    on:touchstart={handleTouchStart}
    on:touchmove={handleTouchMove}
    role="dialog"
    aria-modal="true"
    aria-labelledby={title ? "modal-title" : undefined}
    tabindex="-1"
  >
    <div bind:this={modalBoxElement} class="modal-box {sizeClass} modal-box-mobile">
      {#if title}
        <h3 id="modal-title" class="font-bold text-lg mb-4">{title}</h3>
      {/if}

      {#if type === "confirm" && message}
        <p class="py-4">{message}</p>
      {/if}

      <slot />

      <div class="modal-action">
        {#if type === "confirm"}
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
        {:else if $$slots.actions}
          <slot name="actions" />
        {:else if closable}
          <button class="btn" on:click={closeModal}>Close</button>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  :global(.modal-mobile-optimized) {
    overscroll-behavior: contain;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
  }

  :global(.modal-box-mobile) {
    max-height: 90vh;
    margin: auto;
    position: relative;
    margin-bottom: env(safe-area-inset-bottom, 0px);
    margin-left: env(safe-area-inset-left, 0px);
    margin-right: env(safe-area-inset-right, 0px);
  }

  @media (max-width: 768px) {
    :global(.modal-box-mobile) {
      width: 95vw;
      max-width: 95vw;
      margin: auto;
      min-height: fit-content;
      max-height: 90vh;
      overflow-y: auto;
      -webkit-overflow-scrolling: touch;
      overscroll-behavior: contain;
    }
    
    :global(.modal-mobile-optimized) {
      -webkit-overflow-scrolling: touch;
      overflow-y: auto;
      overflow-x: hidden;
      align-items: center;
      justify-content: center;
      padding: 1rem;
    }
  }

  @media (max-width: 480px) {
    :global(.modal-box-mobile) {
      width: 98vw;
      margin: auto;
      max-height: 95vh;
    }
    
    :global(.modal-mobile-optimized) {
      padding: 0.5rem;
    }
  }
</style>