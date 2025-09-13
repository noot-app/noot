<script lang="ts">
  export let title: string
  export let subtitle: string = ""
  export let showBack: boolean = false
  export let className: string = ""
  
  import { goto } from "$app/navigation"
  import BaseIcon from "../ui/BaseIcon.svelte"
  
  function handleBack() {
    if (showBack) {
      goto(-1) || goto("/")
    }
  }
</script>

<header class="safe-area-content {className}">
  <div class="flex items-center gap-4 mb-6">
    {#if showBack}
      <button
        on:click={handleBack}
        class="btn btn-ghost btn-sm btn-circle touch-target"
        aria-label="Go back"
      >
        <BaseIcon size={20} ariaLabel="Back">
          <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18"/>
        </BaseIcon>
      </button>
    {/if}
    
    <div class="flex-1">
      <h1 class="text-2xl font-bold text-base-content">{title}</h1>
      {#if subtitle}
        <p class="text-base-content-lighter mt-1">{subtitle}</p>
      {/if}
    </div>
    
    {#if $$slots.actions}
      <div class="flex items-center gap-2">
        <slot name="actions" />
      </div>
    {/if}
  </div>
</header>