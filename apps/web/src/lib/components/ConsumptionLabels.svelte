<script lang="ts">
  import { createEventDispatcher } from "svelte"
  import { apiClient } from "$lib/api/client"
  import Label from "./Label.svelte"
  import TagIcon from "./icons/Tag.svelte"

  // Props
  export let consumptionId: string
  export let initialLabels: any[] = [] // Labels already on the consumption
  export let autoShowEdit = false // Whether to automatically show label editing
  export let editable = true // Whether labels can be edited

  // Local state
  let availableLabels: any[] = []
  let isLoadingLabels = false
  let isUpdatingLabels = false
  let labelsLoaded = false
  let selectedLabels: string[] = [] // Simple array of label names - what we want the final state to be
  let currentLabels: string[] = [] // What's currently on the server
  let isEditingLabels = false
  let error = ""

  // Event dispatcher for parent component communication
  const dispatch = createEventDispatcher<{
    labelsUpdated: { labels: any[] }
    error: { message: string }
  }>()

  // Reactive variables - much simpler now
  $: hasChanges = JSON.stringify([...selectedLabels].sort()) !== JSON.stringify([...currentLabels].sort())

  // Initialize edit state based on autoShowEdit
  $: if (autoShowEdit && !isEditingLabels && consumptionId) {
    isEditingLabels = true
  }

  // Load labels when consumption ID is available
  $: if (consumptionId && !labelsLoaded) {
    loadLabels()
  }

  // Sync currentLabels from initial labels prop
  $: if (initialLabels) {
    currentLabels = initialLabels.map((l: any) => l.name as string).filter(Boolean)
    // Initialize selectedLabels from currentLabels if not editing yet
    if (!isEditingLabels) {
      selectedLabels = [...currentLabels]
    }
  }

  // Load user's available labels
  async function loadLabels() {
    if (isLoadingLabels || labelsLoaded || !consumptionId) return

    try {
      isLoadingLabels = true
      const response = await apiClient.GET("/labels")

      if (response.error) {
        const errorMsg = `Failed to load labels: ${response.error}`
        console.error(errorMsg)
        error = errorMsg
        dispatch('error', { message: errorMsg })
        return
      }

      availableLabels = response.data?.labels || []

      // Load labels assigned to this consumption if not already provided
      if (!initialLabels || initialLabels.length === 0) {
        try {
          const assigned = await apiClient.GET("/consumption/{id}/labels", {
            params: { path: { id: consumptionId } },
          })
          if (!assigned.error) {
            const assignedLabels = assigned.data?.labels || []
            // Update currentLabels from server
            currentLabels = assignedLabels.map((l: any) => l.name as string)
            // Initialize selectedLabels if not editing
            if (!isEditingLabels) {
              selectedLabels = [...currentLabels]
            }
            // Notify parent of initial labels
            dispatch('labelsUpdated', { labels: assignedLabels })
          }
        } catch (e) {
          console.warn("Error loading consumption labels:", e)
        }
      }
    } catch (err) {
      const errorMsg = `Error loading labels: ${err}`
      console.error(errorMsg)
      error = errorMsg
      dispatch('error', { message: errorMsg })
    } finally {
      isLoadingLabels = false
      labelsLoaded = true
    }
  }

  function retryLoadLabels() {
    if (!consumptionId) return
    labelsLoaded = false
    error = ""
    loadLabels()
  }

  // Toggle label in selection (GitHub Issues style)
  function toggleLabelSelection(labelName: string) {
    if (selectedLabels.includes(labelName)) {
      selectedLabels = selectedLabels.filter(name => name !== labelName)
    } else {
      selectedLabels = [...selectedLabels, labelName]
    }
  }

  // Apply label changes - send the complete desired state to server
  async function applyLabelChanges() {
    if (!consumptionId || isUpdatingLabels || !hasChanges) return

    try {
      isUpdatingLabels = true
      error = ""

      // Get the label IDs for the desired final state
      const labelIds = selectedLabels
        .map((name: string) => {
          const label = availableLabels.find((l) => l.name === name)
          return label?.id
        })
        .filter(Boolean)

      // Send complete desired state - like GitHub Issues does
      const response = await apiClient.POST("/consumption/{id}/labels", {
        params: { path: { id: consumptionId } },
        body: { ids: labelIds },
      })

      if (response.error) {
        throw new Error(`Failed to update labels: ${response.error}`)
      }

      // Update from server response
      const updatedLabels = response.data?.labels || []
      currentLabels = updatedLabels.map((l: any) => l.name as string)
      
      // Notify parent component of the update
      dispatch('labelsUpdated', { labels: updatedLabels })

      // Keep selectedLabels as is - user's current selection stays active
      
      // Only close editor if not in auto-show mode
      if (!autoShowEdit) {
        isEditingLabels = false
      }
    } catch (err) {
      const errorMsg = `Error applying label changes: ${err}`
      error = errorMsg
      console.error("Apply label changes error:", err)
      dispatch('error', { message: errorMsg })
      // Reset selectedLabels to current server state on error
      selectedLabels = [...currentLabels]
    } finally {
      isUpdatingLabels = false
    }
  }

  function cancelLabelChanges() {
    // Reset to current server state
    selectedLabels = [...currentLabels]
    isEditingLabels = false
    error = ""
  }

  function startLabelEditing() {
    isEditingLabels = true
    // Initialize selectedLabels from current server state when starting to edit
    selectedLabels = [...currentLabels]
  }

  // Export functions for parent component to call
  export function startEditing() {
    startLabelEditing()
  }

  export function isEditing() {
    return isEditingLabels
  }

  export function hasErrors() {
    return !!error
  }

  export function getCurrentLabels() {
    return [...currentLabels]
  }
