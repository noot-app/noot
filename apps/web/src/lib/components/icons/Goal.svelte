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
  <path d="M12 13V2l8 4-8 4"/>
  <path d="M20.561 10.222a9 9 0 1 1-12.55-5.29"/>
  <path d="M8.002 9.997a5 5 0 1 0 8.9 2.02"/>
</svg>
