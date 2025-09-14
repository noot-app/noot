<script lang="ts">
  import BaseModal from './shared/ui/BaseModal.svelte'

  export let show: boolean = false
  export let title: string
  export let message: string = ""
  export let confirmText: string = "Confirm"
  export let cancelText: string = "Cancel"
  export let confirmVariant:
    | "primary"
    | "secondary"
    | "accent"
    | "error"
    | "warning"
    | "info"
    | "success" = "error"
  export let loading: boolean = false

  // Event handlers
  export let onConfirm: () => void | Promise<void>
  export let onCancel: (() => void) | undefined = undefined
</script>

<BaseModal 
  bind:show 
  {title} 
  type="confirm"
  {message}
  {confirmText}
  {cancelText}
  {confirmVariant}
  {loading}
  {onConfirm}
  {onCancel}
  closable={!loading}
>
  <slot />
</BaseModal>
