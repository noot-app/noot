<script lang="ts">
  export let title: string = '';
  export let compact: boolean = false;
  export let variant: 'default' | 'primary' | 'secondary' | 'accent' | 'success' | 'warning' | 'error' | 'info' = 'default';
  export let className: string = '';
  
  $: variantClass = variant === 'default' ? 'bg-base-200' : `bg-${variant}/10`;
  $: titleClass = variant === 'default' ? '' : `text-${variant}`;
</script>

<div class="card {variantClass} shadow-lg {className}">
  <div class="card-body {compact ? 'p-4' : ''}">
    {#if title}
      <h3 class="card-title {titleClass} {compact ? 'text-sm' : ''}">{title}</h3>
    {/if}
    
    <slot />
  </div>
</div>