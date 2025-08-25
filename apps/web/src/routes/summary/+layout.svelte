<script lang="ts">
  import AuthGuard from '$lib/components/AuthGuard.svelte';
  
  interface Props {
    children?: import("svelte").Snippet;
  }

  let { children }: Props = $props();
</script>

<AuthGuard>
  {@render children?.()}
</AuthGuard>