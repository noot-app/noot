<script lang="ts">
  export let label: string
  export let id: string
  export let value: string | number = ""
  export let required: boolean = false
  export let disabled: boolean = false
  export let size: "xs" | "sm" | "md" | "lg" = "md"
  export let helpText: string = ""
  export let error: string = ""
  export let options: Array<{ value: string | number; label: string }> = []

  // Additional CSS classes
  export let className: string = ""
  export let labelClass: string = ""
  export let selectClass: string = ""

  $: sizeClass = size === "md" ? "" : `select-${size}`
  $: hasError = error !== ""
</script>

<div class="form-control {className}">
  <label class="label {labelClass}" for={id}>
    <span class="label-text {size === 'sm' ? 'text-sm' : ''}">{label}</span>
    {#if required}
      <span class="text-error ml-1">*</span>
    {/if}
  </label>

  <select
    {id}
    {required}
    {disabled}
    class="select select-bordered {sizeClass} {selectClass} {hasError
      ? 'select-error'
      : ''}"
    bind:value
  >
    {#each options as option}
      <option value={option.value}>{option.label}</option>
    {/each}
  </select>

  {#if helpText}
    <div class="label">
      <span class="label-text-alt">{helpText}</span>
    </div>
  {/if}

  {#if error}
    <div class="label">
      <span class="label-text-alt text-error">{error}</span>
    </div>
  {/if}
</div>