</script>

<div class="card bg-base-200 shadow-lg">
  <div class="card-body">
    <div class="flex justify-between items-center mb-4">
      <h3 class="card-title text-sm flex items-center gap-2">
        <TagIcon className="w-4 h-4" />
        Labels
      </h3>

      {#if !isEditingLabels && !autoShowEdit && editable}
        <button
          class="btn btn-outline btn-sm"
          on:click={startLabelEditing}
          disabled={isUpdatingLabels}
          data-testid="edit-labels-button"
        >
          <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
          </svg>
          Edit Labels
        </button>
      {/if}
    </div>

    {#if error}
      <div class="alert alert-error mb-4" data-testid="error-alert">
        <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <span>{error}</span>
      </div>
    {/if}

    {#if autoShowEdit && isEditingLabels}
      <div class="alert alert-info mb-4 text-sm" data-testid="auto-edit-info">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        Select labels to apply to this meal. Click "Apply" to save changes, or "Done" when finished.
      </div>
    {/if}

    <!-- Applied Labels Display (when not editing) -->
    {#if !isEditingLabels}
      <div data-testid="applied-labels">
        {#if currentLabels.length > 0}
          <div class="flex flex-wrap gap-2">
            {#each currentLabels as labelName}
              {@const labelData = availableLabels.find((l) => l.name === labelName)}
              {#if labelData}
                <Label name={labelData.name} color={labelData.color} />
              {/if}
            {/each}
          </div>
        {:else}
          <div class="flex items-center gap-2" data-testid="no-labels-message">
            <p class="text-sm text-base-content/70">No labels applied to this meal</p>
            {#if availableLabels.length === 0 && !isLoadingLabels}
              <button class="btn btn-xs" on:click={retryLoadLabels} data-testid="retry-load-button">Load labels</button>
            {/if}
          </div>
        {/if}
      </div>
    {/if}

    <!-- Label Selection Interface -->
    {#if isEditingLabels}
      <div class="space-y-4" data-testid="label-editor">
        {#if availableLabels.length === 0 && !isLoadingLabels}
          <div class="flex items-center gap-2" data-testid="no-available-labels">
            <p class="text-sm text-base-content/70">No labels available for selection.</p>
            <button class="btn btn-xs" on:click={retryLoadLabels} data-testid="retry-button">Retry</button>
          </div>
        {/if}
        
        <div class="grid grid-cols-1 gap-2 max-h-64 overflow-y-auto" data-testid="label-list">
          {#each availableLabels as label}
            {@const isSelected = selectedLabels.includes(label.name)}
            {@const isCurrentlyApplied = currentLabels.includes(label.name)}
            {@const isNewSelection = isSelected && !isCurrentlyApplied}
            {@const willBeRemoved = isCurrentlyApplied && !isSelected}
            <button
              class="flex items-center justify-between p-3 rounded-lg border transition-all hover:bg-base-300 {isSelected
                ? 'bg-base-300 border-primary'
                : 'bg-base-100 border-base-300'}"
              on:click={() => toggleLabelSelection(label.name)}
              disabled={isUpdatingLabels}
              data-testid="label-option"
              data-label-name={label.name}
              data-is-selected={isSelected}
            >
              <div class="flex items-center gap-3">
                <div class="checkbox-wrapper">
                  <input
                    type="checkbox"
                    class="checkbox checkbox-primary checkbox-sm"
                    checked={isSelected}
                    readonly
                    data-testid="label-checkbox"
                  />
                </div>
                <div class="flex items-center gap-2">
                  <div
                    class="w-3 h-3 rounded-full"
                    style="background-color: {label.color?.startsWith('#')
                      ? label.color
                      : `#${label.color}`};"
                    data-testid="label-color"
                  ></div>
                  <span class="font-medium">{label.name}</span>
                  {#if isCurrentlyApplied && isSelected}
                    <span class="badge badge-xs badge-success" data-testid="applied-badge">Applied</span>
                  {:else if isNewSelection}
                    <span class="badge badge-xs badge-warning" data-testid="add-badge">+Add</span>
                  {:else if willBeRemoved}
                    <span class="badge badge-xs badge-error" data-testid="remove-badge">-Remove</span>
                  {/if}
                </div>
              </div>
              {#if label.description}
                <span class="text-xs text-base-content/60 truncate ml-2" data-testid="label-description">
                  {label.description}
                </span>
              {/if}
            </button>
          {/each}
        </div>

        {#if hasChanges}
          <div class="bg-base-200 p-3 rounded-lg text-sm" data-testid="pending-changes">
            <div class="font-medium mb-1">Pending changes:</div>
            <div class="flex flex-wrap gap-2">
              {#each selectedLabels.filter(name => !currentLabels.includes(name)) as labelName}
                <span class="badge badge-warning badge-sm" data-testid="add-change">+{labelName}</span>
              {/each}
              {#each currentLabels.filter(name => !selectedLabels.includes(name)) as labelName}
                <span class="badge badge-error badge-sm" data-testid="remove-change">-{labelName}</span>
              {/each}
            </div>
          </div>
        {:else if currentLabels.length > 0}
          <div class="bg-success/10 p-3 rounded-lg text-sm text-success-content" data-testid="no-changes-message">
            <div class="flex items-center gap-2">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
              </svg>
              All selected labels are already applied
            </div>
          </div>
        {/if}

        <!-- Action Buttons -->
        <div class="flex justify-end gap-2 pt-2 border-t border-base-300" data-testid="action-buttons">
          {#if !autoShowEdit}
            <button
              class="btn btn-ghost btn-sm"
              on:click={cancelLabelChanges}
              disabled={isUpdatingLabels}
              data-testid="cancel-button"
            >
              Cancel
            </button>
          {/if}
          <button
            class="btn btn-primary btn-sm"
            on:click={applyLabelChanges}
            disabled={isUpdatingLabels || !hasChanges}
            data-testid="apply-button"
          >
            {#if isUpdatingLabels}
              <span class="loading loading-spinner loading-sm mr-1"></span>
              Applying...
            {:else if !hasChanges}
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
              </svg>
              No Changes
            {:else}
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
              </svg>
              Apply Changes
            {/if}
          </button>
          {#if autoShowEdit}
            <button
              class="btn btn-outline btn-sm"
              on:click={() => { isEditingLabels = false; }}
              disabled={isUpdatingLabels}
              data-testid="done-button"
            >
              Done
            </button>
          {/if}
        </div>
      </div>
    {/if}

    {#if isLoadingLabels}
      <div class="flex items-center gap-2 text-sm text-base-content/70" data-testid="loading-labels">
        <span class="loading loading-spinner loading-sm"></span>
        Loading labels...
      </div>
    {/if}
  </div>
</div>
