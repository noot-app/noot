<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { onMount } from "svelte"
  import { getAppName } from "$lib/utils/app-info"
  import DashboardHeader from "$lib/components/dashboard/DashboardHeader.svelte"
  import KpiCards from "$lib/components/dashboard/KpiCards.svelte"
  import ChartsGrid from "$lib/components/dashboard/ChartsGrid.svelte"

  // Get app name from runtime environment
  $: appName = getAppName()

  let isLoading = false
  let error = ""
  let dateRange = "7d" // Default to 7 days
  let consumptionsData: any = null
  let goalsData: any = null
  let eventsData: any = null

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
    <DashboardHeader bind:dateRange onDateRangeChange={handleDateRangeChange} />

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
        <KpiCards {consumptionsData} {eventsData} {dateRange} />

        <!-- Charts Section -->
        <ChartsGrid {consumptionsData} {goalsData} {eventsData} {dateRange} />
      </div>
    {:else}
      <div class="text-center py-16">
        <p class="text-base-content-lighter">No data available</p>
      </div>
    {/if}
  </div>
</div>
