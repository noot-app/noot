<script lang="ts">
  export let className: string = "w-6 h-6"
  export let title: string | undefined = undefined
  // Wheat icon for carbohydrates
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
  <path d="M 12 3 C 8.686 3 6 5.686 6 9 L 6 18 C 6 19.657 7.343 21 9 21 L 15 21 C 16.657 21 18 19.657 18 18 L 18 9 C 18 5.686 15.314 3 12 3 Z M 9 8 L 15 8 C 15.552 8 16 8.448 16 9 C 16 9.552 15.552 10 15 10 L 9 10 C 8.448 10 8 9.552 8 9 C 8 8.448 8.448 8 9 8 Z M 9 12 L 15 12 C 15.552 12 16 12.448 16 13 C 16 13.552 15.552 14 15 14 L 9 14 C 8.448 14 8 13.552 8 13 C 8 12.448 8.448 12 9 12 Z" />
</svg>
