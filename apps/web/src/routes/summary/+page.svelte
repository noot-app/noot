<script lang="ts">
  import { goto } from "$app/navigation"
  import { apiClient } from "$lib/api/client"
  import { onMount } from "svelte"
  import NutritionStats from "$lib/components/NutritionStats.svelte"
  import Goals from "$lib/components/Goals.svelte"

  import Label from "$lib/components/Label.svelte"
  import ChartBarIcon from "$lib/components/icons/ChartBar.svelte"

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
            <div class="card-body p-4 sm:p-6">
              <h2 class="card-title mb-6">
                Today's Consumptions
              </h2>
              <div class="space-y-3">
                {#each consumptionsData.consumptions as consumption}
                  <!-- svelte-ignore a11y-click-events-have-key-events -->
                  <!-- svelte-ignore a11y-no-static-element-interactions -->
                  <div 
                    class="consumption-item group cursor-pointer"
                    on:click={() => goto(`/consumptions/${consumption.id}`)}
                  >
                    <!-- Time indicator (desktop only) -->
                    <div class="time-indicator desktop-only">
                      <div class="time-text">
                        {new Date(consumption.created_at).toLocaleTimeString('en-US', { 
                          hour: 'numeric', 
                          minute: '2-digit',
                          hour12: true 
                        })}
                      </div>
                    </div>
                    
                    <!-- Main content -->
                    <div class="consumption-content">
                      <!-- Header with title -->
                      <div class="consumption-header">
                        <h3 class="consumption-title">
                          {consumption.title || consumption.transcript}
                        </h3>
                        <!-- Quick nutrition stats -->
                        <div class="nutrition-quick">
                          <span class="nutrition-stat calories">
                            {consumption.summary?.totals?.calories || 0} cal
                          </span>
                          <span class="nutrition-stat protein">
                            {(consumption.summary?.totals?.protein_g || 0).toFixed(1)}g protein
                          </span>
                          <span class="nutrition-stat carbs">
                            {(consumption.summary?.totals?.total_carbs_g || 0).toFixed(1)}g carbs
                          </span>
                          <span class="nutrition-stat fat">
                            {(consumption.summary?.totals?.total_fat_g || 0).toFixed(1)}g fat
                          </span>
                        </div>
                      </div>
                      
                      <!-- Labels and mobile timestamp -->
                      <div class="bottom-section">
                        <!-- Labels (always takes up left space, even if empty) -->
                        <div class="consumption-labels">
                          {#if consumption.labels && consumption.labels.length > 0}
                            {#each consumption.labels.slice(0, 6) as label}
                              <Label 
                                name={label.name} 
                                color={label.color} 
                                size="xs"
                              />
                            {/each}
                            {#if consumption.labels.length > 6}
                              <span class="label-more">+{consumption.labels.length - 6}</span>
                            {/if}
                          {/if}
                        </div>
                        
                        <!-- Time indicator (mobile only) - always on the right -->
                        <div class="time-indicator mobile-only">
                          <div class="time-text mobile-time">
                            {new Date(consumption.created_at).toLocaleTimeString('en-US', { 
                              hour: 'numeric', 
                              minute: '2-digit',
                              hour12: true 
                            })}
                          </div>
                        </div>
                      </div>
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

<style>
  /* === CONSUMPTION TIMELINE STYLES === */
  .consumption-item {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.75rem;
    padding: 1rem;
    background: rgba(255, 255, 255, 0.7);
    border: 1px solid rgba(0, 0, 0, 0.06);
    border-radius: 1rem;
    transition: all 0.2s ease;
    position: relative;
    overflow: hidden;
  }

  .consumption-item:hover {
    background: rgba(255, 255, 255, 0.9);
    border-color: rgba(0, 0, 0, 0.12);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
    transform: translateY(-1px);
  }

  /* === TIME INDICATOR STYLES === */
  .time-indicator {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
    padding-top: 0.125rem;
    min-width: 60px;
  }

  .time-text {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--color-base-content);
    text-align: center;
    white-space: nowrap;
    background: rgba(var(--color-primary-rgb), 0.1);
    padding: 0.25rem 0.5rem;
    border-radius: 0.5rem;
    border: 1px solid rgba(var(--color-primary-rgb), 0.2);
  }

  /* === RESPONSIVE VISIBILITY === */
  .mobile-only { display: none; }
  .desktop-only { display: flex; }

  /* === CONTENT LAYOUT === */
  .consumption-content { min-width: 0; }
  .consumption-header { margin-bottom: 0.75rem; }
  .bottom-section { 
    display: flex; 
    flex-direction: column; 
    gap: 0.5rem; 
  }

  .consumption-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-base-content);
    margin-bottom: 0.5rem;
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  /* === NUTRITION STATS === */
  .nutrition-quick {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  .nutrition-stat {
    font-size: 0.8125rem;
    font-weight: 600;
    padding: 0.25rem 0.5rem;
    border-radius: 0.5rem;
    border: 1px solid;
    background: rgba(255, 255, 255, 0.8);
  }

  .nutrition-stat.calories {
    color: var(--color-honey);
    border-color: rgba(var(--color-honey-rgb), 0.3);
  }
  .nutrition-stat.protein {
    color: var(--color-secondary);
    border-color: rgba(var(--color-secondary-rgb), 0.3);
  }
  .nutrition-stat.carbs {
    color: var(--color-warning);
    border-color: rgba(var(--color-warning-rgb), 0.3);
  }
  .nutrition-stat.fat {
    color: var(--color-accent);
    border-color: rgba(var(--color-accent-rgb), 0.3);
  }

  /* === LABELS === */
  .consumption-labels {
    display: flex;
    flex-wrap: wrap;
    gap: 0.375rem;
    margin-top: 0.5rem;
  }

  .label-more {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-base-content-lighter);
    padding: 0.1875rem 0.5rem;
  }

  /* === MOBILE RESPONSIVE === */
  @media (max-width: 640px) {
    .desktop-only { display: none; }
    .mobile-only { display: flex; }
    
    .consumption-item {
      grid-template-columns: 1fr;
      gap: 0.5rem;
      padding: 0.75rem;
    }
    
    .consumption-header { margin-bottom: 0.5rem; }
    .consumption-title {
      font-size: 1rem;
      margin-bottom: 0.375rem;
      -webkit-line-clamp: 1;
      line-clamp: 1;
    }
    
    .nutrition-quick {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 0.375rem;
    }
    
    .nutrition-stat {
      font-size: 0.6875rem;
      padding: 0.1875rem 0.375rem;
      text-align: center;
    }
    
    .bottom-section {
      flex-direction: row;
      justify-content: space-between;
      align-items: flex-end;
      gap: 0.75rem;
      margin-top: 0.5rem;
    }
    
    .consumption-labels {
      flex: 1;
      margin-top: 0;
    }
    
    .mobile-time {
      font-size: 0.6875rem;
      padding: 0.1875rem 0.375rem;
      flex-shrink: 0;
    }
  }

  @media (max-width: 440px) {
    .nutrition-stat {
      font-size: 0.625rem;
      padding: 0.125rem 0.25rem;
    }
    
    .mobile-time {
      font-size: 0.625rem;
      padding: 0.125rem 0.25rem;
    }
  }
</style>
