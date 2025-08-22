<script lang="ts">
  import { toastStore } from '$lib/stores/toast';
  import type { Toast } from '$lib/stores/toast';

  export let position: 'top-start' | 'top-center' | 'top-end' | 'bottom-start' | 'bottom-center' | 'bottom-end' = 'bottom-end';
  
  // Convert position prop to DaisyUI classes
  $: positionClass = {
    'top-start': 'toast-top toast-start',
    'top-center': 'toast-top toast-center', 
    'top-end': 'toast-top toast-end',
    'bottom-start': 'toast-bottom toast-start',
    'bottom-center': 'toast-bottom toast-center',
    'bottom-end': 'toast-bottom toast-end'
  }[position];

  function handleDismiss(id: number) {
    toastStore.dismiss(id);
  }

  function handleKeydown(event: KeyboardEvent, id: number) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleDismiss(id);
    }
  }
</script>

{#if $toastStore.length > 0}
<div class="toast {positionClass}">
  {#each $toastStore as toast (toast.id)}
    <div 
      class="alert {toast.type === 'success' ? 'alert-success' : toast.type === 'error' ? 'alert-error' : toast.type === 'warning' ? 'alert-warning' : 'alert-info'} cursor-pointer shadow-lg" 
      role="button"
      tabindex="0"
      on:click={() => handleDismiss(toast.id)}
      on:keydown={(e) => handleKeydown(e, toast.id)}
    >
      {#if toast.type === 'success'}
        <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      {:else if toast.type === 'error'}
        <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      {:else if toast.type === 'warning'}
        <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
      {:else}
        <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      {/if}
      <span class="text-sm">{toast.message}</span>
      <button 
        class="btn btn-sm btn-circle btn-ghost" 
        on:click|stopPropagation={() => handleDismiss(toast.id)}
        aria-label="Dismiss notification"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  {/each}
</div>
{/if}
