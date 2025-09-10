<script lang="ts">
  import { dev } from "$app/environment"
  import Alert from './Alert.svelte'
</script>

{#if dev}
  <Alert type="info">
    <strong>Development Mode:</strong> Authentication may be using development
    settings. Check your environment variables for production deployment.
  </Alert>
{/if}
