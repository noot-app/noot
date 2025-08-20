<script lang="ts">
  import { apiClient } from "$lib/api/client";
  import { PUBLIC_APP_NAME } from "$env/static/public";
  import { onMount } from "svelte";
  import { dev } from '$app/environment';
  import type { paths } from "$lib/api/schema";

  type GoalsResponse = paths["/goals"]["get"]["responses"]["200"]["content"]["application/json"];
  type Goals = GoalsResponse["goals"];

  let goals: Goals | null = null;
  let loading = true;
  let error = "";
  let success = "";
  let saving = false;

  // Form state
  let customName = "";
  let customTargets: Record<string, number> = {};
  let showImperialModal = false;
  let selectedUnits = "metric"; // Track unit system selection

  onMount(async () => {
    await loadGoals();
  });

  async function loadGoals() {
    try {
      loading = true;
      const response = await apiClient.GET("/goals");
      
      if (response.error) {
        throw new Error(`API Error: ${response.error}`);
      }

      goals = response.data.goals;
      
      // Initialize form with current values
      customName = goals.custom_name || "";
      // Note: We'll initialize custom targets as empty and let users override as needed
      customTargets = {};
    } catch (err) {
      error = `Failed to load goals: ${err}`;
      console.error("Goals error:", err);
    } finally {
      loading = false;
    }
  }

  function parseErrorMessage(error: any): string {
    // If it's a plain string, return it
    if (typeof error === 'string') {
      return error;
    }

    // If it has a message property, use that
    if (error?.message && typeof error.message === 'string') {
      return error.message;
    }

    // If it's an Error object, use its message
    if (error instanceof Error) {
      return error.message;
    }

    // Try to extract meaningful error from API response
    if (error?.error && typeof error.error === 'string') {
      return error.error;
    }

    // Fallback to JSON representation
    return JSON.stringify(error);
  }

  function formatErrorForUser(rawError: any): string {
    const errorMessage = parseErrorMessage(rawError);
    
    // Common backend error messages that need user-friendly translations
    const errorTranslations: Record<string, string> = {
      'At least one override must be provided': 'At least one nutrition target override must be provided',
      'Invalid request': 'Please check your input and try again',
      'Unauthorized': 'You need to be logged in to save goals',
      'Forbidden': 'You don\'t have permission to perform this action'
    };

    // Check if we have a translation for this exact message
    const translation = errorTranslations[errorMessage];
    if (translation) {
      return dev ? `${translation} (Dev: ${errorMessage})` : translation;
    }

    // Clean up nested error messages
    const cleanError = errorMessage
      .replace(/^Error: Failed to save goals: /, '')
      .replace(/^Error saving goals: Error: /, '')
      .replace(/^Error: /, '');

    // If in development, add context
    if (dev) {
      return `${cleanError} (check browser console for details)`;
    }

    return cleanError;
  }

  async function saveCustomGoals() {
    if (!goals) return;
    
    try {
      saving = true;
      error = "";
      success = "";

      // Prepare the request payload
      const payload: any = {
        name: customName.trim() || undefined,
        overrides: Object.keys(customTargets).length > 0 ? customTargets : undefined
      };

      const response = await apiClient.PUT("/goals", {
        body: payload
      });

      if (response.error) {
        throw response.error;
      }

      success = "Goals saved successfully!";
      await loadGoals(); // Reload to get updated data
    } catch (err) {
      error = formatErrorForUser(err);
      if (dev) {
        console.error("Save error details:", err);
      }
    } finally {
      saving = false;
    }
  }

  function resetToDefaults() {
    customTargets = {};
    customName = "";
  }

  // Key nutrients that users might want to customize
  // Generate dynamically from available goals instead of hardcoding
  $: editableNutrients = goals ? Object.keys(goals.targets).map(key => ({
    key,
    label: formatNutrientName(key),
    unit: goals.units[key] || ""
  })).sort((a, b) => a.label.localeCompare(b.label)) : [];

  function formatNutrientName(key: string): string {
    return key
      .replace(/_/g, " ")
      .replace(/\b\w/g, l => l.toUpperCase())
      // Remove unit suffixes since they're shown separately
      .replace(/ Mcg$/, "")
      .replace(/ Mg$/, "")
      .replace(/ G$/, "");
  }

  function getNutrientValue(key: string): number {
    return customTargets[key] || goals?.targets[key] || 0;
  }

  function updateNutrient(key: string, value: number) {
    if (value <= 0) {
      delete customTargets[key];
    } else {
      customTargets[key] = value;
    }
    customTargets = { ...customTargets }; // Trigger reactivity
  }

  function openImperialModal() {
    showImperialModal = true;
  }

  function closeImperialModal() {
    showImperialModal = false;
    selectedUnits = "metric"; // Force selection back to metric
  }
