<script lang="ts">
  import { onMount } from "svelte"
  import { apiClient } from "$lib/api/client"
  import NutrientCategoryDisplay from "./NutrientCategoryDisplay.svelte"
  import InfoButton from "./InfoButton.svelte"
  import type { paths } from "$lib/api/schema"

  type GoalsResponse =
    paths["/goals"]["get"]["responses"]["200"]["content"]["application/json"]
  type Goals = GoalsResponse["goals"]

  export let currentNutrition: Record<string, number> = {}
  export let showMealContribution = false // New prop to indicate meal-specific view
  export let isSharedView = false // New prop for shareable link context (non-logged-in users)
  export let title = "Nutrition Goals" // Customizable title
  // Optional: preloaded goals for 'auto' and 'dri' sources; if provided, component will use them
  export let preloadGoalsAuto: Goals | null = null
  export let preloadGoalsDri: Goals | null = null
  // Optional: force source selection ('auto' uses user's active goal set; 'dri' forces DRI)
  export let source: "auto" | "dri" | null = null

  let goals: Goals | null = null
  let loading = true
  let error = ""

  // Compute dynamic title based on context
  $: dynamicTitle = (() => {
    if (showMealContribution) {
      if (isSharedView || goals?.source === "dri") {
        return "How this meal contributes to DRI Nutrition Targets"
      }
      return "How this meal contributes to your daily goals"
    }
    return title
  })()

  // Helper to safely get current nutrient value
  function getCurrentNutrient(nutrient: string): number {
    const value = currentNutrition?.[nutrient] || 0
    return isFinite(value) ? value : 0
  }

  // Simple in-instance memo to prevent duplicate network requests
  let goalsLoaded = false

  async function loadGoals() {
    if (goalsLoaded) return
    try {
      loading = true
      error = ""

      // If goals were preloaded, honor props and avoid network
      if (preloadGoalsAuto || preloadGoalsDri) {
        if (source === "dri" || isSharedView) {
          goals = preloadGoalsDri ?? preloadGoalsAuto
        } else {
          goals = preloadGoalsAuto ?? preloadGoalsDri
        }
      } else {
        // Fallback: fetch based on desired source
        const desired = source ?? (isSharedView ? "dri" : "auto")
        const response = await apiClient.GET("/goals", {
          params: { query: { source: desired } },
        })
        if (response.error) {
          throw new Error(`API Error: ${response.error}`)
        }
        goals = response.data.goals
      }
      goalsLoaded = true
    } catch (err) {
      error = `Failed to load nutrition goals: ${err}`
      console.error("Goals error:", err)
    } finally {
      loading = false
    }
  }

  onMount(async () => {
    await loadGoals()
  })

  // All nutrients to display - organized by category for complete DRI coverage
  const keyNutrients = [
    // Essential macronutrients
    "calories",
    "protein_g",
    "total_carbs_g",
    "total_fat_g",
    "saturated_fat_g",
    "trans_fat_g",
    "monounsaturated_fat_g",
    "polyunsaturated_fat_g",
    "omega3_ala_g",
    "omega3_epa_g",
    "omega3_dha_g",
    "omega6_g",
    "dietary_fiber_g",
    "total_sugars_g",
    "added_sugars_g",
    "cholesterol_mg",
    "sodium_mg",
    "alcohol_g",

    // B-Complex vitamins
    "thiamine_mg",
    "riboflavin_mg",
    "niacin_mg",
    "vitamin_b6_mg",
    "folate_mcg",
    "vitamin_b12_mcg",
    "biotin_mcg",
    "pantothenic_acid_mg",

    // Fat-soluble vitamins
    "vitamin_a_mcg",
    "vitamin_d_mcg",
    "vitamin_e_mg",
    "vitamin_k_mcg",

    // Water-soluble vitamins
    "vitamin_c_mg",
    "choline_mg",

    // Essential minerals
    "calcium_mg",
    "iron_mg",
    "magnesium_mg",
    "phosphorus_mg",
    "potassium_mg",
    "zinc_mg",
    "copper_mg",
    "manganese_mg",
    "selenium_mcg",
    "iodine_mcg",
    "molybdenum_mcg",
    "chromium_mcg",
    "fluoride_mg",
    "chloride_mg",

    // Other compounds
    "caffeine_mg",
    "creatine_mg",
  ]
</script>

