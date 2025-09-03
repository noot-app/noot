<script lang="ts">
  import { createEventDispatcher } from "svelte"
  import { apiClient } from "$lib/api/client"
  import NutritionStats from "./NutritionStats.svelte"
  import Goals from "./Goals.svelte"
  import NutrientComposition from "./NutrientComposition.svelte"
  import Card from "./Card.svelte"
  import Label from "./Label.svelte"
  import TagIcon from "./icons/Tag.svelte"

  // Props
  export let consumption: any = null // The consumption data
  export let transcript = "" // Optional transcript display
  export let showRedoButton = false // Whether to show redo button (vs delete)
  export let autoShowLabelEdit = false // Whether to automatically show label editing
  export let editable = true // Whether the consumption can be edited
  export let preloadGoalsAuto: any = null // Preloaded auto goals
  export let preloadGoalsDri: any = null // Preloaded DRI goals

  // Local state
  let isEditing = false
  let isSubmitting = false
  let error = ""
  let availableLabels: any[] = []
  let isLoadingLabels = false
  let isUpdatingLabels = false
  let labelsLoaded = false
  let selectedLabels: string[] = [] // Simple array of label names - what we want the final state to be
  let currentLabels: string[] = [] // What's currently on the server
  let isEditingLabels = false

  // Reactive variables - much simpler now
  $: hasChanges = JSON.stringify([...selectedLabels].sort()) !== JSON.stringify([...currentLabels].sort())

  // Event dispatcher
  const dispatch = createEventDispatcher()

  // Initialize edit state based on autoShowLabelEdit
  $: if (autoShowLabelEdit && !isEditingLabels && consumption?.id) {
    isEditingLabels = true
  }

  // Load labels when consumption is available
  $: if (consumption?.id && !labelsLoaded) {
    loadLabels()
  }

  // Sync currentLabels from server state whenever consumption changes
  $: if (consumption?.labels) {
    currentLabels = consumption.labels.map((l: any) => l.name as string).filter(Boolean)
    // Initialize selectedLabels from currentLabels if not editing yet
    if (!isEditingLabels) {
      selectedLabels = [...currentLabels]
    }
  }

  // Load user's available labels
  async function loadLabels() {
    if (isLoadingLabels || labelsLoaded || !consumption?.id) return

    try {
      isLoadingLabels = true
      const response = await apiClient.GET("/labels")

      if (response.error) {
        console.error("Failed to load labels:", response.error)
        return
      }

      availableLabels = response.data?.labels || []

      // Load labels assigned to this consumption if not already in consumption data
      if (!consumption?.labels) {
        try {
          const assigned = await apiClient.GET("/consumption/{id}/labels", {
            params: { path: { id: consumption.id } },
          })
          if (!assigned.error) {
            const assignedLabels = assigned.data?.labels || []
            // Update currentLabels from server
            currentLabels = assignedLabels.map((l: any) => l.name as string)
            // Initialize selectedLabels if not editing
            if (!isEditingLabels) {
              selectedLabels = [...currentLabels]
            }
          }
        } catch (e) {
          console.warn("Error loading consumption labels:", e)
        }
      }
    } catch (err) {
      console.error("Error loading labels:", err)
    } finally {
      isLoadingLabels = false
      labelsLoaded = true
    }
  }

  function retryLoadLabels() {
    if (!consumption?.id) return
    labelsLoaded = false
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
    if (!consumption?.id || isUpdatingLabels || !hasChanges) return

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

      // Send complete desired state - like GitHub Issues does with PUT/PATCH
      const response = await apiClient.POST("/consumption/{id}/labels", {
        params: { path: { id: consumption.id } },
        body: { ids: labelIds },
      })

      if (response.error) {
        throw new Error(`Failed to update labels: ${response.error}`)
      }

      // Update from server response
      const updatedLabels = response.data?.labels || []
      currentLabels = updatedLabels.map((l: any) => l.name as string)
      
      // Update consumption object
      if (consumption) {
        consumption.labels = updatedLabels
        consumption = { ...consumption }
      }

      // Keep selectedLabels as is - user's current selection stays active
      
      // Only close editor if not in auto-show mode
      if (!autoShowLabelEdit) {
        isEditingLabels = false
      }
    } catch (err) {
      error = `Error applying label changes: ${err}`
      console.error("Apply label changes error:", err)
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

  // Editing functions
  function startEdit() {
    isEditing = true
    dispatch('edit')
  }

  function cancelEdit() {
    isEditing = false
  }

  async function saveEdit() {
    if (!consumption?.id || !consumption?.items) return

    try {
      isSubmitting = true
      error = ""

      const updateResponse = await apiClient.PUT("/consumption/{id}", {
        params: { path: { id: consumption.id } },
        body: { items: consumption.items },
      })

      if (updateResponse.error) {
        throw new Error(`Update failed: ${updateResponse.error}`)
      }

      consumption = updateResponse.data
      isEditing = false
      dispatch('save', { consumption })
    } catch (err) {
      error = `Error updating consumption: ${err}`
      console.error("Update error:", err)
    } finally {
      isSubmitting = false
    }
  }

  // Function to update quantity and recalculate nutrition
  function updateItemQuantity(itemIndex: number, newQuantity: number) {
    if (!consumption?.items || !consumption.items[itemIndex]) return

    const item = consumption.items[itemIndex]
    const currentQuantity = item.item.user_quantity || 1
    const scalingFactor = newQuantity / currentQuantity

    // Scale all nutrition values
    if (item.item.nutrients) {
      const nutrients = item.item.nutrients
      Object.keys(nutrients).forEach((key) => {
        if (typeof nutrients[key] === "number") {
          nutrients[key] *= scalingFactor
        }
      })
    }

    // Update user quantity and grams
    item.item.user_quantity = newQuantity
    item.item.grams = item.item.grams * scalingFactor

    // Trigger reactivity
    consumption = { ...consumption, items: [...consumption.items] }
  }

  // Function to aggregate nutrition from consumption for goals
  function getMealNutrition(): Record<string, number> {
    if (!consumption?.items) {
      return {}
    }

    const aggregated: Record<string, number> = {}

    consumption.items.forEach((item: any) => {
      if (item.item?.nutrients) {
        Object.keys(item.item.nutrients).forEach((key) => {
          const value = item.item.nutrients[key]
          if (typeof value === "number") {
            aggregated[key] = (aggregated[key] || 0) + value
          }
        })
      }
    })

    return aggregated
  }

  $: currentMealNutrition = consumption?.items ? getMealNutrition() : {}

  // Action handlers
  function handleRedo() {
    dispatch('redo')
  }

  function handleDelete() {
    dispatch('delete')
  }
</script>

<div class="space-y-6">
  <!-- Error display -->
  {#if error}
    <div class="alert alert-error">
      <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
      </svg>
      <span>{error}</span>
    </div>
  {/if}

  <!-- Action buttons -->
  {#if consumption?.id && editable}
    <div class="flex justify-center gap-4 mb-6">
      {#if !isEditing}
        <button
          class="btn btn-outline btn-primary"
          on:click={startEdit}
          disabled={isSubmitting}
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
          </svg>
          ✏️ Edit
        </button>
        
        {#if showRedoButton}
          <button
            class="btn btn-outline btn-error"
            on:click={handleRedo}
            disabled={isSubmitting}
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
            </svg>
            🔄 Redo
          </button>
        {:else}
          <button
            class="btn btn-outline btn-error"
            on:click={handleDelete}
            disabled={isSubmitting}
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
            </svg>
            🗑️ Delete
          </button>
        {/if}
      {:else}
        <button
          class="btn btn-primary"
          on:click={saveEdit}
          disabled={isSubmitting}
        >
          {#if isSubmitting}
            <span class="loading loading-spinner loading-sm mr-2"></span>
          {:else}
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
            </svg>
          {/if}
          Save Changes
        </button>
        <button
          class="btn btn-outline btn-ghost"
          on:click={cancelEdit}
          disabled={isSubmitting}
        >
          Cancel
        </button>
      {/if}
    </div>
  {/if}

  <!-- Transcript -->
  {#if transcript}
    <Card title="What you said:" compact>
      <p class="text-lg italic">"{transcript}"</p>
    </Card>
  {/if}

  <!-- Nutrition Summary -->
  {#if consumption?.summary}
    <Card title="Nutrition Summary" variant="primary">
      <NutritionStats
        calories={consumption.summary.totals.calories}
        protein={consumption.summary.totals.protein_g}
        carbs={consumption.summary.totals.total_carbs_g}
        fat={consumption.summary.totals.total_fat_g}
        size="compact"
        className="bg-transparent shadow-none"
      />
    </Card>
  {/if}

  <!-- Labels Section -->
  {#if consumption?.id}
    <div class="card bg-base-200 shadow-lg">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <h3 class="card-title text-sm flex items-center gap-2">
            <TagIcon className="w-4 h-4" />
            Labels
          </h3>

          {#if !isEditingLabels && !autoShowLabelEdit}
            <button
              class="btn btn-outline btn-sm"
              on:click={startLabelEditing}
              disabled={isUpdatingLabels}
            >
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
              </svg>
              Edit Labels
            </button>
          {/if}
        </div>

        {#if autoShowLabelEdit && isEditingLabels}
          <div class="alert alert-info mb-4 text-sm">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            Select labels to apply to this meal. Click "Apply" to save changes, or "Done" when finished.
          </div>
        {/if}

                <!-- Applied Labels Display (when not editing) -->
        {#if !isEditingLabels}
          {#if currentLabels.length > 0}
            <div class="flex flex-wrap gap-2">
              {#each currentLabels as labelName}
                {@const labelData = availableLabels.find(
                  (l) => l.name === labelName,
                )}
                {#if labelData}
                  <Label name={labelData.name} color={labelData.color} />
                {/if}
              {/each}
            </div>
          {:else}
            <div class="flex items-center gap-2">
              <p class="text-sm text-base-content/70">No labels applied to this meal</p>
              {#if availableLabels.length === 0 && !isLoadingLabels}
                <button class="btn btn-xs" on:click={retryLoadLabels}>Load labels</button>
              {/if}
            </div>
          {/if}
        {/if}

        <!-- Label Selection Interface -->
        {#if isEditingLabels}
          <div class="space-y-4">
            {#if availableLabels.length === 0 && !isLoadingLabels}
              <div class="flex items-center gap-2">
                <p class="text-sm text-base-content/70">No labels available for selection.</p>
                <button class="btn btn-xs" on:click={retryLoadLabels}>Retry</button>
              </div>
            {/if}
            <div class="grid grid-cols-1 gap-2 max-h-64 overflow-y-auto">
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
                >
                  <div class="flex items-center gap-3">
                    <div class="checkbox-wrapper">
                      <input
                        type="checkbox"
                        class="checkbox checkbox-primary checkbox-sm"
                        checked={isSelected}
                        readonly
                      />
                    </div>
                    <div class="flex items-center gap-2">
                      <div
                        class="w-3 h-3 rounded-full"
                        style="background-color: {label.color?.startsWith('#')
                          ? label.color
                          : `#${label.color}`};"
                      ></div>
                      <span class="font-medium">{label.name}</span>
                      {#if isCurrentlyApplied && isSelected}
                        <span class="badge badge-xs badge-success">Applied</span>
                      {:else if isNewSelection}
                        <span class="badge badge-xs badge-warning">+Add</span>
                      {:else if willBeRemoved}
                        <span class="badge badge-xs badge-error">-Remove</span>
                      {/if}
                    </div>
                  </div>
                  {#if label.description}
                    <span class="text-xs text-base-content/60 truncate ml-2"
                      >{label.description}</span
                    >
                  {/if}
                </button>
              {/each}
            </div>

            {#if hasChanges}
              <div class="bg-base-200 p-3 rounded-lg text-sm">
                <div class="font-medium mb-1">Pending changes:</div>
                <div class="flex flex-wrap gap-2">
                  {#each selectedLabels.filter(name => !currentLabels.includes(name)) as labelName}
                    <span class="badge badge-warning badge-sm">+{labelName}</span>
                  {/each}
                  {#each currentLabels.filter(name => !selectedLabels.includes(name)) as labelName}
                    <span class="badge badge-error badge-sm">-{labelName}</span>
                  {/each}
                </div>
              </div>
            {:else if currentLabels.length > 0}
              <div class="bg-success/10 p-3 rounded-lg text-sm text-success-content">
                <div class="flex items-center gap-2">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                  </svg>
                  All selected labels are already applied
                </div>
              </div>
            {/if}

            <!-- Action Buttons -->
            <div class="flex justify-end gap-2 pt-2 border-t border-base-300">
              {#if !autoShowLabelEdit}
                <button
                  class="btn btn-ghost btn-sm"
                  on:click={cancelLabelChanges}
                  disabled={isUpdatingLabels}
                >
                  Cancel
                </button>
              {/if}
              <button
                class="btn btn-primary btn-sm"
                on:click={applyLabelChanges}
                disabled={isUpdatingLabels || !hasChanges}
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
                  {autoShowLabelEdit ? `Apply Changes` : 'Apply Changes'}
                {/if}
              </button>
              {#if autoShowLabelEdit}
                <button
                  class="btn btn-outline btn-sm"
                  on:click={() => { isEditingLabels = false; }}
                  disabled={isUpdatingLabels}
                >
                  Done
                </button>
              {/if}
            </div>
          </div>
        {/if}

        {#if isLoadingLabels}
          <div class="flex items-center gap-2 text-sm text-base-content/70">
            <span class="loading loading-spinner loading-sm"></span>
            Loading labels...
          </div>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Food Items -->
  {#if consumption?.items && consumption.items.length > 0}
    <div class="card bg-base-200 shadow-lg">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <h3 class="card-title text-sm">Items in this consumption</h3>
          {#if isEditing}
            <span class="badge badge-warning">Editing Mode</span>
          {/if}
        </div>
        <div class="space-y-3">
          {#each consumption.items as item, index}
            <div class="card bg-base-100 shadow">
              <div class="card-body p-4">
                <h4 class="font-semibold">{item.item.name}</h4>

                <!-- Quantity controls -->
                <div class="flex items-center gap-2 mt-2">
                  {#if isEditing}
                    <div class="flex items-center gap-2">
                      <label for="quantity-{index}" class="text-sm font-medium">Quantity:</label>
                      <button
                        class="btn btn-circle btn-sm btn-outline"
                        on:click={() =>
                          updateItemQuantity(
                            index,
                            Math.max(0.1, (item.item.user_quantity || 1) - 0.5),
                          )}
                      >
                        -
                      </button>
                      <input
                        id="quantity-{index}"
                        type="number"
                        class="input input-sm input-bordered w-20 text-center"
                        value={item.item.user_quantity}
                        on:input={(e) => {
                          const target = e.target as HTMLInputElement
                          updateItemQuantity(index, parseFloat(target.value) || 1)
                        }}
                        min="0.1"
                        step="0.5"
                      />
                      <button
                        class="btn btn-circle btn-sm btn-outline"
                        on:click={() =>
                          updateItemQuantity(index, (item.item.user_quantity || 1) + 0.5)}
                      >
                        +
                      </button>
                      {#if item.item.user_unit}
                        <span class="text-sm text-base-content/70">{item.item.user_unit}</span>
                      {/if}
                    </div>
                  {:else}
                    <div class="flex items-center gap-2">
                      {#if item.item.user_quantity && item.item.user_unit}
                        <p class="text-sm text-base-content/70">
                          {item.item.user_quantity}
                          {item.item.user_unit}
                          {#if item.item.user_unit !== "g" && item.item.user_unit !== "gram" && item.item.user_unit !== "grams"}
                            <span class="text-xs text-base-content/50">
                              ({Math.round(item.item.grams)}g)
                            </span>
                          {/if}
                        </p>
                      {:else}
                        <p class="text-sm text-base-content/70">{Math.round(item.item.grams)}g</p>
                      {/if}
                    </div>
                  {/if}
                </div>

                {#if item.item.nutrients}
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm mt-3">
                    <div class="bg-base-200 rounded p-2">
                      <span class="font-medium">Calories:</span>
                      {Math.round(item.item.nutrients.calories)}
                    </div>
                    <div class="bg-base-200 rounded p-2">
                      <span class="font-medium">Protein:</span>
                      {item.item.nutrients.protein_g.toFixed(1)}g
                    </div>
                    <div class="bg-base-200 rounded p-2">
                      <span class="font-medium">Carbohydrates:</span>
                      {item.item.nutrients.total_carbs_g.toFixed(1)}g
                    </div>
                    <div class="bg-base-200 rounded p-2">
                      <span class="font-medium">Total Fat:</span>
                      {item.item.nutrients.total_fat_g.toFixed(1)}g
                    </div>
                  </div>

                  <!-- Nutrient Composition Dropdown -->
                  <NutrientComposition nutrients={item.item.nutrients} />
                {/if}
              </div>
            </div>
          {/each}
        </div>
      </div>
    </div>
  {/if}

  <!-- Nutrition Targets - Modified to remove Goals Met section -->
  {#if consumption?.items && consumption.items.length > 0}
    <div class="space-y-4">
      <Goals
        currentNutrition={currentMealNutrition}
        showMealContribution={true}
        hideGoalsMet={true}
        {preloadGoalsAuto}
        {preloadGoalsDri}
      />
    </div>
  {/if}
</div>
