<script lang="ts">
  export let label: string
  export let id: string
  export let type: string = "text"
  export let value: string | number = ""
  export let placeholder: string = ""
  export let required: boolean = false
  export let disabled: boolean = false
  export let size: "xs" | "sm" | "md" | "lg" = "md"
  export let min: string | number | undefined = undefined
  export let max: string | number | undefined = undefined
  export let step: string | number | undefined = undefined
  export let maxlength: number | undefined = undefined
  export let helpText: string = ""
  export let error: string = ""

  // Additional CSS classes
  export let className: string = ""
  export let labelClass: string = ""
  export let inputClass: string = ""

  $: sizeClass = size === "md" ? "" : `input-${size}`
  $: hasError = error !== ""
</script>

<div class="form-control {className}">
  <label class="label {labelClass}" for={id}>
    <span class="label-text {size === 'sm' ? 'text-sm' : ''}">{label}</span>
    {#if required}
      <span class="text-error ml-1">*</span>
    {/if}
  </label>

  <input
    {id}
    {type}
    {placeholder}
    {required}
    {disabled}
    {min}
    {max}
    {step}
    {maxlength}
    class="input input-bordered {sizeClass} {inputClass} {hasError
      ? 'input-error'
      : ''}"
    bind:value
  />

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
