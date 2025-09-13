<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { toast } from "$lib/stores/toast"
  import { formatErrorForUser, handleApiCallWithAuthRedirect } from "$lib/utils/error-handling"
  import ConsumptionDisplay from "$lib/components/ConsumptionDisplay.svelte"
  import ExclamationTriangle from "$lib/components/icons/ExclamationTriangle.svelte"
  import type { PageData } from "./$types"
  import Square2Stack from "$lib/components/icons/Square2Stack.svelte"
  import Lock from "$lib/components/icons/Lock.svelte"
  import Star from "$lib/components/icons/Star.svelte"

  export let data: PageData

  // State for favorites
  let isFavorited = false
  let isUpdatingFavorite = false

  // State for public functionality
  let isUpdatingPublic = false
  let showPublicConfirmModal = false

  // Check if consumption is favorited on page load
  async function checkFavoriteStatus() {
    if (!data.consumption?.id || data.isUnauthenticated) {
      // Skip favorites check for unauthenticated users
      isFavorited = false
      return
    }

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
    if (!data.consumption?.id || isUpdatingFavorite || data.isUnauthenticated) return

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
          toast.error(formatErrorForUser(result.error))
          return
        }

        isFavorited = false
        toast.success("Removed from favorites")
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
          toast.error(formatErrorForUser(result.error))
          return
        }

        isFavorited = true
        toast.success("Added to favorites")
      }
    } catch (err) {
      console.error("Failed to toggle favorite:", err)
      toast.error("Failed to update favorites. Please try again.")
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

  // Toggle public status with confirmation
  function handlePublicToggle() {
    if (!data.consumption?.is_public) {
      // Making public - show confirmation modal
      showPublicConfirmModal = true
    } else {
      // Making private - toggle directly
      togglePublicStatus(false)
    }
  }

  function confirmMakePublic() {
    showPublicConfirmModal = false
    togglePublicStatus(true)
  }

  function cancelMakePublic() {
    showPublicConfirmModal = false
  }

  async function togglePublicStatus(isPublic: boolean) {
    if (!data.consumption?.id || isUpdatingPublic) return

    isUpdatingPublic = true

    try {
      const result = await handleApiCallWithAuthRedirect(async () => {
        return await apiClient.PUT("/consumption/{id}", {
          params: { path: { id: data.consumption!.id } },
          body: {
            items: data.consumption!.items,
            is_public: isPublic,
          }
        })
      })

      if (result.error) {
        toast.error(formatErrorForUser(result.error))
        return
      }

      // Update local state
      if (result.data) {
        data = { ...data, consumption: result.data }
      }
      toast.success(isPublic ? "Consumption is now public" : "Consumption is now private")
    } catch (err) {
      console.error("Failed to toggle public status:", err)
      toast.error("Failed to update public status. Please try again.")
    } finally {
      isUpdatingPublic = false
    }
  }

  // Copy public link to clipboard
  async function copyPublicLink() {
    if (!data.consumption?.is_public || !data.consumption?.id) return

    const publicUrl = `${window.location.origin}/consumptions/${data.consumption.id}`
    
    try {
      await navigator.clipboard.writeText(publicUrl)
      toast.success("Public link copied to clipboard")
    } catch (err) {
      console.error("Failed to copy to clipboard:", err)
      toast.error("Failed to copy link. Please try again.")
    }
  }

  // Check favorite status when page loads
  checkFavoriteStatus()
</script>

<svelte:head>
  <title>Consumption Details</title>
</svelte:head>

<div class="min-h-full bg-base-100">
  <div class="container mx-auto px-4 py-6 max-w-4xl">
    <!-- Back button (only show for authenticated users) -->
    {#if !data.isUnauthenticated}
      <div class="mb-4">
        <a class="btn btn-ghost btn-sm" href="/summary">← Back</a>
      </div>
    {/if}

    <!-- Error handling for consumption not found -->
    {#if data.error || !data.consumption}
      <div class="card bg-error/10 text-error shadow-lg">
        <div class="card-body">
          <div class="flex items-center gap-3 mb-4">
            <ExclamationTriangle className="w-10 h-10"/>
            <div>
              <h2 class="card-title text-lg">Consumption Not Found</h2>
              <p class="text-sm text-error/80 mt-1">
                {data.error || "The consumption you're looking for doesn't exist or you don't have permission to view it."}
              </p>
            </div>
          </div>
          
          <div class="text-sm text-error/70 mb-4">
            <p>This could happen if:</p>
            <ul class="list-disc list-inside mt-2 space-y-1">
              <li>The consumption is private and you're not logged in</li>
              <li>The consumption doesn't exist or has been deleted</li>
              <li>You don't have permission to view this consumption</li>
            </ul>
          </div>

          <div class="card-actions justify-end">
            {#if data.isUnauthenticated}
              <a class="btn btn-outline btn-sm" href="/login">Log In</a>
              <a class="btn btn-primary btn-sm" href="/">Go to Home</a>
            {:else}
              <a class="btn btn-primary btn-sm" href="/summary">Return to Summary</a>
            {/if}
          </div>
        </div>
      </div>
    {:else}

    <div class="mb-4 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">Consumption Details</h1>
        {#if data.consumption?.created_at}
          <p class="text-sm text-base-content/60">
            {new Date(data.consumption.created_at).toLocaleString()}
          </p>
        {/if}
        {#if data.consumption?.is_public}
          <div class="badge badge-success badge-sm mt-1">Public</div>
        {/if}
      </div>
      
      <div class="flex items-center gap-2">
        <!-- Public/Private toggle and copy link (only for authenticated users) -->
        {#if data.consumption?.id && !data.isUnauthenticated}
          <!-- Copy link button (only shown if public) -->
          {#if data.consumption.is_public}
            <button
              class="btn btn-outline btn-sm"
              on:click={copyPublicLink}
              title="Copy public link"
            >
              <Square2Stack className="w-4 h-4"/>
              Copy Link
            </button>
          {/if}

          <!-- Public/Private toggle -->
          <button
            class="btn btn-outline btn-sm"
            class:loading={isUpdatingPublic}
            on:click={handlePublicToggle}
            disabled={isUpdatingPublic}
            title={data.consumption.is_public ? "Make private" : "Make public"}
          >
            {#if isUpdatingPublic}
              <span class="loading loading-spinner loading-sm"></span>
            {:else}
              {#if data.consumption.is_public}
                <Lock className="w-4 h-4"/>
                Make Private
              {:else}
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
                Make Public
              {/if}
            {/if}
          </button>
        {/if}

        <!-- Star button for favorites (only show for authenticated users) -->
        {#if data.consumption?.id && !data.isUnauthenticated}
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
              <Star 
                className={isFavorited ? "w-6 h-6 text-honey" : "w-6 h-6"} 
                filled={isFavorited}
                title={isFavorited ? "Remove from favorites" : "Add to favorites"}
              />
            {/if}
          </button>
        {/if}
      </div>
    </div>

    <ConsumptionDisplay
      consumption={data.consumption}
      transcript={data.consumption?.transcript || ""}
      showRedoButton={false}
      autoShowLabelEdit={false}
      editable={!data.isUnauthenticated}
      buttonsAtBottom={true}
      preloadGoalsAuto={data.goalsAuto}
      preloadGoalsDri={data.goalsDri}
      isUnauthenticated={data.isUnauthenticated}
      on:delete={handleDelete}
      on:save={handleSave}
    />
    {/if}
  </div>
</div>

<!-- Confirmation modal for making consumption public -->
{#if showPublicConfirmModal}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg">Make Consumption Public?</h3>
      <p class="py-4">
        This will make your consumption viewable by anyone with the link. 
        The following information will be publicly visible:
      </p>
      <ul class="list-disc list-inside space-y-1 text-sm text-base-content/80 mb-4">
        <li>Food items and nutrition data</li>
        <li>Title (if set)</li>
        <li>Consumption date</li>
      </ul>
      <p class="text-sm text-base-content/80 mb-4">
        <strong>Private information like labels and notes will remain hidden.</strong>
      </p>
      <div class="modal-action">
        <button class="btn btn-outline" on:click={cancelMakePublic}>Cancel</button>
        <button class="btn btn-primary" on:click={confirmMakePublic}>Make Public</button>
      </div>
    </div>
  </div>
{/if}
