<script lang="ts">
  import BaseModal from './shared/ui/BaseModal.svelte'

  export let show: boolean = false
  export let title: string = ""
  export let size: "sm" | "md" | "lg" | "xl" = "md"
  export let closable: boolean = true
</script>

<BaseModal 
  bind:show 
  {title} 
  {size} 
  {closable}
>
  <slot />
  <slot name="actions" slot="actions" />
</BaseModal>
