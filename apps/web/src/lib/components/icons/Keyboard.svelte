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
  <path d="M10 8h.01"/>
  <path d="M12 12h.01"/>
  <path d="M14 8h.01"/>
  <path d="M16 12h.01"/>
  <path d="M18 8h.01"/>
  <path d="M6 8h.01"/>
  <path d="M7 16h10"/>
  <path d="M8 12h.01"/>
  <rect width="20" height="16" x="2" y="4" rx="2"/>
</svg>
