<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { onMount } from "svelte"
  import NutritionStats from "$lib/components/NutritionStats.svelte"
  import Goals from "$lib/components/Goals.svelte"
  import ConsumptionCard from "$lib/components/ConsumptionCard.svelte"
  import ChartBarIcon from "$lib/components/icons/ChartBar.svelte"
  import { user } from "$lib/auth/store"
  import { getAppName } from "$lib/utils/app-info"

  // Get app name from runtime environment
  $: appName = getAppName()

  let isLoading = false
  let error = ""
  let consumptionsData: any = null

  onMount(() => {
    loadSummary()
  })

  // Function to get time-based greeting
  function getTimeBasedGreeting(): string {
    const hour = new Date().getHours()
    
    if (hour >= 5 && hour < 12) {
      return "Good morning"
    } else if (hour >= 12 && hour < 17) {
      return "Good afternoon"
    } else {
      return "Good evening"
    }
  }

  // Function to extract first name from user
  function getFirstName(user: any): string {
    if (!user) return ""
    
    // Try user_metadata.first_name first
    if (user.user_metadata?.first_name) {
      return user.user_metadata.first_name
    }
    
    // Try user_metadata.full_name and extract first part
    if (user.user_metadata?.full_name) {
      return user.user_metadata.full_name.split(' ')[0]
    }
    
    // Fallback to extracting from email
    if (user.email) {
      const emailPrefix = user.email.split('@')[0]
      // Capitalize first letter
      return emailPrefix.charAt(0).toUpperCase() + emailPrefix.slice(1)
    }
    
    return ""
  }

  // Reactive greeting message
  $: greeting = (() => {
    const timeGreeting = getTimeBasedGreeting()
    const firstName = getFirstName($user)
    return firstName ? `${timeGreeting}, ${firstName}` : timeGreeting
  })()

  async function loadSummary() {
    isLoading = true
    error = ""

    try {
      // Always load today's data only
      const today = new Date()
      const startOfDay = new Date(
        today.getFullYear(),
        today.getMonth(),
        today.getDate(),
        0,
        0,
        0,
      )
      const endOfDay = new Date(
        today.getFullYear(),
        today.getMonth(),
        today.getDate(),
        23,
        59,
        59,
        999,
      )

      const startDate = startOfDay.toISOString()
      const endDate = endOfDay.toISOString()

      const response = await apiClient.GET("/consumptions", {
        params: {
          query: {
            start: startDate,
            end: endDate,
          },
        },
      })

      if (response.error) {
        throw new Error(`API Error: ${response.error}`)
      }

      consumptionsData = response.data
    } catch (err) {
      error = `Error loading summary: ${err}`
      console.error("Summary error:", err)
    } finally {
      isLoading = false
    }
  }

  // Calculate nutrition totals from consumptions data
  $: nutritionTotals = consumptionsData?.consumptions
    ? consumptionsData.consumptions.reduce((totals: any, consumption: any) => {
        const t = consumption?.summary?.totals || {}
        return {
          calories: (totals.calories || 0) + (t.calories || 0),
          protein_g: (totals.protein_g || 0) + (t.protein_g || 0),
          total_carbs_g: (totals.total_carbs_g || 0) + (t.total_carbs_g || 0),
          total_fat_g: (totals.total_fat_g || 0) + (t.total_fat_g || 0),
          dietary_fiber_g:
            (totals.dietary_fiber_g || 0) + (t.dietary_fiber_g || 0),
          sodium_mg: (totals.sodium_mg || 0) + (t.sodium_mg || 0),

          // Fat types
          saturated_fat_g:
            (totals.saturated_fat_g || 0) + (t.saturated_fat_g || 0),
          trans_fat_g: (totals.trans_fat_g || 0) + (t.trans_fat_g || 0),
          monounsaturated_fat_g:
            (totals.monounsaturated_fat_g || 0) +
            (t.monounsaturated_fat_g || 0),
          polyunsaturated_fat_g:
            (totals.polyunsaturated_fat_g || 0) +
            (t.polyunsaturated_fat_g || 0),
          omega3_ala_g: (totals.omega3_ala_g || 0) + (t.omega3_ala_g || 0),
          omega3_epa_g: (totals.omega3_epa_g || 0) + (t.omega3_epa_g || 0),
          omega3_dha_g: (totals.omega3_dha_g || 0) + (t.omega3_dha_g || 0),
          omega6_g: (totals.omega6_g || 0) + (t.omega6_g || 0),
          cholesterol_mg:
            (totals.cholesterol_mg || 0) + (t.cholesterol_mg || 0),
          alcohol_g: (totals.alcohol_g || 0) + (t.alcohol_g || 0),

          // Sugar types
          total_sugars_g:
            (totals.total_sugars_g || 0) + (t.total_sugars_g || 0),
          added_sugars_g:
            (totals.added_sugars_g || 0) + (t.added_sugars_g || 0),

          // B-Complex vitamins
          thiamine_mg: (totals.thiamine_mg || 0) + (t.thiamine_mg || 0),
          riboflavin_mg: (totals.riboflavin_mg || 0) + (t.riboflavin_mg || 0),
          niacin_mg: (totals.niacin_mg || 0) + (t.niacin_mg || 0),
          vitamin_b6_mg:
            (totals.vitamin_b6_mg || 0) + (t.vitamin_b6_mg || 0),
          folate_mcg: (totals.folate_mcg || 0) + (t.folate_mcg || 0),
          vitamin_b12_mcg:
            (totals.vitamin_b12_mcg || 0) + (t.vitamin_b12_mcg || 0),
          biotin_mcg: (totals.biotin_mcg || 0) + (t.biotin_mcg || 0),
          pantothenic_acid_mg:
            (totals.pantothenic_acid_mg || 0) +
            (t.pantothenic_acid_mg || 0),

          // Fat-soluble vitamins
          vitamin_a_mcg: (totals.vitamin_a_mcg || 0) + (t.vitamin_a_mcg || 0),
          vitamin_d_mcg: (totals.vitamin_d_mcg || 0) + (t.vitamin_d_mcg || 0),
          vitamin_e_mg: (totals.vitamin_e_mg || 0) + (t.vitamin_e_mg || 0),
          vitamin_k_mcg: (totals.vitamin_k_mcg || 0) + (t.vitamin_k_mcg || 0),

          // Water-soluble vitamins
          vitamin_c_mg: (totals.vitamin_c_mg || 0) + (t.vitamin_c_mg || 0),
          choline_mg: (totals.choline_mg || 0) + (t.choline_mg || 0),

          // Essential minerals
          calcium_mg: (totals.calcium_mg || 0) + (t.calcium_mg || 0),
          iron_mg: (totals.iron_mg || 0) + (t.iron_mg || 0),
          magnesium_mg: (totals.magnesium_mg || 0) + (t.magnesium_mg || 0),
          phosphorus_mg:
            (totals.phosphorus_mg || 0) + (t.phosphorus_mg || 0),
          potassium_mg: (totals.potassium_mg || 0) + (t.potassium_mg || 0),
          zinc_mg: (totals.zinc_mg || 0) + (t.zinc_mg || 0),
          copper_mg: (totals.copper_mg || 0) + (t.copper_mg || 0),
          manganese_mg:
            (totals.manganese_mg || 0) + (t.manganese_mg || 0),
          selenium_mcg:
            (totals.selenium_mcg || 0) + (t.selenium_mcg || 0),
          iodine_mcg: (totals.iodine_mcg || 0) + (t.iodine_mcg || 0),
          molybdenum_mcg:
            (totals.molybdenum_mcg || 0) + (t.molybdenum_mcg || 0),
          chromium_mcg:
            (totals.chromium_mcg || 0) + (t.chromium_mcg || 0),
          fluoride_mg: (totals.fluoride_mg || 0) + (t.fluoride_mg || 0),
          chloride_mg: (totals.chloride_mg || 0) + (t.chloride_mg || 0),

          // Other compounds
          caffeine_mg: (totals.caffeine_mg || 0) + (t.caffeine_mg || 0),
          creatine_mg: (totals.creatine_mg || 0) + (t.creatine_mg || 0),
        }
      }, {})
    : undefined

  // Use calculated totals for goals comparison (same structure as before)
  $: currentNutrition = nutritionTotals
