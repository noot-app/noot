<script lang="ts">
  export let label: string
  export let id: string
  export let type: "text" | "email" | "password" | "number" | "tel" | "url" | "textarea" | "select" = "text"
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
  export let rows: number = 3

  // For select fields
  export let options: Array<{ value: string | number; label: string }> = []

  // Additional CSS classes
  export let className: string = ""
  export let labelClass: string = ""
  export let inputClass: string = ""

  $: sizeClass = size === "md" ? "" : `${getBaseClass()}-${size}`
  $: hasError = error !== ""

  function getBaseClass(): string {
    switch (type) {
      case "textarea":
        return "textarea"
      case "select":
        return "select"
      default:
        return "input"
    }
  }

  function getClasses(): string {
    const baseClass = getBaseClass()
    const errorClass = hasError ? `${baseClass}-error` : ""
    return `${baseClass} ${baseClass}-bordered ${sizeClass} ${inputClass} ${errorClass}`.trim()
  }
</script>

<div class="form-control {className}">
  <label class="label {labelClass}" for={id}>
    <span class="label-text {size === 'sm' ? 'text-sm' : ''}">{label}</span>
    {#if required}
      <span class="text-error ml-1">*</span>
    {/if}
  </label>

  {#if type === "textarea"}
    <textarea
      {id}
      {placeholder}
      {required}
      {disabled}
      {maxlength}
      {rows}
      class={getClasses()}
      bind:value
    ></textarea>
  {:else if type === "select"}
    <select
      {id}
      {required}
      {disabled}
      class={getClasses()}
      bind:value
    >
      {#each options as option}
        <option value={option.value}>{option.label}</option>
      {/each}
    </select>
  {:else}
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
      class={getClasses()}
      bind:value
    />
  {/if}

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