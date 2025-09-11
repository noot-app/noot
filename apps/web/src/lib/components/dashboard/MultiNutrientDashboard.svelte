<script lang="ts">
  import PerNutrientDashboard from "./PerNutrientDashboard.svelte"
  
  export let consumptionsData: any = null
  export let goalsData: any = null
  export let timeWindow: "1d" | "7d" | "30d" = "7d"
  export let selectedNutrients: string[] = []

  // Complete list of nutrients that can be tracked
  const allNutrients = [
    { key: 'calories', name: 'Calories', unit: 'kcal', category: 'Energy' },
    { key: 'protein_g', name: 'Protein', unit: 'g', category: 'Macronutrients' },
    { key: 'total_fat_g', name: 'Total Fat', unit: 'g', category: 'Macronutrients' },
    { key: 'saturated_fat_g', name: 'Saturated Fat', unit: 'g', category: 'Macronutrients' },
    { key: 'trans_fat_g', name: 'Trans Fat', unit: 'g', category: 'Macronutrients' },
    { key: 'monounsaturated_fat_g', name: 'Monounsaturated Fat', unit: 'g', category: 'Macronutrients' },
    { key: 'polyunsaturated_fat_g', name: 'Polyunsaturated Fat', unit: 'g', category: 'Macronutrients' },
    { key: 'cholesterol_mg', name: 'Cholesterol', unit: 'mg', category: 'Macronutrients' },
    { key: 'sodium_mg', name: 'Sodium', unit: 'mg', category: 'Minerals' },
    { key: 'total_carbs_g', name: 'Total Carbs', unit: 'g', category: 'Macronutrients' },
    { key: 'dietary_fiber_g', name: 'Dietary Fiber', unit: 'g', category: 'Macronutrients' },
    { key: 'total_sugars_g', name: 'Total Sugars', unit: 'g', category: 'Macronutrients' },
    { key: 'added_sugars_g', name: 'Added Sugars', unit: 'g', category: 'Macronutrients' },
    { key: 'vitamin_a_mcg', name: 'Vitamin A', unit: 'mcg', category: 'Vitamins' },
    { key: 'vitamin_c_mg', name: 'Vitamin C', unit: 'mg', category: 'Vitamins' },
    { key: 'vitamin_d_mcg', name: 'Vitamin D', unit: 'mcg', category: 'Vitamins' },
    { key: 'vitamin_e_mg', name: 'Vitamin E', unit: 'mg', category: 'Vitamins' },
    { key: 'vitamin_k_mcg', name: 'Vitamin K', unit: 'mcg', category: 'Vitamins' },
    { key: 'thiamine_mg', name: 'Thiamine (B1)', unit: 'mg', category: 'Vitamins' },
    { key: 'riboflavin_mg', name: 'Riboflavin (B2)', unit: 'mg', category: 'Vitamins' },
    { key: 'niacin_mg', name: 'Niacin (B3)', unit: 'mg', category: 'Vitamins' },
    { key: 'vitamin_b6_mg', name: 'Vitamin B6', unit: 'mg', category: 'Vitamins' },
    { key: 'folate_mcg', name: 'Folate', unit: 'mcg', category: 'Vitamins' },
    { key: 'vitamin_b12_mcg', name: 'Vitamin B12', unit: 'mcg', category: 'Vitamins' },
    { key: 'biotin_mcg', name: 'Biotin', unit: 'mcg', category: 'Vitamins' },
    { key: 'pantothenic_acid_mg', name: 'Pantothenic Acid', unit: 'mg', category: 'Vitamins' },
    { key: 'choline_mg', name: 'Choline', unit: 'mg', category: 'Vitamins' },
    { key: 'calcium_mg', name: 'Calcium', unit: 'mg', category: 'Minerals' },
    { key: 'iron_mg', name: 'Iron', unit: 'mg', category: 'Minerals' },
    { key: 'magnesium_mg', name: 'Magnesium', unit: 'mg', category: 'Minerals' },
    { key: 'phosphorus_mg', name: 'Phosphorus', unit: 'mg', category: 'Minerals' },
    { key: 'potassium_mg', name: 'Potassium', unit: 'mg', category: 'Minerals' },
    { key: 'zinc_mg', name: 'Zinc', unit: 'mg', category: 'Minerals' },
    { key: 'copper_mg', name: 'Copper', unit: 'mg', category: 'Minerals' },
    { key: 'manganese_mg', name: 'Manganese', unit: 'mg', category: 'Minerals' },
    { key: 'selenium_mcg', name: 'Selenium', unit: 'mcg', category: 'Minerals' },
    { key: 'iodine_mcg', name: 'Iodine', unit: 'mcg', category: 'Minerals' },
    { key: 'molybdenum_mcg', name: 'Molybdenum', unit: 'mcg', category: 'Minerals' },
    { key: 'chromium_mcg', name: 'Chromium', unit: 'mcg', category: 'Minerals' },
    { key: 'fluoride_mg', name: 'Fluoride', unit: 'mg', category: 'Minerals' },
    { key: 'chloride_mg', name: 'Chloride', unit: 'mg', category: 'Minerals' },
    { key: 'omega3_ala_g', name: 'Omega-3 ALA', unit: 'g', category: 'Fatty Acids' },
    { key: 'omega3_epa_g', name: 'Omega-3 EPA', unit: 'g', category: 'Fatty Acids' },
    { key: 'omega3_dha_g', name: 'Omega-3 DHA', unit: 'g', category: 'Fatty Acids' },
    { key: 'omega6_g', name: 'Omega-6', unit: 'g', category: 'Fatty Acids' },
    { key: 'caffeine_mg', name: 'Caffeine', unit: 'mg', category: 'Compounds' },
    { key: 'creatine_mg', name: 'Creatine', unit: 'mg', category: 'Compounds' }
  ]

  // Default popular nutrients if none selected
  const defaultNutrients = [
    'calories', 'protein_g', 'total_carbs_g', 'total_fat_g', 
    'dietary_fiber_g', 'sodium_mg', 'vitamin_c_mg', 'calcium_mg',
    'iron_mg', 'potassium_mg', 'added_sugars_g', 'saturated_fat_g'
  ]

  // Determine which nutrients to display
  $: displayNutrients = selectedNutrients.length > 0 ? selectedNutrients : defaultNutrients
  
  // Filter nutrients based on what's being displayed and what has data
  $: nutrientsToShow = allNutrients.filter(nutrient => {
    const isSelected = displayNutrients.includes(nutrient.key)
    const hasGoal = goalsData?.targets?.[nutrient.key] || goalsData?.upper_limits?.[nutrient.key]
    const hasData = consumptionsData?.consumptions?.some((c: any) => 
      c.summary?.totals?.[nutrient.key] && c.summary.totals[nutrient.key] > 0
    )
    
    return isSelected && (hasGoal || hasData)
  })

  // Group nutrients by category for better organization
  $: nutrientsByCategory = nutrientsToShow.reduce((acc: any, nutrient) => {
    if (!acc[nutrient.category]) {
      acc[nutrient.category] = []
    }
    acc[nutrient.category].push(nutrient)
    return acc
  }, {})

  // Category display order
  const categoryOrder = ['Energy', 'Macronutrients', 'Vitamins', 'Minerals', 'Fatty Acids', 'Compounds']
