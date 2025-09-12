<script lang="ts">
  import { dev } from "$app/environment"
  import Alert from './Alert.svelte'
</script>

{#if dev}
  <Alert type="info">
    <strong>Development Mode</strong>
  </Alert>
{/if}
