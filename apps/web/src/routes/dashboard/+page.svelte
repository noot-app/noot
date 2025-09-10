<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { onMount } from "svelte"
  import { getAppName } from "$lib/utils/app-info"

  // Get app name from runtime environment
  $: appName = getAppName()

  let isLoading = false
  let error = ""
  let dateRange = "7d" // Default to 7 days
  let consumptionsData: any = null
  let goalsData: any = null
  let eventsData: any = null

  // Date range options
  const dateRangeOptions = [
    { value: "1d", label: "Today" },
    { value: "7d", label: "7 Days" },
    { value: "30d", label: "30 Days" },
  ]

  onMount(() => {
    loadDashboardData()
  })

  async function loadDashboardData() {
    isLoading = true
    error = ""

    try {
      // Get date range
      const days = parseInt(dateRange.replace('d', ''))
      
      // Load data in parallel
      const [consumptions, goals, events] = await Promise.all([
        apiClient.GET("/consumptions", {
          params: {
            query: { days, limit: 100 }
          }
        }),
        apiClient.GET("/goals"),
        apiClient.GET("/events", {
          params: {
            query: { days, limit: 100 }
          }
        })
      ])

      if (consumptions.error) {
        throw new Error(consumptions.error.error || "Failed to load consumptions")
      }
      if (goals.error) {
        throw new Error(goals.error.error || "Failed to load goals")
      }
      if (events.error) {
        throw new Error(events.error.error || "Failed to load events")
      }

      consumptionsData = consumptions.data
      goalsData = goals.data
      eventsData = events.data

    } catch (err) {
      console.error("Error loading dashboard data:", err)
      error = err instanceof Error ? err.message : "An error occurred"
    } finally {
      isLoading = false
    }
  }

  // Handle date range change
  function handleDateRangeChange() {
    loadDashboardData()
  }
</script>

