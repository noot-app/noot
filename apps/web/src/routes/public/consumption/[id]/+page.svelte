<script lang="ts">
  import ConsumptionDisplay from "$lib/components/ConsumptionDisplay.svelte"
  import Alert from "$lib/components/Alert.svelte"
  import type { PageData } from "./$types"

  export let data: PageData
</script>

<svelte:head>
  <title>Public Consumption</title>
  <meta name="description" content="View public consumption details" />
</svelte:head>

<div class="min-h-full bg-base-100">
  <div class="container mx-auto px-4 py-6 max-w-4xl">
    {#if data.error}
      <Alert type="error">
        {#if data.error.code === 404}
          This consumption was not found or is not publicly available.
        {:else}
          Failed to load consumption: {data.error.error || "Unknown error"}
        {/if}
      </Alert>
      <div class="mt-4">
        <a href="/" class="btn btn-primary">Go Home</a>
      </div>
    {:else if data.consumption}
      <div class="mb-4">
        <div class="flex items-center gap-2 mb-2">
          <h1 class="text-2xl font-bold">Public Consumption</h1>
          <div class="badge badge-success badge-sm">Public</div>
        </div>
        {#if data.consumption.created_at}
          <p class="text-sm text-base-content/60">
            Logged on {new Date(data.consumption.created_at).toLocaleString()}
          </p>
        {/if}
        <p class="text-xs text-base-content/50 mt-1">
          This is a public view - personal information like labels and notes are hidden.
        </p>
      </div>

      <!-- Public consumption display (read-only, no editing, no labels/notes) -->
      <ConsumptionDisplay
        consumption={data.consumption}
        transcript=""
        showRedoButton={false}
        autoShowLabelEdit={false}
        editable={false}
        buttonsAtBottom={false}
        preloadGoalsAuto={null}
        preloadGoalsDri={null}
      />
      
      <div class="mt-8 text-center">
        <p class="text-sm text-base-content/60 mb-4">
          Want to track your own nutrition? 
        </p>
        <a href="/" class="btn btn-primary">Get Started with Noot</a>
      </div>
    {:else}
      <Alert type="error">
        Consumption not found or not available.
      </Alert>
    {/if}
  </div>
</div>