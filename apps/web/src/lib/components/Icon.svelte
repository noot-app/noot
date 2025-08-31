<script lang="ts">
  export let name: string
  export let size: "sm" | "md" | "lg" = "md"
  export let className: string = ""

  $: sizeClass = {
    sm: "w-4 h-4",
    md: "w-6 h-6",
    lg: "w-8 h-8",
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
    zap: { path: "M13 2L3 14h7l-1 8 10-12h-7l1-8z", fill: true },
    dumbbell: {
      path: "M 1 9 C 1 8.448 1.448 8 2 8 L 2.5 8 C 3.052 8 3.5 8.448 3.5 9 L 3.5 15 C 3.5 15.552 3.052 16 2.5 16 L 2 16 C 1.448 16 1 15.552 1 15 Z M 4 7 C 4 6.448 4.448 6 5 6 L 6 6 C 6.552 6 7 6.448 7 7 L 7 17 C 7 17.552 6.552 18 6 18 L 5 18 C 4.448 18 4 17.552 4 17 Z M 17 7 C 17 6.448 17.448 6 18 6 L 19 6 C 19.552 6 20 6.448 20 7 L 20 17 C 20 17.552 19.552 18 19 18 L 18 18 C 17.448 18 17 17.552 17 17 Z M 20.5 9 C 20.5 8.448 20.948 8 21.5 8 L 22 8 C 22.552 8 23 8.448 23 9 L 23 15 C 23 15.552 22.552 16 22 16 L 21.5 16 C 20.948 16 20.5 15.552 20.5 15 Z M 7 11 L 17 11 L 17 13 L 7 13 Z",
      fill: true,
    },
    avocado: {
      path: "M 12 2 C 9 2 6 8 6 14 C 6 19 9 22 12 22 C 15 22 18 19 18 14 C 18 8 15 2 12 2 Z M 12 12 C 13.657 12 15 13.343 15 15 C 15 16.657 13.657 18 12 18 C 10.343 18 9 16.657 9 15 C 9 13.343 10.343 12 12 12 Z",
      fill: true,
    },
    wheat: {
      path: "M 12 3 C 8.686 3 6 5.686 6 9 L 6 18 C 6 19.657 7.343 21 9 21 L 15 21 C 16.657 21 18 19.657 18 18 L 18 9 C 18 5.686 15.314 3 12 3 Z M 9 8 L 15 8 C 15.552 8 16 8.448 16 9 C 16 9.552 15.552 10 15 10 L 9 10 C 8.448 10 8 9.552 8 9 C 8 8.448 8.448 8 9 8 Z M 9 12 L 15 12 C 15.552 12 16 12.448 16 13 C 16 13.552 15.552 14 15 14 L 9 14 C 8.448 14 8 13.552 8 13 C 8 12.448 8.448 12 9 12 Z",
      fill: true,
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
