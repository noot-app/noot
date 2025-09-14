<script lang="ts">
  import { onMount } from "svelte"
  import type { paths } from "$lib/api/schema"
  import { isRestrictedNutrient } from "$lib/utils/nutrients"
  import {
    getNutrientValue,
    formatValue,
    getProgress,
    getActualProgress,
    getProgressBarClass,
    getDailyText,
    getOverageText
  } from "$lib/utils/nutrition-display"
  import Scale from "$lib/components/icons/Scale.svelte"

  import Blocks from "$lib/components/icons/Blocks.svelte"
  import Pyramid from "$lib/components/icons/Pyramid.svelte"
  import Shapes from "$lib/components/icons/Shapes.svelte"
  import Boxes from "$lib/components/icons/Boxes.svelte"

  type GoalsResponse =
    paths["/goals"]["get"]["responses"]["200"]["content"]["application/json"]
  type Goals = GoalsResponse["goals"]

  // Props for configuring the display
  export let nutrients: Record<string, number> = {}
  export let showProgress = false // Whether to show progress bars
  export let showGoals = false // Whether to show progress against goals
  export let isExpandable = false // Whether the component is collapsible
  export let title = "Nutrient Profile"
  export let className = ""
  export let showMealContribution = false // New prop for meal-specific view (different progress bar styling)
  export let showLimitsOnly = false // Only show nutrients that have upper limits
  export let showAllCategories = false // Force show all categories regardless of values (for summary views)
  export let goalsData: Goals | null = null // Injected goals to avoid duplicate fetches
  export let disabledNutrients: string[] = [] // List of disabled nutrients to filter out

  let goals: Goals | null = null
  let loading = false
  let error = ""
  let isExpanded = !isExpandable // If not expandable, always show content

  // Nutrient categories with icons - standardized categorization
  const nutrientCategories = {
    macronutrients: {
      title: "Macronutrients",
      icon: Shapes,
      nutrients: [
        { key: "calories", label: "Calories", unit: "" },
        { key: "protein_g", label: "Protein", unit: "g" },
        { key: "total_carbs_g", label: "Total Carbohydrates", unit: "g" },
        { key: "dietary_fiber_g", label: "Dietary Fiber", unit: "g" },
        { key: "added_sugars_g", label: "Added Sugars", unit: "g" },
        { key: "total_fat_g", label: "Total Fat", unit: "g" },
        { key: "saturated_fat_g", label: "Saturated Fat", unit: "g" },
        {
          key: "monounsaturated_fat_g",
          label: "Monounsaturated Fat",
          unit: "g",
        },
        {
          key: "polyunsaturated_fat_g",
          label: "Polyunsaturated Fat",
          unit: "g",
        },
        { key: "trans_fat_g", label: "Trans Fat", unit: "g" },
        { key: "cholesterol_mg", label: "Cholesterol", unit: "mg" },
      ],
    },
    vitamins: {
      title: "Vitamins",
      icon: Boxes,
      nutrients: [
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
        {
          key: "pantothenic_acid_mg",
          label: "Pantothenic Acid (B5)",
          unit: "mg",
        },
        { key: "choline_mg", label: "Choline", unit: "mg" },
      ],
    },
    minerals: {
      title: "Minerals",
      icon: Scale,
      nutrients: [
        { key: "calcium_mg", label: "Calcium", unit: "mg" },
        { key: "iron_mg", label: "Iron", unit: "mg" },
        { key: "magnesium_mg", label: "Magnesium", unit: "mg" },
        { key: "phosphorus_mg", label: "Phosphorus", unit: "mg" },
        { key: "potassium_mg", label: "Potassium", unit: "mg" },
        { key: "sodium_mg", label: "Sodium", unit: "mg" },
        { key: "zinc_mg", label: "Zinc", unit: "mg" },
        { key: "copper_mg", label: "Copper", unit: "mg" },
        { key: "manganese_mg", label: "Manganese", unit: "mg" },
        { key: "selenium_mcg", label: "Selenium", unit: "mcg" },
        { key: "iodine_mcg", label: "Iodine", unit: "mcg" },
        { key: "molybdenum_mcg", label: "Molybdenum", unit: "mcg" },
        { key: "chromium_mcg", label: "Chromium", unit: "mcg" },
        { key: "fluoride_mg", label: "Fluoride", unit: "mg" },
        { key: "chloride_mg", label: "Chloride", unit: "mg" },
      ],
    },
    emerging_nutrients: {
      title: 'Emerging Nutrients',
      icon: Pyramid,
      nutrients: [
        { key: "omega3_ala_g", label: "Omega-3 ALA", unit: "g" },
        { key: "omega3_epa_g", label: "Omega-3 EPA", unit: "g" },
        { key: "omega3_dha_g", label: "Omega-3 DHA", unit: "g" },
        { key: "omega6_g", label: "Omega-6 Fatty Acids", unit: "g" },
      ],
    },
    functional_compounds: {
      title: "Functional Compounds",
      icon: Blocks,
      nutrients: [
        { key: "alcohol_g", label: "Alcohol", unit: "g" },
        { key: "caffeine_mg", label: "Caffeine", unit: "mg" },
        { key: "creatine_mg", label: "Creatine", unit: "mg" },
      ],
    },
  }

  // Sync provided goals
  $: goals = goalsData || goals

  function handleToggle() {
    if (!isExpandable) return
    isExpanded = !isExpanded
  }

  onMount(() => {
    // No-op; goals are provided by parent to avoid duplicate requests
  })

  function hasNutrientData(categoryNutrients: any[]): boolean {
    // If showAllCategories is true, always show all categories
    if (showAllCategories) {
      return true
    }

    return categoryNutrients.some((nutrient) => {
      // Skip disabled nutrients
      if (disabledNutrients.includes(nutrient.key)) {
        return false
      }

      const value = getNutrientValue(nutrient.key, nutrients)

      // For summary page (!showMealContribution), always show categories with goals/targets
      // even if all nutrients have zero values
      if (!showMealContribution && goals) {
        const hasTarget = goals?.targets?.[nutrient.key] !== undefined
        const hasLimit = goals?.upper_limits?.[nutrient.key] !== undefined

        if (showLimitsOnly) {
          return hasLimit
        }

        // For regular categorized view, exclude only restricted nutrients
        if (hasLimit && isRestrictedNutrient(nutrient.key)) {
          return false
        }

        return hasTarget || hasLimit
      }

      // For meal contribution view, only show nutrients with values > 0
      if (value <= 0) return false

      // If showLimitsOnly is true, only show the restricted nutrients that have upper limits
      if (showLimitsOnly) {
        return goals?.upper_limits?.[nutrient.key] !== undefined && 
               isRestrictedNutrient(nutrient.key)
      }

      // For regular categorized view, exclude only the restricted nutrients
      if (goals?.upper_limits?.[nutrient.key] !== undefined && 
          isRestrictedNutrient(nutrient.key)) {
        return false
      }

      // Otherwise, show nutrients with values
      return true
    })
  }

  // Using imported getOverageText function

  // Get nutrients that have upper limits and are restricted nutrients
  function getNutrientsWithLimits() {
    if (!goals) return []
    const allNutrients = Object.values(nutrientCategories).flatMap(
      (category) => category.nutrients,
    )
    return allNutrients.filter(
      (nutrient) =>
        goals?.upper_limits?.[nutrient.key] !== undefined &&
        isRestrictedNutrient(nutrient.key) &&
        !disabledNutrients.includes(nutrient.key) &&
        (getNutrientValue(nutrient.key, nutrients) > 0 || !showMealContribution),
    )
  }
