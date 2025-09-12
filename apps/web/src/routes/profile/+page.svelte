<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { onMount } from "svelte"
  import { dev } from "$app/environment"
  import { goto } from "$app/navigation"
  import { toast } from "$lib/stores/toast"
  import { signOut } from "$lib/auth/store"
  import Toast from "$lib/components/Toast.svelte"
  import InfoButton from "$lib/components/InfoButton.svelte"
  import FormField from "$lib/components/FormField.svelte"
  import Label from "$lib/components/Label.svelte"
  import FormSelect from "$lib/components/FormSelect.svelte"
  import ConfirmModal from "$lib/components/ConfirmModal.svelte"
  import TagIcon from "$lib/components/icons/Tag.svelte"
  import StarIcon from "$lib/components/icons/Star.svelte"
  import TrophyIcon from "$lib/components/icons/Trophy.svelte"
  import IdentificationIcon from "$lib/components/icons/Identification.svelte"
  import CalendarDaysIcon from "$lib/components/icons/calendar-days.svelte"
  import KeyIcon from "$lib/components/icons/Key.svelte"
  import UserIcon from "$lib/components/icons/User.svelte"
  import ArrowPath from "$lib/components/icons/ArrowPath.svelte"
  import { getStorageJSON, setStorageJSON } from "$lib/utils/secure-storage"
  import {
    parseErrorMessage,
    formatErrorForUser,
  } from "$lib/utils/error-handling"
  import { getAppName } from "$lib/utils/app-info"
  import type { paths } from "$lib/api/schema"

  // Get app name from runtime environment
  $: appName = getAppName()

  type GoalsResponse =
    paths["/goals"]["get"]["responses"]["200"]["content"]["application/json"]
  type Goals = GoalsResponse["goals"]
  type User = GoalsResponse["user"]
  type BiometricsResponse =
    paths["/biometrics"]["get"]["responses"]["200"]["content"]["application/json"]
  type UserBiometrics =
    paths["/biometrics"]["get"]["responses"]["200"]["content"]["application/json"]["biometrics"]
  type UpdateBiometricsRequest =
    paths["/biometrics"]["put"]["requestBody"]["content"]["application/json"]
  type LabelsResponse =
    paths["/labels"]["get"]["responses"]["200"]["content"]["application/json"]
  type Label = LabelsResponse["labels"][0]
  type EventTypesResponse =
    paths["/event-types"]["get"]["responses"]["200"]["content"]["application/json"]
  type EventType = EventTypesResponse["event_types"][0]

  let goals: Goals | null = null
  let user: User | null = null
  let biometrics: UserBiometrics | null = null
  let calculatedMetrics: BiometricsResponse["calculated_metrics"] | null = null
  let labels: Label[] = []
  let eventTypes: EventType[] = []
  let labelsLoading = false
  let eventTypesLoading = false
  let loading = true
  let biometricsLoading = false
  let error = ""
  let biometricsError = ""
  let success = ""
  let biometricsSuccess = ""

  // Reactive statements for user tier
  $: isProUser = user?.subscription_tier === "pro"
  $: isFreeUser = user?.subscription_tier === "free"
  let saving = false
  let savingBiometrics = false

  // Age visibility state with secure client-side storage
  let showAge = true

  // Load age visibility preference from secure localStorage
  if (typeof window !== "undefined") {
    showAge = getStorageJSON("noot-show-age", true)
  }

  // Function to toggle age visibility and save preference securely
  function toggleAgeVisibility() {
    showAge = !showAge
    if (typeof window !== "undefined") {
      setStorageJSON("noot-show-age", showAge)
    }
  }

  // Form state
  let customGoalName = ""
  let selectedUnits = "metric" // Track unit system selection
  let showImperialModal = false

  // UI state for modal
  let showEditModal = false
  let editingGoalName = ""

  // Delete confirmation modal state
  let showDeleteModal = false
  let goalToDelete = ""

  // Reset DRI confirmation modal state
  let showResetDRIModal = false

  // Disabled nutrients modal state
  let showDisabledNutrientsModal = false
  let tempDisabledNutrients: string[] = []

  // All available nutrients for disabling
  type NutrientOption = {
    key: string;
    label: string;
    category: string;
  }

  const ALL_NUTRIENTS: NutrientOption[] = [
    // Macronutrients
    { key: "calories", label: "Calories", category: "Macronutrients" },
    { key: "protein_g", label: "Protein", category: "Macronutrients" },
    { key: "total_carbs_g", label: "Total Carbohydrates", category: "Macronutrients" },
    { key: "total_fat_g", label: "Total Fat", category: "Macronutrients" },
    { key: "saturated_fat_g", label: "Saturated Fat", category: "Macronutrients" },
    { key: "trans_fat_g", label: "Trans Fat", category: "Macronutrients" },
    { key: "monounsaturated_fat_g", label: "Monounsaturated Fat", category: "Macronutrients" },
    { key: "polyunsaturated_fat_g", label: "Polyunsaturated Fat", category: "Macronutrients" },
    { key: "omega3_ala_g", label: "Omega-3 ALA", category: "Macronutrients" },
    { key: "omega3_epa_g", label: "Omega-3 EPA", category: "Macronutrients" },
    { key: "omega3_dha_g", label: "Omega-3 DHA", category: "Macronutrients" },
    { key: "omega6_g", label: "Omega-6", category: "Macronutrients" },
    { key: "dietary_fiber_g", label: "Dietary Fiber", category: "Macronutrients" },
    { key: "total_sugars_g", label: "Total Sugars", category: "Macronutrients" },
    { key: "added_sugars_g", label: "Added Sugars", category: "Macronutrients" },
    { key: "cholesterol_mg", label: "Cholesterol", category: "Macronutrients" },
    { key: "sodium_mg", label: "Sodium", category: "Macronutrients" },
    { key: "alcohol_g", label: "Alcohol", category: "Macronutrients" },

    // B-Complex Vitamins
    { key: "thiamine_mg", label: "Thiamine (B1)", category: "B-Complex Vitamins" },
    { key: "riboflavin_mg", label: "Riboflavin (B2)", category: "B-Complex Vitamins" },
    { key: "niacin_mg", label: "Niacin (B3)", category: "B-Complex Vitamins" },
    { key: "vitamin_b6_mg", label: "Vitamin B6", category: "B-Complex Vitamins" },
    { key: "folate_mcg", label: "Folate", category: "B-Complex Vitamins" },
    { key: "vitamin_b12_mcg", label: "Vitamin B12", category: "B-Complex Vitamins" },
    { key: "biotin_mcg", label: "Biotin", category: "B-Complex Vitamins" },
    { key: "pantothenic_acid_mg", label: "Pantothenic Acid", category: "B-Complex Vitamins" },

    // Fat-Soluble Vitamins
    { key: "vitamin_a_mcg", label: "Vitamin A", category: "Fat-Soluble Vitamins" },
    { key: "vitamin_d_mcg", label: "Vitamin D", category: "Fat-Soluble Vitamins" },
    { key: "vitamin_e_mg", label: "Vitamin E", category: "Fat-Soluble Vitamins" },
    { key: "vitamin_k_mcg", label: "Vitamin K", category: "Fat-Soluble Vitamins" },

    // Water-Soluble Vitamins
    { key: "vitamin_c_mg", label: "Vitamin C", category: "Water-Soluble Vitamins" },
    { key: "choline_mg", label: "Choline", category: "Water-Soluble Vitamins" },

    // Essential Minerals
    { key: "calcium_mg", label: "Calcium", category: "Essential Minerals" },
    { key: "iron_mg", label: "Iron", category: "Essential Minerals" },
    { key: "magnesium_mg", label: "Magnesium", category: "Essential Minerals" },
    { key: "phosphorus_mg", label: "Phosphorus", category: "Essential Minerals" },
    { key: "potassium_mg", label: "Potassium", category: "Essential Minerals" },
    { key: "zinc_mg", label: "Zinc", category: "Essential Minerals" },
    { key: "copper_mg", label: "Copper", category: "Essential Minerals" },
    { key: "manganese_mg", label: "Manganese", category: "Essential Minerals" },
    { key: "selenium_mcg", label: "Selenium", category: "Essential Minerals" },
    { key: "iodine_mcg", label: "Iodine", category: "Essential Minerals" },
    { key: "molybdenum_mcg", label: "Molybdenum", category: "Essential Minerals" },
    { key: "chromium_mcg", label: "Chromium", category: "Essential Minerals" },
    { key: "fluoride_mg", label: "Fluoride", category: "Essential Minerals" },
    { key: "chloride_mg", label: "Chloride", category: "Essential Minerals" },

    // Other Compounds
    { key: "caffeine_mg", label: "Caffeine", category: "Other Compounds" },
    { key: "creatine_mg", label: "Creatine", category: "Other Compounds" },
  ]

  // Sign out state
  let signingOut = false

  // Goal sets data
  let goalSets: Array<{
    name: string
    category: string
    created_at: string
    updated_at: string
  }> = []
  let activeGoalName = ""
  let loadingGoalSets = false
  let savingGoals = false

  // Goal data for modal editing - now using dynamic approach with separate targets and upper limits
  let customTargets: Record<string, number> = {}
  let customUpperLimits: Record<string, number> = {}
  let customName = ""
  let customCategory: "weight" | "fitness" | "health" | "custom" = "custom"

  // Biometrics form state
  let birthDate = ""
  let sex: "male" | "female" | "other" | "prefer_not_to_say" =
    "prefer_not_to_say"
  let heightCm = ""
  let weightKg = ""
  let activityLevel:
    | "sedentary"
    | "lightly_active"
    | "moderately_active"
    | "very_active"
    | "extra_active" = "lightly_active"

  onMount(async () => {
    // Load all data in parallel for better performance
    await Promise.all([
      loadGoals(),
      loadBiometrics(),
      loadGoalSets(),
      loadLabels(),
      loadEventTypes(),
    ])
  })

  function resetToDefaults() {
    customTargets = {}
    customUpperLimits = {}
    customName = ""
    customCategory = "custom"
  }

  // All nutrients to display organized by category for complete coverage
  // Include ALL nutrients that the system knows about, allowing both targets and upper limits
  $: allKnownNutrients = goals
    ? [
        ...Object.keys(goals.targets || {}),
        ...Object.keys(goals.upper_limits || {}),
      ]
        .filter((key, index, array) => array.indexOf(key) === index) // Remove duplicates
        .sort()
    : []

  $: editableNutrients = allKnownNutrients
    .map((key) => ({
      key,
      label: formatNutrientName(key),
      unit: goals?.units[key] || "",
      hasTarget: !!goals?.targets?.[key],
      hasUpperLimit: !!goals?.upper_limits?.[key],
    }))
    .sort((a, b) => a.label.localeCompare(b.label))

  function formatNutrientName(key: string): string {
    return (
      key
        .replace(/_/g, " ")
        .replace(/\b\w/g, (l) => l.toUpperCase())
        // Remove unit suffixes since they're shown separately
        .replace(/ Mcg$/, "")
        .replace(/ Mg$/, "")
        .replace(/ G$/, "")
    )
  }

  function getNutrientValue(key: string): number {
    return customTargets[key] || goals?.targets[key] || 0
  }

  function getUpperLimitValue(key: string): number {
    return customUpperLimits[key] || goals?.upper_limits?.[key] || 0
  }

  function updateNutrient(key: string, value: number) {
    if (value <= 0) {
      delete customTargets[key]
    } else {
      customTargets[key] = value
    }
    customTargets = { ...customTargets } // Trigger reactivity
  }

  function updateUpperLimit(key: string, value: number) {
    if (value < 0) {
      delete customUpperLimits[key]
    } else {
      customUpperLimits[key] = value
    }
    customUpperLimits = { ...customUpperLimits } // Trigger reactivity
  }

  async function loadGoals() {
    try {
      loading = true
      const response = await apiClient.GET("/goals")

      if (response.error) {
        throw new Error(`API Error: ${response.error}`)
      }

      goals = response.data.goals
      user = response.data.user
    } catch (err) {
      toast.error(`Failed to load goals: ${err}`)
      console.error("Goals error:", err)
    } finally {
      loading = false
    }
  }

  // Called when the active goal changes
  async function handleGoalChanged() {
    await Promise.all([loadGoals(), loadGoalSets()])
    toast.success("Active goal switched successfully!")
  }

  async function loadGoalSets() {
    try {
      loadingGoalSets = true

      const response = await apiClient.GET("/goals/sets")

      if (response.error) {
        throw new Error(`API Error: ${response.error}`)
      }

      goalSets = response.data.goal_sets || []
      activeGoalName = response.data.active_goal_name || ""
      
      // Update user data if not already loaded (for parallel loading)
      if (!user) {
        user = response.data.user
      }
    } catch (err) {
      console.error("Goal sets error:", err)
      // Don't show error if user just doesn't have multiple goals yet
    } finally {
      loadingGoalSets = false
    }
  }

  async function switchToGoal(goalName: string) {
    if (goalName === activeGoalName) return

    if (!isProUser) {
      toast.error("Pro subscription required for goal set management")
      return
    }

    try {
      const response = await apiClient.PUT("/goals/active", {
        body: { name: goalName },
      })

      if (response.error) {
        throw response.error
      }

      await handleGoalChanged()
    } catch (err) {
      toast.error(`Failed to switch goal: ${err}`)
      console.error("Switch goal error:", err)
    }
  }

  async function deleteGoalSet(goalName: string) {
    if (!goalName) return

    if (!isProUser) {
      toast.error("Pro subscription required for goal set management")
      return
    }

    // Show modal instead of using confirm()
    goalToDelete = goalName
    showDeleteModal = true
  }

  async function confirmDeleteGoalSet() {
    if (!goalToDelete) return

    if (!isProUser) {
      toast.error("Pro subscription required for goal set management")
      showDeleteModal = false
      goalToDelete = ""
      return
    }

    try {
      const isLastGoal = goalSets.length === 1
      const isActiveGoal = goalToDelete === activeGoalName

      const response = await apiClient.DELETE("/goals/sets/{name}", {
        params: {
          path: { name: goalToDelete },
        },
      })

      if (response.error) {
        throw response.error
      }

      if (isLastGoal && isActiveGoal) {
        toast.success(
          `Goal set "${goalToDelete}" deleted successfully! You're now using DRI nutrition defaults.`,
        )
      } else {
        toast.success(`Goal set "${goalToDelete}" deleted successfully!`)
      }

      // Reload both goals and goal sets to update the UI and show DRI fallback
      await Promise.all([loadGoals(), loadGoalSets()])
    } catch (err) {
      toast.error(`Failed to delete goal set: ${parseErrorMessage(err)}`)
      console.error("Delete goal error:", err)
    } finally {
      // Close modal and reset state
      showDeleteModal = false
      goalToDelete = ""
    }
  }

  function cancelDeleteGoalSet() {
    showDeleteModal = false
    goalToDelete = ""
  }

  async function resetToDRIDefaults() {
    showResetDRIModal = true
  }

  async function confirmResetToDRIDefaults() {
    try {
      // Delete ALL custom goal sets, including the active one
      // We'll delete them all in one go to avoid issues with active goal switching
      for (const goalSet of goalSets) {
        const response = await apiClient.DELETE("/goals/sets/{name}", {
          params: {
            path: { name: goalSet.name },
          },
        })

        if (response.error) {
          console.warn(
            `Failed to delete goal set ${goalSet.name}:`,
            response.error,
          )
        }
      }

      toast.success(
        "Successfully reset to DRI defaults! All custom goal sets have been deleted.",
        5000,
      )

      // Reload everything to reflect the changes
      await Promise.all([loadGoals(), loadGoalSets()])
    } catch (err) {
      toast.error(`Failed to reset to DRI defaults: ${parseErrorMessage(err)}`)
      console.error("Reset to DRI error:", err)
    } finally {
      showResetDRIModal = false
    }
  }

  function cancelResetToDRIDefaults() {
    showResetDRIModal = false
  }

  function openEditModal(goalName: string) {
    editingGoalName = goalName

    // Set the goal name in the modal
    if (goalName === "New Goal") {
      customName = ""
      customCategory = "custom"
    } else {
      customName = goalName
      // Try to find existing goal to get its category
      const existingGoal = goalSets.find(g => g.name === goalName)
      const goalCategory = existingGoal?.category
      customCategory = (goalCategory === "weight" || goalCategory === "fitness" || goalCategory === "health" || goalCategory === "custom") 
        ? goalCategory as "weight" | "fitness" | "health" | "custom"
        : "custom"
    }

    // Reset custom targets - will fall back to current values via getNutrientValue()
    customTargets = {}
    customUpperLimits = {}

    showEditModal = true
  }

  function closeEditModal() {
    showEditModal = false
    editingGoalName = ""
    customName = ""
    customCategory = "custom"
    customTargets = {}
    customUpperLimits = {}
  }

  function openDisabledNutrientsModal() {
    // Initialize with current disabled nutrients
    tempDisabledNutrients = [...(goals?.disabled_nutrients || [])]
    showDisabledNutrientsModal = true
  }

  function closeDisabledNutrientsModal() {
    showDisabledNutrientsModal = false
    tempDisabledNutrients = []
  }

  function toggleDisabledNutrient(nutrientKey: string) {
    if (tempDisabledNutrients.includes(nutrientKey)) {
      tempDisabledNutrients = tempDisabledNutrients.filter(n => n !== nutrientKey)
    } else {
      tempDisabledNutrients = [...tempDisabledNutrients, nutrientKey]
    }
  }

  async function saveDisabledNutrients() {
    if (!isProUser) {
      toast.error("Pro subscription required for custom goals")
      return
    }

    try {
      saving = true

      // Get current active goal or create a new one
      let goalName = activeGoalName || "Custom Goals"
      
      // Prepare the request payload
      const payload: any = {
        name: goalName,
        category: "custom",
        disabled_nutrients: tempDisabledNutrients
      }

      // Include existing targets and upper limits if they exist
      const currentGoal = goalSets.find(g => g.name === goalName)
      if (currentGoal) {
        // We need to get the current goal details to preserve targets/upper_limits
        // For now, just include the disabled nutrients
      }

      const response = await apiClient.PUT("/goals", {
        body: payload,
      })

      if (response.error) {
        throw response.error
      }

      toast.success("Disabled nutrients updated successfully!")
      await Promise.all([loadGoals(), loadGoalSets()]) // Reload to get updated data
      closeDisabledNutrientsModal()
    } catch (err) {
      toast.error(formatErrorForUser(err))
      if (dev) {
        console.error("Save disabled nutrients error:", err)
      }
    } finally {
      saving = false
    }
  }

  async function loadBiometrics() {
    try {
      biometricsLoading = true

      const response = await apiClient.GET("/biometrics")

      if (response.error) {
        throw new Error(`API Error: ${response.error}`)
      }

      biometrics = response.data.biometrics
      calculatedMetrics = response.data.calculated_metrics

      // Initialize form with current values
      if (biometrics) {
        birthDate = biometrics.birth_date || ""
        sex = biometrics.sex || "prefer_not_to_say"
        heightCm = biometrics.height_cm?.toString() || ""
        weightKg = biometrics.weight_kg?.toString() || ""
        activityLevel = biometrics.activity_level || "lightly_active"
      }
    } catch (err) {
      // Don't show error if biometrics just don't exist yet
      if (
        !err?.toString().includes("404") &&
        !err?.toString().includes("not found")
      ) {
        toast.error(`Failed to load biometrics: ${err}`)
        console.error("Biometrics error:", err)
      }
    } finally {
      biometricsLoading = false
    }
  }

  async function loadLabels() {
    try {
      labelsLoading = true

      const response = await apiClient.GET("/labels")

      if (response.error) {
        throw new Error(`API Error: ${response.error}`)
      }

      labels = response.data?.labels || []
    } catch (err) {
      // Don't show error for labels - they're not critical
      console.warn("Failed to load labels:", err)
      labels = []
    } finally {
      labelsLoading = false
    }
  }

  async function loadEventTypes() {
    try {
      eventTypesLoading = true

      const response = await apiClient.GET("/event-types")

      if (response.error) {
        throw new Error(`API Error: ${response.error}`)
      }

      eventTypes = response.data?.event_types || []
    } catch (err) {
      // Don't show error for event types - they're not critical
      console.warn("Failed to load event types:", err)
      eventTypes = []
    } finally {
      eventTypesLoading = false
    }
  }

  async function saveCustomGoals() {
    if (!goals) return

    if (!isProUser) {
      toast.error("Pro subscription required for custom goals")
      return
    }

    try {
      saving = true

      // Use custom goal name from modal
      const goalName = customName.trim()
      
      // Validate required fields
      if (!goalName) {
        toast.error("Goal name is required")
        return
      }

      // Prepare the request payload using the new structure
      const payload: any = {
        name: goalName,
        category: customCategory,
      }

      // Only include targets and upper_limits if they have values
      if (Object.keys(customTargets).length > 0) {
        payload.targets = customTargets
      }

      if (Object.keys(customUpperLimits).length > 0) {
        payload.upper_limits = customUpperLimits
      }

      // Validate that we have at least some overrides
      if (!payload.targets && !payload.upper_limits) {
        toast.error("At least one target or upper limit must be set")
        return
      }

      const response = await apiClient.PUT("/goals", {
        body: payload,
      })

      if (response.error) {
        throw response.error
      }

      toast.success("Goals saved successfully!")
      await Promise.all([loadGoals(), loadGoalSets()]) // Reload to get updated data

      // Close modal
      closeEditModal()
    } catch (err) {
      toast.error(formatErrorForUser(err))
      if (dev) {
        console.error("Save error details:", err)
      }
    } finally {
      saving = false
    }
  }

  async function saveBiometrics() {
    try {
      savingBiometrics = true

      // Prepare the request payload
      const payload: UpdateBiometricsRequest = {}

      if (birthDate.trim()) {
        payload.birth_date = birthDate.trim()
      }

      if (sex && sex !== "prefer_not_to_say") {
        payload.sex = sex
      }

      if (heightCm && !isNaN(parseFloat(heightCm.toString()))) {
        payload.height_cm = parseFloat(heightCm.toString())
      }

      if (weightKg && !isNaN(parseFloat(weightKg.toString()))) {
        payload.weight_kg = parseFloat(weightKg.toString())
      }

      if (activityLevel) {
        payload.activity_level = activityLevel
      }

      const response = await apiClient.PUT("/biometrics", {
        body: payload,
      })

      if (response.error) {
        throw response.error
      }

      toast.success("Biometrics saved successfully!")
      await Promise.all([loadBiometrics(), loadGoals()]) // Reload both since goals may have changed
    } catch (err) {
      toast.error(formatErrorForUser(err))
      if (dev) {
        console.error("Biometrics save error details:", err)
      }
    } finally {
      savingBiometrics = false
    }
  }

  async function deleteBiometrics() {
    if (
      !confirm(
        "Are you sure you want to delete all your biometric data? This action cannot be undone.",
      )
    ) {
      return
    }

    try {
      savingBiometrics = true

      const response = await apiClient.DELETE("/biometrics")

      if (response.error) {
        throw response.error
      }

      toast.success("Biometrics deleted successfully!")

      // Clear form
      birthDate = ""
      sex = "prefer_not_to_say"
      heightCm = ""
      weightKg = ""
      activityLevel = "lightly_active"

      await Promise.all([loadBiometrics(), loadGoals()]) // Reload both since goals may have changed
    } catch (err) {
      toast.error(formatErrorForUser(err))
      if (dev) {
        console.error("Biometrics delete error details:", err)
      }
    } finally {
      savingBiometrics = false
    }
  }

  function openImperialModal() {
    showImperialModal = true
  }

  function closeImperialModal() {
    showImperialModal = false
    selectedUnits = "metric" // Force selection back to metric
  }

  async function handleSignOut() {
    try {
      signingOut = true

      const { error } = await signOut()

      if (error) {
        toast.error("Failed to sign out. Please try again.")
        if (dev) {
          console.error("Sign out error:", error)
        }
        return
      }

      toast.success("Successfully signed out!")

      // Redirect to login page
      await goto("/login")
    } catch (err) {
      toast.error("An unexpected error occurred during sign out.")
      if (dev) {
        console.error("Sign out error:", err)
      }
    } finally {
      signingOut = false
    }
  }

  function getActivityLevelDisplay(level: string): string {
    const activityLevels: Record<string, string> = {
      sedentary: "L1",
      lightly_active: "L2",
      moderately_active: "L3",
      very_active: "L4",
      extra_active: "L5",
    }
    return activityLevels[level] || level
  }

  function getActivityLevelDescription(level: string): string {
    const descriptions: Record<string, string> = {
      sedentary: "Sedentary (little/no exercise)",
      lightly_active: "Lightly Active (1-3 days/week)",
      moderately_active: "Moderately Active (3-5 days/week)",
      very_active: "Very Active (6-7 days/week)",
      extra_active: "Extra Active (very hard exercise daily)",
    }
    return descriptions[level] || level
  }
