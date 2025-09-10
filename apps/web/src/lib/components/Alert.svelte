<script lang="ts">
  import Icon from './Icon.svelte'

  export let type: 'info' | 'success' | 'warning' | 'error' = 'info'
  export let icon: string | null = null
  export let size: 'sm' | 'md' = 'md'
  export let className: string = ''

  // Map alert types to their default icons and DaisyUI classes
  const alertConfig = {
    info: {
      class: 'alert-info',
      defaultIcon: 'info'
    },
    success: {
      class: 'alert-success',
      defaultIcon: 'check-circle'
    },
    warning: {
      class: 'alert-warning',
      defaultIcon: 'exclamation-triangle'
    },
    error: {
      class: 'alert-error',
      defaultIcon: 'exclamation-circle'
    }
  }

  $: config = alertConfig[type]
  $: alertIcon = icon !== null ? icon : config.defaultIcon
  $: sizeClass = size === 'sm' ? 'alert-sm' : ''
</script>

<div class="alert {config.class} {sizeClass} {className}" role="alert" {...$$restProps}>
  {#if alertIcon}
    <Icon name={alertIcon} size="md" />
  {/if}
  <div>
    <slot />
  </div>
</div>
