<script lang="ts">
  export let name: string
  export let size: "sm" | "md" | "lg" | "xl" = "md"
  export let className: string = ""

  $: sizeClass = {
    sm: "w-4 h-4",
    md: "w-6 h-6",
    lg: "w-8 h-8",
    xl: "w-12 h-12",
  }[size]

  const icons = {
    info: {
      path: "M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z",
      fill: false,
    },
    edit: {
      path: "M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z",
      fill: false,
    },
    delete: {
      path: "M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16",
      fill: false,
    },
    save: {
      path: "M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4",
      fill: false,
    },
    close: { path: "M6 18L18 6M6 6l12 12", fill: false },
    external: {
      path: "M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14",
      fill: false,
    },
    "check-circle": {
      path: "M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z",
      fill: false,
    },
    "exclamation-circle": {
      path: "M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z",
      fill: false,
    },
    "exclamation-triangle": {
      path: "M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z",
      fill: false,
    },
  } as const

  type IconName = keyof typeof icons

  $: icon = icons[name as IconName] || icons.info
</script>

<svg
  class="{sizeClass} {className}"
  fill={icon.fill ? "currentColor" : "none"}
  stroke={icon.fill ? "none" : "currentColor"}
  stroke-width={icon.fill ? "0" : "2"}
  stroke-linecap="round"
  stroke-linejoin="round"
  viewBox="0 0 24 24"
>
  <path d={icon.path} />
</svg>
