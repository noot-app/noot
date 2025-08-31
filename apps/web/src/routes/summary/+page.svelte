<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { onMount } from "svelte"
  import NutritionStats from "$lib/components/NutritionStats.svelte"
  import Goals from "$lib/components/Goals.svelte"
  import { getAppName } from "$lib/utils/app-info"

  // Get app name from runtime environment
  $: appName = getAppName()

  let currentView: "today" | "week" = "today"
  let isLoading = false
  let error = ""
  let consumptionsData: any = null
  let expandedMeals: Set<string> = new Set()

  onMount(() => {
    loadSummary()
  })

  async function loadSummary() {
    isLoading = true
    error = ""

    try {
      let startDate: string, endDate: string

      if (currentView === "today") {
        // Get today in local timezone
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

        startDate = startOfDay.toISOString()
        endDate = endOfDay.toISOString()
      } else {
        // Get week (7 days back from today) in local timezone
        const today = new Date()
        const weekAgo = new Date(today.getTime() - 6 * 24 * 60 * 60 * 1000) // 6 days ago + today = 7 days

        const startOfWeek = new Date(
          weekAgo.getFullYear(),
          weekAgo.getMonth(),
          weekAgo.getDate(),
          0,
          0,
          0,
        )
        const endOfToday = new Date(
          today.getFullYear(),
          today.getMonth(),
          today.getDate(),
          23,
          59,
          59,
          999,
        )

        startDate = startOfWeek.toISOString()
        endDate = endOfToday.toISOString()
      }

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

  function switchView(view: "today" | "week") {
    if (currentView !== view) {
      currentView = view
      loadSummary()
    }
  }

  function toggleMealExpansion(mealId: string) {
    if (expandedMeals.has(mealId)) {
      expandedMeals.delete(mealId)
    } else {
      expandedMeals.add(mealId)
    }
    expandedMeals = expandedMeals // trigger reactivity
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

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-8 max-w-6xl">
    <!-- Header -->
    <div class="text-center mb-8">
      <h1 class="text-3xl font-bold text-primary mb-4">Nutrition Summary</h1>
      <p class="text-base-content/70">Track your nutrition intake over time</p>
    </div>

    <!-- View Toggle -->
    <div class="flex justify-center mb-8">
      <div class="btn-group">
        <button
          class="btn {currentView === 'today' ? 'btn-primary' : 'btn-ghost'}"
          on:click={() => switchView("today")}
        >
          Today
        </button>
        <button
          class="btn {currentView === 'week' ? 'btn-primary' : 'btn-ghost'}"
          on:click={() => switchView("week")}
        >
          This Week
        </button>
      </div>
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
            No data for {currentView === "today" ? "today" : "this week"}
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
          <Goals {currentNutrition} />

          <!-- Today's Summary Stats (for today view) -->
          {#if currentView === "today"}
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
          {/if}

          <!-- Individual Meals -->
          <div class="card bg-base-200 shadow-xl">
            <div class="card-body">
              <h2 class="card-title mb-4">
                {currentView === "today" ? "Today's" : "This Week's"} Meals
              </h2>
              <div class="space-y-3">
                {#each consumptionsData.consumptions as consumption}
                  <div class="card bg-base-100 shadow">
                    <div class="card-body p-4">
                      <!-- Meal Header - always visible -->
                      <div 
                        class="flex justify-between items-start cursor-pointer"
                        role="button"
                        tabindex="0"
                        on:click={() => toggleMealExpansion(consumption.id)}
                        on:keydown={(e) => e.key === 'Enter' && toggleMealExpansion(consumption.id)}
                      >
                        <div class="flex-1">
                          <p class="font-medium mb-1">
                            "{consumption.transcript}"
                          </p>
                          <div class="flex items-center gap-4 text-sm text-base-content/60">
                            <span>
                              {new Date(consumption.created_at).toLocaleString()}
                            </span>
                            {#if consumption.labels && consumption.labels.length > 0}
                              <div class="flex gap-1">
                                {#each consumption.labels as label}
                                  <span class="badge badge-sm" style="background-color: {label.color}20; color: {label.color}; border: 1px solid {label.color};">
                                    {label.name}
                                  </span>
                                {/each}
                              </div>
                            {/if}
                          </div>
                        </div>
                        <div class="flex items-center gap-2">
                          {#if consumption.summary?.totals?.calories}
                            <div class="badge badge-primary">
                              {Math.round(consumption.summary.totals.calories)} cal
                            </div>
                          {/if}
                          <button class="btn btn-sm btn-ghost">
                            {expandedMeals.has(consumption.id) ? '▼' : '▶'}
                          </button>
                        </div>
                      </div>

                      <!-- Expanded Details -->
                      {#if expandedMeals.has(consumption.id)}
                        <div class="mt-4 border-t pt-4">
                          <!-- Macronutrients Summary -->
              <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
                            <div class="text-center">
                <div class="text-lg font-bold text-primary">{Math.round(consumption.summary?.totals?.calories || 0)}</div>
                              <div class="text-xs text-base-content/60">Calories</div>
                            </div>
                            <div class="text-center">
                <div class="text-lg font-bold text-secondary">{(consumption.summary?.totals?.protein_g || 0).toFixed(1)}g</div>
                              <div class="text-xs text-base-content/60">Protein</div>
                            </div>
                            <div class="text-center">
                <div class="text-lg font-bold text-accent">{(consumption.summary?.totals?.total_carbs_g || 0).toFixed(1)}g</div>
                              <div class="text-xs text-base-content/60">Carbs</div>
                            </div>
                            <div class="text-center">
                <div class="text-lg font-bold text-warning">{(consumption.summary?.totals?.total_fat_g || 0).toFixed(1)}g</div>
                              <div class="text-xs text-base-content/60">Fat</div>
                            </div>
                          </div>

                          <!-- Individual Items in this meal -->
                          {#if consumption.items && consumption.items.length > 0}
                            <div class="bg-base-200 rounded-lg p-3">
                              <h4 class="font-medium mb-2 text-sm">Items in this meal:</h4>
                              <div class="space-y-2">
                                {#each consumption.items as item}
                                  <div class="flex justify-between items-center text-sm">
                                    <div class="flex-1">
                                      <span class="font-medium">{item.item?.name || ''}</span>
                                      {#if item.item?.brand}
                                        <span class="text-base-content/60">({item.item?.brand})</span>
                                      {/if}
                                      {#if item.item?.grams !== undefined}
                                        <span class="text-base-content/60">- {item.item?.grams}g</span>
                                      {/if}
                                    </div>
                                    <div class="text-right">
                                      <div class="font-medium">{Math.round(item.item?.nutrients?.calories || 0)} cal</div>
                                      <div class="text-xs text-base-content/60">
                                        P: {(item.item?.nutrients?.protein_g || 0).toFixed(1)}g |
                                        C: {(item.item?.nutrients?.total_carbs_g || 0).toFixed(1)}g |
                                        F: {(item.item?.nutrients?.total_fat_g || 0).toFixed(1)}g
                                      </div>
                                    </div>
                                  </div>
                                {/each}
                              </div>
                            </div>
                          {/if}
                          <div class="mt-4 flex justify-end">
                            <a href={`/consumptions/${consumption.id}`} class="btn btn-sm btn-outline">View</a>
                          </div>
                        </div>
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          </div>
        </div>
      {/if}
    {/if}
  </div>
</div>
