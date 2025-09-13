<script lang="ts">
  import BaseFormField from './shared/forms/BaseFormField.svelte'

  export let label: string
  export let id: string
  export let type: "text" | "email" | "password" | "number" | "tel" | "url" | "date" | "datetime-local" = "text"
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
  export let multiline: boolean = false

  // Additional CSS classes
  export let className: string = ""
  export let labelClass: string = ""
  export let inputClass: string = ""

  $: fieldType: "text" | "email" | "password" | "number" | "tel" | "url" | "date" | "datetime-local" | "textarea" | "select" = multiline ? "textarea" : type
</script>

<BaseFormField 
  {label}
  {id}
  type={fieldType}
  bind:value
  {placeholder}
  {required}
  {disabled}
  {size}
  {min}
  {max}
  {step}
  {maxlength}
  {helpText}
  {error}
  {className}
  {labelClass}
  {inputClass}
/>
