<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { toast } from "$lib/stores/toast"
  import { formatErrorForUser } from "$lib/utils/error-handling"
  import FormSelect from "./FormSelect.svelte"
  import type { paths } from "$lib/api/schema"

  type EventType = paths["/event-types"]["get"]["responses"]["200"]["content"]["application/json"]["event_types"][0]
  type CreateEventRequest = paths["/events"]["post"]["requestBody"]["content"]["application/json"]
  type ConsumptionsResponse = paths["/consumptions"]["get"]["responses"]["200"]["content"]["application/json"]
  type Consumption = ConsumptionsResponse["consumptions"][0]
  type EventLinkCreateRequest = paths["/events/{id}/links"]["post"]["requestBody"]["content"]["application/json"]

  export let eventTypes: EventType[] = []
  export let onEventCreated: () => void = () => {}

  let selectedEventTypeId = ""
  $: selectedEventType = eventTypes.find(et => et.id === selectedEventTypeId)
  
  // Default to "symptom" event type when available
  $: if (eventTypes.length > 0 && !selectedEventTypeId) {
    const symptomType = eventTypes.find(et => et.name.toLowerCase() === 'symptom')
    if (symptomType) {
      selectedEventTypeId = symptomType.id
    }
  }
  
  // Event type options for dropdown
  $: eventTypeOptions = [
    { value: "", label: "Select event type..." },
    ...eventTypes.map(type => ({ 
      value: type.id, 
      label: type.name 
    }))
  ]

  let eventLevel: number | undefined = undefined
  let eventNote = ""
  let eventTitle = ""
  let hasEndTime = false
  let eventEndedAt = ""
  let isCreating = false

  // Quick link to last consumption
  let linkToLastConsumption = false
  let lastConsumption: Consumption | null = null
  let isLoadingConsumption = false
  let consumptionError = ""

  // Helper functions for time handling
  function toLocalDateTimeString(utcDate: Date): string {
    const localDate = new Date(utcDate.getTime() - (utcDate.getTimezoneOffset() * 60000))
    return localDate.toISOString().slice(0, 16)
  }

  function fromLocalDateTimeString(localDateTimeString: string): Date {
    return new Date(localDateTimeString)
  }

  // Default to current time
  let eventStartedAt = toLocalDateTimeString(new Date())
  
  // Reset derived values when event type changes
  $: if (selectedEventType) {
    if (!eventTitle.trim()) {
      eventTitle = selectedEventType.default_name || selectedEventType.name
    }
    if (!eventNote.trim() && selectedEventType.description) {
      eventNote = selectedEventType.description
    }
  }
  
  // Clear overrides when switching event types
  $: if (selectedEventTypeId) {
    eventTitle = selectedEventType?.default_name || selectedEventType?.name || ""
    eventNote = selectedEventType?.description || ""
    hasEndTime = false
    eventEndedAt = ""
  }

  function incrementLevel() {
    if (eventLevel === undefined) {
      eventLevel = 1
    } else if (eventLevel < 10) {
      eventLevel++
    }
  }

  function decrementLevel() {
    if (eventLevel === undefined) {
      eventLevel = 0
    } else if (eventLevel > 0) {
      eventLevel--
    } else {
      eventLevel = undefined
    }
  }

  function resetToNow() {
    eventStartedAt = toLocalDateTimeString(new Date())
  }

  // Auto-set end time when hasEndTime is toggled
  $: if (hasEndTime && !eventEndedAt) {
    // Default end time to 1 hour after start time
    const startDate = fromLocalDateTimeString(eventStartedAt)
    const endDate = new Date(startDate.getTime() + (60 * 60 * 1000))
    eventEndedAt = toLocalDateTimeString(endDate)
  }

  // Fetch latest consumption when quick link option is toggled
  async function fetchLatestConsumption() {
    if (!linkToLastConsumption || lastConsumption) return
    
    try {
      isLoadingConsumption = true
      consumptionError = ""
      
      const response = await apiClient.GET("/consumptions", {
        params: { query: { days: 1 } } // Get consumptions from last 24 hours
      })
      
      if (response.error) {
        consumptionError = "Failed to fetch recent consumptions"
        return
      }
      
      if (!response.data?.consumptions || response.data.consumptions.length === 0) {
        consumptionError = "No recent consumptions found"
        lastConsumption = null
        return
      }
      
      // Get the most recent consumption (they should be sorted by created_at desc)
      lastConsumption = response.data.consumptions[0]
    } catch (err) {
      console.error("Error fetching latest consumption:", err)
      consumptionError = "Failed to load recent consumptions"
    } finally {
      isLoadingConsumption = false
    }
  }

  // Watch for changes in linkToLastConsumption toggle
  $: if (linkToLastConsumption) {
    fetchLatestConsumption()
  } else {
    // Clear consumption data when toggled off
    lastConsumption = null
    consumptionError = ""
  }

  async function quickAddEvent() {
    if (isCreating || !selectedEventType) return

    try {
      isCreating = true
      
      const eventData: CreateEventRequest = {
        name: eventTitle.trim() || selectedEventType.default_name || selectedEventType.name,
        event_type_id: selectedEventType.id,
        started_at: fromLocalDateTimeString(eventStartedAt).toISOString(),
        ended_at: hasEndTime && eventEndedAt ? fromLocalDateTimeString(eventEndedAt).toISOString() : undefined,
        level: eventLevel,
        note: eventNote.trim() || undefined,
        color: selectedEventType.color,
      }

      const response = await apiClient.POST("/events", {
        body: eventData,
      })

      if (response.error) {
        toast.error(formatErrorForUser(response.error))
        return
      }

      const createdEvent = response.data
      let linkSuccess = true

      // Create link to last consumption if requested and available
      if (linkToLastConsumption && lastConsumption && createdEvent) {
        try {
          const linkData: EventLinkCreateRequest = {
            consumption_id: lastConsumption.id
          }

          const linkResponse = await apiClient.POST("/events/{id}/links", {
            params: { path: { id: createdEvent.id } },
            body: linkData
          })

          if (linkResponse.error) {
            console.error("Failed to create event-consumption link:", linkResponse.error)
            linkSuccess = false
          }
        } catch (linkErr) {
          console.error("Error creating event-consumption link:", linkErr)
          linkSuccess = false
        }
      }

      // Show appropriate success message
      if (linkToLastConsumption && lastConsumption) {
        if (linkSuccess) {
          toast.success(`${selectedEventType.name} event created and linked to consumption`)
        } else {
          toast.success(`${selectedEventType.name} event created (linking failed)`)
        }
      } else {
        toast.success(`${selectedEventType.name} event created successfully`)
      }
      
      // Reset form but keep event type selected
      eventLevel = undefined
      eventNote = selectedEventType?.description || ""
      eventTitle = selectedEventType?.default_name || selectedEventType?.name || ""
      hasEndTime = false
      eventEndedAt = ""
      linkToLastConsumption = false
      lastConsumption = null
      consumptionError = ""
      resetToNow()
      
      onEventCreated()
    } catch (err) {
      console.error("Error creating event:", err)
      toast.error("Failed to create event. Please try again.")
    } finally {
      isCreating = false
    }
  }

  // Form validation
  $: isValidLevel = eventLevel === undefined || (eventLevel >= 0 && eventLevel <= 10)
  $: isValidNote = eventNote.length <= 1000
  $: isValidTitle = eventTitle.length <= 100
  $: isValidEndTime = !hasEndTime || !eventEndedAt || fromLocalDateTimeString(eventEndedAt) >= fromLocalDateTimeString(eventStartedAt)
  $: canAdd = isValidLevel && isValidNote && isValidTitle && isValidEndTime && selectedEventType
