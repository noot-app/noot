<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { onMount } from "svelte"
  import { getAppName } from "$lib/utils/app-info"
  import NutritionTrendsChart from "$lib/components/dashboard/NutritionTrendsChart.svelte"
  import MacroCompositionChart from "$lib/components/dashboard/MacroCompositionChart.svelte"
  import GoalProgressChart from "$lib/components/dashboard/GoalProgressChart.svelte"
  import EventCorrelationChart from "$lib/components/dashboard/EventCorrelationChart.svelte"
  import Zap from "$lib/components/icons/Zap.svelte"
  import Dumbbell from "$lib/components/icons/Dumbbell.svelte"
  import Cube from "$lib/components/icons/Cube.svelte"
  import CalendarIcon from "$lib/components/icons/calendar-days.svelte"

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
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 lg:gap-6">
          <!-- Total Calories -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-4 lg:p-6">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-xs lg:text-sm font-medium text-base-content-lighter">
                    Total Calories
                  </h3>
                  <p class="text-xl lg:text-2xl font-bold text-base-content mt-1">
                    {consumptionsData.consumptions
                      .reduce((total: number, c: any) => total + (c.summary?.totals?.calories || 0), 0)
                      .toLocaleString()}
                  </p>
                </div>
                <div class="text-honey">
                  <Zap className="w-6 h-6 lg:w-8 lg:h-8" />
                </div>
              </div>
            </div>
          </div>

          <!-- Average Daily Protein -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-4 lg:p-6">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-xs lg:text-sm font-medium text-base-content-lighter">
                    Avg Daily Protein
                  </h3>
                  <p class="text-xl lg:text-2xl font-bold text-base-content mt-1">
                    {Math.round(
                      consumptionsData.consumptions
                        .reduce((total: number, c: any) => total + (c.summary?.totals?.protein_g || 0), 0) /
                      Math.max(parseInt(dateRange.replace('d', '')), 1)
                    )}g
                  </p>
                </div>
                <div class="text-secondary">
                  <Dumbbell className="w-6 h-6 lg:w-8 lg:h-8" />
                </div>
              </div>
            </div>
          </div>

          <!-- Total Meals -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-4 lg:p-6">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-xs lg:text-sm font-medium text-base-content-lighter">
                    Total Meals
                  </h3>
                  <p class="text-xl lg:text-2xl font-bold text-base-content mt-1">
                    {consumptionsData.consumptions.length}
                  </p>
                </div>
                <div class="text-secondary">
                  <Cube className="w-6 h-6 lg:w-8 lg:h-8" />
                </div>
              </div>
            </div>
          </div>

          <!-- Events Tracked -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-4 lg:p-6">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-xs lg:text-sm font-medium text-base-content-lighter">
                    Events Tracked
                  </h3>
                  <p class="text-xl lg:text-2xl font-bold text-base-content mt-1">
                    {eventsData?.events?.length || 0}
                  </p>
                </div>
                <div class="text-info">
                  <CalendarIcon className="w-6 h-6 lg:w-8 lg:h-8" />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Charts Section -->
        <div class="space-y-4 lg:space-y-6">
          <!-- Main Nutrition Chart - Full Width -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-4 lg:p-6">
              <NutritionTrendsChart consumptions={consumptionsData.consumptions} />
            </div>
          </div>

          <!-- Secondary Charts - Responsive Grid -->
          <div class="grid grid-cols-1 xl:grid-cols-2 gap-4 lg:gap-6">
            <!-- Macro Composition -->
            <div class="card bg-base-200 shadow-sm">
              <div class="card-body p-4 lg:p-6">
                <MacroCompositionChart consumptions={consumptionsData.consumptions} />
              </div>
            </div>

            <!-- Goal Progress -->
            <div class="card bg-base-200 shadow-sm">
              <div class="card-body p-4 lg:p-6">
                <GoalProgressChart consumptions={consumptionsData.consumptions} goals={goalsData} />
              </div>
            </div>

            <!-- Event Correlation - Full Width on Mobile, Spans 2 cols on Desktop -->
            <div class="card bg-base-200 shadow-sm xl:col-span-2">
              <div class="card-body p-4 lg:p-6">
                <EventCorrelationChart 
                  consumptions={consumptionsData.consumptions} 
                  events={eventsData?.events || []} 
                />
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
                        {consumption.title || consumption.transcript || 'Untitled meal'}
                      </p>
                      <p class="text-sm text-base-content-lighter">
                        {new Date(consumption.created_at).toLocaleDateString()} • 
                        {consumption.summary?.totals?.calories || 0} cal
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
