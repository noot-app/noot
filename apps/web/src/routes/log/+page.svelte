<script lang="ts">
  import { onMount } from "svelte"
  import { apiClient } from "$lib/api/client"
  import { toast } from "$lib/stores/toast"
  import { formatErrorForUser } from "$lib/utils/error-handling"
  import TimelineIcon from "$lib/components/icons/Timeline.svelte"
  import type { paths } from "$lib/api/schema"

  // Type definitions
  type ConsumptionsResponse = paths["/consumptions"]["get"]["responses"]["200"]["content"]["application/json"]
  type EventsResponse = paths["/events"]["get"]["responses"]["200"]["content"]["application/json"]
  type Consumption = ConsumptionsResponse["consumptions"][0]
  type Event = EventsResponse["events"][0]
  type EventLinksResponse = paths["/events/{id}/links"]["get"]["responses"]["200"]["content"]["application/json"]
  type EventLink = NonNullable<EventLinksResponse["links"]>[0]

  // Timeline entry type combining consumptions and events
  interface TimelineEntry {
    id: string
    type: "consumption" | "event"
    created_at: string
    data: Consumption | Event
  }

  // State
  let timelineEntries: TimelineEntry[] = []
  let eventLinks: Map<string, EventLink[]> = new Map()
  let isLoading = false
  let error = ""
  let hasMore = true
  let currentPage = 0
  const pageSize = 20

  // Load event links for events
  async function loadEventLinks(events: Event[]) {
    try {
      const eventIds = events.map(event => event.id)
      
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
      
      // Update the existing eventLinks map
      results.forEach(({ eventId, links }) => {
        eventLinks.set(eventId, links || [])
      })
      eventLinks = eventLinks // Trigger reactivity
    } catch (err) {
      console.error("Error loading event links:", err)
    }
  }

  // Helper functions for event links
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

  // Check if a consumption is linked to any events
  function isConsumptionLinked(consumptionId: string): boolean {
    for (const [eventId, links] of eventLinks) {
      if (links.some(link => link.consumption_id === consumptionId)) {
        return true
      }
    }
    return false
  }

  // Load timeline data
  async function loadTimelineData(offset = 0, append = false) {
    if (isLoading) return
    
    isLoading = true
    error = ""

    try {
      // Fetch both consumptions and events in parallel
      const [consumptionsResponse, eventsResponse] = await Promise.all([
        apiClient.GET("/consumptions", {
          params: { 
            query: { 
              limit: pageSize, 
              offset: offset 
            } 
          }
        }),
        apiClient.GET("/events", {
          params: { 
            query: { 
              limit: pageSize, 
              offset: offset 
            } 
          }
        })
      ])

      // Handle errors
      if (consumptionsResponse.error) {
        error = formatErrorForUser(consumptionsResponse.error)
        return
      }
      if (eventsResponse.error) {
        error = formatErrorForUser(eventsResponse.error)
        return
      }

      // Transform data into timeline entries
      const consumptionEntries: TimelineEntry[] = (consumptionsResponse.data?.consumptions || []).map((consumption) => ({
        id: consumption.id,
        type: "consumption" as const,
        created_at: consumption.created_at,
        data: consumption
      }))

      const eventEntries: TimelineEntry[] = (eventsResponse.data?.events || []).map((event) => ({
        id: event.id,
        type: "event" as const,
        created_at: event.started_at, // Use started_at as the timeline timestamp
        data: event
      }))

      // Combine and sort by created_at (newest first)
      const newEntries = [...consumptionEntries, ...eventEntries].sort((a, b) => 
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      )

      if (append) {
        timelineEntries = [...timelineEntries, ...newEntries]
      } else {
        timelineEntries = newEntries
      }

      // Load event links for the new events
      const events = eventsResponse.data?.events || []
      if (events.length > 0) {
        await loadEventLinks(events)
      }

      // Check if there's more data to load
      const totalFetched = (consumptionsResponse.data?.consumptions?.length || 0) + (eventsResponse.data?.events?.length || 0)
      hasMore = totalFetched >= pageSize

    } catch (err) {
      console.error("Error loading timeline data:", err)
      error = "Failed to load timeline data"
    } finally {
      isLoading = false
    }
  }

  // Load more entries for pagination
  async function loadMore() {
    if (!hasMore || isLoading) return
    currentPage++
    await loadTimelineData(currentPage * pageSize, true)
  }

  // Format date for display
  function formatDate(dateString: string): string {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffDays === 0) {
      return `Today at ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
    } else if (diffDays === 1) {
      return `Yesterday at ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
    } else if (diffDays < 7) {
      return `${diffDays} days ago at ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
    } else {
      return date.toLocaleDateString()
    }
  }

  // Get nutrition highlights for consumption
  function getNutritionHighlights(consumption: Consumption) {
    const summary = consumption.summary?.totals
    if (!summary) return ""
    
    const highlights = []
    if (summary.calories) highlights.push(`${Math.round(summary.calories)} cal`)
    if (summary.protein_g) highlights.push(`${Math.round(summary.protein_g)}g protein`)
    if (summary.total_carbs_g) highlights.push(`${Math.round(summary.total_carbs_g)}g carbs`)
    if (summary.total_fat_g) highlights.push(`${Math.round(summary.total_fat_g)}g fat`)
    
    return highlights.join(" • ")
  }

  onMount(() => {
    loadTimelineData()
  })
</script>

<svelte:head>
  <title>Timeline Log - Noot</title>
  <meta name="description" content="View your complete timeline of meals and events" />
</svelte:head>

<div class="container mx-auto px-4 py-8 max-w-4xl">
  <!-- Header -->
  <div class="flex items-center gap-3 mb-8">
    <TimelineIcon class="w-8 h-8 text-primary" />
    <div>
      <h1 class="text-3xl font-bold">Timeline Log</h1>
      <p class="text-base-content/70">Your complete timeline of meals and events</p>
    </div>
  </div>

  <!-- Error State -->
  {#if error}
    <div class="alert alert-error mb-6">
      <svg class="stroke-current shrink-0 w-6 h-6" fill="none" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span>{error}</span>
      <button class="btn btn-sm" on:click={() => loadTimelineData()}>Retry</button>
    </div>
  {/if}

  <!-- Timeline -->
  {#if timelineEntries.length === 0 && !isLoading}
    <!-- Empty State -->
    <div class="text-center py-12">
      <div class="text-6xl mb-4">📚</div>
      <h3 class="text-2xl font-bold mb-2">No Timeline Entries</h3>
      <p class="text-base-content/70 mb-6">
        Start logging meals and creating events to see your timeline
      </p>
      <div class="flex gap-4 justify-center">
        <a href="/record" class="btn btn-primary">Record a Meal</a>
        <a href="/events" class="btn btn-outline">Create Event</a>
      </div>
    </div>
  {:else}
    <!-- Timeline Entries -->
    <div class="timeline timeline-snap-icon max-md:timeline-compact timeline-vertical">
      {#each timelineEntries as entry, index (entry.id)}
        <li>
          <div class="timeline-middle">
            {#if entry.type === "consumption"}
              <div class="w-3 h-3 bg-success rounded-full"></div>
            {:else}
              <div class="w-3 h-3 bg-info rounded-full"></div>
            {/if}
          </div>
          <div class="timeline-start md:text-end mb-10">
            <time class="font-mono italic text-sm text-base-content/60">
              {formatDate(entry.created_at)}
            </time>
          </div>
          <div class="timeline-end timeline-box">
            {#if entry.type === "consumption"}
              {@const consumption = entry.data}
              <div class="space-y-3">
                <div class="flex items-start gap-2">
                  <div class="w-2 h-2 bg-success rounded-full mt-2 shrink-0"></div>
                  <div class="flex-1">
                    <div class="flex items-center gap-2 mb-1">
                      <h3 class="font-semibold text-success">Meal Logged</h3>
                      <!-- Show lock icon for linked consumptions -->
                      {#if isConsumptionLinked(consumption.id)}
                        <svg class="w-3 h-3 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24" title="Linked to event">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
                        </svg>
                      {/if}
                    </div>
                    <p class="text-sm italic mb-2">"{consumption.transcript}"</p>
                    
                    <!-- Nutrition highlights -->
                    {#if getNutritionHighlights(consumption)}
                      <div class="text-sm text-base-content/80 mb-2">
                        {getNutritionHighlights(consumption)}
                      </div>
                    {/if}
                    
                    <!-- Labels -->
                    {#if consumption.labels && consumption.labels.length > 0}
                      <div class="flex gap-1 flex-wrap">
                        {#each consumption.labels as label}
                          <span 
                            class="badge badge-xs"
                            style="background-color: {label.color}20; color: {label.color}; border: 1px solid {label.color};"
                          >
                            {label.name}
                          </span>
                        {/each}
                      </div>
                    {/if}
                  </div>
                </div>
              </div>
            {:else}
              {@const event = entry.data}
              <div class="space-y-3">
                <div class="flex items-start gap-2">
                  <div class="w-2 h-2 bg-info rounded-full mt-2 shrink-0"></div>
                  <div class="flex-1">
                    <div class="flex items-center gap-2 mb-1">
                      <h3 class="font-semibold text-info">{event.name}</h3>
                      {#if event.level !== null && event.level !== undefined}
                        <span class="badge badge-xs badge-outline">Level {event.level}</span>
                      {/if}
                      <!-- Show lock icon for linked events -->
                      {#if hasConsumptionLinks(event.id)}
                        <svg class="w-3 h-3 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24" title="Linked to consumption">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
                        </svg>
                      {/if}
                    </div>
                    
                    <p class="text-sm text-base-content/60 mb-2">{event.event_type?.name || "Event"}</p>
                    
                    {#if event.note}
                      <p class="text-sm mb-2">{event.note}</p>
                    {/if}
                    
                    <!-- Show consumption links -->
                    {#if hasConsumptionLinks(event.id)}
                      {@const consumptionLinks = getConsumptionLinks(event.id)}
                      <div class="flex items-center gap-2 mt-2">
                        <svg class="w-3 h-3 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.102m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/>
                        </svg>
                        <span class="text-primary font-medium text-xs">
                          Linked to {consumptionLinks.length === 1 ? 'consumption' : `${consumptionLinks.length} consumptions`}
                        </span>
                      </div>
                    {/if}
                    
                    <!-- Event timing -->
                    <div class="text-xs text-base-content/50 mt-2">
                      {#if event.ended_at}
                        Duration: {formatDate(event.started_at)} → {formatDate(event.ended_at)}
                      {:else}
                        Started: {formatDate(event.started_at)}
                      {/if}
                    </div>
                  </div>
                </div>
              </div>
            {/if}
          </div>
        </li>
      {/each}
    </div>

    <!-- Load More Button -->
    {#if hasMore}
      <div class="text-center mt-8">
        <button 
          class="btn btn-outline"
          class:loading={isLoading}
          disabled={isLoading}
          on:click={loadMore}
        >
          {isLoading ? "Loading..." : "Load More"}
        </button>
      </div>
    {/if}
  {/if}

  <!-- Loading State -->
  {#if isLoading && timelineEntries.length === 0}
    <div class="flex justify-center py-12">
      <span class="loading loading-spinner loading-lg"></span>
    </div>
  {/if}
</div>