</script>

<div class="space-y-8">
  <!-- Header with time window selector -->
  <div class="flex justify-between items-center">
    <div>
      <h2 class="text-xl font-bold text-base-content">Per-Nutrient Dashboards</h2>
      <p class="text-sm text-base-content/70 mt-1">
        Individual tracking for {nutrientsToShow.length} nutrients with goal boundaries
      </p>
    </div>
    
    <!-- Time Window Selector -->
    <div class="flex gap-2">
      <button 
        class="btn btn-sm {timeWindow === '1d' ? 'btn-primary' : 'btn-outline'}"
        on:click={() => timeWindow = '1d'}
      >
        1 Day
      </button>
      <button 
        class="btn btn-sm {timeWindow === '7d' ? 'btn-primary' : 'btn-outline'}"
        on:click={() => timeWindow = '7d'}
      >
        7 Days
      </button>
      <button 
        class="btn btn-sm {timeWindow === '30d' ? 'btn-primary' : 'btn-outline'}"
        on:click={() => timeWindow = '30d'}
      >
        30 Days
      </button>
    </div>
  </div>

  <!-- Nutrients organized by category -->
  {#each categoryOrder as category}
    {#if nutrientsByCategory[category] && nutrientsByCategory[category].length > 0}
      <div class="space-y-4">
        <h3 class="text-lg font-semibold text-base-content border-b border-base-300 pb-2">
          {category}
        </h3>
        
        <!-- Responsive grid for nutrient charts -->
        <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-4">
          {#each nutrientsByCategory[category] as nutrient}
            <div class="card bg-base-200 shadow-sm">
              <div class="card-body p-4">
                <PerNutrientDashboard
                  nutrientKey={nutrient.key}
                  nutrientName={nutrient.name}
                  unit={nutrient.unit}
                  consumptions={consumptionsData?.consumptions || []}
                  {goalsData}
                  {timeWindow}
                />
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  {/each}

  <!-- Empty state -->
  {#if nutrientsToShow.length === 0}
    <div class="text-center py-16">
      <svg class="w-16 h-16 mx-auto mb-4 opacity-50" fill="currentColor" viewBox="0 0 20 20">
        <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h6a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h6a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h4a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
      </svg>
      <h3 class="text-lg font-semibold text-base-content mb-2">No Nutrient Data Available</h3>
      <p class="text-base-content/70 max-w-md mx-auto">
        Start logging some consumption data or set up custom nutrition goals to see individual nutrient tracking charts.
      </p>
    </div>
  {/if}
</div>