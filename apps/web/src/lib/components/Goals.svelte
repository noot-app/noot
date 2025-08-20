<script lang="ts">
  import { onMount } from "svelte";
  import { apiClient } from "$lib/api/client";
  import type { paths } from "$lib/api/schema";

  type GoalsResponse = paths["/goals"]["get"]["responses"]["200"]["content"]["application/json"];
  type Goals = GoalsResponse["goals"];

  export let currentNutrition: Record<string, number> = {};
  export let showMealContribution = false; // New prop to indicate meal-specific view
  export let title = "Nutrition Goals"; // Customizable title

  let goals: Goals | null = null;
  let loading = true;
  let error = "";

  // Helper to safely get current nutrient value
  function getCurrentNutrient(nutrient: string): number {
    const value = currentNutrition?.[nutrient] || 0;
    return isFinite(value) ? value : 0;
  }

  onMount(async () => {
    try {
      const response = await apiClient.GET("/goals");
      
      if (response.error) {
        throw new Error(`API Error: ${response.error}`);
      }

      goals = response.data.goals;
    } catch (err) {
      error = `Failed to load nutrition goals: ${err}`;
      console.error("Goals error:", err);
    } finally {
      loading = false;
    }
  });

  function getProgress(nutrient: string, current: number): number {
    // Check if this is an upper limit (should be minimized)
    if (goals?.upper_limits?.[nutrient] !== undefined) {
      const limit = goals.upper_limits[nutrient];
      if (limit === 0) {
        // For zero limits (like trans fat), any amount is over
        return current > 0 ? 100 : 0;
      }
      // For upper limits, "progress" is how close to the limit (inverted logic)
      const progress = (current / limit) * 100;
      return isFinite(progress) ? Math.min(progress, 100) : 0;
    }
    
    // Regular target logic
    if (goals?.targets[nutrient] === undefined) return 0;
    const progress = (current / goals.targets[nutrient]) * 100;
    return isFinite(progress) ? Math.min(progress, 100) : 0;
  }

  function getActualProgress(nutrient: string, current: number): number {
    // Check if this is an upper limit
    if (goals?.upper_limits?.[nutrient] !== undefined) {
      const limit = goals.upper_limits[nutrient];
      if (limit === 0) {
        return current > 0 ? 200 : 0; // Show high percentage for any trans fat
      }
      const progress = (current / limit) * 100;
      return isFinite(progress) ? progress : 0;
    }
    
    // Regular target logic
    if (goals?.targets[nutrient] === undefined) return 0;
    const progress = (current / goals.targets[nutrient]) * 100;
    return isFinite(progress) ? progress : 0;
  }

  function getOverageText(nutrient: string, current: number): string {
    // Check if this is an upper limit
    if (goals?.upper_limits?.[nutrient] !== undefined) {
      const limit = goals.upper_limits[nutrient];
      if (limit === 0 && current > 0) {
        return "⚠️ Should be 0";
      }
      if (current > limit) {
        const overage = ((current / limit) - 1) * 100;
        return isFinite(overage) ? `⚠️ ${overage.toFixed(0)}% over limit` : "⚠️ Over limit";
      }
      return "";
    }
    
    // Regular target logic
    if (goals?.targets[nutrient] === undefined) return "";
    const actualProgress = (current / goals.targets[nutrient]) * 100;
    if (!isFinite(actualProgress) || actualProgress <= 100) return "";
    const overage = actualProgress - 100;
    return isFinite(overage) ? `+${overage.toFixed(0)}% over` : "+Over";
  }

  function isUpperLimit(nutrient: string): boolean {
    return goals?.upper_limits?.[nutrient] !== undefined;
  }

  function getGoalValue(nutrient: string): number {
    if (goals?.upper_limits?.[nutrient] !== undefined) {
      return goals.upper_limits[nutrient];
    }
    return goals?.targets?.[nutrient] || 0;
  }

  function formatNutrientName(key: string): string {
    return key
      .replace(/_/g, " ")
      .replace(/\b\w/g, l => l.toUpperCase())
      // Remove unit suffixes since they're shown separately
      .replace(/ Mcg$/, "")
      .replace(/ Mg$/, "")
      .replace(/ G$/, "");
  }

  function formatValue(value: number, unit: string): string {
    if (!isFinite(value) || value === 0) return "0";
    if (value < 1) return value.toFixed(1);
    return value.toFixed(0);
  }

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
    <h2 class="card-title flex items-center gap-2">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
      </svg>
      {title}
    </h2>

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
            <span class="badge badge-outline">
              {goals.life_stage.sex} • {goals.life_stage.age_bracket}
            </span>
            {#if goals.source === "custom"}
              <span class="badge badge-primary ml-2">
                {goals.custom_name || "Custom Goals"}
              </span>
            {:else}
              <div class="tooltip tooltip-bottom ml-2" data-tip="Dietary Reference Intakes (DRI) are nutrient reference values developed by health experts. Visit your profile to customize your nutrition goals.">
                <span class="badge badge-primary">
                  DRI Guidelines
                </span>
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
                <h4 class="font-semibold text-base mb-3 text-primary">Nutrition Targets</h4>
                <div class="space-y-3">
                  {#each targetNutrients as nutrient}
                    {@const current = getCurrentNutrient(nutrient)}
                    {@const target = goals.targets[nutrient]}
                    {@const progress = getProgress(nutrient, current)}
                    {@const unit = goals.units[nutrient] || ""}
                    
                    <div class="space-y-1">
                      <div class="flex justify-between items-center text-sm">
                        <span class="font-medium">
                          {formatNutrientName(nutrient)}
                          {#if getOverageText(nutrient, current)}
                            <span class="text-xs text-info ml-1">{getOverageText(nutrient, current)}</span>
                          {/if}
                        </span>
                        <span class="text-base-content/70">
                          {formatValue(current, unit)}/{formatValue(target, unit)} {unit}
                        </span>
                      </div>
                      <div class="flex items-center gap-2">
                        {#if showMealContribution}
                          <!-- Stacked progress bar showing meal contribution -->
                          <div class="flex-1 relative">
                            <progress 
                              class="progress progress-accent absolute inset-0"
                              value={progress} 
                              max="100"
                              title="This meal's contribution: {progress.toFixed(0)}%"
                            ></progress>
                          </div>
                        {:else}
                          <!-- Standard progress bar for targets -->
                          <progress 
                            class="progress flex-1"
                            class:progress-success={progress >= 80}
                            class:progress-warning={progress >= 50 && progress < 80}
                            class:progress-primary={progress < 50}
                            value={progress} 
                            max="100"
                          ></progress>
                        {/if}
                        <span class="text-xs text-base-content/60 min-w-[3rem]">
                          {getActualProgress(nutrient, current).toFixed(0)}%
                        </span>
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}

            <!-- Upper Limits (Minimize These) -->
            {@const limitNutrients = keyNutrients.filter(n => goals?.upper_limits?.[n] !== undefined)}
            {#if limitNutrients.length > 0}
              <div>
                <h4 class="font-semibold text-base mb-3 text-warning">Upper Limits (Minimize These)</h4>
                <p class="text-xs text-base-content/70 mb-3">These nutrients should be consumed as little as possible for optimal health.</p>
                <div class="space-y-3">
                  {#each limitNutrients as nutrient}
                    {@const current = getCurrentNutrient(nutrient)}
                    {@const limit = goals.upper_limits?.[nutrient] || 0}
                    {@const progress = getProgress(nutrient, current)}
                    {@const unit = goals.units[nutrient] || ""}
                    
                    <div class="space-y-1">
                      <div class="flex justify-between items-center text-sm">
                        <span class="font-medium">
                          {formatNutrientName(nutrient)}
                          <span class="badge badge-outline badge-info badge-xs ml-1">Limit</span>
                          {#if getOverageText(nutrient, current)}
                            <span class="text-xs text-warning ml-1">{getOverageText(nutrient, current)}</span>
                          {/if}
                        </span>
                        <span class="text-base-content/70">
                          {formatValue(current, unit)}/{formatValue(limit, unit)} {unit}
                        </span>
                      </div>
                      <div class="flex items-center gap-2">
                        {#if showMealContribution}
                          <!-- Stacked progress bar showing meal contribution -->
                          <div class="flex-1 relative">
                            <progress 
                              class="progress progress-warning absolute inset-0"
                              value={progress} 
                              max="100"
                              title="This meal's contribution: {progress.toFixed(0)}% of limit"
                            ></progress>
                          </div>
                        {:else}
                          <!-- Progress bar for limits (red = bad, green = good) -->
                          <progress 
                            class="progress flex-1"
                            class:progress-success={progress <= 50}
                            class:progress-warning={progress > 50 && progress <= 80}
                            class:progress-error={progress > 80}
                            value={progress} 
                            max="100"
                          ></progress>
                        {/if}
                        <span class="text-xs text-base-content/60 min-w-[3rem]">
                          {getActualProgress(nutrient, current).toFixed(0)}%
                        </span>
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
          {/if}
        </div>

        <!-- Summary Stats -->
        {#if goals}
          {@const availableNutrients = keyNutrients.filter(n => goals?.targets[n] !== undefined || goals?.upper_limits?.[n] !== undefined)}
          {@const metGoals = availableNutrients.filter(n => {
            const current = getCurrentNutrient(n);
            const progress = getProgress(n, current);
            // For upper limits, "success" means staying under the limit (low percentage)
            if (goals?.upper_limits?.[n] !== undefined) {
              return progress <= 80; // Success if under 80% of the upper limit
            }
            // For regular targets, success is reaching 80% or more
            return progress >= 80;
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

<style>
  progress.progress-success {
    --progress-color: oklch(var(--su));
  }
  progress.progress-warning {
    --progress-color: oklch(var(--wa));
  }
  progress.progress-primary {
    --progress-color: var(--color-dark);
  }
</style>
