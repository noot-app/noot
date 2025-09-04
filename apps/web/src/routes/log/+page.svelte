<script lang="ts">
  import { onMount } from "svelte"
  import { apiClient } from "$lib/api/client"
  import { toast } from "$lib/stores/toast"
  import { formatErrorForUser, handleApiCallWithAuthRedirect } from "$lib/utils/error-handling"
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
      // Fetch both consumptions and events in parallel with auth redirect handling
      const [consumptionsResult, eventsResult] = await Promise.all([
        handleApiCallWithAuthRedirect(async () => {
          return await apiClient.GET("/consumptions", {
            params: { 
              query: { 
                limit: pageSize, 
                offset: offset 
              } 
            }
          })
        }),
        handleApiCallWithAuthRedirect(async () => {
          return await apiClient.GET("/events", {
            params: { 
              query: { 
                limit: pageSize, 
                offset: offset 
              } 
            }
          })
        })
      ])

      // Handle errors
      if (consumptionsResult.error) {
        error = consumptionsResult.error
        return
      }
      if (eventsResult.error) {
        error = eventsResult.error
        return
      }

      // Transform data into timeline entries
      const consumptionEntries: TimelineEntry[] = (consumptionsResult.data?.consumptions || []).map((consumption) => ({
        id: consumption.id,
        type: "consumption" as const,
        created_at: consumption.created_at,
        data: consumption
      }))

      const eventEntries: TimelineEntry[] = (eventsResult.data?.events || []).map((event) => ({
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
      const events = eventsResult.data?.events || []
      if (events.length > 0) {
        await loadEventLinks(events)
      }

      // Check if there's more data to load
      const totalFetched = (consumptionsResult.data?.consumptions?.length || 0) + (eventsResult.data?.events?.length || 0)
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

  // Format time only
  function formatTime(dateString: string): string {
    const date = new Date(dateString)
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  // Format relative date
  function formatRelativeDate(dateString: string): string {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffDays === 0) return 'Today'
    if (diffDays === 1) return 'Yesterday'
    if (diffDays < 7) return `${diffDays} days ago`
    if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`
    return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
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

  // Get nutrition stats array for better display
  function getNutritionStats(consumption: Consumption) {
    const summary = consumption.summary?.totals
    if (!summary) return []
    
    const stats = []
    if (summary.calories) stats.push({ label: 'Calories', value: Math.round(summary.calories), unit: '' })
    if (summary.protein_g) stats.push({ label: 'Protein', value: Math.round(summary.protein_g), unit: 'g' })
    if (summary.total_carbs_g) stats.push({ label: 'Carbs', value: Math.round(summary.total_carbs_g), unit: 'g' })
    if (summary.total_fat_g) stats.push({ label: 'Fat', value: Math.round(summary.total_fat_g), unit: 'g' })
    
    return stats
  }

  // Get meal type from transcript
  function getMealType(consumption: Consumption): string {
    const transcript = consumption.transcript?.toLowerCase() || ''
    if (transcript.includes('breakfast') || transcript.includes('morning')) return 'Breakfast'
    if (transcript.includes('lunch')) return 'Lunch'
    if (transcript.includes('dinner')) return 'Dinner'
    if (transcript.includes('snack')) return 'Snack'
    if (transcript.includes('coffee') || transcript.includes('tea') || transcript.includes('drink')) return 'Beverage'
    return 'Meal'
  }

  // Format event type for display
  function formatEventType(eventTypeId: string | null | undefined): string {
    if (!eventTypeId) return 'Event'
    return eventTypeId.charAt(0).toUpperCase() + eventTypeId.slice(1)
  }

  // Get level color and description
  function getLevelInfo(level: number | null | undefined) {
    if (level === null || level === undefined) return null
    
    if (level <= 2) return { color: 'text-green-600 bg-green-100', desc: 'Mild' }
    if (level <= 4) return { color: 'text-yellow-600 bg-yellow-100', desc: 'Moderate' }
    if (level <= 6) return { color: 'text-orange-600 bg-orange-100', desc: 'Significant' }
    if (level <= 8) return { color: 'text-red-600 bg-red-100', desc: 'High' }
    return { color: 'text-purple-600 bg-purple-100', desc: 'Intense' }
  }

  onMount(() => {
    loadTimelineData()
  })
</script>

<svelte:head>
  <title>Timeline Log - Noot</title>
  <meta name="description" content="View your complete timeline of meals and events" />
</svelte:head>

<!-- Clean, consistent background matching other pages -->
<div class="min-h-screen bg-base-100">
  <div class="max-w-3xl mx-auto px-6 py-12">
    
    <!-- Simple, clean header -->
    <div class="mb-16">
      <h1 class="text-2xl font-medium text-base-content mb-2">Timeline</h1>
      <p class="text-base-content/70">Your complete log of meals and wellness events</p>
    </div>

    <!-- Error State -->
    {#if error}
      <div class="mb-8 p-4 bg-error/10 border border-error/20 rounded-lg">
        <div class="flex items-center justify-between">
          <p class="text-error">{error}</p>
          <button class="text-error hover:text-error/80 font-medium" on:click={() => loadTimelineData()}>
            Retry
          </button>
        </div>
      </div>
    {/if}

    <!-- Timeline Content -->
    {#if timelineEntries.length === 0 && !isLoading}
      <!-- Clean empty state -->
      <div class="text-center py-24">
        <h3 class="text-lg font-medium text-base-content mb-2">No entries yet</h3>
        <p class="text-base-content/70 mb-8 max-w-sm mx-auto">
          Start logging meals and tracking events to see your timeline here
        </p>
        <div class="flex gap-3 justify-center">
          <a href="/record" class="px-4 py-2 bg-primary text-primary-content rounded-lg hover:bg-primary/90 transition-colors">
            Record a Meal
          </a>
          <a href="/events" class="px-4 py-2 border border-base-300 text-base-content rounded-lg hover:bg-base-200 transition-colors">
            Create Event
          </a>
        </div>
      </div>
    {:else}
      <!-- Clean Timeline -->
      <div class="space-y-8">
        {#each timelineEntries as entry, index (entry.id)}
          {@const isConsumption = entry.type === "consumption"}
          
          <div class="flex gap-4">
            <!-- Simple timestamp -->
            <div class="flex-shrink-0 w-20 pt-1">
              <div class="text-xs text-base-content/60 font-mono">
                {formatTime(entry.created_at)}
              </div>
              <div class="text-xs text-base-content/40">
                {formatRelativeDate(entry.created_at)}
              </div>
            </div>

            <!-- Content -->
            <div class="flex-1 min-w-0">
              {#if isConsumption}
                {@const consumption = entry.data as Consumption}
                {@const nutritionStats = getNutritionStats(consumption)}
                
                <div class="bg-base-100 border border-base-300 rounded-lg p-6 hover:border-base-content/20 transition-colors">
                  <!-- Header -->
                  <div class="flex items-center justify-between mb-4">
                    <div class="flex items-center gap-3">
                      <div class="w-2 h-2 bg-primary rounded-full"></div>
                      <div>
                        <h3 class="font-medium text-base-content">
                          {getMealType(consumption)}
                        </h3>
                        {#if isConsumptionLinked(consumption.id)}
                          <div class="flex items-center gap-1 mt-1">
                            <svg class="w-3 h-3 text-base-content/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.102m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/>
                            </svg>
                            <span class="text-xs text-base-content/60">Linked to event</span>
                          </div>
                        {/if}
                      </div>
                    </div>
                  </div>
                  
                  <!-- Meal description -->
                  <div class="mb-4">
                    <p class="text-base-content/80 leading-relaxed">
                      {consumption.transcript}
                    </p>
                  </div>
                  
                  <!-- Nutrition stats -->
                  {#if nutritionStats.length > 0}
                    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
                      {#each nutritionStats as stat}
                        <div class="text-center">
                          <div class="text-sm font-medium text-base-content">{stat.value}{stat.unit}</div>
                          <div class="text-xs text-base-content/60">{stat.label}</div>
                        </div>
                      {/each}
                    </div>
                  {/if}
                  
                  <!-- Labels -->
                  {#if consumption.labels && consumption.labels.length > 0}
                    <div class="flex flex-wrap gap-2">
                      {#each consumption.labels as label}
                        <span class="px-2 py-1 text-xs bg-base-200 text-base-content/80 rounded-md">
                          {label.name}
                        </span>
                      {/each}
                    </div>
                  {/if}
                </div>

              {:else}
                {@const event = entry.data as Event}
                {@const levelInfo = getLevelInfo(event.level)}
                
                <div class="bg-base-100 border border-base-300 rounded-lg p-6 hover:border-base-content/20 transition-colors">
                  <!-- Header -->
                  <div class="flex items-center justify-between mb-4">
                    <div class="flex items-center gap-3">
                      <div class="w-2 h-2 bg-secondary rounded-full"></div>
                      <div>
                        <h3 class="font-medium text-base-content">
                          {event.name}
                        </h3>
                        <div class="flex items-center gap-2 mt-1">
                          <span class="text-xs text-base-content/60">
                            {formatEventType(event.event_type_id)}
                          </span>
                          {#if levelInfo}
                            <span class="text-xs text-base-content/60">
                              • Level {event.level}
                            </span>
                          {/if}
                          {#if hasConsumptionLinks(event.id)}
                            <div class="flex items-center gap-1">
                              <svg class="w-3 h-3 text-base-content/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.102m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/>
                              </svg>
                              <span class="text-xs text-base-content/60">Linked to meal</span>
                            </div>
                          {/if}
                        </div>
                      </div>
                    </div>
                  </div>
                  
                  <!-- Event details -->
                  {#if event.note}
                    <div class="mb-4">
                      <p class="text-base-content/80 leading-relaxed">
                        {event.note}
                      </p>
                    </div>
                  {/if}
                  
                  <!-- Event duration -->
                  {#if event.ended_at}
                    <div class="text-xs text-base-content/60">
                      Duration: {formatTime(event.started_at)} → {formatTime(event.ended_at)}
                    </div>
                  {/if}
                </div>
              {/if}
            </div>
          </div>
        {/each}
      </div>

      <!-- Load More Button -->
      {#if hasMore}
        <div class="text-center mt-12">
          <button 
            class="px-6 py-2 border border-base-300 text-base-content rounded-lg hover:bg-base-200 transition-colors {isLoading ? 'opacity-50' : ''}"
            disabled={isLoading}
            on:click={loadMore}
          >
            {isLoading ? "Loading..." : "Load more entries"}
          </button>
        </div>
      {:else if timelineEntries.length > 0}
        <div class="text-center mt-16 py-8">
          <p class="text-base-content/60 text-sm">End of timeline</p>
        </div>
      {/if}
    {/if}

    <!-- Loading State -->
    {#if isLoading && timelineEntries.length === 0}
      <div class="flex flex-col items-center justify-center py-24">
        <div class="animate-spin w-6 h-6 border-2 border-base-300 border-t-base-content rounded-full mb-4"></div>
        <p class="text-base-content/70">Loading your timeline...</p>
      </div>
    {/if}
  </div>
</div>
