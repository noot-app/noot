<script lang="ts">
  export let className: string = "w-6 h-6"
  export let title: string | undefined = undefined
  // Avocado icon for healthy fats
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
  <path d="M 12 2 C 9 2 6 8 6 14 C 6 19 9 22 12 22 C 15 22 18 19 18 14 C 18 8 15 2 12 2 Z M 12 12 C 13.657 12 15 13.343 15 15 C 15 16.657 13.657 18 12 18 C 10.343 18 9 16.657 9 15 C 9 13.343 10.343 12 12 12 Z" />
</svg>
