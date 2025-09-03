<script lang="ts">
  import { onMount } from "svelte"
  import { apiClient } from "$lib/api/client"
  import { toast } from "$lib/stores/toast"
  import Toast from "$lib/components/Toast.svelte"
  import FormField from "$lib/components/FormField.svelte"
  import FormSelect from "$lib/components/FormSelect.svelte"
  import ConfirmModal from "$lib/components/ConfirmModal.svelte"
  import ColorPicker from "$lib/components/ColorPicker.svelte"
  import QuickAddEvent from "$lib/components/QuickAddEvent.svelte"
  import CalendarIcon from "$lib/components/icons/calendar-days.svelte"
  import { formatErrorForUser } from "$lib/utils/error-handling"
  import type { paths } from "$lib/api/schema"

  type EventsResponse =
    paths["/events"]["get"]["responses"]["200"]["content"]["application/json"]
  type Event = EventsResponse["events"][0]
  type CreateEventRequest =
    paths["/events"]["post"]["requestBody"]["content"]["application/json"]
  type EventTypesResponse = 
    paths["/event-types"]["get"]["responses"]["200"]["content"]["application/json"]
  type EventType = EventTypesResponse["event_types"][0]
  type CreateEventTypeRequest =
    paths["/event-types"]["post"]["requestBody"]["content"]["application/json"]
  type EventLinksResponse = 
    paths["/events/{id}/links"]["get"]["responses"]["200"]["content"]["application/json"]
  type EventLink = NonNullable<EventLinksResponse["links"]>[0]

  let events: Event[] = []
  let eventTypes: EventType[] = []
  let loading = true
  let loadingEventTypes = false
  let error = ""

  // Event links cache
  let eventLinks: Map<string, EventLink[]> = new Map()

  // Filtering and sorting
  let filterEventType = ""
  let filterStartDate = ""
  let filterEndDate = ""
  let sortBy = "started_at"
  let sortOrder = "desc"

  // New/Edit event modal state
  let showModal = false
  let editingEvent: Event | null = null
  let modalTitle = ""
  let eventName = ""
  let eventTypeId = ""
  let eventStartedAt = ""
  let eventEndedAt = ""
  let eventLevel: string | number | undefined = undefined
  let eventNote = ""
  let eventColor = ""

  // Consumption linking for modal
  let linkToLastConsumption = false
  let lastConsumption: any = null
  let isLoadingConsumption = false
  let consumptionError = ""

  // New event type modal state
  let showEventTypeModal = false
  let newEventTypeName = ""
  let newEventTypeDescription = ""
  let newEventTypeColor = "#FFD700"
  let newEventTypeDefaultName = ""

  // Delete confirmation modal
  let showDeleteModal = false
  let eventToDelete: Event | null = null

  // Derived event type options for dropdowns
  $: eventTypeOptions = [
    { value: "", label: "All Event Types" },
    ...eventTypes.map(type => ({ value: type.id, label: type.name }))
  ]

  $: createEventTypeOptions = eventTypes.map(type => ({ 
    value: type.id, 
    label: `${type.name}${type.event_count ? ` (${type.event_count} events)` : ''}` 
  }))

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

  function setCurrentTime(isEndTime = false) {
    const now = new Date()
    const localTimeString = toLocalDateTimeString(now)
    
    if (isEndTime) {
      eventEndedAt = localTimeString
    } else {
      eventStartedAt = localTimeString
    }
  }

  // Form validation
  $: isValidName = eventName.trim().length > 0 && eventName.length <= 100
  $: isValidStartDate = eventStartedAt.length > 0
  $: isValidEndDate = !eventEndedAt || eventEndedAt >= eventStartedAt
  $: isValidLevel = eventLevel === undefined || eventLevel === "" || (Number(eventLevel) >= 0 && Number(eventLevel) <= 10)
  $: isValidNote = eventNote.length <= 1000
  $: isValidColor = !eventColor || /^#?[0-9a-f]{6}$/i.test(eventColor)
  
  // Validate that events are not in the future
  $: isStartDateInFuture = eventStartedAt && fromLocalDateTimeString(eventStartedAt) > new Date()
  $: isEndDateInFuture = eventEndedAt && fromLocalDateTimeString(eventEndedAt) > new Date()
  $: isValidFutureDate = !isStartDateInFuture && !isEndDateInFuture
  
  $: canSave =
    isValidName &&
    isValidStartDate &&
    isValidEndDate &&
    isValidLevel &&
    isValidNote &&
    isValidColor &&
    isValidFutureDate

  // Event Type form validation
  $: isValidEventTypeName = 
    newEventTypeName.trim().length > 0 && 
    newEventTypeName.length <= 39 &&
    /^[a-zA-Z0-9]([a-zA-Z0-9]|-(?=[a-zA-Z0-9]))*$/.test(newEventTypeName.trim())
  $: isValidEventTypeDescription = newEventTypeDescription.length <= 250
  $: isValidEventTypeColor = /^#?[0-9a-f]{6}$/i.test(newEventTypeColor)
  $: isValidEventTypeDefaultName = newEventTypeDefaultName.length <= 100
  
  $: canSaveEventType = 
    isValidEventTypeName &&
    isValidEventTypeDescription &&
    isValidEventTypeColor &&
    isValidEventTypeDefaultName

  async function loadEventTypes() {
    try {
      loadingEventTypes = true
      const response = await apiClient.GET("/event-types")

      if (response.error) {
        console.error("Error loading event types:", response.error)
        return
      }

      eventTypes = response.data?.event_types || []
    } catch (err) {
      console.error("Error loading event types:", err)
    } finally {
      loadingEventTypes = false
    }
  }

  async function loadEventLinks(eventIds: string[]) {
    try {
      const linkPromises = eventIds.map(async (eventId) => {
        const response = await apiClient.GET("/events/{id}/links", {
          params: { path: { id: eventId } }
        })
        
        if (!response.error && response.data) {
          return { eventId, links: response.data.links }
        }
        return { eventId, links: [] }
      })

      const results = await Promise.all(linkPromises)
      eventLinks = new Map()
      results.forEach(({ eventId, links }) => {
        eventLinks.set(eventId, links || [])
      })
    } catch (err) {
      console.error("Error loading event links:", err)
    }
  }

  async function loadEvents() {
    try {
      loading = true
      const params: Record<string, string> = {}

      if (filterEventType) params.event_type_id = filterEventType
      if (filterStartDate) params.start_date = filterStartDate
      if (filterEndDate) params.end_date = filterEndDate

      const response = await apiClient.GET("/events", {
        params: { query: params },
      })

      if (response.error) {
        error = formatErrorForUser(response.error)
        return
      }

      let eventList = response.data?.events || []

      // Client-side sorting since API might not support all sorting options
      eventList.sort((a, b) => {
        let aVal: string | number = a[sortBy as keyof Event] as string | number
        let bVal: string | number = b[sortBy as keyof Event] as string | number

        if (sortBy === "started_at" || sortBy === "created_at" || sortBy === "updated_at") {
          aVal = new Date(aVal as string).getTime()
          bVal = new Date(bVal as string).getTime()
        }

        if (sortOrder === "desc") {
          return aVal > bVal ? -1 : aVal < bVal ? 1 : 0
        } else {
          return aVal > bVal ? 1 : aVal < bVal ? -1 : 0
        }
      })

      events = eventList
      
      // Load event links for all events
      if (events.length > 0) {
        await loadEventLinks(events.map(e => e.id))
      }
      
      error = ""
    } catch (err) {
      console.error("Error loading events:", err)
      error = "Failed to load events. Please try again."
    } finally {
      loading = false
    }
  }

  function openCreateModal() {
    editingEvent = null
    modalTitle = "Create Event"
    eventName = ""
    eventTypeId = ""
    eventStartedAt = toLocalDateTimeString(new Date()) // Current local time
    eventEndedAt = ""
    eventLevel = undefined
    eventNote = ""
    eventColor = "#FFD700" // Default to gold
    showModal = true
  }

  function openEditModal(event: Event) {
    editingEvent = event
    modalTitle = "Edit Event"
    eventName = event.name
    eventTypeId = event.event_type_id || ""
    // Convert UTC times to local time for editing
    eventStartedAt = toLocalDateTimeString(new Date(event.started_at))
    eventEndedAt = event.ended_at ? toLocalDateTimeString(new Date(event.ended_at)) : ""
    eventLevel = event.level ?? undefined
    eventNote = event.note || ""
    eventColor = event.color ? (event.color.startsWith("#") ? event.color : `#${event.color}`) : ""
    showModal = true
  }

  function closeModal() {
    showModal = false
    editingEvent = null
    resetForm()
  }

  function resetForm() {
    eventName = ""
    eventTypeId = ""
    eventStartedAt = ""
    eventEndedAt = ""
    eventLevel = undefined
    eventNote = ""
    eventColor = ""
    linkToLastConsumption = false
    lastConsumption = null
    consumptionError = ""
  }

  async function saveEvent() {
    if (!canSave) return

    try {
      const eventData: CreateEventRequest = {
        name: eventName.trim(),
        event_type_id: eventTypeId || undefined,
        // Convert local times to UTC for server storage
        started_at: fromLocalDateTimeString(eventStartedAt).toISOString(),
        ended_at: eventEndedAt ? fromLocalDateTimeString(eventEndedAt).toISOString() : undefined,
        level: eventLevel ? Number(eventLevel) : undefined,
        note: eventNote.trim() || undefined,
        color: eventColor ? eventColor.replace("#", "") : undefined,
      }

      if (editingEvent) {
        // Update existing event
        const response = await apiClient.PUT("/events/{id}", {
          params: { path: { id: editingEvent.id } },
          body: eventData,
        })

        if (response.error) {
          toast.error(formatErrorForUser(response.error))
          return
        }

        toast.success("Event updated successfully")
      } else {
        // Create new event
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
            const linkData = {
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
            toast.success("Event created and linked to consumption")
          } else {
            toast.success("Event created (linking failed)")
          }
        } else {
          toast.success("Event created successfully")
        }
      }

      closeModal()
      loadEvents() // Reload the events list
    } catch (err) {
      console.error("Error saving event:", err)
      toast.error("Failed to save event. Please try again.")
    }
  }

  function openEventTypeModal() {
    newEventTypeName = ""
    newEventTypeDescription = ""
    newEventTypeColor = "#FFD700"
    newEventTypeDefaultName = ""
    showEventTypeModal = true
  }

  function closeEventTypeModal() {
    showEventTypeModal = false
    newEventTypeName = ""
    newEventTypeDescription = ""
    newEventTypeColor = "#FFD700"
    newEventTypeDefaultName = ""
  }

  async function saveEventType() {
    if (!canSaveEventType) return

    try {
      const eventTypeData: CreateEventTypeRequest = {
        name: newEventTypeName.trim(),
        description: newEventTypeDescription.trim() || undefined,
        color: newEventTypeColor.replace("#", ""),
        default_name: newEventTypeDefaultName.trim() || undefined,
      }

      const response = await apiClient.POST("/event-types", {
        body: eventTypeData,
      })

      if (response.error) {
        toast.error(formatErrorForUser(response.error))
        return
      }

      toast.success("Event type created successfully")
      closeEventTypeModal()
      loadEventTypes() // Reload event types
    } catch (err) {
      console.error("Error saving event type:", err)
      toast.error("Failed to create event type. Please try again.")
    }
  }

  function openDeleteModal(event: Event) {
    eventToDelete = event
    showDeleteModal = true
  }

  function closeDeleteModal() {
    showDeleteModal = false
    eventToDelete = null
  }

  async function deleteEvent() {
    if (!eventToDelete) return

    try {
      const response = await apiClient.DELETE("/events/{id}", {
        params: { path: { id: eventToDelete.id } },
      })

      if (response.error) {
        toast.error(formatErrorForUser(response.error))
        return
      }

      toast.success("Event deleted successfully")
      closeDeleteModal()
      loadEvents() // Reload the events list
    } catch (err) {
      console.error("Error deleting event:", err)
      toast.error("Failed to delete event. Please try again.")
    }
  }

  function formatDateTime(dateStr: string) {
    return new Date(dateStr).toLocaleString()
  }

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString()
  }

  function getEventTypeName(eventTypeId: string | null | undefined) {
    if (!eventTypeId) return null
    const eventType = eventTypes.find(type => type.id === eventTypeId)
    return eventType?.name || null
  }

  function getEventTypeColor(eventTypeId: string | null | undefined) {
    if (!eventTypeId) return "#9CA3AF"
    const eventType = eventTypes.find(type => type.id === eventTypeId)
    return eventType?.color ? `#${eventType.color}` : "#9CA3AF"
  }

  function getEventLinks(eventId: string): EventLink[] {
    return eventLinks.get(eventId) || []
  }

  function hasConsumptionLinks(eventId: string): boolean {
    const links = getEventLinks(eventId)
    return links.some(link => link.consumption_id)
  }

  function getConsumptionLinks(eventId: string): EventLink[] {
    const links = getEventLinks(eventId)
    return links.filter(link => link.consumption_id)
  }

  function navigateToConsumption(consumptionId: string) {
    window.open(`/consumptions/${consumptionId}`, '_blank')
  }

  // Fetch latest consumption when modal quick link option is toggled
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

  // Watch for filter changes and reload
  $: filterEventType, filterStartDate, filterEndDate, sortBy, sortOrder, loadEvents()

  onMount(() => {
    loadEventTypes().then(() => {
      loadEvents()
    })
  })