</script>

<div class="card bg-base-100 shadow-xl border-2 border-primary/20">
  <div class="card-body">
    <div class="flex items-center justify-between mb-4">
      <h3 class="font-bold text-lg">Quick Add Event</h3>
      <button
        class="btn btn-primary"
        disabled={!canAdd || isCreating}
        on:click={quickAddEvent}
      >
        {#if isCreating}
          <span class="loading loading-spinner loading-sm"></span>
          Adding...
        {:else}
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 6v6m0 0v6m0-6h6m-6 0H6"
            />
          </svg>
          Add Event
        {/if}
      </button>
    </div>

    <!-- Event Type Selection -->
    <div class="mb-4">
      <FormSelect
        id="quickAddEventType"
        label="Event Type"
        bind:value={selectedEventTypeId}
        options={eventTypeOptions}
        required
      />
    </div>

    {#if selectedEventType}
      <!-- Show event type specific content -->
      <div class="bg-base-200 rounded-lg border border-base-300 p-4">
        <div class="flex items-center gap-3 mb-6">
          <div 
            class="w-4 h-4 rounded-full flex-shrink-0" 
            style="background-color: #{selectedEventType.color}"
          ></div>
          <div class="flex-1 min-w-0">
            <h4 class="font-semibold text-base">
              {selectedEventType.name}
            </h4>
            <p class="text-sm text-base-content/70 truncate">
              {eventTitle.trim() || selectedEventType.default_name || selectedEventType.name}
            </p>
          </div>
        </div>

        <!-- PROMINENT LEVEL CONTROLS - Mobile First Design -->
        <div class="mb-8">
          <div class="text-center mb-4">
            <h5 class="text-lg font-bold text-base-content/80">Level</h5>
            <p class="text-sm text-base-content/60">Tap to adjust intensity (0-10)</p>
          </div>
          
          <div class="flex items-center justify-center gap-8">
            <button
              type="button"
              class="btn btn-circle btn-xl btn-ghost hover:btn-primary text-3xl"
              on:click={decrementLevel}
              disabled={eventLevel === 0}
              aria-label="Decrease level"
            >
              <svg class="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="3"
                  d="M20 12H4"
                />
              </svg>
            </button>
            
            <div class="flex flex-col items-center">
              <div class="w-20 h-20 bg-base-100 border-2 border-primary/30 rounded-full flex items-center justify-center shadow-lg">
                <span 
                  id="quickAddLevel"
                  class="text-4xl font-black font-mono text-primary"
                  role="textbox"
                  aria-readonly="true"
                >
                  {eventLevel ?? '–'}
                </span>
              </div>
            </div>
            
            <button
              type="button"
              class="btn btn-circle btn-xl btn-ghost hover:btn-primary text-3xl"
              on:click={incrementLevel}
              disabled={eventLevel === 10}
              aria-label="Increase level"
            >
              <svg class="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="3"
                  d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                />
              </svg>
            </button>
          </div>
        </div>

        <!-- COLLAPSIBLE ADVANCED OPTIONS -->
        <div class="collapse collapse-arrow bg-base-100 border border-base-300">
          <input type="checkbox" /> 
          <div class="collapse-title text-base font-medium">
            <div class="flex items-center gap-2">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              Advanced Options
            </div>
          </div>
          <div class="collapse-content">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-2">
              <!-- Left Column: Input Fields -->
              <div class="space-y-4">
                <!-- Event Title Override -->
                <div class="form-control">
                  <label class="label" for="quickAddTitle">
                    <span class="label-text font-medium">Event Title</span>
                    <span class="label-text-alt text-xs opacity-60">Optional override</span>
                  </label>
                  <input
                    id="quickAddTitle"
                    type="text"
                    class="input input-bordered input-sm"
                    bind:value={eventTitle}
                    placeholder={selectedEventType.default_name || selectedEventType.name}
                    maxlength={100}
                  />
                  {#if !isValidTitle}
                    <div class="label">
                      <span class="label-text-alt text-error">Title must be 100 characters or less</span>
                    </div>
                  {/if}
                </div>

                <!-- Start Time -->
                <div class="form-control">
                  <label class="label" for="quickAddStartTime">
                    <span class="label-text font-medium">Start Time</span>
                  </label>
                  <div class="flex gap-2">
                    <input
                      id="quickAddStartTime"
                      type="datetime-local"
                      class="input input-bordered input-sm flex-1"
                      bind:value={eventStartedAt}
                    />
                    <button
                      type="button"
                      class="btn btn-ghost btn-sm btn-square"
                      on:click={resetToNow}
                      title="Set to now"
                      aria-label="Set to now"
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                      </svg>
                    </button>
                  </div>
                </div>

                <!-- End Time (Optional) -->
                <div class="form-control">
                  <label class="label" for="quickAddEndTimeToggle">
                    <span class="label-text font-medium">End Time</span>
                    <span class="label-text-alt text-xs opacity-60">Optional</span>
                  </label>
                  <div class="space-y-2">
                    <label class="cursor-pointer label justify-start gap-2 py-1">
                      <input
                        id="quickAddEndTimeToggle"
                        type="checkbox"
                        class="checkbox checkbox-sm"
                        bind:checked={hasEndTime}
                      />
                      <span class="label-text">Set end time</span>
                    </label>
                    {#if hasEndTime}
                      <input
                        type="datetime-local"
                        class="input input-bordered input-sm w-full"
                        bind:value={eventEndedAt}
                      />
                      {#if !isValidEndTime}
                        <div class="label">
                          <span class="label-text-alt text-error">End time must be after start time</span>
                        </div>
                      {/if}
                    {/if}
                  </div>
                </div>

                <!-- Note -->
                <div class="form-control">
                  <div class="mb-2">
                    <span class="label-text font-medium">Note</span>
                    <span class="label-text-alt text-xs opacity-60 ml-2">Optional</span>
                  </div>
                  <textarea
                    id="quickAddNote"
                    class="textarea textarea-bordered textarea-sm h-20"
                    placeholder={selectedEventType.description || "Add any additional notes..."}
                    bind:value={eventNote}
                    maxlength={1000}
                  ></textarea>
                  {#if !isValidNote}
                    <div class="label">
                      <span class="label-text-alt text-error">Note must be 1000 characters or less</span>
                    </div>
                  {/if}
                </div>

                <!-- Quick Link to Last Consumption -->
                <div class="form-control">
                  <div class="mb-2">
                    <span class="label-text font-medium">Link to Consumption</span>
                    <span class="label-text-alt text-xs opacity-60 ml-2">Optional</span>
                  </div>
                  <label class="cursor-pointer label justify-start gap-2 py-1">
                    <input
                      type="checkbox"
                      class="checkbox checkbox-sm"
                      bind:checked={linkToLastConsumption}
                      disabled={isLoadingConsumption}
                    />
                    <span class="label-text">Link to last consumption</span>
                    {#if isLoadingConsumption}
                      <span class="loading loading-spinner loading-xs"></span>
                    {/if}
                  </label>
                  
                  {#if linkToLastConsumption}
                    {#if consumptionError}
                      <div class="alert alert-warning alert-sm mt-2">
                        <svg class="stroke-current shrink-0 w-4 h-4" fill="none" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
                        </svg>
                        <span class="text-xs">{consumptionError}</span>
                      </div>
                    {:else if lastConsumption}
                      <div class="bg-base-200 border border-base-300 rounded p-3 mt-2">
                        <div class="text-xs font-medium mb-1">Will link to:</div>
                        <div class="text-sm font-semibold">"{lastConsumption.transcript}"</div>
                        <div class="text-xs text-base-content/60 mt-1">
                          {new Date(lastConsumption.created_at).toLocaleString()}
                        </div>
                        {#if lastConsumption.summary?.totals?.calories}
                          <div class="text-xs text-base-content/60 mt-1">
                            {Math.round(lastConsumption.summary.totals.calories)} cal
                          </div>
                        {/if}
                      </div>
                    {/if}
                  {/if}
                </div>
              </div>

              <!-- Right Column: Preview -->
              <div class="space-y-4">
                <div class="sticky top-4">
                  <h5 class="font-semibold text-base mb-3">Preview</h5>
                  <div class="bg-base-200 border border-base-300 rounded-lg p-4 space-y-3">
                    <div class="flex items-start gap-3">
                      <div 
                        class="w-3 h-3 rounded-full flex-shrink-0 mt-1" 
                        style="background-color: #{selectedEventType.color}"
                      ></div>
                      <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm break-words">
                          {eventTitle.trim() || selectedEventType.default_name || selectedEventType.name}
                        </div>
                        <div class="text-xs text-base-content/60 mt-1">
                          {selectedEventType.name}
                        </div>
                      </div>
                    </div>
                    
                    <div class="divider my-2"></div>
                    
                    <div class="space-y-2 text-xs">
                      <div class="flex justify-between">
                        <span class="text-base-content/60">Start:</span>
                        <span class="font-mono">
                          {new Date(eventStartedAt).toLocaleString()}
                        </span>
                      </div>
                      
                      <div class="flex justify-between">
                        <span class="text-base-content/60">End:</span>
                        <span class="font-mono">
                          {hasEndTime && eventEndedAt ? new Date(eventEndedAt).toLocaleString() : 'Not set'}
                        </span>
                      </div>
                      
                      <div class="flex justify-between">
                        <span class="text-base-content/60">Level:</span>
                        <span class="font-mono">
                          {eventLevel !== undefined ? eventLevel : 'Not set'}
                        </span>
                      </div>
                      
                      <div class="flex justify-between items-start">
                        <span class="text-base-content/60 flex-shrink-0">Note:</span>
                        <span class="text-right break-words ml-2 max-w-40">
                          {eventNote.trim() || 'None'}
                        </span>
                      </div>
                      
                      <div class="flex justify-between items-start">
                        <span class="text-base-content/60 flex-shrink-0">Link:</span>
                        <span class="text-right break-words ml-2 max-w-40">
                          {#if linkToLastConsumption && lastConsumption}
                            Last consumption
                          {:else if linkToLastConsumption && consumptionError}
                            Error loading
                          {:else if linkToLastConsumption && isLoadingConsumption}
                            Loading...
                          {:else}
                            None
                          {/if}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    {:else}
      <!-- Show placeholder when no event type is selected -->
      <div class="bg-base-200 rounded-lg border border-base-300 p-8 text-center">
        <svg class="w-12 h-12 mx-auto text-base-content/30 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
          />
        </svg>
        <p class="text-base-content/60">
          Select an event type above to quickly add an event with pre-filled defaults
        </p>
      </div>
    {/if}
  </div>
</div>