</script>

<svelte:head>
  <title>Profile - {PUBLIC_APP_NAME}</title>
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-8 max-w-4xl">
    <!-- Header -->
    <div class="text-center mb-8">
      <h1 class="text-3xl font-bold text-primary mb-4">Profile & Goals</h1>
      <p class="text-base-content/70">Customize your nutrition goals and preferences</p>
    </div>

    {#if loading}
      <div class="text-center py-12">
        <span class="loading loading-spinner loading-lg"></span>
        <p class="text-base-content/70 mt-4">Loading your profile...</p>
      </div>
    {:else if error}
      <div class="alert alert-error mb-6">
        <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{error}</span>
      </div>
    {/if}

    {#if success}
      <div class="alert alert-success mb-6">
        <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{success}</span>
      </div>
    {/if}

    {#if goals}
      <div class="grid gap-8 lg:grid-cols-2">
        <!-- Current Goals Overview -->
        <div class="card bg-base-200 shadow-lg">
          <div class="card-body">
            <h2 class="card-title">Current Goals</h2>
            <div class="space-y-4">
              <div class="stats bg-base-100 shadow-sm">
                <div class="stat">
                  <div class="stat-title">Profile</div>
                  <div class="stat-value text-sm">
                    {goals.life_stage.sex} • {goals.life_stage.age_bracket}
                  </div>
                  <div class="stat-desc">Demographic info</div>
                </div>
              </div>
              
              <div class="badge badge-primary badge-lg">
                {goals.source === "custom" 
                  ? (goals.custom_name || "Custom Goals") 
                  : "DRI Guidelines"}
              </div>
              
              {#if goals.source === "dri"}
                <p class="text-sm text-base-content/70">
                  Currently using Dietary Reference Intakes (DRI) based on your profile. 
                  You can create custom goals below to override specific nutrients.
                </p>
              {/if}
            </div>
          </div>
        </div>

        <!-- Custom Goals Form -->
        <div class="card bg-base-200 shadow-lg">
          <div class="card-body">
            <h2 class="card-title">Customize Goals</h2>
            
            <div class="form-control w-full">
              <label class="label" for="goalName">
                <span class="label-text">Goal Set Name (Optional)</span>
              </label>
              <input 
                id="goalName"
                type="text" 
                placeholder="e.g., 'My Fitness Goals', 'Cutting Diet'"
                class="input input-bordered w-full" 
                maxlength="50"
                bind:value={customName}
              />
              <div class="label">
                <span class="label-text-alt">Give your custom goals a memorable name</span>
              </div>
            </div>

            <div class="divider">Nutrition Targets</div>
            
            <div class="space-y-4 max-h-96 overflow-y-auto">
              {#each editableNutrients as nutrient}
                <div class="form-control">
                  <label class="label" for={nutrient.key}>
                    <span class="label-text">{nutrient.label}</span>
                    <span class="label-text-alt">{nutrient.unit}</span>
                  </label>
                  <div class="flex gap-2 items-center">
                    <input 
                      id={nutrient.key}
                      type="number" 
                      class="input input-bordered flex-1" 
                      min="0"
                      step="0.1"
                      value={getNutrientValue(nutrient.key)}
                      on:input={(e) => updateNutrient(nutrient.key, parseFloat(e.currentTarget.value) || 0)}
                    />
                    {#if customTargets[nutrient.key]}
                      <button 
                        class="btn btn-ghost btn-sm"
                        on:click={() => updateNutrient(nutrient.key, 0)}
                        title="Reset to default"
                      >
                        ↺
                      </button>
                    {/if}
                  </div>
                  {#if !customTargets[nutrient.key] && goals.targets[nutrient.key]}
                    <div class="label">
                      <span class="label-text-alt">Default: {goals.targets[nutrient.key]} {nutrient.unit}</span>
                    </div>
                  {/if}
                </div>
              {/each}
            </div>

            <div class="card-actions justify-between mt-6">
              <button 
                class="btn btn-outline"
                on:click={resetToDefaults}
                disabled={saving}
              >
                Reset All
              </button>
              <button 
                class="btn btn-primary"
                on:click={saveCustomGoals}
                disabled={saving}
              >
                {#if saving}
                  <span class="loading loading-spinner loading-sm"></span>
                  Saving...
                {:else}
                  💾 Save Goals
                {/if}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Additional Settings Section -->
      <div class="card bg-base-200 shadow-lg mt-8">
        <div class="card-body">
          <h2 class="card-title">Additional Settings</h2>
          <div class="grid gap-4 md:grid-cols-2">
            <div class="card bg-base-100">
              <div class="card-body">
                <h3 class="card-title text-lg">Units Preference</h3>
                <p class="text-sm text-base-content/70 mb-4">
                  Choose your preferred units for nutrition display
                </p>
                <div class="form-control">
                  <label class="label cursor-pointer">
                    <span class="label-text">Metric (grams, milligrams)</span> 
                    <input 
                      type="radio" 
                      name="units" 
                      class="radio radio-primary" 
                      bind:group={selectedUnits} 
                      value="metric"
                    />
                  </label>
                </div>
                <div class="form-control">
                  <label class="label cursor-pointer">
                    <span class="label-text">Imperial (ounces, pounds)</span> 
                    <input 
                      type="radio" 
                      name="units" 
                      class="radio radio-primary" 
                      bind:group={selectedUnits} 
                      value="imperial"
                      on:change={openImperialModal}
                    />
                  </label>
                </div>
              </div>
            </div>

            <div class="card bg-base-100">
              <div class="card-body">
                <h3 class="card-title text-lg">Data & Privacy</h3>
                <p class="text-sm text-base-content/70 mb-4">
                  Manage your data and privacy settings
                </p>
                <div class="space-y-2">
                  <button class="btn btn-outline btn-sm w-full">Export Data</button>
                  <button class="btn btn-outline btn-sm w-full">Clear History</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<!-- Imperial System Meme Modal -->
{#if showImperialModal}
  <div class="modal modal-open">
    <div class="modal-box max-w-2xl">
      <div class="text-center space-y-6">
        <!-- Meme Header -->
        <div class="text-6xl">🚫</div>
        <h3 class="font-bold text-2xl" style="color: var(--color-dark);">LOL NO.</h3>
        
        <!-- Meme Content -->
        <div class="space-y-4 text-lg">
          <p>You can't use Imperial units on this site.</p>
          <p class="font-semibold text-primary">This site uses the METRIC SYSTEM because it is SUPERIOR! 🧑‍🔬</p>
          
          <div class="bg-base-200 p-4 rounded-box space-y-2">
            <p class="text-sm">🌍 Used by 95% of the world</p>
            <p class="text-sm">🧮 Base-10, actually makes sense</p>
            <p class="text-sm">🚀 Used by NASA (even though they're American)</p>
            <p class="text-sm">🔬 All scientific research uses metric</p>
            <p class="text-sm">💊 Your medicine dosages? Metric.</p>
            <p class="text-sm">🏃‍♂️ Olympic records? Metric.</p>
          </div>
          
          <div class="text-base space-y-2">
            <p>Imperial is just...</p>
            <p class="italic">"12 inches in a foot, 3 feet in a yard, 1760 yards in a mile"</p>
            <p class="font-bold">vs.</p>
            <p class="italic">"10mm = 1cm, 100cm = 1m, 1000m = 1km"</p>
            <p class="text-primary font-semibold">See the difference? 🤯</p>
          </div>
          
          <div class="text-sm text-base-content/70">
            <p>Even the UK switched to metric for most things.</p>
            <p>It's time to let go of the past. 📏➡️📐</p>
          </div>
        </div>
        
        <!-- Acknowledgment Button -->
        <div class="modal-action justify-center">
          <button 
            class="btn btn-success btn-lg"
            on:click={closeImperialModal}
          >
            I acknowledge that the metric system is better
          </button>
        </div>
      </div>
    </div>
    <div 
      class="modal-backdrop" 
      on:click={closeImperialModal}
      on:keydown={(e) => e.key === 'Escape' && closeImperialModal()}
      role="button" 
      tabindex="0"
      aria-label="Close modal"
    ></div>
  </div>
{/if}

<style>
  /* Better focus styling for inputs and interactive elements */
  .input:focus,
  .input:focus-visible {
    outline: none;
    border-color: oklch(var(--p));
    box-shadow: 0 0 0 2px oklch(var(--p) / 0.2);
  }

  /* Card focus styling */
  .card:focus,
  .card:focus-visible {
    outline: none;
    border: 2px solid oklch(var(--p) / 0.3);
    box-shadow: 0 0 0 1px oklch(var(--p) / 0.1);
  }

  /* Button focus improvements */
  .btn:focus,
  .btn:focus-visible {
    outline: none;
    box-shadow: 0 0 0 2px oklch(var(--p) / 0.3);
  }

  /* Radio button focus */
  .radio:focus,
  .radio:focus-visible {
    outline: none;
    box-shadow: 0 0 0 2px oklch(var(--p) / 0.4);
  }

  /* Modal backdrop - remove focus styles since it's not meant to be keyboard navigable in normal use */
  .modal-backdrop:focus {
    outline: none;
  }

  /* Smooth transitions for focus states */
  .input,
  .card,
  .btn,
  .radio {
    transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;
  }
</style>