</script>

<svelte:head>
  <title>Summary - {appName}</title>
</svelte:head>

<div class="min-h-full bg-base-100">
  <div class="container mx-auto px-4 py-8 max-w-6xl safe-area-page-top">
    <!-- Header -->
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-base-content flex items-center gap-3">
        <ChartBarIcon className="w-8 h-8" />
        Nutrition Summary
      </h1>
      <p class="text-base-content-lighter mt-2">
        {greeting}
      </p>
    </div>

    <!-- Loading -->
    {#if isLoading}
      <div class="flex justify-center items-center py-12">
        <span class="loading loading-spinner loading-lg"></span>
        <span class="ml-4 text-lg">Loading summary...</span>
      </div>
    {/if}

    <!-- Error -->
    {#if error}
      <div class="alert alert-error mb-6">
        <svg
          class="w-6 h-6"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>{error}</span>
        <button class="btn btn-sm" on:click={loadSummary}>Retry</button>
      </div>
    {/if}

    <!-- Summary Data -->
    {#if consumptionsData && !isLoading}
      <!-- Check if we have any actual data -->
      {#if !consumptionsData.consumptions || consumptionsData.consumptions.length === 0}
        <!-- No Data State -->
        <div class="text-center py-12">
          <div class="text-6xl mb-4">🍽️</div>
          <h3 class="text-2xl font-bold mb-2">
            No data for today
          </h3>
          <p class="text-base-content/70 mb-6">
            Start logging your meals to see your nutrition summary
          </p>
          <a href="/record" class="btn btn-primary"> Record Your First Meal </a>
        </div>
      {:else}
        <!-- We have data - show the nutrition components -->
        <div class="space-y-8">
          <!-- Overview Stats -->
          <NutritionStats
            calories={nutritionTotals?.calories || 0}
            protein={nutritionTotals?.protein_g || 0}
            carbs={nutritionTotals?.total_carbs_g || 0}
            fat={nutritionTotals?.total_fat_g || 0}
            size="normal"
          />

          <!-- Nutrition Goals -->
          <Goals {currentNutrition} showAllCategories={true} />

          <!-- Today's Summary Stats -->
          <div class="card bg-base-200 shadow-xl">
            <div class="card-body">
              <h2 class="card-title mb-4">Today's Summary</h2>
              <div
                class="stats stats-vertical lg:stats-horizontal shadow w-full"
              >
                <div class="stat">
                  <div class="stat-title">Meals Logged</div>
                  <div class="stat-value text-lg">
                      {consumptionsData.consumptions.length}
                    </div>
                    <div class="stat-desc">Today</div>
                  </div>
                  <div class="stat">
                    <div class="stat-title">Avg per Meal</div>
                    <div class="stat-value text-lg">
                      {consumptionsData.consumptions.length > 0 ? Math.round((nutritionTotals?.calories || 0) / consumptionsData.consumptions.length) : 0}
                    </div>
                    <div class="stat-desc">Calories</div>
                  </div>
                  <div class="stat">
                    <div class="stat-title">Fiber</div>
                    <div class="stat-value text-lg">
                      {(nutritionTotals?.dietary_fiber_g || 0).toFixed(1)}g
                    </div>
                    <div class="stat-desc">Total</div>
                  </div>
                  <div class="stat">
                    <div class="stat-title">Sodium</div>
                    <div class="stat-value text-lg">
                      {Math.round(nutritionTotals?.sodium_mg || 0)} mg
                    </div>
                    <div class="stat-desc">Total</div>
                  </div>
                </div>
              </div>
            </div>

          <!-- Individual Meals -->
          <div class="card bg-base-200 shadow-xl">
            <div class="card-body">
              <h2 class="card-title mb-4">
                Today's Meals
              </h2>
              <div class="space-y-4">
                {#each consumptionsData.consumptions as consumption}
                  <ConsumptionCard
                    {consumption}
                    mode="readonly"
                    clickable={true}
                    variant="compact"
                    showNutrition={true}
                    showLabels={true}
                    showTimestamp={true}
                    showNote={true}
                  />
                {/each}
              </div>
            </div>
          </div>
        </div>
      {/if}
    {/if}
  </div>
</div>
