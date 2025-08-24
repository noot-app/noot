<script lang="ts">
  import { onMount } from "svelte";
  import { apiClient } from "$lib/api/client";
  import type { components, paths } from "$lib/api/schema";

  type GoalsResponse = paths["/goals"]["get"]["responses"]["200"]["content"]["application/json"];
  type Goals = GoalsResponse["goals"];
  type CompleteNutrient = components["schemas"]["CompleteNutrient"];

  export let nutrients: CompleteNutrient;
  export let foodName: string = "";

  let goals: Goals | null = null;
  let isExpanded = false;
  let loading = false;
  let error = "";

  // Nutrient categories - organized for clear display
  const nutrientCategories = {
    macronutrients: [
      { key: "calories", label: "Calories", unit: "" },
      { key: "protein_g", label: "Protein", unit: "g" },
      { key: "total_carbs_g", label: "Total Carbohydrates", unit: "g" },
      { key: "dietary_fiber_g", label: "Dietary Fiber", unit: "g" },
      { key: "total_sugars_g", label: "Total Sugars", unit: "g" },
      { key: "added_sugars_g", label: "Added Sugars", unit: "g" },
      { key: "total_fat_g", label: "Total Fat", unit: "g" },
      { key: "saturated_fat_g", label: "Saturated Fat", unit: "g" },
      { key: "trans_fat_g", label: "Trans Fat", unit: "g" },
      { key: "cholesterol_mg", label: "Cholesterol", unit: "mg" },
      { key: "sodium_mg", label: "Sodium", unit: "mg" }
    ],
    vitamins: [
      { key: "vitamin_a_mcg", label: "Vitamin A", unit: "mcg" },
      { key: "vitamin_c_mg", label: "Vitamin C", unit: "mg" },
      { key: "vitamin_d_mcg", label: "Vitamin D", unit: "mcg" },
      { key: "vitamin_e_mg", label: "Vitamin E", unit: "mg" },
      { key: "vitamin_k_mcg", label: "Vitamin K", unit: "mcg" },
      { key: "thiamine_mg", label: "Thiamine (B1)", unit: "mg" },
      { key: "riboflavin_mg", label: "Riboflavin (B2)", unit: "mg" },
      { key: "niacin_mg", label: "Niacin (B3)", unit: "mg" },
      { key: "vitamin_b6_mg", label: "Vitamin B6", unit: "mg" },
      { key: "folate_mcg", label: "Folate", unit: "mcg" },
      { key: "vitamin_b12_mcg", label: "Vitamin B12", unit: "mcg" },
      { key: "biotin_mcg", label: "Biotin", unit: "mcg" },
      { key: "pantothenic_acid_mg", label: "Pantothenic Acid (B5)", unit: "mg" },
      { key: "choline_mg", label: "Choline", unit: "mg" }
    ],
    minerals: [
      { key: "calcium_mg", label: "Calcium", unit: "mg" },
      { key: "iron_mg", label: "Iron", unit: "mg" },
      { key: "magnesium_mg", label: "Magnesium", unit: "mg" },
      { key: "phosphorus_mg", label: "Phosphorus", unit: "mg" },
      { key: "potassium_mg", label: "Potassium", unit: "mg" },
      { key: "zinc_mg", label: "Zinc", unit: "mg" },
      { key: "copper_mg", label: "Copper", unit: "mg" },
      { key: "manganese_mg", label: "Manganese", unit: "mg" },
      { key: "selenium_mcg", label: "Selenium", unit: "mcg" },
      { key: "iodine_mcg", label: "Iodine", unit: "mcg" },
      { key: "molybdenum_mcg", label: "Molybdenum", unit: "mcg" },
      { key: "chromium_mcg", label: "Chromium", unit: "mcg" },
      { key: "fluoride_mg", label: "Fluoride", unit: "mg" },
      { key: "chloride_mg", label: "Chloride", unit: "mg" }
    ]
  };

  async function loadGoals() {
    if (goals) return; // Already loaded

    try {
      loading = true;
      error = "";
      
      const response = await apiClient.GET("/goals");
      
      if (response.error) {
        throw new Error(`API Error: ${response.error}`);
      }
      
      goals = response.data.goals;
    } catch (err) {
      error = `Error loading goals: ${err}`;
      console.error("Goals error:", err);
    } finally {
      loading = false;
    }
  }

  function handleToggle() {
    isExpanded = !isExpanded;
    if (isExpanded && !goals) {
      loadGoals();
    }
  }

  function getNutrientValue(key: string): number {
    const value = (nutrients as any)[key];
    return (typeof value === 'number' && isFinite(value)) ? value : 0;
  }

  function getProgress(key: string): number {
    if (!goals) return 0;
    
    const current = getNutrientValue(key);
    const target = goals.targets?.[key] || goals.upper_limits?.[key];
    
    if (!target || target === 0) return 0;
    
    const progress = (current / target) * 100;
    return isFinite(progress) ? Math.min(progress, 200) : 0; // Cap at 200% for display
  }

  function formatValue(value: number): string {
    if (!isFinite(value) || value === 0) return "0";
    if (value < 1) return value.toFixed(1);
    return value.toFixed(0);
  }

  function getProgressBarClass(key: string, progress: number): string {
    const isLimit = goals?.upper_limits?.[key] !== undefined;
    
    if (isLimit) {
      // For limits (minimize these)
      if (progress <= 50) return "progress-success";
      if (progress <= 100) return "progress-warning";
      return "progress-error";
    } else {
      // For targets (aim to meet these)
      if (progress >= 80) return "progress-success";
      if (progress >= 50) return "progress-warning";
      return "progress-primary";
    }
  }

  function getDailyText(key: string): string {
    if (!goals) return "";
    
    const target = goals.targets?.[key];
    const limit = goals.upper_limits?.[key];
    
    if (target) {
      return `${formatValue(target)} ${goals.units[key] || ""}`;
    } else if (limit) {
      return `${formatValue(limit)} ${goals.units[key] || ""} limit`;
    }
    
    return "";
  }

  function hasNutrientData(category: any[]): boolean {
    return category.some(nutrient => getNutrientValue(nutrient.key) > 0);
  }
