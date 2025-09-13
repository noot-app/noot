<script lang="ts">
  export let size: "xs" | "sm" | "md" | "lg" = "md"
  export let color: "primary" | "secondary" | "accent" | "neutral" | "info" | "success" | "warning" | "error" = "primary"
  export let text: string = ""
  export let centered: boolean = true
  export let className: string = ""

  $: sizeClass = size === "md" ? "" : `loading-${size}`
  $: colorClass = color === "primary" ? "" : `text-${color}`
  $: containerClass = centered ? "flex items-center justify-center" : ""
</script>

<div class="{containerClass} {className}">
  <div class="flex flex-col items-center gap-3">
    <span class="loading loading-spinner {sizeClass} {colorClass}"></span>
    {#if text}
      <p class="text-sm text-base-content-lighter">{text}</p>
    {/if}
  </div>
</div>