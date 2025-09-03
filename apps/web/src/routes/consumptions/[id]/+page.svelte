<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import ConsumptionDisplay from "$lib/components/ConsumptionDisplay.svelte"
  import type { PageData } from "./$types"

  export let data: PageData

  async function handleDelete() {
    if (!data.consumption?.id) return

    const confirmed = confirm(
      "Delete this consumption? This action cannot be undone.",
    )
    if (!confirmed) return

    try {
      const deleteResponse = await apiClient.DELETE("/consumption/{id}", {
        params: { path: { id: data.consumption.id } },
      })

      if (deleteResponse.error) {
        throw new Error(`Delete failed: ${deleteResponse.error}`)
      }

      // Redirect to summary after successful deletion
      window.location.href = "/summary"
    } catch (err) {
      console.error("Delete error:", err)
      alert(`Error deleting consumption: ${err}`)
    }
  }

  function handleSave(event: CustomEvent) {
    // Update consumption data with saved changes
    data.consumption = event.detail.consumption
  }
</script>

<svelte:head>
  <title>Consumption Details</title>
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-6 max-w-4xl">
    <div class="mb-4">
      <a class="btn btn-ghost btn-sm" href="/summary">← Back</a>
    </div>

    <div class="mb-4">
      <h1 class="text-2xl font-bold">Consumption Details</h1>
      {#if data.consumption?.created_at}
        <p class="text-sm text-base-content/60">
          {new Date(data.consumption.created_at).toLocaleString()}
        </p>
      {/if}
    </div>

    <ConsumptionDisplay
      consumption={data.consumption}
      transcript={data.consumption?.transcript || ""}
      showRedoButton={false}
      autoShowLabelEdit={false}
      editable={true}
      preloadGoalsAuto={data.goalsAuto}
      preloadGoalsDri={data.goalsDri}
      on:delete={handleDelete}
      on:save={handleSave}
    />
  </div>
</div>
