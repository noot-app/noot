<script lang="ts">
  export let size: number | string = 24
  export let color: string = "currentColor"
  export let className: string = ""
  export let strokeWidth: number = 1.5
  export let fill: string = "none"
  export let viewBox: string = "0 0 24 24"
  export let ariaLabel: string = ""

  $: sizeValue = typeof size === 'number' ? `${size}px` : size
  $: classes = `inline-block ${className}`.trim()
</script>

<svg
  width={sizeValue}
  height={sizeValue}
  {viewBox}
  {fill}
  stroke={color}
  stroke-width={strokeWidth}
  class={classes}
  role="img"
  aria-label={ariaLabel}
>
  <slot />
</svg>
</script>