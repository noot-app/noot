<script lang="ts">
  export let modalId: string = ""
  export let size: "sm" | "md" | "lg" = "sm"
  export let className = ""
  export let iconClassName = "text-info"
  export let standalone = false // For use in alerts or inline contexts

  const sizeClasses = {
    sm: "w-4 h-4",
    md: "w-5 h-5",
    lg: "w-6 h-6",
  }

  // Default class for standalone icons
  $: defaultStandaloneClasses = standalone ? "shrink-0" : ""
  $: finalIconClasses =
    `${sizeClasses[size]} ${iconClassName} ${defaultStandaloneClasses} ${className}`.trim()
</script>

{#if standalone}
  <!-- Standalone icon for alerts and inline use -->
  <svg
    xmlns="http://www.w3.org/2000/svg"
    fill="none"
    viewBox="0 0 24 24"
    stroke-width="1.5"
    stroke="currentColor"
    class={finalIconClasses}
  >
    <path
      stroke-linecap="round"
      stroke-linejoin="round"
      d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z"
    />
  </svg>
{:else}
  <!-- Button that triggers modal -->
  <label
    for={modalId}
    class="btn btn-ghost btn-circle btn-{size} cursor-pointer {className}"
  >
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      stroke-width="1.5"
      stroke="currentColor"
      class="{sizeClasses[size]} {iconClassName}"
    >
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z"
      />
    </svg>
  </label>
{/if}
