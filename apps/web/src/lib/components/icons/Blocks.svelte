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
  <path d="M10 22V7a1 1 0 0 0-1-1H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-5a1 1 0 0 0-1-1H2"/>
  <rect x="14" y="2" width="8" height="8" rx="1"/>
</svg>
