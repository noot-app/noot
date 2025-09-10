<script lang="ts">
  export let className: string = "w-6 h-6"
  export let strokeWidth: number = 1.8
  export let title: string | undefined = undefined
  // Heroicons outline: bolt
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  fill="none"
  viewBox="0 0 24 24"
  stroke-width={strokeWidth}
  stroke="currentColor"
  class={className}
  aria-hidden={title ? undefined : true}
  aria-label={title}
>
  <path
    stroke-linecap="round"
    stroke-linejoin="round"
    d="m3.75 13.5 10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75Z"
  />
</svg>
