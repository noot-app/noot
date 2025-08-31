<script lang="ts">
  export let name: string;
  export let color: string; // hex with or without leading '#'
  export let size: 'xs' | 'sm' | 'md' | 'lg' = 'lg';
  export let className: string = '';
  export let ariaLabel: string | undefined = undefined;

  function formatColor(c: string | null | undefined): string {
    if (!c) return '#000000';
    return c.startsWith('#') ? c : `#${c}`;
  }

  function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
    const h = formatColor(hex).replace('#', '');
    if (h.length !== 6) return null;
    const r = parseInt(h.slice(0, 2), 16);
    const g = parseInt(h.slice(2, 4), 16);
    const b = parseInt(h.slice(4, 6), 16);
    if (Number.isNaN(r) || Number.isNaN(g) || Number.isNaN(b)) return null;
    return { r, g, b };
  }

  function rgba(hex: string, alpha: number): string {
    const rgb = hexToRgb(hex);
    if (!rgb) return 'transparent';
    const { r, g, b } = rgb;
    return `rgba(${r}, ${g}, ${b}, ${alpha})`;
  }

  // Simple, readable scheme: light tint background, subtle border, theme-aware text color
  $: base = formatColor(color);
  $: bgSoft = rgba(base, 0.18);
  $: borderSoft = rgba(base, 0.40);
  $: textColor = 'oklch(var(--n))'; // slightly darker than base-content for better contrast
  $: sizeClass = size === 'md' ? '' : `badge-${size}`;
</script>

<div
  class="badge badge-soft font-medium px-3 py-2 {sizeClass} {className}"
  style="background-color: {bgSoft}; border-color: {borderSoft}; color: {textColor};"
  aria-label={ariaLabel}
>
  {name}
  <slot />
  
</div>

<style>
  /* Consumers can override sizing via className */
</style>
