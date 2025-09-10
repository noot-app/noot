<script lang="ts">
  export let className: string = "w-6 h-6"
  export let title: string | undefined = undefined
  // Dumbbell icon for protein/strength
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
  <path d="M 1 9 C 1 8.448 1.448 8 2 8 L 2.5 8 C 3.052 8 3.5 8.448 3.5 9 L 3.5 15 C 3.5 15.552 3.052 16 2.5 16 L 2 16 C 1.448 16 1 15.552 1 15 Z M 4 7 C 4 6.448 4.448 6 5 6 L 6 6 C 6.552 6 7 6.448 7 7 L 7 17 C 7 17.552 6.552 18 6 18 L 5 18 C 4.448 18 4 17.552 4 17 Z M 17 7 C 17 6.448 17.448 6 18 6 L 19 6 C 19.552 6 20 6.448 20 7 L 20 17 C 20 17.552 19.552 18 19 18 L 18 18 C 17.448 18 17 17.552 17 17 Z M 20.5 9 C 20.5 8.448 20.948 8 21.5 8 L 22 8 C 22.552 8 23 8.448 23 9 L 23 15 C 23 15.552 22.552 16 22 16 L 21.5 16 C 20.948 16 20.5 15.552 20.5 15 Z M 7 11 L 17 11 L 17 13 L 7 13 Z" />
</svg>
