<script lang="ts">
  export let className: string = "w-6 h-6"
  export let strokeWidth: number = 1.8
  export let title: string | undefined = undefined
  // Heroicons outline: scale
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
    d="M12 3v17.25m0 0c-1.472 0-2.882.265-4.185.75M12 20.25c1.472 0 2.882.265 4.185.75M18.75 4.97A48.254 48.254 0 0 0 12 4.5c-2.291 0-4.545.16-6.75.47m13.5 0c1.01.143 2.01.317 3 .52m-3-.52 2.062 10.25a8.25 8.25 0 0 1-.306 1.645c-.51 1.54-2.4 2.135-4.756 2.135-2.357 0-4.246-.595-4.756-2.135a8.25 8.25 0 0 1-.306-1.645L8.25 5.47m10.5 0L16.687 15.4c-.11.54-.582.6-.582.6s-.472-.06-.582-.6L13.462 5.47M18.75 4.97l-1.687 10.43c-.11.54-.582.6-.582.6s-.472-.06-.582-.6L14.25 4.97"
  />
</svg>
