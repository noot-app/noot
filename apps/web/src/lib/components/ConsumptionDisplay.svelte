<script lang="ts">
  import { createEventDispatcher } from "svelte"
  import { apiClient } from "$lib/api/client"
  import NutritionStats from "./NutritionStats.svelte"
  import Goals from "./Goals.svelte"
  import NutrientComposition from "./NutrientComposition.svelte"
  import Card from "./Card.svelte"
  import ConsumptionLabels from "./ConsumptionLabels.svelte"
  import PencilSquare from "./icons/PencilSquare.svelte"
  import Trash from "./icons/Trash.svelte"
  import Alert from "./Alert.svelte"

  // Props
  export let consumption: any = null // The consumption data
  export let transcript = "" // Optional transcript display
  export let showRedoButton = false // Whether to show redo button (vs delete)
  export let autoShowLabelEdit = false // Whether to automatically show label editing
  export let editable = true // Whether the consumption can be edited
  export let preloadGoalsAuto: any = null // Preloaded auto goals
  export let preloadGoalsDri: any = null // Preloaded DRI goals
  export let buttonsAtBottom = false // Whether to show action buttons at bottom instead of top

  // Local state
  let isEditing = false
  let isSubmitting = false
  let error = ""
  let editableNote = ""
  let editableTitle = ""
  let editableTimestamp = ""

  // Helper functions for local time handling
  function toLocalDateTimeString(utcDate: Date): string {
    // Convert UTC date to local datetime-local input format
    const localDate = new Date(utcDate.getTime() - (utcDate.getTimezoneOffset() * 60000))
    return localDate.toISOString().slice(0, 16)
  }

  function fromLocalDateTimeString(localDateTimeString: string): Date {
    // Convert local datetime-local input to UTC Date
    return new Date(localDateTimeString)
  }

  function setCurrentTime() {
    const now = new Date()
    editableTimestamp = toLocalDateTimeString(now)
  }

  // Form validation
  $: isValidTimestamp = !editableTimestamp || fromLocalDateTimeString(editableTimestamp) <= new Date()
  $: isTimestampInFuture = Boolean(editableTimestamp && fromLocalDateTimeString(editableTimestamp) > new Date())

    // Event dispatcher
  const dispatch = createEventDispatcher()

  // Handle label updates from ConsumptionLabels component
  function handleLabelsUpdated(event: CustomEvent<{ labels: any[] }>) {
    if (consumption) {
      consumption.labels = event.detail.labels
      consumption = { ...consumption } // Trigger reactivity
    }
  }

  // Handle label errors from ConsumptionLabels component  
  function handleLabelError(event: CustomEvent<{ message: string }>) {
    error = event.detail.message
  }

  // Editing functions
  function startEdit() {
    isEditing = true
    editableNote = consumption?.note || ""
    editableTitle = consumption?.title || ""
    // Initialize timestamp from consumption's created_at
    if (consumption?.created_at) {
      editableTimestamp = toLocalDateTimeString(new Date(consumption.created_at))
    }
    dispatch('edit')
  }

  function cancelEdit() {
    isEditing = false
    editableTimestamp = "" // Clear the editable timestamp when canceling
  }

  async function saveEdit() {
    if (!consumption?.id || !consumption?.items) return

    // Validate timestamp if provided
    if (editableTimestamp && isTimestampInFuture) {
      error = "Consumption timestamp cannot be in the future"
      return
    }

    try {
      isSubmitting = true
      error = ""

      // Prepare the update request body
      const updateBody: any = { 
        items: consumption.items,
        note: editableNote.trim() || null,
        title: editableTitle.trim() || null
      }

      // Add timestamp if it was edited
      if (editableTimestamp) {
        updateBody.consumed_at = fromLocalDateTimeString(editableTimestamp).toISOString()
      }

      const updateResponse = await apiClient.PUT("/consumption/{id}", {
        params: { path: { id: consumption.id } },
        body: updateBody,
      })

      if (updateResponse.error) {
        throw new Error(`Update failed: ${updateResponse.error}`)
      }

      consumption = updateResponse.data
      isEditing = false
      editableTimestamp = "" // Clear the editable timestamp after successful save
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

<!-- Action buttons block - defined once and used conditionally -->
{#snippet actionButtons()}
  {#if consumption?.id && editable}
    <div class="flex justify-center gap-4 mb-6">
      {#if !isEditing}
        <button
          class="btn btn-outline btn-primary"
          on:click={startEdit}
          disabled={isSubmitting}
        >
          <PencilSquare className="w-4 h-4 mr-2" />
          Edit
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
            <Trash className="w-4 h-4 mr-2" />
            Delete
          </button>
        {/if}
      {:else}
        <button
          class="btn btn-primary"
          on:click={saveEdit}
          disabled={isSubmitting || isTimestampInFuture}
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
{/snippet}

<div class="space-y-6">
  <!-- Error display -->
  {#if error}
    <Alert type="error">
      {error}
    </Alert>
  {/if}

  <!-- Show buttons at top if buttonsAtBottom is false (default behavior) -->
  {#if !buttonsAtBottom}
    {@render actionButtons()}
  {/if}

  <!-- Nutrition Summary -->
  {#if consumption?.summary}
    <Card title="Nutrition Summary">
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

  <!-- Transcript -->
  {#if transcript}
    <Card title="What you said:" compact>
      <p class="text-lg italic">"{transcript}"</p>
    </Card>
  {/if}

  <!-- Title -->
  {#if consumption?.title || (editable && isEditing)}
    <Card title="Title:" compact>
      {#if isEditing && editable}
        <div class="space-y-2">
          <input
            type="text"
            class="input input-bordered w-full"
            placeholder="Give this meal a custom title..."
            maxlength="100"
            bind:value={editableTitle}
          />
          <div class="text-xs text-base-content/60">
            {editableTitle.length}/100 characters
          </div>
        </div>
      {:else if consumption?.title}
        <p class="text-lg font-semibold">{consumption.title}</p>
      {:else}
        <p class="text-base-content/60 italic">No title added</p>
      {/if}
    </Card>
  {/if}

  <!-- Timestamp editing (only shown when editing) -->
  {#if isEditing && editable}
    <Card title="When:" compact>
      <div class="space-y-2">
        <label for="consumption-timestamp" class="block text-sm font-medium">
          Date & Time
        </label>
        <div class="flex gap-2 items-start">
          <input
            id="consumption-timestamp"
            type="datetime-local"
            class="input input-bordered flex-1 max-w-sm"
            class:input-error={!isValidTimestamp || isTimestampInFuture}
            bind:value={editableTimestamp}
            required
          />
          <button
            type="button"
            class="btn btn-outline btn-sm"
            on:click={setCurrentTime}
            title="Set to current time"
          >
            Now
          </button>
        </div>
        {#if isTimestampInFuture}
          <div class="text-xs text-error">
            Consumption time cannot be in the future
          </div>
        {:else}
          <div class="text-xs text-base-content/60">
            Enter the date and time when this meal was consumed
          </div>
        {/if}
      </div>
    </Card>
  {/if}

  <!-- Note -->
  {#if consumption?.note || (editable && isEditing)}
    <Card title="Note:" compact>
      {#if isEditing && editable}
        <div class="space-y-2">
          <textarea
            class="textarea textarea-bordered w-full"
            placeholder="Add a note about this meal..."
            maxlength="1000"
            bind:value={editableNote}
          ></textarea>
          <div class="text-xs text-base-content/60">
            {editableNote.length}/1000 characters
          </div>
        </div>
      {:else if consumption?.note}
        <p class="text-lg">{consumption.note}</p>
      {:else}
        <p class="text-base-content/60 italic">No note added</p>
      {/if}
    </Card>
  {/if}

  <!-- Labels Section -->
  {#if consumption?.id}
    <ConsumptionLabels
      consumptionId={consumption.id}
      initialLabels={consumption.labels || []}
      autoShowEdit={autoShowLabelEdit}
      editable={editable}
      on:labelsUpdated={handleLabelsUpdated}
      on:error={handleLabelError}
    />
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

  <!-- Show buttons at bottom if buttonsAtBottom is true -->
  {#if buttonsAtBottom}
    {@render actionButtons()}
  {/if}
</div>
