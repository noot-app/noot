<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { onMount } from "svelte"
  import { getAppName } from "$lib/utils/app-info"
  import NutritionTrendsChart from "$lib/components/dashboard/NutritionTrendsChart.svelte"
  import MacroCompositionChart from "$lib/components/dashboard/MacroCompositionChart.svelte"
  import GoalProgressChart from "$lib/components/dashboard/GoalProgressChart.svelte"

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

        <!-- Charts Section -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Main Nutrition Chart -->
          <div class="card bg-base-200 shadow-sm lg:col-span-2">
            <div class="card-body p-6">
              <NutritionTrendsChart consumptions={consumptionsData.consumptions} />
            </div>
          </div>

          <!-- Macro Composition -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-6">
              <MacroCompositionChart consumptions={consumptionsData.consumptions} />
            </div>
          </div>

          <!-- Goal Progress -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-6">
              <GoalProgressChart consumptions={consumptionsData.consumptions} goals={goalsData} />
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