</script>

<svelte:head>
  <title>Profile - {appName}</title>
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-8 max-w-6xl">
    <!-- Header -->
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-base-content flex items-center gap-3">
        <UserIcon className="w-8 h-8" />
        Profile
      </h1>
      <p class="text-base-content-lighter mt-2">
        Customize your profile, nutrition goals, and preferences
      </p>
    </div>

    {#if loading}
      <div class="text-center py-12">
        <span class="loading loading-spinner loading-lg"></span>
        <p class="text-base-content/70 mt-4">Loading your profile...</p>
      </div>
    {:else if goals}
      <!-- Main Profile Grid -->
      <div class="grid gap-8 xl:grid-cols-3 lg:grid-cols-2 md:grid-cols-1">
        <!-- Current Goals Overview -->
        <div class="card bg-base-200 shadow-lg">
          <div class="card-body p-6">
            <h2 class="card-title flex items-center gap-2">
              <StarIcon className="w-5 h-5" /> Current Goals
              <div class="badge badge-primary badge-sm">
                {goals.source === "custom"
                  ? goals.custom_name || "Custom"
                  : "DRI"}
              </div>
            </h2>

            <div class="space-y-4">
              <!-- Key Macros Display -->
              <div class="grid grid-cols-2 gap-3">
                <div class="stat bg-base-100 rounded-box p-3">
                  <div class="stat-title text-xs">Calories</div>
                  <div class="stat-value text-lg">
                    {goals.targets.calories || 2000}
                  </div>
                  <div class="stat-desc text-xs">kcal/day</div>
                </div>
                <div class="stat bg-base-100 rounded-box p-3">
                  <div class="stat-title text-xs">Protein</div>
                  <div class="stat-value text-lg">
                    {goals.targets.protein_g || 0}
                  </div>
                  <div class="stat-desc text-xs">grams/day</div>
                </div>
                <div class="stat bg-base-100 rounded-box p-3">
                  <div class="stat-title text-xs">Carbs</div>
                  <div class="stat-value text-lg">
                    {goals.targets.total_carbs_g || 0}
                  </div>
                  <div class="stat-desc text-xs">grams/day</div>
                </div>
                <div class="stat bg-base-100 rounded-box p-3">
                  <div class="stat-title text-xs">Fat</div>
                  <div class="stat-value text-lg">
                    {goals.targets.total_fat_g || 0}
                  </div>
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
              <TrophyIcon className="w-5 h-5" /> Goals
              {#if isProUser}
                <div class="flex gap-2 ml-auto">
                  {#if goalSets.length > 0}
                    <button
                      class="btn btn-warning btn-sm min-h-[44px]"
                      on:click={resetToDRIDefaults}
                      title="Delete all custom goals to return to DRI defaults"
                    >
                      <ArrowPath className="w-4 h-4 mr-1" />
                      Reset to DRI
                    </button>
                  {/if}
                  <button
                    class="btn btn-outline btn-sm min-h-[44px]"
                    on:click={() => openEditModal("New Goal")}
                  >
                    + New
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
                    <h3 class="font-semibold text-sm">
                      Using DRI Nutrition Guidelines
                    </h3>
                    <p class="text-sm text-base-content/70 mt-1">
                      Your nutrition targets are based on Dietary Reference
                      Intakes (DRI) tailored to your profile.
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
                  <p class="text-sm text-base-content/70 mt-2">
                    Loading goal sets...
                  </p>
                </div>
              {:else if goalSets.length > 0}
                <div class="space-y-2">
                  {#each goalSets as goalSet}
                    <div
                      class="flex items-center justify-between p-3 rounded-lg {goalSet.name ===
                      activeGoalName
                        ? 'bg-primary/10 border border-primary/20'
                        : 'bg-base-100'}"
                    >
                      <div class="flex items-center gap-3">
                        <div class="flex flex-col">
                          <span class="font-medium">{goalSet.name}</span>
                          {#if goalSet.name === activeGoalName}
                            <span class="badge badge-primary badge-xs"
                              >Active</span
                            >
                          {/if}
                        </div>
                      </div>

                      <div class="flex gap-2">
                        {#if goalSet.name !== activeGoalName}
                          <button
                            class="btn btn-primary btn-sm min-h-[44px]"
                            on:click={() => switchToGoal(goalSet.name)}
                          >
                            Activate
                          </button>
                        {/if}
                        <button
                          class="btn btn-outline btn-sm min-h-[44px]"
                          on:click={() => openEditModal(goalSet.name)}
                        >
                          Edit
                        </button>
                        <button
                          class="btn btn-error btn-sm min-h-[44px] min-w-[44px]"
                          on:click={() => deleteGoalSet(goalSet.name)}
                          title="Delete this goal set"
                          aria-label="Delete goal set"
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
                        <h3 class="font-semibold text-sm">
                          Using DRI Nutrition Guidelines
                        </h3>
                        <p class="text-sm text-base-content/70 mt-1">
                          You're currently using <a
                            href="https://www.nal.usda.gov/human-nutrition-and-food-safety/dietary-guidance"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="link link-info"
                            >Dietary Reference Intakes (DRI)</a
                          >
                          based on your profile. As a
                          <span class="font-semibold">Pro</span> user, you can create
                          custom goal sets to override specific targets.
                        </p>
                      </div>
                    </div>
                  </div>
                  <div class="text-center py-4">
                    <p class="text-sm text-base-content/70">
                      Ready to create your first custom goal set?
                    </p>
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

        <!-- Disabled Nutrients (Pro only) -->
        {#if isProUser}
          <div class="card bg-base-200 shadow-lg">
            <div class="card-body p-6">
              <h2 class="card-title flex items-center gap-2">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Disabled Nutrients
                <div class="badge badge-info badge-sm">Pro</div>
              </h2>
              
              <p class="text-sm text-base-content/70 mb-4">
                Hide specific nutrients from all charts, summaries, and goal tracking. Nutrients will still be tracked but not displayed.
              </p>

              {#if goals?.disabled_nutrients && goals.disabled_nutrients.length > 0}
                <div class="mb-4">
                  <h4 class="font-medium mb-2">Currently Disabled:</h4>
                  <div class="flex flex-wrap gap-2">
                    {#each goals.disabled_nutrients as nutrient}
                      <span class="badge badge-outline badge-sm">
                        {nutrient.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
                      </span>
                    {/each}
                  </div>
                </div>
              {/if}

              <button 
                class="btn btn-primary btn-sm"
                on:click={openDisabledNutrientsModal}
              >
                Manage Disabled Nutrients
              </button>
            </div>
          </div>
        {/if}

        <!-- User Biometrics -->
        <div class="card bg-base-200 shadow-lg">
          <div class="card-body p-6">
            <h2 class="card-title flex items-center gap-2">
              <IdentificationIcon className="w-5 h-5" />
              Biometrics
            </h2>

            {#if biometricsLoading}
              <div class="text-center py-4">
                <span class="loading loading-spinner loading-sm"></span>
                <p class="text-sm text-base-content/70 mt-2">
                  Loading biometrics...
                </p>
              </div>
            {:else}
              <div class="space-y-4">
                <!-- Current Biometrics Display -->
                {#if biometrics}
                  <div class="grid grid-cols-2 gap-4">
                    {#if calculatedMetrics?.bmi}
                      <div
                        class="stat bg-base-100 rounded-box p-3 tooltip tooltip-top"
                        data-tip="Body Mass Index - A measure of body fat based on height and weight"
                      >
                        <div class="stat-title text-xs">BMI</div>
                        <div class="stat-value text-lg">
                          {calculatedMetrics.bmi}
                        </div>
                      </div>
                    {/if}
                    {#if calculatedMetrics?.bmr}
                      <div
                        class="stat bg-base-100 rounded-box p-3 tooltip tooltip-top"
                        data-tip="Basal Metabolic Rate - Calories your body burns at rest for basic functions"
                      >
                        <div class="stat-title text-xs">BMR</div>
                        <div class="stat-value text-lg">
                          {Math.round(calculatedMetrics.bmr)}
                        </div>
                        <div class="stat-desc text-xs">kcal/day</div>
                      </div>
                    {/if}
                    {#if calculatedMetrics?.tdee}
                      <div
                        class="stat bg-base-100 rounded-box p-3 tooltip tooltip-top"
                        data-tip="Total Daily Energy Expenditure - Total calories burned including exercise and daily activities"
                      >
                        <div class="stat-title text-xs">TDEE</div>
                        <div class="stat-value text-lg">
                          {Math.round(calculatedMetrics.tdee)}
                        </div>
                        <div class="stat-desc text-xs">kcal/day</div>
                      </div>
                    {/if}
                    {#if biometrics.activity_level}
                      <div
                        class="stat bg-base-100 rounded-box p-3 tooltip tooltip-top"
                        data-tip={getActivityLevelDescription(
                          biometrics.activity_level,
                        )}
                      >
                        <div class="stat-title text-xs">Activity</div>
                        <div class="stat-value text-lg">
                          {getActivityLevelDisplay(biometrics.activity_level)}
                        </div>
                      </div>
                    {/if}
                    {#if calculatedMetrics?.age_years && showAge}
                      <div
                        class="stat bg-base-100 rounded-box p-3 tooltip tooltip-top"
                        data-tip="Your current age based on birth date"
                      >
                        <div class="stat-title text-xs flex items-center gap-1">
                          Age
                          <button
                            class="btn btn-ghost btn-xs p-0 h-auto min-h-0"
                            on:click={toggleAgeVisibility}
                            aria-label="Hide age"
                          >
                            <svg
                              xmlns="http://www.w3.org/2000/svg"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke-width="1.5"
                              stroke="currentColor"
                              class="w-3 h-3"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M3.98 8.223A10.477 10.477 0 0 0 1.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.451 10.451 0 0 1 12 4.5c4.756 0 8.773 3.162 10.065 7.498a10.522 10.522 0 0 1-4.293 5.774M6.228 6.228 3 3m3.228 3.228 3.65 3.65m7.894 7.894L21 21m-3.228-3.228-3.65-3.65m0 0a3 3 0 1 1-4.243-4.243m4.242 4.242L9.88 9.88"
                              />
                            </svg>
                          </button>
                        </div>
                        <div class="stat-value text-lg">
                          {calculatedMetrics.age_years}
                        </div>
                        <div class="stat-desc text-xs">years</div>
                      </div>
                    {/if}
                    {#if calculatedMetrics?.age_years && !showAge}
                      <div
                        class="stat bg-base-100 rounded-box p-3 tooltip tooltip-top"
                        data-tip="Age is hidden - click to show"
                      >
                        <div class="stat-title text-xs flex items-center gap-1">
                          Age
                          <button
                            class="btn btn-ghost btn-xs p-0 h-auto min-h-0"
                            on:click={toggleAgeVisibility}
                            aria-label="Show age"
                          >
                            <svg
                              xmlns="http://www.w3.org/2000/svg"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke-width="1.5"
                              stroke="currentColor"
                              class="w-3 h-3"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z"
                              />
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"
                              />
                            </svg>
                          </button>
                        </div>
                        <div class="stat-value text-lg opacity-20">••</div>
                        <div class="stat-desc text-xs opacity-50">years</div>
                      </div>
                    {/if}
                  </div>
                {:else}
                  <div class="alert alert-info">
                    <InfoButton 
                      standalone={true} 
                      size="lg" 
                      iconClassName="text-info-content"
                    />
                    <span
                      >No biometric data yet. Add your details below for
                      personalized nutrition goals!</span
                    >
                  </div>
                {/if}

                <!-- Quick Form -->
                <div class="space-y-3">
                  <!-- Birth Date -->
                  <FormField
                    label="Birth Date"
                    id="birthDate"
                    type="date"
                    size="sm"
                    max={new Date().toISOString().split("T")[0]}
                    bind:value={birthDate}
                  />

                  <!-- Sex -->
                  <FormSelect
                    label="Sex (for DRI calculations)"
                    id="sex"
                    size="sm"
                    bind:value={sex}
                    options={[
                      {
                        value: "prefer_not_to_say",
                        label: "Prefer not to say",
                      },
                      { value: "male", label: "Male" },
                      { value: "female", label: "Female" },
                      { value: "other", label: "Other" },
                    ]}
                  />

                  <!-- Height & Weight Row -->
                  <div class="grid grid-cols-2 gap-2">
                    <FormField
                      label="Height (cm)"
                      id="height"
                      type="number"
                      size="sm"
                      placeholder="175"
                      min={50}
                      max={300}
                      step={0.1}
                      bind:value={heightCm}
                    />
                    <FormField
                      label="Weight (kg)"
                      id="weight"
                      type="number"
                      size="sm"
                      placeholder="70"
                      min={20}
                      max={500}
                      step={0.1}
                      bind:value={weightKg}
                    />
                  </div>

                  <!-- Activity Level -->
                  <FormSelect
                    label="Activity Level"
                    id="activity"
                    size="sm"
                    bind:value={activityLevel}
                    options={[
                      {
                        value: "sedentary",
                        label: "Level 1 - Sedentary (little/no exercise)",
                      },
                      {
                        value: "lightly_active",
                        label: "Level 2 - Lightly Active (1-3 days/week)",
                      },
                      {
                        value: "moderately_active",
                        label: "Level 3 - Moderately Active (3-5 days/week)",
                      },
                      {
                        value: "very_active",
                        label: "Level 4 - Very Active (6-7 days/week)",
                      },
                      {
                        value: "extra_active",
                        label:
                          "Level 5 - Extra Active (very hard exercise daily)",
                      },
                    ]}
                  />
                </div>

                <!-- Actions -->
                <div class="flex gap-2 justify-between">
                  {#if biometrics}
                    <button
                      class="btn btn-outline btn-error btn-sm"
                      on:click={deleteBiometrics}
                      disabled={savingBiometrics}
                    >
                      Delete
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
                      Save
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
              {editingGoalName === "New Goal"
                ? "Create New Goal"
                : `Edit ${editingGoalName}`}
            </h3>

            <div class="alert alert-info mb-4">
              <InfoButton
                standalone={true}
                size="lg"
                iconClassName="text-info-content"
              />
              <div>
                <div class="text-sm">
                  Need help setting your nutrition goals?
                </div>
                <div class="text-xs mt-1">
                  Use the official <a
                    href="https://www.nal.usda.gov/human-nutrition-and-food-safety/dri-calculator"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="link link-info font-semibold text-accent-content"
                    >USDA DRI Calculator</a
                  > to determine appropriate targets for your age, sex, and activity
                  level.
                </div>
              </div>
            </div>

            <div class="space-y-4">
              <!-- Goal Name Input -->
              <div class="form-control w-full">
                <label class="label" for="goalName">
                  <span class="label-text">Goal Name <span class="text-error">*</span></span>
                </label>
                <input
                  id="goalName"
                  type="text"
                  placeholder="e.g., Bulking, Cutting, Maintenance"
                  class="input input-bordered w-full"
                  class:input-error={!customName.trim()}
                  bind:value={customName}
                  required
                />
              {#if !customName.trim()}
                  <div class="label">
                    <span class="label-text-alt text-error">Goal name is required</span>
                  </div>
              {:else if Object.keys(customTargets).length === 0 && Object.keys(customUpperLimits).length === 0}
                  <div class="label">
                    <span class="label-text-alt text-warning">At least one target or upper limit must be set</span>
                  </div>
              {/if}
              </div>

              <!-- Category Selection -->
              <div class="form-control w-full">
                <label class="label" for="goalCategory">
                  <span class="label-text">Category <span class="text-base-content/50">(optional)</span></span>
                </label>
                <select
                  id="goalCategory"
                  class="select select-bordered w-full"
                  bind:value={customCategory}
                >
                  <option value="custom">Custom</option>
                  <option value="weight">Weight Management</option>
                  <option value="fitness">Fitness & Performance</option>
                  <option value="health">Health & Wellness</option>
                </select>
                <div class="label">
                  <span class="label-text-alt">Choose a category to help organize your goals</span>
                </div>
              </div>

              <div class="divider">Custom Nutrition Goals</div>

              <div class="alert alert-info mb-4">
                <InfoButton
                  standalone={true}
                  size="sm"
                  iconClassName="text-info-content"
                />
                <div class="text-sm">
                  <div><strong>Daily Targets:</strong> The minimum amount of each nutrient you would like to consume per day.</div>
                  <div><strong>Upper Limits:</strong> The maximum amount of a given nutrient you would like to consume per day (optional).</div>
                </div>
              </div>

              <div class="space-y-6 max-h-96 overflow-y-auto">
                <!-- Single Nutrient List with Both Target and Upper Limit Options -->
                {#if editableNutrients.length > 0}
                  <div class="space-y-4">
                    {#each editableNutrients as nutrient}
                      <div class="card bg-base-100 border border-base-300">
                        <div class="card-body p-4">
                          <h5 class="font-medium text-sm mb-3 flex items-center justify-between">
                            <span>{nutrient.label}</span>
                            <span class="text-xs text-base-content/60">{nutrient.unit}</span>
                          </h5>
                          
                          <div class="grid grid-cols-2 gap-3">
                            <!-- Daily Target -->
                            <div class="form-control">
                              <label class="label justify-start gap-2" for={`target_${nutrient.key}`}>
                                <span class="label-text text-xs text-primary">Daily Target</span>
                                {#if nutrient.hasTarget}
                                  <span class="badge badge-primary badge-xs">DRI: {goals?.targets[nutrient.key]}</span>
                                {/if}
                              </label>
                              <input
                                type="number"
                                id={`target_${nutrient.key}`}
                                class="input input-bordered input-sm"
                                min="0"
                                step="0.1"
                                placeholder={nutrient.hasTarget ? goals?.targets[nutrient.key]?.toString() : "Optional"}
                                value={getNutrientValue(nutrient.key) || ""}
                                on:input={(e) =>
                                  updateNutrient(
                                    nutrient.key,
                                    parseFloat(e.currentTarget.value) || 0,
                                  )}
                              />
                            </div>

                            <!-- Upper Limit -->
                            <div class="form-control">
                              <label class="label justify-start gap-2" for={`limit_${nutrient.key}`}>
                                <span class="label-text text-xs text-warning">Upper Limit</span>
                                {#if nutrient.hasUpperLimit}
                                  <span class="badge badge-warning badge-xs">DRI: {goals?.upper_limits[nutrient.key]}</span>
                                {/if}
                              </label>
                              <input
                                type="number"
                                id={`limit_${nutrient.key}`}
                                class="input input-bordered input-warning input-sm"
                                min="0"
                                step="0.1"
                                placeholder={nutrient.hasUpperLimit ? goals?.upper_limits[nutrient.key]?.toString() : "Optional"}
                                value={getUpperLimitValue(nutrient.key) || ""}
                                on:input={(e) =>
                                  updateUpperLimit(
                                    nutrient.key,
                                    parseFloat(e.currentTarget.value) || 0,
                                  )}
                              />
                            </div>
                          </div>
                        </div>
                      </div>
                    {/each}
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
                disabled={saving || !customName.trim() || (Object.keys(customTargets).length === 0 && Object.keys(customUpperLimits).length === 0)}
              >
                {#if saving}
                  <span class="loading loading-spinner loading-xs"></span>
                  Saving...
                {:else}
                  Save Goals
                {/if}
              </button>

              <button class="btn btn-outline" on:click={closeEditModal}>
                Cancel
              </button>
            </div>
          </div>
        </div>
      {/if}

      <!-- Labels Section -->
      <div class="card bg-base-200 shadow-lg mt-8">
        <div class="card-body p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="card-title flex items-center gap-2">
              <TagIcon className="w-5 h-5" />
              Your Labels
            </h2>
            <a href="/labels" class="btn btn-primary btn-sm"> Manage Labels </a>
          </div>

          {#if labelsLoading}
            <div class="flex justify-center py-6">
              <span class="loading loading-spinner loading-sm"></span>
              <span class="ml-2 text-sm text-base-content/70"
                >Loading labels...</span
              >
            </div>
          {:else if labels.length === 0}
            <div class="text-center py-8">
              <TagIcon
                className="w-12 h-12 mx-auto text-base-content/30 mb-3"
              />
              <p class="text-base-content/50 text-sm mb-3">No labels found</p>
              <a href="/labels" class="btn btn-primary btn-sm">
                Create Your First Label
              </a>
            </div>
          {:else}
            <div class="space-y-3">
              <p class="text-sm text-base-content/70">
                Organize your meals with labels you've created:
              </p>

              <!-- Labels Grid -->
              <div class="flex flex-wrap gap-2">
                {#each labels as label}
                  <div
                    class="tooltip"
                    data-tip={label.description ||
                      `${label.consumption_count} ${label.consumption_count === 1 ? "consumption" : "consumptions"}`}
                  >
                    <Label
                      name={label.name}
                      color={label.color}
                      className="cursor-help px-3 py-2"
                    >
                      <span class="ml-1 text-xs opacity-80"
                        >{label.consumption_count}</span
                      >
                    </Label>
                  </div>
                {/each}
              </div>

              {#if labels.length > 6}
                <div class="text-center pt-2">
                  <a href="/labels" class="btn btn-ghost btn-sm">
                    View All {labels.length} Labels →
                  </a>
                </div>
              {/if}
            </div>
          {/if}
        </div>
      </div>

      <!-- Events Section -->
      <div class="card bg-base-200 shadow-lg mt-8">
        <div class="card-body p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="card-title flex items-center gap-2">
              <CalendarDaysIcon className="w-5 h-5" />
              Your Events
            </h2>
            <a href="/events" class="btn btn-primary btn-sm"> Manage Events </a>
          </div>

          {#if eventTypesLoading}
            <div class="flex justify-center py-6">
              <span class="loading loading-spinner loading-sm"></span>
              <span class="ml-2 text-sm text-base-content/70"
                >Loading event types...</span
              >
            </div>
          {:else if eventTypes.length === 0}
            <div class="text-center py-8">
              <CalendarDaysIcon
                className="w-12 h-12 mx-auto text-base-content/30 mb-3"
              />
              <p class="text-base-content/50 text-sm mb-3">No event types found</p>
              <a href="/events" class="btn btn-primary btn-sm">
                Create Your First Event Type
              </a>
            </div>
          {:else}
            <div class="space-y-3">
              <p class="text-sm text-base-content/70">
                Track different types of events you've created:
              </p>

              <!-- Event Types Grid -->
              <div class="flex flex-wrap gap-2">
                {#each eventTypes as eventType}
                  <div
                    class="tooltip"
                    data-tip={eventType.description ||
                      `${eventType.event_count} ${eventType.event_count === 1 ? "event" : "events"}`}
                  >
                    <Label
                      name={eventType.name}
                      color={eventType.color}
                      className="cursor-help px-3 py-2"
                    >
                      <span class="ml-1 text-xs opacity-80"
                        >{eventType.event_count}</span
                      >
                    </Label>
                  </div>
                {/each}
              </div>

              {#if eventTypes.length > 6}
                <div class="text-center pt-2">
                  <a href="/events" class="btn btn-ghost btn-sm">
                    View All {eventTypes.length} Event Types →
                  </a>
                </div>
              {/if}
            </div>
          {/if}
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
                  <a
                    href="/api-keys"
                    class="btn btn-outline btn-sm w-full flex items-center gap-2"
                  >
                    <KeyIcon className="w-4 h-4" />
                    API Keys
                    {#if !isProUser}
                      <span class="badge badge-primary badge-xs ml-auto">Pro</span>
                    {/if}
                  </a>
                  <button class="btn btn-outline btn-sm w-full"
                    >Export Data</button
                  >
                  <button class="btn btn-outline btn-sm w-full"
                    >Clear History</button
                  >
                  <button
                    class="btn btn-error btn-sm w-full"
                    disabled={signingOut}
                    on:click={handleSignOut}
                  >
                    {#if signingOut}
                      <span class="loading loading-spinner loading-xs"></span>
                      Signing out...
                    {:else}
                      Sign Out
                    {/if}
                  </button>
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
<!-- Delete Goal Set Confirmation Modal -->
<ConfirmModal
  bind:show={showDeleteModal}
  title="Delete Goal Set"
  confirmText="Delete Goal Set"
  confirmVariant="error"
  onConfirm={confirmDeleteGoalSet}
  onCancel={cancelDeleteGoalSet}
>
  <div class="space-y-4">
    <p>
      Are you sure you want to delete the goal set <strong
        >"{goalToDelete}"</strong
      >?
    </p>
    <div class="alert alert-warning">
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
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.268 16.5c-.77.833.192 2.5 1.732 2.5z"
        />
      </svg>
      <span>This action cannot be undone.</span>
    </div>
    {#if goalSets.length === 1}
      <div class="bg-info/10 p-3 rounded-lg">
        <p class="text-sm text-info-content">
          🧬 This is your last custom goal set. Deleting it will return you to
          DRI (Dietary Reference Intakes) defaults.
        </p>
      </div>
    {/if}
  </div>
</ConfirmModal>

<!-- Reset to DRI Defaults Confirmation Modal -->
{#if showResetDRIModal}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg mb-4">Reset to DRI Defaults</h3>

      <div class="space-y-4">
        <p>
          Are you sure you want to delete <strong>ALL</strong> custom goal sets and
          return to DRI defaults?
        </p>
        <div class="alert alert-warning">
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
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.268 16.5c-.77.833.192 2.5 1.732 2.5z"
            />
          </svg>
          <span
            >This will delete all {goalSets.length} custom goal sets. This action
            cannot be undone.</span
          >
        </div>
        <div class="bg-info/10 p-3 rounded-lg">
          <p class="text-sm text-info-content">
            🧬 You'll return to DRI (Dietary Reference Intakes) defaults based
            on your profile demographics.
          </p>
        </div>
      </div>

      <div class="modal-action">
        <button class="btn btn-ghost" on:click={cancelResetToDRIDefaults}>
          Cancel
        </button>
        <button class="btn btn-warning" on:click={confirmResetToDRIDefaults}>
          <ArrowPath className="w-4 h-4 mr-1" />
          Reset to DRI Defaults
        </button>
      </div>
    </div>
    <div
      class="modal-backdrop"
      on:click={cancelResetToDRIDefaults}
      on:keydown={(e) => e.key === "Escape" && cancelResetToDRIDefaults()}
      role="button"
      tabindex="0"
      aria-label="Close modal"
    ></div>
  </div>
{/if}

<!-- Imperial System Meme Modal -->
{#if showImperialModal}
  <div class="modal modal-open">
    <div class="modal-box max-w-2xl max-h-[90vh] overflow-y-auto">
      <div class="text-center space-y-4 sm:space-y-6">
        <!-- Meme Header -->
        <div class="text-4xl sm:text-6xl">🚫</div>
        <h3 class="font-bold text-xl sm:text-2xl" style="color: var(--color-dark);">
          LOL NO.
        </h3>

        <!-- Meme Content -->
        <div class="space-y-3 sm:space-y-4 text-base sm:text-lg">
          <p>You can't use Imperial units on this site.</p>
          <p class="font-semibold text-primary">
            This site uses the METRIC SYSTEM because it is SUPERIOR! 🧑‍🔬
          </p>

          <div class="bg-base-200 p-3 sm:p-4 rounded-box space-y-1 sm:space-y-2">
            <p class="text-sm">🌍 Used by 95% of the world</p>
            <p class="text-sm">🧮 Base-10, actually makes sense</p>
            <p class="text-sm">
              🚀 Used by NASA (even though they're American)
            </p>
            <p class="text-sm">🔬 All scientific research uses metric</p>
            <p class="text-sm">💊 Your medicine dosages? Metric.</p>
            <p class="text-sm">🏃‍♂️ Olympic records? Metric.</p>
          </div>

          <div class="text-sm sm:text-base space-y-1 sm:space-y-2">
            <p>Imperial is just...</p>
            <p class="italic">
              "12 inches in a foot, 3 feet in a yard, 1760 yards in a mile"
            </p>
            <p class="font-bold">vs.</p>
            <p class="italic">"10mm = 1cm, 100cm = 1m, 1000m = 1km"</p>
            <p class="text-primary font-semibold">See the difference? 🤯</p>
          </div>

          <div class="text-sm text-base-content/70 space-y-1">
            <p>Even the UK switched to metric for most things.</p>
            <p>It's time to let go of the past. 📏➡️📐</p>
          </div>
        </div>

        <!-- Acknowledgment Button -->
        <div class="modal-action justify-center pt-2">
          <button
            class="btn btn-success btn-lg min-h-[44px] w-full sm:w-auto"
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
      on:keydown={(e) => e.key === "Escape" && closeImperialModal()}
      role="button"
      tabindex="0"
      aria-label="Close modal"
    ></div>
  </div>
{/if}

<!-- Disabled Nutrients Modal -->
{#if showDisabledNutrientsModal}
  <div class="modal modal-open">
    <div class="modal-box max-w-4xl max-h-[90vh] overflow-y-auto">
      <h3 class="font-bold text-lg flex items-center gap-2 mb-4">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        Manage Disabled Nutrients
      </h3>
      
      <div class="alert alert-info mb-6">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <div>
          <div class="font-semibold">What are disabled nutrients?</div>
          <div class="text-sm mt-1">
            Disabled nutrients will still be tracked when you log meals, but they won't appear in any charts, summaries, or goal calculations. This helps you focus on the nutrients that matter most to you.
          </div>
        </div>
      </div>

      <div class="space-y-6">
        <!-- Macronutrients -->
        <div class="card bg-base-100 border border-base-300">
          <div class="card-body p-4">
            <h4 class="font-medium text-base mb-3 flex items-center justify-between border-b border-base-300 pb-2">
              <span>Macronutrients</span>
              <span class="text-xs text-base-content/60">
                {ALL_NUTRIENTS.filter(n => n.category === "Macronutrients" && tempDisabledNutrients.includes(n.key)).length} of {ALL_NUTRIENTS.filter(n => n.category === "Macronutrients").length} disabled
              </span>
            </h4>
            
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {#each ALL_NUTRIENTS.filter(n => n.category === "Macronutrients") as nutrient}
                <div class="form-control">
                  <label class="label cursor-pointer p-2 hover:bg-base-200 rounded-lg transition-colors">
                    <span class="label-text text-sm">{nutrient.label}</span>
                    <input 
                      type="checkbox" 
                      class="checkbox checkbox-error" 
                      checked={tempDisabledNutrients.includes(nutrient.key)}
                      on:change={() => toggleDisabledNutrient(nutrient.key)}
                    />
                  </label>
                </div>
              {/each}
            </div>
          </div>
        </div>

        <!-- B-Complex Vitamins -->
        <div class="card bg-base-100 border border-base-300">
          <div class="card-body p-4">
            <h4 class="font-medium text-base mb-3 flex items-center justify-between border-b border-base-300 pb-2">
              <span>B-Complex Vitamins</span>
              <span class="text-xs text-base-content/60">
                {ALL_NUTRIENTS.filter(n => n.category === "B-Complex Vitamins" && tempDisabledNutrients.includes(n.key)).length} of {ALL_NUTRIENTS.filter(n => n.category === "B-Complex Vitamins").length} disabled
              </span>
            </h4>
            
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {#each ALL_NUTRIENTS.filter(n => n.category === "B-Complex Vitamins") as nutrient}
                <div class="form-control">
                  <label class="label cursor-pointer p-2 hover:bg-base-200 rounded-lg transition-colors">
                    <span class="label-text text-sm">{nutrient.label}</span>
                    <input 
                      type="checkbox" 
                      class="checkbox checkbox-error" 
                      checked={tempDisabledNutrients.includes(nutrient.key)}
                      on:change={() => toggleDisabledNutrient(nutrient.key)}
                    />
                  </label>
                </div>
              {/each}
            </div>
          </div>
        </div>

        <!-- Fat-Soluble Vitamins -->
        <div class="card bg-base-100 border border-base-300">
          <div class="card-body p-4">
            <h4 class="font-medium text-base mb-3 flex items-center justify-between border-b border-base-300 pb-2">
              <span>Fat-Soluble Vitamins</span>
              <span class="text-xs text-base-content/60">
                {ALL_NUTRIENTS.filter(n => n.category === "Fat-Soluble Vitamins" && tempDisabledNutrients.includes(n.key)).length} of {ALL_NUTRIENTS.filter(n => n.category === "Fat-Soluble Vitamins").length} disabled
              </span>
            </h4>
            
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {#each ALL_NUTRIENTS.filter(n => n.category === "Fat-Soluble Vitamins") as nutrient}
                <div class="form-control">
                  <label class="label cursor-pointer p-2 hover:bg-base-200 rounded-lg transition-colors">
                    <span class="label-text text-sm">{nutrient.label}</span>
                    <input 
                      type="checkbox" 
                      class="checkbox checkbox-error" 
                      checked={tempDisabledNutrients.includes(nutrient.key)}
                      on:change={() => toggleDisabledNutrient(nutrient.key)}
                    />
                  </label>
                </div>
              {/each}
            </div>
          </div>
        </div>

        <!-- Water-Soluble Vitamins -->
        <div class="card bg-base-100 border border-base-300">
          <div class="card-body p-4">
            <h4 class="font-medium text-base mb-3 flex items-center justify-between border-b border-base-300 pb-2">
              <span>Water-Soluble Vitamins</span>
              <span class="text-xs text-base-content/60">
                {ALL_NUTRIENTS.filter(n => n.category === "Water-Soluble Vitamins" && tempDisabledNutrients.includes(n.key)).length} of {ALL_NUTRIENTS.filter(n => n.category === "Water-Soluble Vitamins").length} disabled
              </span>
            </h4>
            
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {#each ALL_NUTRIENTS.filter(n => n.category === "Water-Soluble Vitamins") as nutrient}
                <div class="form-control">
                  <label class="label cursor-pointer p-2 hover:bg-base-200 rounded-lg transition-colors">
                    <span class="label-text text-sm">{nutrient.label}</span>
                    <input 
                      type="checkbox" 
                      class="checkbox checkbox-error" 
                      checked={tempDisabledNutrients.includes(nutrient.key)}
                      on:change={() => toggleDisabledNutrient(nutrient.key)}
                    />
                  </label>
                </div>
              {/each}
            </div>
          </div>
        </div>

        <!-- Essential Minerals -->
        <div class="card bg-base-100 border border-base-300">
          <div class="card-body p-4">
            <h4 class="font-medium text-base mb-3 flex items-center justify-between border-b border-base-300 pb-2">
              <span>Essential Minerals</span>
              <span class="text-xs text-base-content/60">
                {ALL_NUTRIENTS.filter(n => n.category === "Essential Minerals" && tempDisabledNutrients.includes(n.key)).length} of {ALL_NUTRIENTS.filter(n => n.category === "Essential Minerals").length} disabled
              </span>
            </h4>
            
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {#each ALL_NUTRIENTS.filter(n => n.category === "Essential Minerals") as nutrient}
                <div class="form-control">
                  <label class="label cursor-pointer p-2 hover:bg-base-200 rounded-lg transition-colors">
                    <span class="label-text text-sm">{nutrient.label}</span>
                    <input 
                      type="checkbox" 
                      class="checkbox checkbox-error" 
                      checked={tempDisabledNutrients.includes(nutrient.key)}
                      on:change={() => toggleDisabledNutrient(nutrient.key)}
                    />
                  </label>
                </div>
              {/each}
            </div>
          </div>
        </div>

        <!-- Other Compounds -->
        <div class="card bg-base-100 border border-base-300">
          <div class="card-body p-4">
            <h4 class="font-medium text-base mb-3 flex items-center justify-between border-b border-base-300 pb-2">
              <span>Other Compounds</span>
              <span class="text-xs text-base-content/60">
                {ALL_NUTRIENTS.filter(n => n.category === "Other Compounds" && tempDisabledNutrients.includes(n.key)).length} of {ALL_NUTRIENTS.filter(n => n.category === "Other Compounds").length} disabled
              </span>
            </h4>
            
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {#each ALL_NUTRIENTS.filter(n => n.category === "Other Compounds") as nutrient}
                <div class="form-control">
                  <label class="label cursor-pointer p-2 hover:bg-base-200 rounded-lg transition-colors">
                    <span class="label-text text-sm">{nutrient.label}</span>
                    <input 
                      type="checkbox" 
                      class="checkbox checkbox-error" 
                      checked={tempDisabledNutrients.includes(nutrient.key)}
                      on:change={() => toggleDisabledNutrient(nutrient.key)}
                    />
                  </label>
                </div>
              {/each}
            </div>
          </div>
        </div>
      </div>

      <!-- Summary -->
      {#if tempDisabledNutrients.length > 0}
        <div class="alert alert-warning mt-6">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <div>
            <div class="font-semibold">
              {tempDisabledNutrients.length} nutrient{tempDisabledNutrients.length !== 1 ? 's' : ''} will be disabled
            </div>
            <div class="text-sm mt-1">
              These nutrients will be hidden from all charts, summaries, and goal calculations.
            </div>
          </div>
        </div>
      {/if}

      <div class="modal-action">
        <button 
          class="btn btn-outline" 
          on:click={closeDisabledNutrientsModal}
          disabled={saving}
        >
          Cancel
        </button>
        <button 
          class="btn btn-primary" 
          on:click={saveDisabledNutrients}
          disabled={saving}
        >
          {#if saving}
            <span class="loading loading-spinner loading-xs"></span>
            Saving...
          {:else}
            Save Changes
          {/if}
        </button>
      </div>
    </div>
    <div 
      class="modal-backdrop" 
      on:click={closeDisabledNutrientsModal}
      on:keydown={(e) => e.key === "Escape" && closeDisabledNutrientsModal()}
      role="button"
      tabindex="0"
      aria-label="Close modal"
    ></div>
  </div>
{/if}

<!-- Toast notifications -->
<!-- Toast Notifications -->
<Toast position="bottom-end" />

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
    transition:
      border-color 0.15s ease-in-out,
      box-shadow 0.15s ease-in-out;
  }
</style>
