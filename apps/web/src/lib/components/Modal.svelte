<script lang="ts">
  import { browser } from "$app/environment"

  export let show: boolean = false
  export let title: string = ""
  export let size: "sm" | "md" | "lg" | "xl" = "md"
  export let closable: boolean = true

  let modalElement: HTMLDivElement
  let modalBoxElement: HTMLDivElement
  let previouslyFocused: HTMLElement | null = null

  function closeModal() {
    if (closable) {
      show = false
      restoreFocus()
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
    
    // Prevent background scroll when modal is open
    event.preventDefault()
  }

  function handleTouchMove(event: TouchEvent) {
    if (!browser || !modalBoxElement) return
    
    // Check if the touch is on the modal box itself - allow internal scrolling
    const touch = event.touches[0]
    const elementAtTouch = document.elementFromPoint(touch.clientX, touch.clientY)
    
    if (!modalBoxElement.contains(elementAtTouch as Node)) {
      // Touch is outside modal box - prevent scrolling
      event.preventDefault()
    }
  }

  function manageFocus() {
    if (!browser) return

    if (show) {
      // Store currently focused element
      previouslyFocused = document.activeElement as HTMLElement
      
      // Focus the modal
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
      
      // Prevent body scroll on mobile
      document.body.style.overflow = 'hidden'
    } else {
      // Restore body scroll
      document.body.style.overflow = ''
    }
  }

  function restoreFocus() {
    if (!browser) return
    
    // Restore focus to previously focused element
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

  // Handle focus management when show changes
  $: if (browser) {
    manageFocus()
  }
</script>

<svelte:window on:keydown={handleEscapeKey} />

{#if show}
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
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

<style>
  :global(.modal-mobile-optimized) {
    /* Prevent overscroll behavior on mobile */
    overscroll-behavior: contain;
    /* Ensure modal stays within viewport bounds */
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
  }

  :global(.modal-box-mobile) {
    /* Mobile-friendly positioning and sizing */
    max-height: 90vh;
    margin: auto;
    /* Prevent the modal from being dragged beyond reasonable bounds */
    position: relative;
    /* Add safe area padding for mobile devices */
    margin-bottom: env(safe-area-inset-bottom, 0px);
    margin-left: env(safe-area-inset-left, 0px);
    margin-right: env(safe-area-inset-right, 0px);
  }

  /* Mobile-specific styles */
  @media (max-width: 768px) {
    :global(.modal-box-mobile) {
      width: 95vw;
      max-width: 95vw;
      margin: 5vh auto;
      /* Ensure proper touch targets */
      min-height: fit-content;
    }
    
    :global(.modal-mobile-optimized) {
      /* Prevent bounce scrolling on iOS */
      -webkit-overflow-scrolling: touch;
      overflow-y: auto;
      overflow-x: hidden;
    }
  }

  /* For very small screens */
  @media (max-width: 480px) {
    :global(.modal-box-mobile) {
      width: 98vw;
      margin: 2vh auto;
    }
  }
</style>
