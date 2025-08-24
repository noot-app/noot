<script lang="ts">
  import { onMount } from "svelte";
  import { apiClient } from "$lib/api/client";
  import { isPro } from "$lib/auth/store";
  import NutrientCategoryDisplay from "./NutrientCategoryDisplay.svelte";
  import type { paths } from "$lib/api/schema";

  type GoalsResponse = paths["/goals"]["get"]["responses"]["200"]["content"]["application/json"];
  type Goals = GoalsResponse["goals"];

  export let currentNutrition: Record<string, number> = {};
  export let showMealContribution = false; // New prop to indicate meal-specific view
  export let isSharedView = false; // New prop for shareable link context (non-logged-in users)
  export let title = "Nutrition Goals"; // Customizable title

  let goals: Goals | null = null;
  let loading = true;
  let error = "";

  // Compute dynamic title based on context
  $: dynamicTitle = (() => {
    if (showMealContribution) {
      if (isSharedView || (goals?.source === "dri")) {
        return "How this meal contributes to DRI Nutrition Targets";
      }
      return "How this meal contributes to your daily goals";
    }
    return title;
  })();

  // Determine if we're in DRI mode
  $: isDriMode = isSharedView || (goals?.source === "dri");

  // Helper to safely get current nutrient value
  function getCurrentNutrient(nutrient: string): number {
    const value = currentNutrition?.[nutrient] || 0;
    return isFinite(value) ? value : 0;
  }

  async function loadGoals() {
    try {
      loading = true;
      error = "";
      
      // For shared views, always use DRI defaults
      if (isSharedView) {
        const response = await apiClient.GET("/goals");
        
        if (response.error) {
          throw new Error(`API Error: ${response.error}`);
        }
        
        goals = response.data.goals;
      } else {
        // Normal user goals loading with DRI fallback
        const response = await apiClient.GET("/goals");
        
        if (response.error) {
          throw new Error(`API Error: ${response.error}`);
        }

        goals = response.data.goals;
      }
    } catch (err) {
      error = `Failed to load nutrition goals: ${err}`;
      console.error("Goals error:", err);
    } finally {
      loading = false;
    }
  }

  onMount(async () => {
    await loadGoals();
  });

  // All nutrients to display - organized by category for complete DRI coverage
  const keyNutrients = [
    // Essential macronutrients
    "calories", "protein_g", "total_carbs_g", "total_fat_g", "saturated_fat_g", "trans_fat_g", "dietary_fiber_g", "total_sugars_g", "added_sugars_g", "cholesterol_mg", "sodium_mg",
    
    // B-Complex vitamins
    "thiamine_mg", "riboflavin_mg", "niacin_mg", "vitamin_b6_mg", "folate_mcg", "vitamin_b12_mcg", "biotin_mcg", "pantothenic_acid_mg",
    
    // Fat-soluble vitamins  
    "vitamin_a_mcg", "vitamin_d_mcg", "vitamin_e_mg", "vitamin_k_mcg",
    
    // Water-soluble vitamins
    "vitamin_c_mg", "choline_mg",
    
    // Essential minerals
    "calcium_mg", "iron_mg", "magnesium_mg", "phosphorus_mg", "potassium_mg", "zinc_mg", "copper_mg", "manganese_mg", "selenium_mcg", "iodine_mcg", "molybdenum_mcg", "chromium_mcg", "fluoride_mg", "chloride_mg"
  ];
</script>

