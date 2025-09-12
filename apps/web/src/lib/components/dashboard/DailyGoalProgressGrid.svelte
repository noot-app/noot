<script lang="ts">
  import CheckCircle from "$lib/components/icons/CheckCircle.svelte"
  import XCircle from "$lib/components/icons/XCircle.svelte"

  export let consumptions: any[] = []
  export let goals: any = null
  export let dateRange: string = "7d"

  // All nutrients from the nutrients.yml config
  const nutrients = [
    { key: 'calories', name: 'Calories', target: 2000, unit: 'kcal' },
    { key: 'protein_g', name: 'Protein', target: 50, unit: 'g' },
    { key: 'total_fat_g', name: 'Total Fat', target: 65, unit: 'g' },
    { key: 'saturated_fat_g', name: 'Saturated Fat', target: 20, unit: 'g' },
    { key: 'trans_fat_g', name: 'Trans Fat', target: 0, unit: 'g' },
    { key: 'monounsaturated_fat_g', name: 'Monounsaturated Fat', target: 20, unit: 'g' },
    { key: 'polyunsaturated_fat_g', name: 'Polyunsaturated Fat', target: 10, unit: 'g' },
    { key: 'cholesterol_mg', name: 'Cholesterol', target: 300, unit: 'mg' },
    { key: 'sodium_mg', name: 'Sodium', target: 2300, unit: 'mg' },
    { key: 'total_carbs_g', name: 'Total Carbs', target: 300, unit: 'g' },
    { key: 'dietary_fiber_g', name: 'Dietary Fiber', target: 25, unit: 'g' },
    { key: 'total_sugars_g', name: 'Total Sugars', target: 50, unit: 'g' },
    { key: 'added_sugars_g', name: 'Added Sugars', target: 25, unit: 'g' },
    { key: 'vitamin_a_mcg', name: 'Vitamin A', target: 900, unit: 'mcg' },
    { key: 'vitamin_c_mg', name: 'Vitamin C', target: 90, unit: 'mg' },
    { key: 'vitamin_d_mcg', name: 'Vitamin D', target: 20, unit: 'mcg' },
    { key: 'vitamin_e_mg', name: 'Vitamin E', target: 15, unit: 'mg' },
    { key: 'vitamin_k_mcg', name: 'Vitamin K', target: 120, unit: 'mcg' },
    { key: 'thiamine_mg', name: 'Thiamine (B1)', target: 1.2, unit: 'mg' },
    { key: 'riboflavin_mg', name: 'Riboflavin (B2)', target: 1.3, unit: 'mg' },
    { key: 'niacin_mg', name: 'Niacin (B3)', target: 16, unit: 'mg' },
    { key: 'vitamin_b6_mg', name: 'Vitamin B6', target: 1.7, unit: 'mg' },
    { key: 'folate_mcg', name: 'Folate', target: 400, unit: 'mcg' },
    { key: 'vitamin_b12_mcg', name: 'Vitamin B12', target: 2.4, unit: 'mcg' },
    { key: 'biotin_mcg', name: 'Biotin', target: 30, unit: 'mcg' },
    { key: 'pantothenic_acid_mg', name: 'Pantothenic Acid', target: 5, unit: 'mg' },
    { key: 'choline_mg', name: 'Choline', target: 550, unit: 'mg' },
    { key: 'calcium_mg', name: 'Calcium', target: 1000, unit: 'mg' },
    { key: 'iron_mg', name: 'Iron', target: 8, unit: 'mg' },
    { key: 'magnesium_mg', name: 'Magnesium', target: 400, unit: 'mg' },
    { key: 'phosphorus_mg', name: 'Phosphorus', target: 700, unit: 'mg' },
    { key: 'potassium_mg', name: 'Potassium', target: 4700, unit: 'mg' },
    { key: 'zinc_mg', name: 'Zinc', target: 11, unit: 'mg' },
    { key: 'copper_mg', name: 'Copper', target: 0.9, unit: 'mg' },
    { key: 'manganese_mg', name: 'Manganese', target: 2.3, unit: 'mg' },
    { key: 'selenium_mcg', name: 'Selenium', target: 55, unit: 'mcg' },
    { key: 'iodine_mcg', name: 'Iodine', target: 150, unit: 'mcg' },
    { key: 'molybdenum_mcg', name: 'Molybdenum', target: 45, unit: 'mcg' },
    { key: 'chromium_mcg', name: 'Chromium', target: 35, unit: 'mcg' },
    { key: 'fluoride_mg', name: 'Fluoride', target: 4, unit: 'mg' },
    { key: 'chloride_mg', name: 'Chloride', target: 2300, unit: 'mg' },
    { key: 'omega3_ala_g', name: 'Omega-3 ALA', target: 1.6, unit: 'g' },
    { key: 'omega3_epa_g', name: 'Omega-3 EPA', target: 0.25, unit: 'g' },
    { key: 'omega3_dha_g', name: 'Omega-3 DHA', target: 0.25, unit: 'g' },
    { key: 'omega6_g', name: 'Omega-6', target: 12, unit: 'g' },
    { key: 'caffeine_mg', name: 'Caffeine', target: 400, unit: 'mg' },
    { key: 'creatine_mg', name: 'Creatine', target: 3000, unit: 'mg' }
  ]

  // Calculate daily progress data
  $: progressData = calculateDailyProgress(consumptions, goals, dateRange)

  function calculateDailyProgress(data: any[], goalsData: any, dateRange: string) {
    if (!data || data.length === 0) {
      return { days: [], dailyData: [], nutrients }
    }

    const days = parseInt(dateRange.replace('d', ''))
    
    // Generate date labels for the past N days
    const dateLabels = []
    const today = new Date()
    for (let i = days - 1; i >= 0; i--) {
      const date = new Date(today)
      date.setDate(date.getDate() - i)
      dateLabels.push({
        date: date.toISOString().split('T')[0],
        label: i === 0 ? 'Today' : date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
      })
    }

    // Group consumptions by date
    const consumptionsByDate = data.reduce((acc: any, consumption: any) => {
      const date = new Date(consumption.consumed_at || consumption.created_at).toISOString().split('T')[0]
      if (!acc[date]) acc[date] = []
      acc[date].push(consumption)
      return acc
    }, {})

    // Get goal targets
    const targets = goalsData?.goals?.targets || {}

    // Calculate daily totals and goal achievement for each nutrient
    const dailyData = dateLabels.map(({ date, label }) => {
      const dayConsumptions = consumptionsByDate[date] || []
      
      // Sum up nutrients for this day
      const dayTotals = dayConsumptions.reduce((acc: any, consumption: any) => {
        nutrients.forEach(nutrient => {
          const value = consumption.summary?.totals?.[nutrient.key] || 0
          acc[nutrient.key] = (acc[nutrient.key] || 0) + value
        })
        return acc
      }, {})

      // Check goal achievement for each nutrient
      const nutrientStatus = nutrients.map(nutrient => {
        const actual = dayTotals[nutrient.key] || 0
        const target = targets[nutrient.key] || nutrient.target
        const achieved = actual >= target * 0.8 // 80% threshold for "achieved"
        
        return {
          key: nutrient.key,
          name: nutrient.name,
          actual: Math.round(actual * 100) / 100,
          target,
          unit: nutrient.unit,
          achieved,
          percentage: Math.round((actual / target) * 100)
        }
      })

      return {
        date,
        label,
        nutrients: nutrientStatus
      }
    })

    return {
      days: dateLabels,
      dailyData,
      nutrients
    }
  }
