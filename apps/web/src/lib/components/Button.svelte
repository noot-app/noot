<script lang="ts">
  export let variant:
    | "primary"
    | "secondary"
    | "accent"
    | "error"
    | "warning"
    | "info"
    | "success"
    | "ghost" = "primary"
  export let outline: boolean = false
  export let size: "xs" | "sm" | "md" | "lg" = "md"
  export let disabled: boolean = false
  export let loading: boolean = false
  export let circle: boolean = false
  export let wide: boolean = false
  export let type: "button" | "submit" | "reset" = "button"
  export let href: string | undefined = undefined
  export let className: string = ""

  $: baseClasses = "btn"
  $: variantClass = outline ? `btn-outline btn-${variant}` : `btn-${variant}`
  $: sizeClass = size === "md" ? "" : `btn-${size}`
  $: modifierClasses = [
    circle && "btn-circle",
    wide && "btn-wide",
    loading && "loading",
  ]
    .filter(Boolean)
    .join(" ")

  $: classes = [
    baseClasses,
    variantClass,
    sizeClass,
    modifierClasses,
    className,
  ]
    .filter(Boolean)
    .join(" ")
</script>

{#if href}
  <a {href} class={classes} role="button">
    <slot />
  </a>
{:else}
  <button {type} {disabled} class={classes} on:click>
    {#if loading}
      <span class="loading loading-spinner loading-sm"></span>
    {/if}
    <slot />
  </button>
{/if}
