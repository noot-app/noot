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
    if (!goals?.targets[nutrient]) return 0;
    return Math.min((current / goals.targets[nutrient]) * 100, 100);
  }

  function getActualProgress(nutrient: string, current: number): number {
    if (!goals?.targets[nutrient]) return 0;
    return (current / goals.targets[nutrient]) * 100;
  }

  function getOverageText(nutrient: string, current: number): string {
    if (!goals?.targets[nutrient]) return "";
    const actualProgress = (current / goals.targets[nutrient]) * 100;
    if (actualProgress <= 100) return "";
    const overage = actualProgress - 100;
    return `+${overage.toFixed(0)}% over`;
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
    if (value === 0) return "0";
    if (value < 1) return value.toFixed(1);
    return value.toFixed(0);
  }

  // Key nutrients to display - organized by importance
  const keyNutrients = [
    // Essential macronutrients
    "protein_g", "total_carbs_g", "dietary_fiber_g", "sodium_mg",
    
    // Important vitamins
    "vitamin_c_mg", "vitamin_d_mcg", "vitamin_a_mcg", "folate_mcg", "vitamin_b12_mcg",
    
    // Essential minerals
    "calcium_mg", "iron_mg", "magnesium_mg", "potassium_mg", "zinc_mg",
    
    // Additional important nutrients
    "vitamin_e_mg", "thiamine_mg", "riboflavin_mg", "niacin_mg"
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
        <div class="space-y-3">
          {#each keyNutrients as nutrient}
            {#if goals.targets[nutrient]}
              {@const current = currentNutrition[nutrient] || 0}
              {@const target = goals.targets[nutrient]}
              {@const progress = getProgress(nutrient, current)}
              {@const unit = goals.units[nutrient] || ""}
              
              <div class="space-y-1">
                <div class="flex justify-between items-center text-sm">
                  <span class="font-medium">
                    {formatNutrientName(nutrient)}
                    {#if getOverageText(nutrient, current)}
                      <span class="text-xs text-info ml-1">({getOverageText(nutrient, current)})</span>
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
                    <!-- Standard progress bar -->
                    <progress 
                      class="progress flex-1"
                      class:progress-primary={progress < 50}
                      class:progress-warning={progress >= 50 && progress < 80}
                      class:progress-success={progress >= 80}
                      value={progress} 
                      max="100"
                    ></progress>
                  {/if}
                  <span class="text-xs text-base-content/60 min-w-[3rem]">
                    {getActualProgress(nutrient, current).toFixed(0)}%
                  </span>
                </div>
              </div>
            {/if}
          {/each}
        </div>

        <!-- Summary Stats -->
        {#if goals}
          {@const availableNutrients = keyNutrients.filter(n => goals?.targets[n])}
          {@const metGoals = availableNutrients.filter(n => goals?.targets[n] && getProgress(n, currentNutrition[n] || 0) >= 80).length}
          
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
    --progress-color: oklch(var(--er));
  }
</style>