</script>

<div class="w-full">
  <div class="mb-4">
    <h3 class="text-lg font-semibold text-base-content mb-2">Daily Goal Progress</h3>
    <p class="text-sm text-base-content-lighter">
      Track your nutrition goal achievement for the past {parseInt(dateRange.replace('d', ''))} days across {progressData.nutrients?.length || 0} nutrients
    </p>
  </div>

  {#if progressData.dailyData.length > 0}
    <!-- Compact table-style grid -->
    <div class="daily-progress-grid bg-base-200 rounded-lg overflow-hidden border border-base-300">
      <!-- Header row with dates -->
      <div class="daily-progress-header bg-base-300 border-b border-base-300 sticky top-0 z-10">
        <div class="grid gap-0" style="grid-template-columns: 150px repeat({progressData.days.length}, 60px);">
          <div class="p-2 text-xs font-semibold text-base-content border-r border-base-300 bg-base-300 sticky left-0 z-20">
            Nutrient
          </div>
          {#each progressData.days as day}
            <div class="p-2 text-center text-xs font-semibold text-base-content border-r border-base-300 last:border-r-0 bg-base-300">
              <div class="truncate">{day.label}</div>
            </div>
          {/each}
        </div>
      </div>

      <!-- Nutrient rows - show all nutrients -->
      <div class="daily-progress-body max-h-80 overflow-y-auto overscroll-contain">
        {#each progressData.nutrients as nutrient}
          <div class="border-b border-base-300 last:border-b-0 hover:bg-base-100 transition-colors">
            <div class="grid gap-0" style="grid-template-columns: 150px repeat({progressData.days.length}, 60px);">
              <!-- Nutrient name -->
              <div class="p-2 text-xs font-medium text-base-content border-r border-base-300 flex items-center bg-base-200 sticky left-0 z-10">
                <span class="truncate" title={nutrient.name}>{nutrient.name}</span>
              </div>
              
              <!-- Daily status cells -->
              {#each progressData.dailyData as dayData}
                {@const nutrientData = dayData.nutrients.find(n => n.key === nutrient.key)}
                <div class="p-2 border-r border-base-300 last:border-r-0 flex items-center justify-center h-10 touch-manipulation">
                  {#if nutrientData}
                    <div 
                      class="cursor-help w-5 h-5 flex items-center justify-center touch-target"
                      title="{nutrientData.name}: {nutrientData.actual}{nutrientData.unit} / {nutrientData.target}{nutrientData.unit} ({nutrientData.percentage}%)"
                      role="button"
                      tabindex="0"
                      on:keydown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault()
                          // Could show detailed info in future
                        }
                      }}
                    >
                      {#if nutrientData.achieved}
                        <CheckCircle className="w-4 h-4 text-success" />
                      {:else}
                        <XCircle className="w-4 h-4 text-error" />
                      {/if}
                    </div>
                  {:else}
                    <div class="w-4 h-4 bg-base-300 rounded-full opacity-30"></div>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    </div>

    <!-- Compact legend -->
    <div class="flex items-center justify-center gap-6 mt-3 text-xs text-base-content-lighter">
      <div class="flex items-center gap-1">
        <CheckCircle className="w-3 h-3 text-success" />
        <span>Goal met</span>
      </div>
      <div class="flex items-center gap-1">
        <XCircle className="w-3 h-3 text-error" />
        <span>Goal missed</span>
      </div>
      <div class="flex items-center gap-1">
        <div class="w-3 h-3 bg-base-300 rounded-full opacity-30"></div>
        <span>No data</span>
      </div>
    </div>
  {:else}
    <div class="h-48 flex items-center justify-center text-base-content-lighter bg-base-200 rounded-lg">
      <div class="text-center">
        <svg class="w-12 h-12 mx-auto mb-2 opacity-50" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h6a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h6a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h4a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
        </svg>
        <p class="text-sm">No consumption data available</p>
      </div>
    </div>
  {/if}
</div>

<style>
  /* Mobile-optimized grid behavior */
  :global(.daily-progress-grid) {
    /* Prevent excessive bounce scrolling on mobile */
    overscroll-behavior: contain;
    /* Ensure content stays within bounds */
    overflow: hidden;
  }

  :global(.daily-progress-body) {
    /* Controlled scrolling with proper bounds */
    overscroll-behavior: contain;
    -webkit-overflow-scrolling: touch;
  }

  /* Ensure sticky elements work properly */
  :global(.daily-progress-header) {
    position: sticky;
    top: 0;
    z-index: 10;
  }

  /* Touch-friendly targets for mobile */
  :global(.touch-target) {
    min-width: 44px;
    min-height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  /* Overscroll containment utility */
  :global(.overscroll-contain) {
    overscroll-behavior: contain;
  }

  /* Mobile-specific improvements */
  @media (max-width: 768px) {
    :global(.daily-progress-grid) {
      /* More constrained on mobile */
      max-height: 70vh;
    }
    
    :global(.daily-progress-body) {
      /* Better scrolling behavior on mobile */
      scroll-behavior: smooth;
      /* Prevent momentum scrolling from going too far */
      -webkit-overflow-scrolling: touch;
      overscroll-behavior-y: contain;
    }
  }

  /* Extra small screens */
  @media (max-width: 480px) {
    :global(.daily-progress-grid) {
      max-height: 60vh;
    }
  }
</style>
