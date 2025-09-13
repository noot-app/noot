<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { onMount } from "svelte"
  import NutritionStats from "$lib/components/NutritionStats.svelte"
  import Goals from "$lib/components/Goals.svelte"
  import ConsumptionCard from "$lib/components/ConsumptionCard.svelte"
  import ChartBarIcon from "$lib/components/icons/ChartBar.svelte"
  import Bolt from "$lib/components/icons/Bolt.svelte"
  import Wheat from "$lib/components/icons/Wheat.svelte"
  import Sprout from "$lib/components/icons/Sprout.svelte"
  import Scale from "$lib/components/icons/Scale.svelte"
  import Candy from "$lib/components/icons/Candy.svelte"
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
            Start logging your consumptions to see your nutrition summary
          </p>
          <a href="/record" class="btn btn-primary"> Record Your First Consumption </a>
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

          <!-- Quick Stats (utility-based tiles) -->
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3" aria-label="Today's quick nutrition stats">
            <!-- Consumptions -->
            <div class="stat-tile honey-border" aria-label="Consumptions logged">
              <div class="flex items-center justify-center h-full px-2">
                <div class="flex flex-col items-center text-center flex-1">
                  <span class="tile-label">Consumptions</span>
                  <div class="tile-value mt-2">{consumptionsData.consumptions.length}</div>
                </div>
                <div class="flex items-center justify-center ml-4">
                  <ChartBarIcon className="w-7 h-7 text-honey" />
                </div>
              </div>
            </div>
            <!-- Sodium -->
            <div class="stat-tile secondary-border" aria-label="Total sodium milligrams">
              <div class="flex items-center justify-center h-full px-2">
                <div class="flex flex-col items-center text-center flex-1">
                  <span class="tile-label">Sodium</span>
                  <div class="tile-value mt-2">{Math.round(nutritionTotals?.sodium_mg || 0)}<span class="unit ml-1">mg</span></div>
                </div>
                <div class="flex items-center justify-center ml-4">
                  <Scale className="w-7 h-7 text-secondary" />
                </div>
              </div>
            </div>
            <!-- Fiber -->
            <div class="stat-tile accent-border" aria-label="Total fiber grams">
              <div class="flex items-center justify-center h-full px-2">
                <div class="flex flex-col items-center text-center flex-1">
                  <span class="tile-label">Fiber</span>
                  <div class="tile-value mt-2">{(nutritionTotals?.dietary_fiber_g || 0).toFixed(1)}<span class="unit ml-1">g</span></div>
                </div>
                <div class="flex items-center justify-center ml-4">
                  <Sprout className="w-7 h-7 text-accent" />
                </div>
              </div>
            </div>
            <!-- Added Sugars -->
            <div class="stat-tile warning-border" aria-label="Total added sugars grams">
              <div class="flex items-center justify-center h-full px-2">
                <div class="flex flex-col items-center text-center flex-1">
                  <span class="tile-label">Added Sugars</span>
                  <div class="tile-value mt-2">{(nutritionTotals?.added_sugars_g || 0).toFixed(1)}<span class="unit ml-1">g</span></div>
                </div>
                <div class="flex items-center justify-center ml-4">
                  <Candy className="w-7 h-7 text-warning" />
                </div>
              </div>
            </div>
          </div>

          <!-- Nutrition Goals (with integrated footer stats) -->
          <Goals {currentNutrition} showAllCategories={true} />

          <!-- Individual Consumptions -->
          <div class="card bg-base-200 shadow-xl">
            <div class="card-body">
              <h2 class="card-title mb-4">
                Today's Consumptions
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

<style>
  .stat-tile {
    background: var(--color-base-100);
    border: 2px solid transparent;
    border-radius: 0.75rem;
    padding: 0.8rem 0.6rem;
    min-height: 92px;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  }
  .stat-tile.warning-border { 
    border-color: var(--color-warning); 
    border-color: color-mix(in srgb, var(--color-warning) 20%, transparent);
  }
  .stat-tile.accent-border { 
    border-color: var(--color-accent); 
    border-color: color-mix(in srgb, var(--color-accent) 20%, transparent);
  }
  .stat-tile.secondary-border { 
    border-color: var(--color-secondary); 
    border-color: color-mix(in srgb, var(--color-secondary) 20%, transparent);
  }
  .stat-tile.honey-border { 
    border-color: var(--color-honey); 
    border-color: color-mix(in srgb, var(--color-honey) 20%, transparent);
  }
  .tile-label { 
    font-size: 0.58rem; 
    text-transform: uppercase; 
    letter-spacing: 0.08em; 
    font-weight: 600; 
    color: var(--color-base-content-lighter); 
  }
  .tile-value { 
    font-size: 1.35rem; 
    font-weight: 600; 
    line-height: 1.1; 
    display: flex; 
    align-items: baseline; 
    color: var(--color-base-content);
  }
  .tile-value .unit { font-size: 0.65rem; opacity: 0.65; font-weight: 500; }
  @media (max-width: 440px) {
    .stat-tile { min-height: 82px; padding: 0.65rem 0.7rem; }
    .tile-value { font-size: 1.2rem; }
  }
</style>