<svelte:head>
  <title>Dashboard - {appName}</title>
  <meta
    name="description"
    content="Nutrition insights dashboard showing trends, goals, and correlations"
  />
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-8 max-w-7xl">
    <!-- Header -->
    <div class="flex flex-col lg:flex-row lg:items-center lg:justify-between mb-8">
      <div>
        <h1 class="text-3xl font-bold text-base-content">
          Nutrition Dashboard
        </h1>
        <p class="text-base-content-lighter mt-2">
          Track your nutrition trends, goals, and insights
        </p>
      </div>
      
      <!-- Date Range Selector -->
      <div class="mt-4 lg:mt-0">
        <div class="flex items-center gap-2">
          <label class="text-sm font-medium text-base-content" for="date-range">
            Time Period:
          </label>
          <select
            id="date-range"
            bind:value={dateRange}
            on:change={handleDateRangeChange}
            class="select select-bordered select-sm"
          >
            {#each dateRangeOptions as option}
              <option value={option.value}>{option.label}</option>
            {/each}
          </select>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    {#if isLoading}
      <div class="flex justify-center items-center py-16">
        <span class="loading loading-spinner loading-lg"></span>
      </div>
    {:else if error}
      <!-- Error State -->
      <div class="alert alert-error">
        <span>{error}</span>
      </div>
    {:else if consumptionsData}
      <!-- Dashboard Content -->
      <div class="space-y-8">
        
        <!-- KPI Cards Row -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <!-- Total Calories -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-6">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-sm font-medium text-base-content-lighter">
                    Total Calories
                  </h3>
                  <p class="text-2xl font-bold text-base-content mt-1">
                    {consumptionsData.consumptions
                      .reduce((total: number, c: any) => total + (c.summary?.total_calories || 0), 0)
                      .toLocaleString()}
                  </p>
                </div>
                <div class="text-primary">
                  <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <!-- Average Daily Protein -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-6">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-sm font-medium text-base-content-lighter">
                    Avg Daily Protein
                  </h3>
                  <p class="text-2xl font-bold text-base-content mt-1">
                    {Math.round(
                      consumptionsData.consumptions
                        .reduce((total: number, c: any) => total + (c.summary?.total_protein_g || 0), 0) /
                      Math.max(parseInt(dateRange.replace('d', '')), 1)
                    )}g
                  </p>
                </div>
                <div class="text-accent">
                  <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M3 3a1 1 0 000 2v8a2 2 0 002 2h2.586l-1.293 1.293a1 1 0 101.414 1.414L10 15.414l2.293 2.293a1 1 0 001.414-1.414L12.414 15H15a2 2 0 002-2V5a1 1 0 100-2H3zm11.707 4.707a1 1 0 00-1.414-1.414L10 9.586 8.707 8.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <!-- Total Meals -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-6">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-sm font-medium text-base-content-lighter">
                    Total Meals
                  </h3>
                  <p class="text-2xl font-bold text-base-content mt-1">
                    {consumptionsData.consumptions.length}
                  </p>
                </div>
                <div class="text-secondary">
                  <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <!-- Events Tracked -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-6">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-sm font-medium text-base-content-lighter">
                    Events Tracked
                  </h3>
                  <p class="text-2xl font-bold text-base-content mt-1">
                    {eventsData?.events?.length || 0}
                  </p>
                </div>
                <div class="text-info">
                  <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clip-rule="evenodd" />
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Charts Section Placeholder -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Main Nutrition Chart -->
          <div class="card bg-base-200 shadow-sm lg:col-span-2">
            <div class="card-body p-6">
              <h3 class="text-lg font-semibold text-base-content mb-4">
                Nutrition Trends Over Time
              </h3>
              <div class="h-64 flex items-center justify-center text-base-content-lighter">
                <div class="text-center">
                  <svg class="w-16 h-16 mx-auto mb-4 opacity-50" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                  </svg>
                  <p>Charts will be implemented in next phase</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Macro Composition -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-6">
              <h3 class="text-lg font-semibold text-base-content mb-4">
                Macro Composition
              </h3>
              <div class="h-48 flex items-center justify-center text-base-content-lighter">
                <div class="text-center">
                  <svg class="w-12 h-12 mx-auto mb-2 opacity-50" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM4.332 8.027a6.012 6.012 0 011.912-2.706C6.512 5.73 6.974 6 7.5 6A1.5 1.5 0 019 7.5V8a2 2 0 004 0 2 2 0 011.523-1.943A5.977 5.977 0 0116 10c0 .34-.028.675-.083 1H15a2 2 0 00-2 2v2.197A5.973 5.973 0 0110 16v-2a2 2 0 00-2-2 2 2 0 01-2-2 2 2 0 00-1.668-1.973z" clip-rule="evenodd" />
                  </svg>
                  <p class="text-sm">Macro breakdown chart</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Missing Nutrients -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-6">
              <h3 class="text-lg font-semibold text-base-content mb-4">
                Goal Progress
              </h3>
              <div class="h-48 flex items-center justify-center text-base-content-lighter">
                <div class="text-center">
                  <svg class="w-12 h-12 mx-auto mb-2 opacity-50" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M12.395 2.553a1 1 0 00-1.45-.385c-.345.23-.614.558-.822.88-.214.33-.403.713-.57 1.116-.334.804-.614 1.768-.84 2.734a31.365 31.365 0 00-.613 3.58 2.64 2.64 0 01-.945-1.067c-.328-.68-.398-1.534-.398-2.654A1 1 0 005.05 6.05 6.981 6.981 0 003 11a7 7 0 1011.95-4.95c-.592-.591-.98-.985-1.348-1.467-.363-.476-.724-1.063-1.207-2.03zM12.12 15.12A3 3 0 017 13s.879.5 2.5.5c0-1 .5-4 1.25-4.5.5 1 .786 1.293 1.371 1.879A2.99 2.99 0 0113 13a2.99 2.99 0 01-.879 2.121z" clip-rule="evenodd" />
                  </svg>
                  <p class="text-sm">Nutrition goals vs actual</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Recent Activity -->
        <div class="card bg-base-200 shadow-sm">
          <div class="card-body p-6">
            <h3 class="text-lg font-semibold text-base-content mb-4">
              Recent Activity
            </h3>
            {#if consumptionsData.consumptions.length > 0}
              <div class="space-y-3">
                {#each consumptionsData.consumptions.slice(0, 5) as consumption}
                  <div class="flex items-center justify-between py-2 border-b border-base-300 last:border-b-0">
                    <div class="flex-1">
                      <p class="font-medium text-base-content">
                        {consumption.title || 'Untitled meal'}
                      </p>
                      <p class="text-sm text-base-content-lighter">
                        {new Date(consumption.created_at).toLocaleDateString()} • 
                        {consumption.summary?.total_calories || 0} cal
                      </p>
                    </div>
                    <div class="text-right">
                      <div class="badge badge-outline">
                        {consumption.items?.length || 0} items
                      </div>
                    </div>
                  </div>
                {/each}
              </div>
            {:else}
              <p class="text-base-content-lighter text-center py-8">
                No recent activity found
              </p>
            {/if}
          </div>
        </div>
      </div>
    {:else}
      <div class="text-center py-16">
        <p class="text-base-content-lighter">No data available</p>
      </div>
    {/if}
  </div>
</div>