</script>

<svelte:head>
  <title>Events - Noot</title>
  <meta
    name="description"
    content="Track your symptoms, activities, measurements, and other health events."
  />
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-8 max-w-6xl">
    <!-- Header -->
    <div class="mb-8">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-4">
        <div>
          <h1 class="text-3xl font-bold text-base-content">Events</h1>
          <p class="text-base-content/70 mt-2">
            Track symptoms, activities, measurements, and other health events.
          </p>
        </div>
        <div class="flex gap-2 shrink-0">
          <button
            class="btn btn-outline min-h-[44px]"
            on:click={openEventTypeModal}
            disabled={loadingEventTypes}
          >
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
              />
            </svg>
            New Event Type
          </button>
          <button
            class="btn btn-primary min-h-[44px]"
            on:click={openCreateModal}
          >
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 6v6m0 0v6m0-6h6m-6 0H6"
              />
            </svg>
            New Event
          </button>
        </div>
      </div>

      <!-- Quick Add Events -->
      {#if eventTypes.length > 0}
        <div class="mb-6">
          <QuickAddEvent 
            {eventTypes} 
            onEventCreated={loadEvents}
          />
        </div>
      {/if}

      <!-- Filters -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 p-4 bg-base-200 rounded-lg">
        <FormSelect
          id="eventTypeFilter"
          label="Event Type"
          bind:value={filterEventType}
          options={eventTypeOptions}
        />
        
        <FormField
          id="startDate"
          label="Start Date"
          type="date"
          bind:value={filterStartDate}
        />
        
        <FormField
          id="endDate"  
          label="End Date"
          type="date"
          bind:value={filterEndDate}
        />

        <FormSelect
          id="sortBy"
          label="Sort By"
          bind:value={sortBy}
          options={[
            { value: "started_at", label: "Start Date" },
            { value: "created_at", label: "Created" },
            { value: "name", label: "Name" },
          ]}
        />

        <FormSelect
          id="sortOrder"
          label="Order"
          bind:value={sortOrder}
          options={[
            { value: "desc", label: "Newest First" },
            { value: "asc", label: "Oldest First" },
          ]}
        />
      </div>
    </div>

    <!-- Loading state -->
    {#if loading}
      <div class="flex justify-center items-center py-12">
        <span class="loading loading-spinner loading-lg"></span>
      </div>
    {:else if error}
      <!-- Error state -->
      <div class="alert alert-error">
        <svg class="stroke-current shrink-0 w-6 h-6" fill="none" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>{error}</span>
        <div>
          <button class="btn btn-sm btn-ghost" on:click={loadEvents}>Try Again</button>
        </div>
      </div>
    {:else}
      <!-- Events list -->
      <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
          {#if events.length === 0}
            <div class="text-center py-12">
              <CalendarIcon className="w-16 h-16 mx-auto text-base-content/50 mb-4" />
              <h3 class="text-lg font-medium text-base-content/70 mb-2">No events yet</h3>
              <p class="text-base-content/50 mb-4">
                Create your first event to start tracking symptoms, activities, and more.
              </p>
              <button class="btn btn-primary" on:click={openCreateModal}>Create Event</button>
            </div>
          {:else}
            <div class="space-y-4">
              {#each events as event}
                <div
                  class="flex flex-col lg:flex-row lg:items-center lg:justify-between p-4 border border-base-300 rounded-lg hover:bg-base-50 transition-colors gap-4 lg:gap-0"
                >
                  <div class="flex-1">
                    <div class="flex items-center gap-3 mb-2">
                      <!-- Event color indicator -->
                      {#if event.color}
                        <div
                          class="w-4 h-4 rounded-full border border-base-300"
                          style="background-color: #{event.color}"
                        ></div>
                      {:else}
                        <div
                          class="w-4 h-4 rounded-full border border-base-300"
                          style="background-color: {getEventTypeColor(event.event_type_id)}"
                        ></div>
                      {/if}
                      
                      <h3 class="font-semibold text-lg">{event.name}</h3>
                      
                      {#if getEventTypeName(event.event_type_id)}
                        <span class="badge badge-outline text-xs capitalize">{getEventTypeName(event.event_type_id)}</span>
                      {/if}
                      
                      {#if event.level !== null}
                        <span class="badge badge-primary text-xs">Level: {event.level}/10</span>
                      {/if}
                    </div>

                    <div class="text-sm text-base-content/70 space-y-1">
                      <div>
                        <strong>Started:</strong> {formatDateTime(event.started_at)}
                        {#if event.ended_at}
                          <span class="mx-2">•</span>
                          <strong>Ended:</strong> {formatDateTime(event.ended_at)}
                        {/if}
                      </div>
                      
                      {#if event.note}
                        <div><strong>Note:</strong> {event.note}</div>
                      {/if}
                      
                      <!-- Consumption Links Display -->
                      {#if hasConsumptionLinks(event.id)}
                        {@const consumptionLinks = getConsumptionLinks(event.id)}
                        <div class="flex items-center gap-2 mt-2">
                          <svg class="w-4 h-4 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.102m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/>
                          </svg>
                          <span class="text-primary font-medium">
                            Linked to {consumptionLinks.length === 1 ? 'consumption' : `${consumptionLinks.length} consumptions`}
                          </span>
                          <div class="flex gap-1">
                            {#each consumptionLinks as link}
                              {#if link.consumption_id}
                                <button
                                  class="btn btn-xs btn-outline btn-primary"
                                  title="View linked consumption"
                                  on:click={() => navigateToConsumption(link.consumption_id || '')}
                                >
                                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
                                  </svg>
                                  View
                                </button>
                              {/if}
                            {/each}
                          </div>
                        </div>
                      {/if}
                    </div>
                  </div>

                  <div class="flex gap-2 self-start lg:self-center">
                    <button
                      class="btn btn-ghost btn-sm"
                      aria-label={`Edit event ${event.name}`}
                      title={`Edit event ${event.name}`}
                      on:click={() => openEditModal(event)}
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                        />
                      </svg>
                    </button>
                    <button
                      class="btn btn-ghost btn-sm text-error hover:bg-error hover:text-error-content"
                      aria-label={`Delete event ${event.name}`}
                      title={`Delete event ${event.name}`}
                      on:click={() => openDeleteModal(event)}
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                        />
                      </svg>
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</div>

<!-- Create/Edit Event Modal -->
{#if showModal}
  <div class="modal modal-open">
    <div class="modal-box max-w-2xl">
      <h3 class="font-bold text-lg">{modalTitle}</h3>

      <form on:submit|preventDefault={saveEvent} class="space-y-4 mt-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Event Name -->
          <FormField
            id="eventName"
            label="Event Name"
            bind:value={eventName}
            required
            maxlength={100}
            placeholder="e.g. Headache, Morning run, Blood pressure"
            error={!isValidName && eventName.length > 0
              ? "Name must be 1-100 characters"
              : ""}
          />

          <!-- Event Type -->
          <div class="form-control">
            <label class="label" for="eventTypeId">
              <span class="label-text">Event Type (Optional)</span>
            </label>
            <div class="flex gap-2">
              <FormSelect
                id="eventTypeId"
                label=""
                bind:value={eventTypeId}
                options={createEventTypeOptions}
                className="flex-1"
              />
              <button
                type="button"
                class="btn btn-outline btn-sm"
                title="Create new event type"
                aria-label="Create new event type"
                on:click={openEventTypeModal}
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                  />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Start Date/Time -->
          <div class="form-control">
            <label class="label" for="eventStartedAt">
              <span class="label-text">Start Date & Time <span class="text-error">*</span></span>
            </label>
            <div class="flex gap-2">
              <input
                id="eventStartedAt"
                type="datetime-local"
                class="input input-bordered flex-1"
                class:input-error={!isValidStartDate || isStartDateInFuture}
                bind:value={eventStartedAt}
                required
              />
              <button
                type="button"
                class="btn btn-outline btn-sm"
                title="Set to current time"
                on:click={() => setCurrentTime(false)}
              >
                Now
              </button>
            </div>
            {#if !isValidStartDate}
              <div class="label">
                <span class="label-text-alt text-error">Start date is required</span>
              </div>
            {:else if isStartDateInFuture}
              <div class="label">
                <span class="label-text-alt text-error">Start time cannot be in the future</span>
              </div>
            {/if}
            <div class="label">
              <span class="label-text-alt text-base-content/60">Times shown in your local timezone</span>
            </div>
          </div>

          <!-- End Date/Time -->
          <div class="form-control">
            <label class="label" for="eventEndedAt">
              <span class="label-text">End Date & Time (Optional)</span>
            </label>
            <div class="flex gap-2">
              <input
                id="eventEndedAt"
                type="datetime-local"
                class="input input-bordered flex-1"
                class:input-error={!isValidEndDate || isEndDateInFuture}
                bind:value={eventEndedAt}
              />
              <button
                type="button"
                class="btn btn-outline btn-sm"
                title="Set to current time"
                on:click={() => setCurrentTime(true)}
              >
                Now
              </button>
            </div>
            {#if !isValidEndDate}
              <div class="label">
                <span class="label-text-alt text-error">End date must be after start date</span>
              </div>
            {:else if isEndDateInFuture}
              <div class="label">
                <span class="label-text-alt text-error">End time cannot be in the future</span>
              </div>
            {/if}
          </div>
        </div>

        <!-- Level -->
        <FormField
          id="eventLevel"
          label="Level (0-10, Optional)"
          type="number"
          bind:value={eventLevel}
          min="0"
          max="10"
          placeholder="Intensity, severity, or performance level"
          error={!isValidLevel ? "Level must be between 0 and 10" : ""}
        />

        <!-- Note -->
        <FormField
          id="eventNote"
          label="Note (Optional)"
          bind:value={eventNote}
          maxlength={1000}
          placeholder="Additional details about this event"
          multiline
          error={!isValidNote ? "Note must be 1000 characters or less" : ""}
        />

        <!-- Color Selection -->
        <div class="form-control">
          <label class="label" for="eventColor">
            <span class="label-text">Color (Optional)</span>
          </label>

          <ColorPicker 
            bind:selectedColor={eventColor} 
            previewName={eventName || "Event"} 
          />

          {#if !isValidColor && eventColor.length > 0}
            <div class="label">
              <span class="label-text-alt text-error"
                >Please enter a valid hex color (e.g. #FF0000)</span
              >
            </div>
          {/if}
        </div>

        <!-- Quick Link to Last Consumption (only for new events) -->
        {#if !editingEvent}
          <div class="form-control">
            <div class="mb-2 flex items-center gap-2">
              <span class="label-text font-medium">Link to Consumption</span>
              <span class="label-text-alt text-xs opacity-60">Optional</span>
              <div class="tooltip tooltip-top" data-tip="Linking related events and consumptions helps Noot's AI nutritionist better understand patterns and connections in your health data, leading to more accurate insights and personalized recommendations.">
                <svg class="w-4 h-4 text-base-content/50 cursor-help" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
              </div>
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
        {/if}
      </form>

      <div class="modal-action">
        <button class="btn btn-ghost" on:click={closeModal}>Cancel</button>
        <button class="btn btn-primary" disabled={!canSave} on:click={saveEvent}>
          {editingEvent ? "Update" : "Create"}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Modal -->
<ConfirmModal
  show={showDeleteModal}
  title="Delete Event"
  message="Are you sure you want to delete the event '{eventToDelete?.name}'? This action cannot be undone."
  confirmText="Delete"
  confirmVariant="error"
  onConfirm={deleteEvent}
  onCancel={closeDeleteModal}
/>

<!-- Create Event Type Modal -->
{#if showEventTypeModal}
  <div class="modal modal-open">
    <div class="modal-box max-w-lg">
      <h3 class="font-bold text-lg">Create Event Type</h3>

      <form on:submit|preventDefault={saveEventType} class="space-y-6 mt-4">
        <!-- 1. Event Type Name -->
        <div>
          <FormField
            id="eventTypeName"
            label="Event Type Name"
            bind:value={newEventTypeName}
            required
            maxlength={39}
            placeholder="e.g. symptom, activity, measurement"
            error={!isValidEventTypeName && newEventTypeName.length > 0
              ? "Name must be 1-39 characters and contain only letters, numbers, and hyphens"
              : ""}
          />
          <p class="text-sm text-base-content/60 mt-1">
            The category name for this type of event (e.g., "Symptoms", "Medications", "Activities").
          </p>
        </div>

        <!-- 2. Default Event Name -->
        <div>
          <FormField
            id="eventTypeDefaultName"
            label="Default Event Name"
            bind:value={newEventTypeDefaultName}
            maxlength={100}
            placeholder="e.g. Headache, Morning walk, Blood pressure"
            error={!isValidEventTypeDefaultName ? "Default name must be 100 characters or less" : ""}
          />
          <p class="text-sm text-base-content/60 mt-1">
            The default event name when using quick-add. Can be overridden when creating individual events.
          </p>
        </div>

        <!-- 3. Default Description -->
        <div>
          <FormField
            id="eventTypeDescription"
            label="Default Description"
            bind:value={newEventTypeDescription}
            maxlength={250}
            placeholder="Default description for events of this type"
            error={!isValidEventTypeDescription ? "Description must be 250 characters or less" : ""}
          />
          <p class="text-sm text-base-content/60 mt-1">
            The default description when using quick-add. Can be overridden when creating individual events.
          </p>
        </div>

        <!-- 4. Color -->
        <div>
          <label class="label" for="eventTypeColor">
            <span class="label-text">Color <span class="text-error">*</span></span>
          </label>
          <p class="text-sm text-base-content/60 mb-3">
            The color used to identify this event type throughout the app.
          </p>
          
          <ColorPicker 
            bind:selectedColor={newEventTypeColor} 
            previewName={newEventTypeName || "Event Type"} 
          />

          {#if !isValidEventTypeColor && newEventTypeColor.length > 0}
            <div class="label">
              <span class="label-text-alt text-error"
                >Please enter a valid hex color (e.g. #FF0000)</span
              >
            </div>
          {/if}
        </div>
      </form>

      <div class="modal-action">
        <button class="btn btn-ghost" on:click={closeEventTypeModal}>Cancel</button>
        <button class="btn btn-primary" disabled={!canSaveEventType} on:click={saveEventType}>
          Create Event Type
        </button>
      </div>
    </div>
  </div>
{/if}

<Toast />
