<script lang="ts">
  import { onMount } from "svelte"
  import { apiClient } from "$lib/api/client"
  import { toast } from "$lib/stores/toast"
  import { formatErrorForUser, handleApiCallWithAuthRedirect } from "$lib/utils/error-handling"
  import ConsumptionCard from "$lib/components/ConsumptionCard.svelte"
  import type { paths } from "$lib/api/schema"
  import CursorArrowRays from "$lib/components/icons/CursorArrowRays.svelte"

  // Type definitions
  type FavoritesResponse = paths["/favorites"]["get"]["responses"]["200"]["content"]["application/json"]
  type ConsumptionsResponse = paths["/consumptions"]["get"]["responses"]["200"]["content"]["application/json"]
  type FavoriteWithConsumption = FavoritesResponse["favorites"][0]
  type Consumption = ConsumptionsResponse["consumptions"][0]

  // State
  let favorites: FavoriteWithConsumption[] = []
  let recentConsumptions: Consumption[] = []
  let isLoadingFavorites = false
  let isLoadingRecents = false
  let error = ""
  let hasMoreRecents = true
  let currentPage = 0
  const pageSize = 10

  // Load favorites
  async function loadFavorites() {
    if (isLoadingFavorites) return
    
    isLoadingFavorites = true
    error = ""

    try {
      const result = await handleApiCallWithAuthRedirect(async () => {
        return await apiClient.GET("/favorites")
      })

      if (result.error) {
        error = formatErrorForUser(result.error)
        return
      }

      favorites = result.data?.favorites || []
    } catch (err) {
      console.error("Failed to load favorites:", err)
      error = "Failed to load favorites. Please try again."
    } finally {
      isLoadingFavorites = false
    }
  }

  // Load recent consumptions  
  async function loadRecentConsumptions(offset = 0, append = false) {
    if (isLoadingRecents) return
    
    isLoadingRecents = true
    if (!append) error = ""

    try {
      const result = await handleApiCallWithAuthRedirect(async () => {
        return await apiClient.GET("/consumptions", {
          params: { 
            query: { 
              limit: pageSize, 
              offset: offset 
            } 
          }
        })
      })

      if (result.error) {
        error = formatErrorForUser(result.error)
        return
      }

      const newConsumptions = result.data?.consumptions || []
      
      if (append) {
        recentConsumptions = [...recentConsumptions, ...newConsumptions]
      } else {
        recentConsumptions = newConsumptions
      }

      hasMoreRecents = newConsumptions.length === pageSize
    } catch (err) {
      console.error("Failed to load recent consumptions:", err)
      error = "Failed to load recent consumptions. Please try again."
    } finally {
      isLoadingRecents = false
    }
  }

  // Load more recents
  function loadMoreRecents() {
    if (!hasMoreRecents || isLoadingRecents) return
    currentPage++
    loadRecentConsumptions(currentPage * pageSize, true)
  }

  // Quick re-log consumption
  async function quickRelogConsumption(consumption: Consumption) {
    try {
      // Create a new consumption by duplicating the existing one using consumption_id
      const result = await handleApiCallWithAuthRedirect(async () => {
        return await apiClient.POST("/consumption", {
          body: {
            consumption_id: consumption.id
          }
        })
      })

      if (result.error) {
        toast.error(formatErrorForUser(result.error))
        return
      }

      toast.success("Successfully re-logged consumption!")
      
      // Refresh recent consumptions to show the new one
      currentPage = 0
      await loadRecentConsumptions(0, false)
    } catch (err) {
      console.error("Failed to re-log consumption:", err)
      toast.error("Failed to re-log consumption. Please try again.")
    }
  }

  // Load data on mount
  onMount(async () => {
    await Promise.all([
      loadFavorites(),
      loadRecentConsumptions()
    ])
  })
</script>

<svelte:head>
  <title>Quick Add</title>
</svelte:head>

<div class="min-h-full bg-base-100">
  <div class="container mx-auto px-4 py-6 max-w-4xl">
    <!-- Header -->
    <div class="mb-6">
      <h1 class="text-3xl font-bold flex items-center gap-2">
        <CursorArrowRays className="w-8 h-8" />
        Quick Add
      </h1>
      <p class="text-base-content/70 mt-2">
        Quickly re-log your favorite and recent meals
      </p>
    </div>

    <!-- Error Display -->
    {#if error}
      <div class="alert alert-error mb-4">
        <span>{error}</span>
      </div>
    {/if}

    <!-- Favorites Section -->
    <div class="mb-8">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold">Favorites</h2>
        {#if isLoadingFavorites}
          <span class="loading loading-spinner loading-sm"></span>
        {/if}
      </div>

      {#if favorites.length === 0 && !isLoadingFavorites}
        <div class="card bg-base-200">
          <div class="card-body text-center py-8">
            <p class="text-base-content/70">
              No favorites yet. Star some consumptions to see them here!
            </p>
          </div>
        </div>
      {:else}
        <div class="grid gap-4">
          {#each favorites as favorite (favorite.id)}
            <ConsumptionCard
              consumption={favorite.consumption}
              mode="actions"
              variant="compact"
              onReLog={quickRelogConsumption}
            />
          {/each}
        </div>
      {/if}
    </div>

    <!-- Recent Consumptions Section -->
    <div>
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold">Recent</h2>
        {#if isLoadingRecents && currentPage === 0}
          <span class="loading loading-spinner loading-sm"></span>
        {/if}
      </div>

      {#if recentConsumptions.length === 0 && !isLoadingRecents}
        <div class="card bg-base-200">
          <div class="card-body text-center py-8">
            <p class="text-base-content/70">
              No recent consumptions found.
            </p>
          </div>
        </div>
      {:else}
        <div class="grid gap-4">
          {#each recentConsumptions as consumption (consumption.id)}
            <ConsumptionCard
              {consumption}
              mode="actions"
              variant="compact"
              onReLog={quickRelogConsumption}
            />
          {/each}
        </div>

        <!-- Load More Button -->
        {#if hasMoreRecents}
          <div class="text-center mt-6">
            <button
              class="btn btn-outline"
              class:loading={isLoadingRecents && currentPage > 0}
              on:click={loadMoreRecents}
              disabled={isLoadingRecents}
            >
              {isLoadingRecents && currentPage > 0 ? "Loading..." : "Load More"}
            </button>
          </div>
        {/if}
      {/if}
    </div>
  </div>
</div>