<div class="card bg-base-200 shadow-lg">
  <div class="card-body">
    <h2 class="card-title flex items-center gap-2 text-primary">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
      </svg>
      Nutrition Targets
      <!-- DRI tooltip when showing DRI targets -->
      {#if (showMealContribution && isDriMode)}
        <div class="tooltip tooltip-bottom" data-tip="Dietary Reference Intakes (DRI) are nutrient reference values developed by health experts to help individuals achieve adequate nutrition.">
          <a href="https://www.nal.usda.gov/human-nutrition-and-food-safety/dietary-guidance" target="_blank" rel="noopener noreferrer" class="text-info hover:text-info-focus text-sm ml-1" aria-label="Learn more about DRI">
            <svg class="w-4 h-4 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </a>
        </div>
      {/if}
    </h2>
    
    <!-- Subtext for meal contribution -->
    {#if showMealContribution}
      <p class="text-sm text-base-content/70 mb-4">{dynamicTitle}</p>
    {/if}

    {#if loading}
      <div class="text-center py-4">
        <span class="loading loading-spinner loading-md"></span>
        <p class="text-sm text-base-content/70 mt-2">Loading goals...</p>
      </div>
    {:else if error}
      <div class="alert alert-error">
        <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span class="text-sm">{error}</span>
      </div>
    {:else if goals}
      <div class="space-y-3">
        <!-- Goals Header -->
        <div class="flex justify-between items-center">
          <div class="text-sm text-base-content/70">
            {#if goals.source === "custom"}
              <span class="badge badge-primary">
                {goals.custom_name || "Custom Goals"}
              </span>
            {:else}
              <div class="tooltip tooltip-bottom" data-tip="Dietary Reference Intakes (DRI) are nutrient reference values developed by health experts. Learn more at nal.usda.gov">
                <a href="https://www.nal.usda.gov/human-nutrition-and-food-safety/dietary-guidance" target="_blank" rel="noopener noreferrer" class="badge badge-primary hover:badge-primary-focus">
                DRI Guidelines
                <svg class="w-3 h-3 ml-1 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                </svg>
                </a>
              </div>
            {/if}
          </div>
        </div>

        <!-- Key Nutrients Progress -->
        <div class="space-y-6">
          {#if goals}
            <!-- Regular Nutrition Targets -->
            {@const targetNutrients = keyNutrients.filter(n => goals?.targets[n] !== undefined)}
            {#if targetNutrients.length > 0}
              <div>
                <NutrientCategoryDisplay 
                  nutrients={currentNutrition} 
                  showProgress={true}
                  showGoals={true}
                  isExpandable={false}
                  title=""
                  {showMealContribution}
                />
              </div>
            {/if}

            <!-- Upper Limits (Minimize These) -->
            {@const limitNutrients = keyNutrients.filter(n => goals?.upper_limits?.[n] !== undefined)}
            {#if limitNutrients.length > 0}
              <div>
                <h4 class="font-semibold text-base mb-3 text-warning">Upper Limits (Minimize These)</h4>
                <p class="text-xs text-base-content/70 mb-3">These nutrients should be consumed as little as possible for optimal health.</p>
                <NutrientCategoryDisplay 
                  nutrients={currentNutrition} 
                  showProgress={true}
                  showGoals={true}
                  isExpandable={false}
                  title=""
                  {showMealContribution}
                  showLimitsOnly={true}
                />
              </div>
            {/if}
          {/if}
        </div>

        <!-- Summary Stats -->
        {#if goals}
          {@const availableNutrients = keyNutrients.filter(n => goals?.targets[n] !== undefined || goals?.upper_limits?.[n] !== undefined)}
          {@const metGoals = availableNutrients.filter(n => {
            const current = getCurrentNutrient(n);
            
            // Check if this is an upper limit (should be minimized)
            if (goals?.upper_limits?.[n] !== undefined) {
              const limit = goals.upper_limits[n];
              if (limit === 0) {
                return current === 0; // Success if no consumption for zero limits
              }
              const progress = (current / limit) * 100;
              return progress <= 80; // Success if under 80% of the upper limit
            }
            
            // Regular target logic
            if (goals?.targets[n] === undefined) return false;
            const progress = (current / goals.targets[n]) * 100;
            return progress >= 80; // Success if reaching 80% or more of target
          }).length}
          
          <div class="stats stats-vertical lg:stats-horizontal bg-base-100 shadow-sm">
            <div class="stat">
              <div class="stat-title">Goals Met</div>
              <div class="stat-value text-lg">
                {metGoals}
                <span class="text-sm">/{availableNutrients.length}</span>
              </div>
              <div class="stat-desc">≥80% of target</div>
            </div>
            <div class="stat">
              <div class="stat-title">Source</div>
              <div class="stat-value text-lg">
                {goals.source === "custom" 
                  ? (goals.custom_name || "Custom") 
                  : "DRI"}
              </div>
              <div class="stat-desc">Nutrition guidelines</div>
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>


