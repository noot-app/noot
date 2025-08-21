<script lang="ts">
  import { apiClient } from "$lib/api/client";
  import { PUBLIC_APP_NAME } from "$env/static/public";
  import { onMount } from "svelte";
  import { dev } from '$app/environment';
  import type { paths } from "$lib/api/schema";

  type GoalsResponse = paths["/goals"]["get"]["responses"]["200"]["content"]["application/json"];
  type Goals = GoalsResponse["goals"];
  type BiometricsResponse = paths["/biometrics"]["get"]["responses"]["200"]["content"]["application/json"];
  type UserBiometrics = paths["/biometrics"]["get"]["responses"]["200"]["content"]["application/json"]["biometrics"];
  type UpdateBiometricsRequest = paths["/biometrics"]["put"]["requestBody"]["content"]["application/json"];
  
  let goals: Goals | null = null;
  let biometrics: UserBiometrics | null = null;
  let calculatedMetrics: BiometricsResponse["calculated_metrics"] | null = null;
  let loading = true;
  let biometricsLoading = false;
  let error = "";
  let biometricsError = "";
  let success = "";
  let biometricsSuccess = "";
  let saving = false;
  let savingBiometrics = false;

  // Form state
  let customName = "";
  let customTargets: Record<string, number> = {};
  let showImperialModal = false;
  let selectedUnits = "metric"; // Track unit system selection

  // Biometrics form state
  let birthDate = "";
  let sex: "male" | "female" | "other" | "prefer_not_to_say" = "prefer_not_to_say";
  let heightCm = "";
  let weightKg = "";
  let activityLevel: "sedentary" | "lightly_active" | "moderately_active" | "very_active" | "extra_active" = "lightly_active";

  onMount(async () => {
    await Promise.all([loadGoals(), loadBiometrics()]);
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

  async function loadBiometrics() {
    try {
      biometricsLoading = true;
      biometricsError = "";
      
      const response = await apiClient.GET("/biometrics");
      
      if (response.error) {
        throw new Error(`API Error: ${response.error}`);
      }

      biometrics = response.data.biometrics;
      calculatedMetrics = response.data.calculated_metrics;
      
      // Initialize form with current values
      if (biometrics) {
        birthDate = biometrics.birth_date || "";
        sex = biometrics.sex || "prefer_not_to_say";
        heightCm = biometrics.height_cm?.toString() || "";
        weightKg = biometrics.weight_kg?.toString() || "";
        activityLevel = biometrics.activity_level || "lightly_active";
      }
    } catch (err) {
      // Don't show error if biometrics just don't exist yet
      if (!err?.toString().includes("404") && !err?.toString().includes("not found")) {
        biometricsError = `Failed to load biometrics: ${err}`;
        console.error("Biometrics error:", err);
      }
    } finally {
      biometricsLoading = false;
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

  async function saveBiometrics() {
    try {
      savingBiometrics = true;
      biometricsError = "";
      biometricsSuccess = "";

      // Prepare the request payload
      const payload: UpdateBiometricsRequest = {};
      
      if (birthDate.trim()) {
        payload.birth_date = birthDate.trim();
      }
      
      if (sex && sex !== "prefer_not_to_say") {
        payload.sex = sex;
      }
      
      if (heightCm && !isNaN(parseFloat(heightCm.toString()))) {
        payload.height_cm = parseFloat(heightCm.toString());
      }
      
      if (weightKg && !isNaN(parseFloat(weightKg.toString()))) {
        payload.weight_kg = parseFloat(weightKg.toString());
      }
      
      if (activityLevel) {
        payload.activity_level = activityLevel;
      }

      const response = await apiClient.PUT("/biometrics", {
        body: payload
      });

      if (response.error) {
        throw response.error;
      }

      biometricsSuccess = "Biometrics saved successfully!";
      await Promise.all([loadBiometrics(), loadGoals()]); // Reload both since goals may have changed
    } catch (err) {
      biometricsError = formatErrorForUser(err);
      if (dev) {
        console.error("Biometrics save error details:", err);
      }
    } finally {
      savingBiometrics = false;
    }
  }

  async function deleteBiometrics() {
    if (!confirm("Are you sure you want to delete all your biometric data? This action cannot be undone.")) {
      return;
    }

    try {
      savingBiometrics = true;
      biometricsError = "";
      biometricsSuccess = "";

      const response = await apiClient.DELETE("/biometrics");

      if (response.error) {
        throw response.error;
      }

      biometricsSuccess = "Biometrics deleted successfully!";
      
      // Clear form
      birthDate = "";
      sex = "prefer_not_to_say";
      heightCm = "";
      weightKg = "";
      activityLevel = "lightly_active";
      
      await Promise.all([loadBiometrics(), loadGoals()]); // Reload both since goals may have changed
    } catch (err) {
      biometricsError = formatErrorForUser(err);
      if (dev) {
        console.error("Biometrics delete error details:", err);
      }
    } finally {
      savingBiometrics = false;
    }
  }

  function resetToDefaults() {
    customTargets = {};
    customName = "";
  }

  // Key nutrients that users might want to customize
  // Generate dynamically from available goals instead of hardcoding
  $: editableTargets = goals ? Object.keys(goals.targets).map(key => ({
    key,
    label: formatNutrientName(key),
    unit: goals?.units[key] || ""
  })).sort((a, b) => a.label.localeCompare(b.label)) : [];

  $: editableUpperLimits = goals ? Object.keys(goals.upper_limits || {}).map(key => ({
    key,
    label: formatNutrientName(key),
    unit: goals?.units[key] || ""
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

  function getUpperLimitValue(key: string): number {
    return customTargets[key] || goals?.upper_limits?.[key] || 0;
  }

  function updateNutrient(key: string, value: number) {
    if (value <= 0) {
      delete customTargets[key];
    } else {
      customTargets[key] = value;
    }
    customTargets = { ...customTargets }; // Trigger reactivity
  }

  function updateUpperLimit(key: string, value: number) {
    if (value < 0) {
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

  function getActivityLevelDisplay(level: string): string {
    const activityLevels: Record<string, string> = {
      "sedentary": "L1",
      "lightly_active": "L2", 
      "moderately_active": "L3",
      "very_active": "L4",
      "extra_active": "L5"
    };
    return activityLevels[level] || level;
  }

  function getActivityLevelDescription(level: string): string {
    const descriptions: Record<string, string> = {
      "sedentary": "Sedentary (little/no exercise)",
      "lightly_active": "Lightly Active (1-3 days/week)",
      "moderately_active": "Moderately Active (3-5 days/week)", 
      "very_active": "Very Active (6-7 days/week)",
      "extra_active": "Extra Active (very hard exercise daily)"
    };
    return descriptions[level] || level;
  }
</script>

<svelte:head>
  <title>Profile - {PUBLIC_APP_NAME}</title>
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-8 max-w-6xl">
    <!-- Header -->
    <div class="text-center mb-8">
      <h1 class="text-3xl font-bold text-primary mb-4">Profile</h1>
      <p class="text-base-content/70">Customize your profile, nutrition goals, and preferences</p>
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

    {#if biometricsSuccess}
      <div class="alert alert-success mb-6">
        <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{biometricsSuccess}</span>
      </div>
    {/if}

    {#if biometricsError}
      <div class="alert alert-error mb-6">
        <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{biometricsError}</span>
      </div>
    {/if}

    {#if goals}
      <!-- Main Profile Grid -->
      <div class="grid gap-8 xl:grid-cols-3 lg:grid-cols-2 md:grid-cols-1">
        <!-- Current Goals Overview -->
        <div class="card bg-base-200 shadow-lg">
          <div class="card-body p-6">
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

        <!-- User Biometrics -->
        <div class="card bg-base-200 shadow-lg">
          <div class="card-body p-6">
            <h2 class="card-title flex items-center gap-2">
              📊 Biometrics
              {#if calculatedMetrics?.age_years}
                <div class="badge badge-primary badge-sm">{calculatedMetrics.age_years}y</div>
              {/if}
            </h2>
            
            {#if biometricsLoading}
              <div class="text-center py-4">
                <span class="loading loading-spinner loading-sm"></span>
                <p class="text-sm text-base-content/70 mt-2">Loading biometrics...</p>
              </div>
            {:else}
              <div class="space-y-4">
                <!-- Current Biometrics Display -->
                {#if biometrics}
                  <div class="grid grid-cols-2 gap-4">
                    {#if calculatedMetrics?.bmi}
                      <div class="stat bg-base-100 rounded-box p-3 tooltip tooltip-top" data-tip="Body Mass Index - A measure of body fat based on height and weight">
                        <div class="stat-title text-xs">BMI</div>
                        <div class="stat-value text-lg">{calculatedMetrics.bmi}</div>
                      </div>
                    {/if}
                    {#if calculatedMetrics?.bmr}
                      <div class="stat bg-base-100 rounded-box p-3 tooltip tooltip-top" data-tip="Basal Metabolic Rate - Calories your body burns at rest for basic functions">
                        <div class="stat-title text-xs">BMR</div>
                        <div class="stat-value text-lg">{Math.round(calculatedMetrics.bmr)}</div>
                        <div class="stat-desc text-xs">kcal/day</div>
                      </div>
                    {/if}
                    {#if calculatedMetrics?.tdee}
                      <div class="stat bg-base-100 rounded-box p-3 tooltip tooltip-top" data-tip="Total Daily Energy Expenditure - Total calories burned including exercise and daily activities">
                        <div class="stat-title text-xs">TDEE</div>
                        <div class="stat-value text-lg">{Math.round(calculatedMetrics.tdee)}</div>
                        <div class="stat-desc text-xs">kcal/day</div>
                      </div>
                    {/if}
                    {#if biometrics.activity_level}
                      <div class="stat bg-base-100 rounded-box p-3 tooltip tooltip-top" data-tip="{getActivityLevelDescription(biometrics.activity_level)}">
                        <div class="stat-title text-xs">Activity</div>
                        <div class="stat-value text-lg">{getActivityLevelDisplay(biometrics.activity_level)}</div>
                      </div>
                    {/if}
                  </div>
                {:else}
                  <div class="alert alert-info">
                    <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <span>No biometric data yet. Add your details below for personalized nutrition goals!</span>
                  </div>
                {/if}

                <!-- Quick Form -->
                <div class="space-y-3">
                  <!-- Birth Date -->
                  <div class="form-control">
                    <label class="label" for="birthDate">
                      <span class="label-text text-sm">Birth Date</span>
                    </label>
                    <input 
                      id="birthDate"
                      type="date" 
                      class="input input-bordered input-sm"
                      bind:value={birthDate}
                      max={new Date().toISOString().split('T')[0]}
                    />
                  </div>

                  <!-- Sex -->
                  <div class="form-control">
                    <label class="label" for="sex">
                      <span class="label-text text-sm">Sex (for DRI calculations)</span>
                    </label>
                    <select id="sex" class="select select-bordered select-sm" bind:value={sex}>
                      <option value="prefer_not_to_say">Prefer not to say</option>
                      <option value="male">Male</option>
                      <option value="female">Female</option>
                      <option value="other">Other</option>
                    </select>
                  </div>

                  <!-- Height & Weight Row -->
                  <div class="grid grid-cols-2 gap-2">
                    <div class="form-control">
                      <label class="label" for="height">
                        <span class="label-text text-sm">Height (cm)</span>
                      </label>
                      <input 
                        id="height"
                        type="number"
                        class="input input-bordered input-sm"
                        placeholder="175"
                        min="50"
                        max="300"
                        step="0.1"
                        bind:value={heightCm}
                      />
                    </div>
                    <div class="form-control">
                      <label class="label" for="weight">
                        <span class="label-text text-sm">Weight (kg)</span>
                      </label>
                      <input 
                        id="weight"
                        type="number"
                        class="input input-bordered input-sm"
                        placeholder="70"
                        min="20"
                        max="500"
                        step="0.1"
                        bind:value={weightKg}
                      />
                    </div>
                  </div>

                  <!-- Activity Level -->
                  <div class="form-control">
                    <label class="label" for="activity">
                      <span class="label-text text-sm">Activity Level</span>
                    </label>
                    <select id="activity" class="select select-bordered select-sm" bind:value={activityLevel}>
                      <option value="sedentary">Level 1 - Sedentary (little/no exercise)</option>
                      <option value="lightly_active">Level 2 - Lightly Active (1-3 days/week)</option>
                      <option value="moderately_active">Level 3 - Moderately Active (3-5 days/week)</option>
                      <option value="very_active">Level 4 - Very Active (6-7 days/week)</option>
                      <option value="extra_active">Level 5 - Extra Active (very hard exercise daily)</option>
                    </select>
                  </div>
                </div>

                <!-- Actions -->
                <div class="flex gap-2 justify-between">
                  {#if biometrics}
                    <button 
                      class="btn btn-outline btn-error btn-sm"
                      on:click={deleteBiometrics}
                      disabled={savingBiometrics}
                    >
                      🗑️ Delete
                    </button>
                  {:else}
                    <div></div>
                  {/if}
                  
                  <button 
                    class="btn btn-primary btn-sm"
                    on:click={saveBiometrics}
                    disabled={savingBiometrics}
                  >
                    {#if savingBiometrics}
                      <span class="loading loading-spinner loading-xs"></span>
                      Saving...
                    {:else}
                      💾 Save
                    {/if}
                  </button>
                </div>
              </div>
            {/if}
          </div>
        </div>

        <!-- Custom Goals Form -->
        <div class="card bg-base-200 shadow-lg">
          <div class="card-body p-6">
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
                <span class="label-text-alt text-wrap">Give your goals a memorable name</span>
              </div>
            </div>

            <div class="divider">Nutrition Targets</div>
            
            <div class="space-y-6 max-h-96 overflow-y-auto">
              <!-- Regular Nutrition Targets -->
              {#if editableTargets.length > 0}
                <div>
                  <h4 class="font-semibold text-base mb-3 text-primary">Daily Targets</h4>
                  <div class="space-y-4">
                    {#each editableTargets as nutrient}
                      <div class="form-control">
                        <label class="label" for={nutrient.key}>
                          <span class="label-text">{nutrient.label}</span>
                          <span class="label-text-alt">{nutrient.unit}</span>
                        </label>
                        <input 
                          type="number"
                          id={nutrient.key}
                          class="input input-bordered input-sm"
                          min="0"
                          step="0.1"
                          placeholder={getNutrientValue(nutrient.key).toString()}
                          value={getNutrientValue(nutrient.key)}
                          on:input={(e) => updateNutrient(nutrient.key, parseFloat(e.currentTarget.value) || 0)}
                        />
                      </div>
                    {/each}
                  </div>
                </div>
              {/if}

              <!-- Upper Limits (Minimize These) -->
              {#if editableUpperLimits.length > 0}
                <div>
                  <h4 class="font-semibold text-base mb-3 text-warning">Upper Limits</h4>
                  <p class="text-xs text-base-content/70 mb-3">Set maximum daily limits for nutrients that should be minimized.</p>
                  <div class="space-y-4">
                    {#each editableUpperLimits as nutrient}
                      <div class="form-control">
                        <label class="label" for={`limit_${nutrient.key}`}>
                          <span class="label-text">{nutrient.label}</span>
                          <span class="label-text-alt">{nutrient.unit} (max)</span>
                        </label>
                        <input 
                          type="number"
                          id={`limit_${nutrient.key}`}
                          class="input input-bordered input-warning input-sm"
                          min="0"
                          step="0.1"
                          placeholder={getUpperLimitValue(nutrient.key).toString()}
                          value={getUpperLimitValue(nutrient.key)}
                          on:input={(e) => updateUpperLimit(nutrient.key, parseFloat(e.currentTarget.value) || 0)}
                        />
                      </div>
                    {/each}
                  </div>
                </div>
              {/if}
            </div>            <div class="card-actions justify-between mt-6">
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
        <div class="card-body p-6">
          <h2 class="card-title">Additional Settings</h2>
          <div class="grid gap-6 lg:grid-cols-2">
            <div class="card bg-base-100">
              <div class="card-body p-4">
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
              <div class="card-body p-4">
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
