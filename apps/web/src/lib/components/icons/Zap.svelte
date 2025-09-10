<script lang="ts">
  export let className: string = "w-6 h-6"
  export let title: string | undefined = undefined
  // Zap/lightning bolt icon for energy/calories
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  fill="currentColor"
  viewBox="0 0 24 24"
  stroke="none"
  class={className}
  aria-hidden={title ? undefined : true}
  aria-label={title}
>
  <path d="M13 2L3 14h7l-1 8 10-12h-7l1-8z" />
</svg>
