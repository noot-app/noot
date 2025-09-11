<script lang="ts">
  export let className: string = "w-6 h-6"
  export let strokeWidth: number = 1.8
  export let title: string | undefined = undefined
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
  <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
</svg>
