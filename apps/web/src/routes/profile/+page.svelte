<script lang="ts">
  import { apiClient } from "$lib/api/client";
  import { PUBLIC_APP_NAME } from "$env/static/public";
  import { onMount } from "svelte";
  import { dev } from '$app/environment';
  import type { paths } from "$lib/api/schema";

  type GoalsResponse = paths["/goals"]["get"]["responses"]["200"]["content"]["application/json"];
  type Goals = GoalsResponse["goals"];
  type User = GoalsResponse["user"];
  type BiometricsResponse = paths["/biometrics"]["get"]["responses"]["200"]["content"]["application/json"];
  type UserBiometrics = paths["/biometrics"]["get"]["responses"]["200"]["content"]["application/json"]["biometrics"];
  type UpdateBiometricsRequest = paths["/biometrics"]["put"]["requestBody"]["content"]["application/json"];
  
  let goals: Goals | null = null;
  let user: User | null = null;
  let biometrics: UserBiometrics | null = null;
  let calculatedMetrics: BiometricsResponse["calculated_metrics"] | null = null;
  let loading = true;
  let biometricsLoading = false;
  let error = "";
  let biometricsError = "";
  let success = "";
  let biometricsSuccess = "";
  
  // Reactive statements for user tier
  $: isProUser = user?.subscription_tier === "pro";
  $: isFreeUser = user?.subscription_tier === "free";
  let saving = false;
  let savingBiometrics = false;

  // Form state
  let customGoalName = "";
  let selectedUnits = "metric"; // Track unit system selection
  let showImperialModal = false;
  
  // UI state for modal
  let showEditModal = false;
  let editingGoalName = "";
  
  // Delete confirmation modal state
  let showDeleteModal = false;
  let goalToDelete = "";
  
  // Reset DRI confirmation modal state
  let showResetDRIModal = false;
  
  // Goal sets data
  let goalSets: Array<{name: string, created_at: string, updated_at: string}> = [];
  let activeGoalName = "";
  let loadingGoalSets = false;
  let savingGoals = false;
  
  // Goal data for modal editing - now using dynamic approach
  let customTargets: Record<string, number> = {};
  let customName = "";

  // Biometrics form state
  let birthDate = "";
  let sex: "male" | "female" | "other" | "prefer_not_to_say" = "prefer_not_to_say";
  let heightCm = "";
  let weightKg = "";
  let activityLevel: "sedentary" | "lightly_active" | "moderately_active" | "very_active" | "extra_active" = "lightly_active";

  onMount(async () => {
    await Promise.all([loadGoals(), loadBiometrics(), loadGoalSets()]);
  });

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

  async function loadGoals() {
    try {
      loading = true;
      const response = await apiClient.GET("/goals");
      
      if (response.error) {
        throw new Error(`API Error: ${response.error}`);
      }

      goals = response.data.goals;
      user = response.data.user;
    } catch (err) {
      error = `Failed to load goals: ${err}`;
      console.error("Goals error:", err);
    } finally {
      loading = false;
    }
  }

  // Called when the active goal changes
  async function handleGoalChanged() {
    await Promise.all([loadGoals(), loadGoalSets()]);
    success = "Active goal switched successfully!";
    setTimeout(() => success = "", 3000);
  }

  async function loadGoalSets() {
    try {
      loadingGoalSets = true;
      const response = await apiClient.GET("/goals/sets");

      if (response.error) {
        throw new Error(`API Error: ${response.error}`);
      }

      goalSets = response.data.goal_sets || [];
      activeGoalName = response.data.active_goal_name || "";
    } catch (err) {
      console.error("Goal sets error:", err);
      // Don't show error if user just doesn't have multiple goals yet
    } finally {
      loadingGoalSets = false;
    }
  }

  async function switchToGoal(goalName: string) {
    if (goalName === activeGoalName) return;

    try {
      const response = await apiClient.PUT("/goals/active", {
        body: { name: goalName }
      });

      if (response.error) {
        throw response.error;
      }

      await handleGoalChanged();
    } catch (err) {
      error = `Failed to switch goal: ${err}`;
      console.error("Switch goal error:", err);
    }
  }

  async function deleteGoalSet(goalName: string) {
    if (!goalName) return;
    
    // Show modal instead of using confirm()
    goalToDelete = goalName;
    showDeleteModal = true;
  }

  async function confirmDeleteGoalSet() {
    if (!goalToDelete) return;

    try {
      const isLastGoal = goalSets.length === 1;
      const isActiveGoal = goalToDelete === activeGoalName;
      
      const response = await apiClient.DELETE("/goals/sets/{name}", {
        params: {
          path: { name: goalToDelete }
        }
      });

      if (response.error) {
        throw response.error;
      }

      if (isLastGoal && isActiveGoal) {
        success = `Goal set "${goalToDelete}" deleted successfully! You're now using DRI nutrition defaults.`;
      } else {
        success = `Goal set "${goalToDelete}" deleted successfully!`;
      }
      setTimeout(() => success = "", 3000);
      
      // Reload both goals and goal sets to update the UI and show DRI fallback
      await Promise.all([loadGoals(), loadGoalSets()]);
    } catch (err) {
      error = `Failed to delete goal set: ${parseErrorMessage(err)}`;
      console.error("Delete goal error:", err);
    } finally {
      // Close modal and reset state
      showDeleteModal = false;
      goalToDelete = "";
    }
  }

  function cancelDeleteGoalSet() {
    showDeleteModal = false;
    goalToDelete = "";
  }

  async function resetToDRIDefaults() {
    showResetDRIModal = true;
  }

  async function confirmResetToDRIDefaults() {
    try {
      // Delete ALL custom goal sets, including the active one
      // We'll delete them all in one go to avoid issues with active goal switching
      for (const goalSet of goalSets) {
        const response = await apiClient.DELETE("/goals/sets/{name}", {
          params: {
            path: { name: goalSet.name }
          }
        });
        
        if (response.error) {
          console.warn(`Failed to delete goal set ${goalSet.name}:`, response.error);
        }
      }

      success = "Successfully reset to DRI defaults! All custom goal sets have been deleted.";
      setTimeout(() => success = "", 5000);
      
      // Reload everything to reflect the changes
      await Promise.all([loadGoals(), loadGoalSets()]);
    } catch (err) {
      error = `Failed to reset to DRI defaults: ${parseErrorMessage(err)}`;
      console.error("Reset to DRI error:", err);
    } finally {
      showResetDRIModal = false;
    }
  }

  function cancelResetToDRIDefaults() {
    showResetDRIModal = false;
  }

  function openEditModal(goalName: string) {
    editingGoalName = goalName;
    
    // Set the goal name in the modal
    if (goalName === "New Goal") {
      customName = "";
    } else {
      customName = goalName;
    }
    
    // Reset custom targets - will fall back to current values via getNutrientValue()
    customTargets = {};
    
    showEditModal = true;
  }

  function closeEditModal() {
    showEditModal = false;
    editingGoalName = "";
    customName = "";
    customTargets = {};
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

      // Use custom goal name from modal
      const goalName = customName.trim();

      // Prepare the request payload using only the customTargets that have been modified
      const payload: any = {
        name: goalName || undefined,
        overrides: customTargets
      };

      const response = await apiClient.PUT("/goals", {
        body: payload
      });

      if (response.error) {
        throw response.error;
      }

      success = "Goals saved successfully!";
      await Promise.all([loadGoals(), loadGoalSets()]); // Reload to get updated data
      
      // Close modal
      closeEditModal();
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
            <h2 class="card-title flex items-center gap-2">
              🎯 Current Goals
              <div class="badge badge-primary badge-sm">
                {goals.source === "custom" ? (goals.custom_name || "Custom") : "DRI"}
              </div>
            </h2>
            
            <div class="space-y-4">
              <!-- Key Macros Display -->
              <div class="grid grid-cols-2 gap-3">
                <div class="stat bg-base-100 rounded-box p-3">
                  <div class="stat-title text-xs">Calories</div>
                  <div class="stat-value text-lg">{goals.targets.calories || 2000}</div>
                  <div class="stat-desc text-xs">kcal/day</div>
                </div>
                <div class="stat bg-base-100 rounded-box p-3">
                  <div class="stat-title text-xs">Protein</div>
                  <div class="stat-value text-lg">{goals.targets.protein_g || 0}</div>
                  <div class="stat-desc text-xs">grams/day</div>
                </div>
                <div class="stat bg-base-100 rounded-box p-3">
                  <div class="stat-title text-xs">Carbs</div>
                  <div class="stat-value text-lg">{goals.targets.total_carbs_g || 0}</div>
                  <div class="stat-desc text-xs">grams/day</div>
                </div>
                <div class="stat bg-base-100 rounded-box p-3">
                  <div class="stat-title text-xs">Fat</div>
                  <div class="stat-value text-lg">{goals.targets.total_fat_g || 0}</div>
                  <div class="stat-desc text-xs">grams/day</div>
                </div>
              </div>

              <!-- Profile Info -->
              <div class="stats bg-base-100 shadow-sm">
                <div class="stat">
                  <div class="stat-title">Profile</div>
                  <div class="stat-value text-sm">
                    {goals.life_stage.sex} • {goals.life_stage.age_bracket}
                  </div>
                  <div class="stat-desc">Demographic info</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Available Goals -->
        <div class="card bg-base-200 shadow-lg">
          <div class="card-body p-6">
            <h2 class="card-title flex items-center gap-2">
              🏆 Goals              
              {#if isProUser}
                <div class="flex gap-2 ml-auto">
                  {#if goalSets.length > 0}
                    <button 
                      class="btn btn-warning btn-xs"
                      on:click={resetToDRIDefaults}
                      title="Delete all custom goals to return to DRI defaults"
                    >
                      🔄 Reset to DRI
                    </button>
                  {/if}
                  <button 
                    class="btn btn-outline btn-xs"
                    on:click={() => openEditModal("New Goal")}
                  >
                    + New
                  </button>
                </div>
              {:else if isFreeUser}
                <div class="ml-auto">
                  <button class="btn btn-primary btn-xs" disabled>
                    Upgrade to Pro for Custom Goals
                  </button>
                </div>
              {/if}
            </h2>

            {#if isFreeUser}
              <!-- Free Tier: Show DRI information -->
              <div class="bg-info/10 p-4 rounded-lg">
                <div class="flex items-start gap-3">
                  <div class="badge badge-info">Free</div>
                  <div>
                    <h3 class="font-semibold text-sm">Using DRI Nutrition Guidelines</h3>
                    <p class="text-sm text-base-content/70 mt-1">
                      Your nutrition targets are based on Dietary Reference Intakes (DRI) tailored to your profile.
                      <a 
                        href="https://www.nal.usda.gov/human-nutrition-and-food-safety/dietary-guidance" 
                        target="_blank" 
                        rel="noopener noreferrer"
                        class="link link-info"
                      >
                        Learn more about DRI →
                      </a>
                    </p>
                    <div class="mt-3">
                      <button class="btn btn-primary btn-sm">
                        Upgrade to Pro to Create Custom Goals
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            {:else if isProUser}
              <!-- Pro Tier: Show goal management -->
              {#if loadingGoalSets}
                <div class="text-center py-4">
                  <span class="loading loading-spinner loading-sm"></span>
                  <p class="text-sm text-base-content/70 mt-2">Loading goal sets...</p>
                </div>
              {:else if goalSets.length > 0}
                <div class="space-y-2">
                  {#each goalSets as goalSet}
                    <div class="flex items-center justify-between p-3 rounded-lg {goalSet.name === activeGoalName ? 'bg-primary/10 border border-primary/20' : 'bg-base-100'}">
                      <div class="flex items-center gap-3">
                        <div class="flex flex-col">
                          <span class="font-medium">{goalSet.name}</span>
                          {#if goalSet.name === activeGoalName}
                            <span class="badge badge-primary badge-xs">Active</span>
                          {/if}
                        </div>
                      </div>
                      
                      <div class="flex gap-2">
                        {#if goalSet.name !== activeGoalName}
                          <button 
                            class="btn btn-primary btn-xs"
                            on:click={() => switchToGoal(goalSet.name)}
                          >
                            Activate
                          </button>
                        {/if}
                        <button 
                          class="btn btn-outline btn-xs"
                          on:click={() => openEditModal(goalSet.name)}
                        >
                          Edit
                        </button>
                        <button 
                          class="btn btn-error btn-xs"
                          on:click={() => deleteGoalSet(goalSet.name)}
                          title="Delete this goal set"
                        >
                          ×
                        </button>
                      </div>
                    </div>
                  {/each}
                </div>
              {:else}
                <!-- Pro user with no custom goals - show DRI + create option -->
                <div class="space-y-4">
                  <div class="bg-info/10 p-4 rounded-lg">
                    <div class="flex items-start gap-3">
                      <div class="badge badge-info">DRI</div>
                      <div>
                        <h3 class="font-semibold text-sm">Using DRI Nutrition Guidelines</h3>
                        <p class="text-sm text-base-content/70 mt-1">
                          You're currently using <a 
                            href="https://www.nal.usda.gov/human-nutrition-and-food-safety/dietary-guidance" 
                            target="_blank" 
                            rel="noopener noreferrer"
                            class="link link-info"
                          >Dietary Reference Intakes (DRI)</a> based on your profile.
                          As a <span class="font-semibold">Pro</span> user, you can create custom goal sets to override specific targets.
                        </p>
                      </div>
                    </div>
                  </div>
                  <div class="text-center py-4">
                    <p class="text-sm text-base-content/70">Ready to create your first custom goal set?</p>
                    <button 
                      class="btn btn-primary btn-sm mt-2"
                      on:click={() => openEditModal("My Custom Goals")}
                    >
                      Create Your First Goal Set
                    </button>
                  </div>
                </div>
              {/if}
            {/if}
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

      </div>

<!-- Edit Goal Modal -->
{#if showEditModal}
  <div class="modal modal-open">
    <div class="modal-box max-w-2xl">
      <h3 class="font-bold text-lg mb-4">
        {editingGoalName === "New Goal" ? "Create New Goal" : `Edit ${editingGoalName}`}
      </h3>
      
      <div class="alert alert-info mb-4">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
        <div>
          <div class="text-sm">Need help setting your nutrition goals?</div>
          <div class="text-xs mt-1">
            Use the official <a 
              href="https://www.nal.usda.gov/human-nutrition-and-food-safety/dri-calculator" 
              target="_blank" 
              rel="noopener noreferrer"
              class="link link-info font-semibold"
            >USDA DRI Calculator</a> to determine appropriate targets for your age, sex, and activity level.
          </div>
        </div>
      </div>
      
      <div class="space-y-4">
        <!-- Goal Name Input -->
        <div class="form-control w-full">
          <label class="label" for="goalName">
            <span class="label-text">Goal Name</span>
          </label>
          <input 
            id="goalName"
            type="text" 
            placeholder="e.g., Bulking, Cutting, Maintenance" 
            class="input input-bordered w-full" 
            bind:value={customName}
          />
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
        </div>
      </div>

      <div class="modal-action">
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
            <span class="loading loading-spinner loading-xs"></span>
            Saving...
          {:else}
            💾 Save Goals
          {/if}
        </button>
        
        <button 
          class="btn btn-outline" 
          on:click={closeEditModal}
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
{/if}

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

<!-- Delete Goal Confirmation Modal -->
{#if showDeleteModal}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg mb-4">Delete Goal Set</h3>
      
      <div class="space-y-4">
        <p>Are you sure you want to delete the goal set <strong>"{goalToDelete}"</strong>?</p>
        <div class="alert alert-warning">
          <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.268 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
          <span>This action cannot be undone.</span>
        </div>
        {#if goalSets.length === 1}
          <div class="bg-info/10 p-3 rounded-lg">
            <p class="text-sm text-info-content">
              🧬 This is your last custom goal set. Deleting it will return you to DRI (Dietary Reference Intakes) defaults.
            </p>
          </div>
        {/if}
      </div>
      
      <div class="modal-action">
        <button 
          class="btn btn-ghost"
          on:click={cancelDeleteGoalSet}
        >
          Cancel
        </button>
        <button 
          class="btn btn-error"
          on:click={confirmDeleteGoalSet}
        >
          Delete Goal Set
        </button>
      </div>
    </div>
    <div 
      class="modal-backdrop" 
      on:click={cancelDeleteGoalSet}
      on:keydown={(e) => e.key === 'Escape' && cancelDeleteGoalSet()}
      role="button" 
      tabindex="0"
      aria-label="Close modal"
    ></div>
  </div>
{/if}

<!-- Reset to DRI Defaults Confirmation Modal -->
{#if showResetDRIModal}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg mb-4">Reset to DRI Defaults</h3>
      
      <div class="space-y-4">
        <p>Are you sure you want to delete <strong>ALL</strong> custom goal sets and return to DRI defaults?</p>
        <div class="alert alert-warning">
          <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.268 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
          <span>This will delete all {goalSets.length} custom goal sets. This action cannot be undone.</span>
        </div>
        <div class="bg-info/10 p-3 rounded-lg">
          <p class="text-sm text-info-content">
            🧬 You'll return to DRI (Dietary Reference Intakes) defaults based on your profile demographics.
          </p>
        </div>
      </div>
      
      <div class="modal-action">
        <button 
          class="btn btn-ghost"
          on:click={cancelResetToDRIDefaults}
        >
          Cancel
        </button>
        <button 
          class="btn btn-warning"
          on:click={confirmResetToDRIDefaults}
        >
          🔄 Reset to DRI Defaults
        </button>
      </div>
    </div>
    <div 
      class="modal-backdrop" 
      on:click={cancelResetToDRIDefaults}
      on:keydown={(e) => e.key === 'Escape' && cancelResetToDRIDefaults()}
      role="button" 
      tabindex="0"
      aria-label="Close modal"
    ></div>
  </div>
{/if}

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