</script>

<div class="nutrient-category-display {className}">
  {#if isExpandable}
    <div class="border-t border-base-300 pt-3 mt-3">
      <button
        class="flex items-center justify-between w-full text-left hover:bg-base-200 p-2 rounded transition-colors"
        on:click={handleToggle}
      >
        <span class="text-sm font-medium text-primary">{title}</span>
        <svg
          class="w-4 h-4 transform transition-transform {isExpanded
            ? 'rotate-180'
            : ''}"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </button>
    </div>
  {:else if title}
    <h3 class="text-lg font-semibold mb-4">{title}</h3>
  {/if}

  {#if isExpanded}
    <div
      class="nutrient-content space-y-6 {isExpandable
        ? 'mt-4 animate-fade-in'
        : ''}"
    >
      {#if showGoals && loading}
        <div class="text-center py-4">
          <span class="loading loading-spinner loading-sm"></span>
          <p class="text-xs text-base-content/70 mt-2">
            Loading nutrition goals...
          </p>
        </div>
      {:else if showGoals && error}
        <div class="alert alert-error alert-sm">
          <svg
            class="w-4 h-4"
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
          <span class="text-xs">{error}</span>
        </div>
      {:else}
        {#if showLimitsOnly}
          <!-- Flat list for upper limits (no categories) -->
          <div class="space-y-2">
            {#each getNutrientsWithLimits() as nutrient}
              {@const value = getNutrientValue(nutrient.key, nutrients)}
              {@const progress = getProgress(nutrient.key, goals, nutrients, showLimitsOnly)}
              {@const dailyText = getDailyText(nutrient.key, goals)}
              <div class="flex justify-between items-center text-sm">
                <div class="flex-1">
                  <div class="flex justify-between items-center mb-1">
                    <span class="font-medium">
                      {nutrient.label}
                      <span class="badge badge-outline badge-info badge-xs ml-1"
                        >Limit</span
                      >
                      {#if showGoals && getOverageText(nutrient.key, value, goals)}
                        <span class="text-xs text-warning ml-1"
                          >{getOverageText(nutrient.key, value, goals)}</span
                        >
                      {/if}
                    </span>
                    <span class="text-base-content/70">
                      {formatValue(value)}{nutrient.unit}
                      {#if showGoals && dailyText}
                        <span class="text-xs">/ {dailyText}</span>
                      {/if}
                    </span>
                  </div>
                  {#if showProgress && showGoals && goals && (progress > 0 || !showMealContribution)}
                    <div class="flex items-center gap-2">
                      {#if showMealContribution}
                        <!-- Stacked progress bar showing meal contribution -->
                        <div class="flex-1 relative">
                          <progress
                            class="progress progress-sm absolute inset-0 {getProgressBarClass(
                              nutrient.key,
                              progress,
                              goals,
                              showMealContribution,
                              showLimitsOnly,
                              nutrients
                            )}"
                            value={progress}
                            max="100"
                            title="This meal's contribution: {progress.toFixed(
                              0,
                            )}% of limit"
                          ></progress>
                        </div>
                      {:else}
                        <!-- Standard progress bar -->
                        <progress
                          class="progress progress-sm flex-1 {getProgressBarClass(
                            nutrient.key,
                            progress,
                            goals,
                            showMealContribution,
                            showLimitsOnly,
                            nutrients
                          )}"
                          value={Math.min(progress, 100)}
                          max="100"
                        ></progress>
                      {/if}
                      <span class="text-xs text-base-content/60 min-w-[3rem]">
                        {getActualProgress(nutrient.key, goals, nutrients).toFixed(0)}%
                      </span>
                    </div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {:else}
          <!-- Categorized view for regular nutrients -->
          {#each Object.entries(nutrientCategories) as [, category]}
            {#if hasNutrientData(category.nutrients)}
              <div>
                <h4
                  class="font-semibold text-base mb-3 pb-2 border-b border-base-300 flex items-center nutrient-header"
                >
                  {#if typeof category.icon === 'string'}
                    <svg
                      class="w-4 h-4 mr-2"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d={category.icon}
                      />
                    </svg>
                  {:else}
                    <svelte:component this={category.icon} className="w-4 h-4 mr-2" strokeWidth={2} />
                  {/if}
                  {category.title}
                </h4>
                <div class="space-y-2">
                  {#each category.nutrients as nutrient}
                    {@const value = getNutrientValue(nutrient.key, nutrients)}
                    {@const isRestricted = isRestrictedNutrient(nutrient.key)}
                    {@const hasUpperLimit = goals?.upper_limits?.[nutrient.key] !== undefined}
                    {@const hasTarget = goals?.targets?.[nutrient.key] !== undefined}
                    {@const isDisabled = disabledNutrients.includes(nutrient.key)}
                    {#if !isDisabled && (value > 0 || (!showMealContribution && goals && (hasTarget || (showLimitsOnly && hasUpperLimit) || (hasUpperLimit && !isRestricted))))}
                      {#if showLimitsOnly ? hasUpperLimit : !(hasUpperLimit && isRestricted)}
                        {@const progress = getProgress(nutrient.key, goals, nutrients, showLimitsOnly)}
                        {@const dailyText = getDailyText(nutrient.key, goals)}
                        <div class="flex justify-between items-center text-sm">
                          <div class="flex-1">
                            <div class="flex justify-between items-center mb-1">
                              <span class="font-medium">
                                {nutrient.label}
                                {#if showLimitsOnly}
                                  <span
                                    class="badge badge-outline badge-info badge-xs ml-1"
                                    >Limit</span
                                  >
                                {/if}
                                {#if showGoals && getOverageText(nutrient.key, value, goals)}
                                  <span class="text-xs text-info ml-1"
                                    >{getOverageText(nutrient.key, value, goals)}</span
                                  >
                                {/if}
                              </span>
                              <span class="text-base-content/70">
                                {formatValue(value)}{nutrient.unit}
                                {#if showGoals && dailyText}
                                  <span class="text-xs">/ {dailyText}</span>
                                {/if}
                              </span>
                            </div>
                            {#if showProgress && showGoals && goals && (progress > 0 || !showMealContribution)}
                              <div class="flex items-center gap-2">
                                {#if showMealContribution}
                                  <!-- Stacked progress bar showing meal contribution -->
                                  <div class="flex-1 relative">
                                    <progress
                                      class="progress progress-sm absolute inset-0 {getProgressBarClass(
                                        nutrient.key,
                                        progress,
                                        goals,
                                        showMealContribution,
                                        showLimitsOnly,
                                        nutrients
                                      )}"
                                      value={progress}
                                      max="100"
                                      title="This meal's contribution: {progress.toFixed(
                                        0,
                                      )}%"
                                    ></progress>
                                  </div>
                                {:else}
                                  <!-- Standard progress bar -->
                                  <progress
                                    class="progress progress-sm flex-1 {getProgressBarClass(
                                      nutrient.key,
                                      progress,
                                      goals,
                                      showMealContribution,
                                      showLimitsOnly,
                                      nutrients
                                    )}"
                                    value={Math.min(progress, 100)}
                                    max="100"
                                  ></progress>
                                {/if}
                                <span
                                  class="text-xs text-base-content/60 min-w-[3rem]"
                                >
                                  {getActualProgress(nutrient.key, goals, nutrients).toFixed(0)}%
                                </span>
                              </div>
                            {/if}
                          </div>
                        </div>
                      {/if}
                    {/if}
                  {/each}
                </div>
              </div>
            {/if}
          {/each}


        {/if}

        {#if showGoals && !goals}
          <div class="text-center py-2">
            <p class="text-xs text-base-content/60">
              Daily percentages are calculated based on Dietary Reference
              Intakes (DRI)
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

  :global(.progress-neutral) {
    --progress-color: oklch(var(--n));
  }

  .nutrient-header {
    color: var(--color-base-content);
  }
</style>
