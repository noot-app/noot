<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { toast } from "$lib/stores/toast"
  import { formatErrorForUser, handleApiCallWithAuthRedirect } from "$lib/utils/error-handling"
  import ConsumptionDisplay from "$lib/components/ConsumptionDisplay.svelte"
  import type { PageData } from "./$types"

  export let data: PageData

  // State for favorites
  let isFavorited = false
  let isUpdatingFavorite = false

  // Check if consumption is favorited on page load
  async function checkFavoriteStatus() {
    if (!data.consumption?.id) return

    try {
      const result = await handleApiCallWithAuthRedirect(async () => {
        return await apiClient.GET("/favorites")
      })

      if (result.error) {
        console.error("Failed to check favorite status:", result.error)
        return
      }

      const favorites = result.data?.favorites || []
      isFavorited = favorites.some(fav => fav.consumption_id === data.consumption?.id)
    } catch (err) {
      console.error("Failed to check favorite status:", err)
    }
  }

  // Toggle favorite status
  async function toggleFavorite() {
    if (!data.consumption?.id || isUpdatingFavorite) return

    isUpdatingFavorite = true

    try {
      if (isFavorited) {
        // Remove from favorites
        const result = await handleApiCallWithAuthRedirect(async () => {
          return await apiClient.DELETE("/favorites/{consumption_id}", {
            params: { path: { consumption_id: data.consumption!.id } }
          })
        })

        if (result.error) {
          toast.add(formatErrorForUser(result.error), "error")
          return
        }

        isFavorited = false
        toast.add("Removed from favorites", "success")
      } else {
        // Add to favorites
        const result = await handleApiCallWithAuthRedirect(async () => {
          return await apiClient.POST("/favorites", {
            body: {
              consumption_id: data.consumption!.id
            }
          })
        })

        if (result.error) {
          toast.add(formatErrorForUser(result.error), "error")
          return
        }

        isFavorited = true
        toast.add("Added to favorites", "success")
      }
    } catch (err) {
      console.error("Failed to toggle favorite:", err)
      toast.add("Failed to update favorites. Please try again.", "error")
    } finally {
      isUpdatingFavorite = false
    }
  }

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
    // Update consumption data with saved changes and trigger reactivity
    data = { ...data, consumption: event.detail.consumption }
  }

  // Check favorite status when page loads
  checkFavoriteStatus()
</script>

<svelte:head>
  <title>Consumption Details</title>
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-6 max-w-4xl">
    <div class="mb-4">
      <a class="btn btn-ghost btn-sm" href="/summary">← Back</a>
    </div>

    <div class="mb-4 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">Consumption Details</h1>
        {#if data.consumption?.created_at}
          <p class="text-sm text-base-content/60">
            {new Date(data.consumption.created_at).toLocaleString()}
          </p>
        {/if}
      </div>
      
      <!-- Star button for favorites -->
      {#if data.consumption?.id}
        <button
          class="btn btn-ghost"
          class:loading={isUpdatingFavorite}
          on:click={toggleFavorite}
          disabled={isUpdatingFavorite}
          title={isFavorited ? "Remove from favorites" : "Add to favorites"}
        >
          {#if isUpdatingFavorite}
            <span class="loading loading-spinner loading-sm"></span>
          {:else}
            <span class="text-2xl">{isFavorited ? "⭐" : "☆"}</span>
          {/if}
        </button>
      {/if}
    </div>

    <ConsumptionDisplay
      consumption={data.consumption}
      transcript={data.consumption?.transcript || ""}
      showRedoButton={false}
      autoShowLabelEdit={false}
      editable={true}
      buttonsAtBottom={true}
      preloadGoalsAuto={data.goalsAuto}
      preloadGoalsDri={data.goalsDri}
      on:delete={handleDelete}
      on:save={handleSave}
    />
  </div>
</div>