<div class="card bg-base-200 shadow-lg">
  <div class="card-body">
    <h2 class="card-title text-lg flex items-center gap-2 goals-header">
      <svg
        class="w-5 h-5"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M13 10V3L4 14h7v7l9-11h-7z"
        />
      </svg>
      Nutrition Targets
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
        <svg
          class="w-6 h-6 shrink-0"
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
        <span class="text-sm">{error}</span>
      </div>
    {:else if goals}
      <div class="space-y-3">
        <!-- Goals Header -->
        <div class="flex justify-between items-center">
          <!-- Removed badge section - info will be in stats section instead -->
        </div>

        <!-- Key Nutrients Progress -->
        <div class="space-y-6">
          {#if goals}
            <!-- Regular Nutrition Targets -->
            {@const targetNutrients = keyNutrients.filter(
              (n) => goals?.targets[n] !== undefined,
            )}
            {#if targetNutrients.length > 0}
              <div>
                <NutrientCategoryDisplay
                  nutrients={currentNutrition}
                  showProgress={true}
                  showGoals={true}
                  isExpandable={false}
                  title=""
                  {showMealContribution}
                  goalsData={goals}
                />
              </div>
            {/if}

            <!-- Upper Limits (Minimize These) -->
            {@const limitNutrients = keyNutrients.filter(
              (n) =>
                goals?.upper_limits?.[n] !== undefined &&
                (getCurrentNutrient(n) > 0 || !showMealContribution),
            )}
            {#if limitNutrients.length > 0}
              <div>
                <h4 class="font-semibold text-base mb-3 text-warning">
                  Upper Limits (Minimize These)
                </h4>
                <p class="text-xs text-base-content/70 mb-3">
                  These nutrients should be consumed as little as possible for
                  optimal health.
                </p>
                <NutrientCategoryDisplay
                  nutrients={currentNutrition}
                  showProgress={true}
                  showGoals={true}
                  isExpandable={false}
                  title=""
                  {showMealContribution}
                  showLimitsOnly={true}
                  goalsData={goals}
                />
              </div>
            {/if}
          {/if}
        </div>

        <!-- Summary Stats -->
        {#if goals}
          {@const availableNutrients = keyNutrients.filter(
            (n) =>
              goals?.targets[n] !== undefined ||
              goals?.upper_limits?.[n] !== undefined,
          )}
          {@const metGoals = availableNutrients.filter((n) => {
            const current = getCurrentNutrient(n)

            // Check if this is an upper limit (should be minimized)
            if (goals?.upper_limits?.[n] !== undefined) {
              const limit = goals.upper_limits[n]
              if (limit === 0) {
                return current === 0 // Success if no consumption for zero limits
              }
              const progress = (current / limit) * 100
              return progress <= 80 // Success if under 80% of the upper limit
            }

            // Regular target logic
            if (goals?.targets[n] === undefined) return false
            const progress = (current / goals.targets[n]) * 100
            return progress >= 80 // Success if reaching 80% or more of target
          }).length}

          <!-- Sugar Breakdown Section -->
          {@const totalSugars = getCurrentNutrient("total_sugars_g")}
          {@const addedSugars = getCurrentNutrient("added_sugars_g")}
          {@const naturalSugars = Math.max(0, totalSugars - addedSugars)}
          {#if totalSugars > 0}
            <div class="card bg-base-100 shadow-sm mt-6">
              <div class="card-body p-4">
                <div class="flex items-center justify-between mb-3">
                  <h4
                    class="font-semibold text-base flex items-center goals-subheader"
                  >
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
                        d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                      />
                    </svg>
                    Sugar Breakdown
                  </h4>
                  <!-- Info button that opens modal -->
                  <InfoButton modalId="sugar-info-modal" size="sm" />
                </div>

                <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 text-sm">
                  <!-- Total Sugars -->
                  <div class="text-center">
                    <div class="stat-value text-xl text-base-content">
                      {totalSugars.toFixed(1)}g
                    </div>
                    <div class="stat-title">Total Sugars</div>
                  </div>

                  <!-- Natural Sugars -->
                  <div class="text-center">
                    <div class="stat-value text-xl text-success">
                      {naturalSugars.toFixed(1)}g
                    </div>
                    <div class="stat-title">Natural Sugars</div>
                    <div class="stat-desc text-xs text-success/70">
                      From whole foods
                    </div>
                  </div>

                  <!-- Added Sugars -->
                  <div class="text-center">
                    <div class="stat-value text-xl text-warning">
                      {addedSugars.toFixed(1)}g
                    </div>
                    <div class="stat-title">Added Sugars</div>
                    {#if goals?.upper_limits?.added_sugars_g}
                      <div class="stat-desc text-xs">
                        {(
                          (addedSugars / goals.upper_limits.added_sugars_g) *
                          100
                        ).toFixed(0)}% of {goals.upper_limits.added_sugars_g}g
                        limit
                      </div>
                    {/if}
                  </div>

                  <!-- Sugar Ratio -->
                  <div class="text-center">
                    <div class="stat-value text-xl text-info">
                      {totalSugars > 0
                        ? ((naturalSugars / totalSugars) * 100).toFixed(0)
                        : 0}%
                    </div>
                    <div class="stat-title">Natural</div>
                    <div class="stat-desc text-xs text-info/70">
                      vs Added ratio
                    </div>
                  </div>
                </div>

                <!-- Visual ratio bar -->
                {#if totalSugars > 0}
                  <div class="mt-4">
                    <div
                      class="flex items-center text-xs text-base-content/70 mb-1"
                    >
                      <span class="text-success">Natural</span>
                      <span class="flex-1"></span>
                      <span class="text-warning">Added</span>
                    </div>
                    <div
                      class="flex h-2 bg-base-200 rounded-full overflow-hidden"
                    >
                      <div
                        class="bg-success transition-all duration-300 sugar-bar-natural"
                        style:width="{(naturalSugars / totalSugars) * 100}%"
                      ></div>
                      <div
                        class="bg-warning transition-all duration-300 sugar-bar-added"
                        style:width="{(addedSugars / totalSugars) * 100}%"
                      ></div>
                    </div>
                  </div>
                {/if}
              </div>
            </div>
          {/if}

          <div
            class="stats stats-vertical lg:stats-horizontal bg-base-100 shadow-sm mt-6"
          >
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
              <div class="stat-value text-lg flex items-center gap-2">
                {goals.source === "custom"
                  ? goals.custom_name || "Custom"
                  : "DRI"}
                <!-- Info button for modal -->
                {#if goals.source === "custom"}
                  <InfoButton modalId="custom-goals-info" />
                {:else}
                  <InfoButton modalId="dri-info" />
                {/if}
              </div>
              <div class="stat-desc">
                {goals.source === "custom"
                  ? "Custom Nutrition Goal"
                  : "Nutrition guidelines"}
              </div>
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<!-- Sugar Info Modal -->
<input type="checkbox" id="sugar-info-modal" class="modal-toggle" />
<div class="modal">
  <div class="modal-box">
    <label
      for="sugar-info-modal"
      class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</label
    >
    <h3 class="font-bold text-lg mb-4">Natural Sugar vs. Added Sugar</h3>
    <div class="prose prose-sm max-w-none">
      <p>
        Even though the molecules are the same (glucose, fructose, sucrose),
        sugar's health effects depend entirely on the context in which it's
        eaten.
      </p>

      <ul>
        <li>
          <strong>Natural sugars</strong> in whole foods like berries come
          packaged with fiber, water, vitamins, and phytochemicals. That fiber
          slows absorption, helping stabilize blood sugar. Plus, studies show
          that eating whole fruits—especially berries, grapes, and apples—is
          linked to a <strong>lower risk of developing type 2 diabetes</strong>
          (<a
            href="https://pubmed.ncbi.nlm.nih.gov/23990623/"
            target="_blank"
            rel="noopener noreferrer"
            class="link link-primary">PubMed</a
          >).
        </li>
        <li>
          <strong>Added sugars</strong>, such as those in soda, juice, or
          sweets, deliver calories without nutrients. These "empty" calories
          spike blood sugar, promote fat storage, and are strongly associated
          with <strong>weight gain, type 2 diabetes, and heart disease</strong>
          (<a
            href="https://pmc.ncbi.nlm.nih.gov/articles/PMC6723421/"
            target="_blank"
            rel="noopener noreferrer"
            class="link link-primary">PMC</a
          >,
          <a
            href="https://www.nature.com/articles/s41574-021-00627-6"
            target="_blank"
            rel="noopener noreferrer"
            class="link link-primary">Nature</a
          >).
        </li>
      </ul>

      <p>
        <strong>Bottom line:</strong> Sugar from whole, minimally processed foods
        isn't harmful and may even support health. In contrast, excess added sugar—especially
        in sugary drinks—poses clear risks.
      </p>
    </div>
    <div class="modal-action">
      <label for="sugar-info-modal" class="btn btn-primary">Got it!</label>
    </div>
  </div>
  <label class="modal-backdrop" for="sugar-info-modal">Close</label>
</div>

<!-- DRI Info Modal -->
<input type="checkbox" id="dri-info" class="modal-toggle" />
<div class="modal">
  <div class="modal-box">
    <label
      for="dri-info"
      class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</label
    >
    <h3 class="font-bold text-lg mb-4">Dietary Reference Intakes (DRI)</h3>
    <div class="prose prose-sm max-w-none">
      <p>
        The Dietary Reference Intakes (DRI) are a set of reference values used
        to plan and assess nutrient intakes of healthy people. They are
        developed by health experts and include:
      </p>

      <ul>
        <li>
          <strong>Recommended Dietary Allowance (RDA):</strong> The average daily
          dietary nutrient intake level sufficient to meet the nutrient requirements
          of nearly all (97–98 percent) healthy people.
        </li>
        <li>
          <strong>Adequate Intake (AI):</strong> Used when an RDA cannot be determined.
          Based on observed or experimentally-determined estimates of nutrient intake.
        </li>
        <li>
          <strong>Tolerable Upper Intake Level (UL):</strong> The highest average
          daily nutrient intake level likely to pose no risk of adverse health effects.
        </li>
      </ul>

      <p>
        These guidelines help ensure you get adequate nutrition while avoiding
        potentially harmful amounts of nutrients.
      </p>

      <p class="text-sm text-base-content/70 mt-4">
        <strong>Source:</strong> U.S. National Academy of Sciences, Engineering,
        and Medicine
      </p>
    </div>
    <div class="modal-action">
      <a
        href="https://www.nal.usda.gov/human-nutrition-and-food-safety/dietary-guidance"
        target="_blank"
        rel="noopener noreferrer"
        class="btn btn-outline"
      >
        Learn More
        <svg
          class="w-4 h-4 ml-1"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
          />
        </svg>
      </a>
      <label for="dri-info" class="btn btn-primary">Got it!</label>
    </div>
  </div>
  <label class="modal-backdrop" for="dri-info">Close</label>
</div>

<!-- Custom Goals Info Modal -->
<input type="checkbox" id="custom-goals-info" class="modal-toggle" />
<div class="modal">
  <div class="modal-box max-w-lg max-h-[80vh] overflow-y-auto">
    <label
      for="custom-goals-info"
      class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2 z-10"
      >✕</label
    >
    <h3 class="font-bold text-lg mb-4">Custom Nutrition Goals</h3>
    <div class="prose prose-sm max-w-none">
      {#if goals && goals.source === "custom"}
        <p>
          <strong>Goal Name:</strong>
          {goals.custom_name || "Unnamed Custom Goal"}
        </p>

        {#if goals.targets && Object.keys(goals.targets).length > 0}
          <h4 class="font-semibold mt-4 mb-2">Custom Targets:</h4>
          <div
            class="bg-base-200 p-3 rounded text-xs space-y-1 max-h-48 overflow-y-auto"
          >
            {#each Object.entries(goals.targets) as [nutrient, value]}
              <div class="flex justify-between">
                <span class="capitalize">{nutrient.replace(/_/g, " ")}</span>
                <span class="font-mono"
                  >{value}{nutrient.includes("_mcg")
                    ? "μg"
                    : nutrient.includes("_mg")
                      ? "mg"
                      : nutrient.includes("_g")
                        ? "g"
                        : ""}</span
                >
              </div>
            {/each}
          </div>
        {/if}

        {#if goals.upper_limits && Object.keys(goals.upper_limits).length > 0}
          <h4 class="font-semibold mt-4 mb-2">Custom Upper Limits:</h4>
          <div
            class="bg-warning/10 p-3 rounded text-xs space-y-1 max-h-32 overflow-y-auto"
          >
            {#each Object.entries(goals.upper_limits) as [nutrient, value]}
              <div class="flex justify-between">
                <span class="capitalize">{nutrient.replace(/_/g, " ")}</span>
                <span class="font-mono"
                  >{value}{nutrient.includes("_mcg")
                    ? "μg"
                    : nutrient.includes("_mg")
                      ? "mg"
                      : nutrient.includes("_g")
                        ? "g"
                        : ""}</span
                >
              </div>
            {/each}
          </div>
        {/if}

        <p class="text-sm text-base-content/70 mt-4">
          These custom goals override the default DRI recommendations and are
          tailored to your specific needs. You can change them anytime on your
          profile page.
        </p>
      {:else}
        <p>
          Custom nutrition goals allow you to set personalized targets that
          override the default DRI recommendations.
        </p>
        <p>
          Upgrade to Pro to create and use custom nutrition goals tailored to
          your specific dietary needs.
        </p>
      {/if}
    </div>
    <div class="modal-action">
      <label for="custom-goals-info" class="btn btn-primary">Got it!</label>
    </div>
  </div>
  <label class="modal-backdrop" for="custom-goals-info">Close</label>
</div>

<style>
  .goals-header {
    color: var(--color-base-content);
  }

  .goals-subheader {
    color: var(--color-base-content);
  }
</style>
