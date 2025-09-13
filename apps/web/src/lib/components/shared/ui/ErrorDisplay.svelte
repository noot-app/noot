<script lang="ts">
  export let error: string
  export let title: string = "Error"
  export let showRetry: boolean = false
  export let retryText: string = "Retry"
  export let onRetry: (() => void) | undefined = undefined
  export let className: string = ""
  
  import BaseIcon from "./BaseIcon.svelte"
  
  function handleRetry() {
    if (onRetry) {
      onRetry()
    }
  }
</script>

{#if error}
  <div class="alert alert-error {className}">
    <BaseIcon size={20} ariaLabel="Error">
      <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
    </BaseIcon>
    <div class="flex-1">
      <h3 class="font-bold">{title}</h3>
      <div class="text-sm opacity-75">{error}</div>
    </div>
    {#if showRetry && onRetry}
      <button class="btn btn-sm btn-outline" on:click={handleRetry}>
        {retryText}
      </button>
    {/if}
  </div>
{/if}