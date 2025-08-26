<script lang="ts">
  import { toastStore } from '$lib/stores/toast';
  import Icon from './Icon.svelte';

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

  function getToastIcon(type: string): string {
    switch (type) {
      case 'success': return 'check-circle';
      case 'error': return 'exclamation-circle';
      case 'warning': return 'exclamation-triangle';
      default: return 'info';
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
      <Icon name={getToastIcon(toast.type)} size="md" className="flex-shrink-0" />
      <span class="text-sm">{toast.message}</span>
      <button 
        class="btn btn-sm btn-circle btn-ghost" 
        on:click|stopPropagation={() => handleDismiss(toast.id)}
        aria-label="Dismiss notification"
      >
        <Icon name="close" size="sm" />
      </button>
    </div>
  {/each}
</div>
{/if}
