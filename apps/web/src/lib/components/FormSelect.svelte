<script lang="ts">
  import BaseFormField from './shared/forms/BaseFormField.svelte'

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
</script>

<BaseFormField 
  {label}
  {id}
  type="select"
  bind:value
  {required}
  {disabled}
  {size}
  {helpText}
  {error}
  {options}
  {className}
  {labelClass}
  inputClass={selectClass}
/>