</script>

<div class="border-t border-base-300 pt-3 mt-3">
  <button 
    class="flex items-center justify-between w-full text-left hover:bg-base-200 p-2 rounded transition-colors"
    on:click={handleToggle}
  >
    <span class="text-sm font-medium text-primary">
      View Full Nutrient Profile
      {#if foodName}
        for {foodName}
      {/if}
    </span>
    <svg 
      class="w-4 h-4 transform transition-transform {isExpanded ? 'rotate-180' : ''}" 
      fill="none" 
      stroke="currentColor" 
      viewBox="0 0 24 24"
    >
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
    </svg>
  </button>

  {#if isExpanded}
    <div class="mt-4 space-y-6 animate-fade-in">
      {#if loading}
        <div class="text-center py-4">
          <span class="loading loading-spinner loading-sm"></span>
          <p class="text-xs text-base-content/70 mt-2">Loading nutrition goals...</p>
        </div>
      {:else if error}
        <div class="alert alert-error alert-sm">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-xs">{error}</span>
        </div>
      {:else}
        <!-- Macronutrients -->
        {#if hasNutrientData(nutrientCategories.macronutrients)}
          <div>
            <h4 class="font-semibold text-sm mb-3 text-accent flex items-center">
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
              Macronutrients
            </h4>
            <div class="space-y-2">
              {#each nutrientCategories.macronutrients as nutrient}
                {@const value = getNutrientValue(nutrient.key)}
                {#if value > 0}
                  {@const progress = getProgress(nutrient.key)}
                  {@const dailyText = getDailyText(nutrient.key)}
                  <div class="flex justify-between items-center text-sm">
                    <div class="flex-1">
                      <div class="flex justify-between items-center mb-1">
                        <span class="font-medium">{nutrient.label}</span>
                        <span class="text-base-content/70">
                          {formatValue(value)}{nutrient.unit}
                          {#if dailyText}
                            <span class="text-xs">/ {dailyText}</span>
                          {/if}
                        </span>
                      </div>
                      {#if goals && progress > 0}
                        <div class="flex items-center gap-2">
                          <progress 
                            class="progress progress-sm flex-1 {getProgressBarClass(nutrient.key, progress)}"
                            value={progress} 
                            max="100"
                          ></progress>
                          <span class="text-xs text-base-content/60 min-w-[3rem]">
                            {progress.toFixed(0)}%
                          </span>
                        </div>
                      {/if}
                    </div>
                  </div>
                {/if}
              {/each}
            </div>
          </div>
        {/if}

        <!-- Vitamins -->
        {#if hasNutrientData(nutrientCategories.vitamins)}
          <div>
            <h4 class="font-semibold text-sm mb-3 text-success flex items-center">
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              Vitamins
            </h4>
            <div class="space-y-2">
              {#each nutrientCategories.vitamins as nutrient}
                {@const value = getNutrientValue(nutrient.key)}
                {#if value > 0}
                  {@const progress = getProgress(nutrient.key)}
                  {@const dailyText = getDailyText(nutrient.key)}
                  <div class="flex justify-between items-center text-sm">
                    <div class="flex-1">
                      <div class="flex justify-between items-center mb-1">
                        <span class="font-medium">{nutrient.label}</span>
                        <span class="text-base-content/70">
                          {formatValue(value)}{nutrient.unit}
                          {#if dailyText}
                            <span class="text-xs">/ {dailyText}</span>
                          {/if}
                        </span>
                      </div>
                      {#if goals && progress > 0}
                        <div class="flex items-center gap-2">
                          <progress 
                            class="progress progress-sm flex-1 {getProgressBarClass(nutrient.key, progress)}"
                            value={progress} 
                            max="100"
                          ></progress>
                          <span class="text-xs text-base-content/60 min-w-[3rem]">
                            {progress.toFixed(0)}%
                          </span>
                        </div>
                      {/if}
                    </div>
                  </div>
                {/if}
              {/each}
            </div>
          </div>
        {/if}

        <!-- Minerals -->
        {#if hasNutrientData(nutrientCategories.minerals)}
          <div>
            <h4 class="font-semibold text-sm mb-3 text-warning flex items-center">
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 6l3 1m0 0l-3 9a5.002 5.002 0 006.001 0M6 7l3 9M6 7l6-2m6 2l3-1m-3 1l-3 9a5.002 5.002 0 006.001 0M18 7l3 9m-3-9l-6-2m0-2v2m0 16V5m0 16H9m3 0h3" />
              </svg>
              Minerals
            </h4>
            <div class="space-y-2">
              {#each nutrientCategories.minerals as nutrient}
                {@const value = getNutrientValue(nutrient.key)}
                {#if value > 0}
                  {@const progress = getProgress(nutrient.key)}
                  {@const dailyText = getDailyText(nutrient.key)}
                  <div class="flex justify-between items-center text-sm">
                    <div class="flex-1">
                      <div class="flex justify-between items-center mb-1">
                        <span class="font-medium">{nutrient.label}</span>
                        <span class="text-base-content/70">
                          {formatValue(value)}{nutrient.unit}
                          {#if dailyText}
                            <span class="text-xs">/ {dailyText}</span>
                          {/if}
                        </span>
                      </div>
                      {#if goals && progress > 0}
                        <div class="flex items-center gap-2">
                          <progress 
                            class="progress progress-sm flex-1 {getProgressBarClass(nutrient.key, progress)}"
                            value={progress} 
                            max="100"
                          ></progress>
                          <span class="text-xs text-base-content/60 min-w-[3rem]">
                            {progress.toFixed(0)}%
                          </span>
                        </div>
                      {/if}
                    </div>
                  </div>
                {/if}
              {/each}
            </div>
          </div>
        {/if}

        {#if !goals}
          <div class="text-center py-2">
            <p class="text-xs text-base-content/60">
              Daily percentages are calculated based on Dietary Reference Intakes (DRI)
            </p>
          </div>
        {/if}
      {/if}
    </div>
  {/if}
</div>

<style>
  .animate-fade-in {
    animation: fadeInSlide 0.3s ease-out;
  }

  @keyframes fadeInSlide {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>