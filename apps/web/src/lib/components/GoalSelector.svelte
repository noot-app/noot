<script lang="ts">
  import { onMount } from "svelte"
  import { apiClient } from "$lib/api/client"
  import Icon from "./Icon.svelte"

  // Define types based on actual API responses
  interface GoalSet {
    name: string
    created_at: string
    updated_at: string
  }

  interface GoalSetsResponse {
    goal_sets: GoalSet[]
    active_goal_name: string
    user: {
      id: string
      email: string
      subject: string
      provider: string
      subscription_tier: string
      created_at: string
    }
  }

  export let onGoalChanged: () => void = () => {} // Callback when goal changes

  let goalSets: GoalSet[] = []
  let activeGoalName = ""
  let loading = true
  let switching = false
  let error = ""
  let user: GoalSetsResponse["user"] | null = null

  // Reactive subscription check
  $: isProUser = user?.subscription_tier === "pro"

  onMount(async () => {
    await loadGoalSets()
  })

  async function loadGoalSets() {
    try {
      loading = true
      const { data, error } = await apiClient.GET("/goals/sets")

      if (error) {
        throw new Error(`API Error: ${JSON.stringify(error)}`)
      }

      goalSets = data.goal_sets
      activeGoalName = data.active_goal_name || ""
      // Map the API user to our frontend User interface
      if (data.user) {
        user = {
          id: data.user.id,
          email: data.user.email,
          subscription_tier: data.user.subscription_tier,
          provider: "api",
          subject: data.user.id,
          created_at: data.user.created_at,
        }
      }
    } catch (err) {
      error = `Failed to load goal sets: ${err}`
      console.error("Goal sets error:", err)
    } finally {
      loading = false
    }
  }

  async function switchGoal(goalName: string) {
    if (switching || goalName === activeGoalName) return

    if (!isProUser) {
      error = "Pro subscription required for goal set management"
      return
    }

    try {
      switching = true
      const { error } = await apiClient.PUT("/goals/active", {
        body: { name: goalName },
      })

      if (error) {
        throw new Error(`API Error: ${JSON.stringify(error)}`)
      }

      activeGoalName = goalName
      // Notify parent component that the goal has changed
      onGoalChanged()
    } catch (err) {
      error = `Failed to switch goal: ${err}`
      console.error("Switch goal error:", err)
    } finally {
      switching = false
    }
  }
</script>

{#if goalSets.length > 1}
  <div class="flex items-center gap-3 mb-4">
    <span class="text-sm font-medium">Active Goal:</span>

    {#if loading}
      <div class="skeleton h-8 w-32"></div>
    {:else}
      <select
        class="select select-sm select-bordered bg-base-100"
        bind:value={activeGoalName}
        on:change={(e) => switchGoal(e.currentTarget.value)}
        disabled={switching}
        aria-label="Select active goal"
      >
        {#each goalSets as goalSet}
          <option value={goalSet.name}>{goalSet.name}</option>
        {/each}
      </select>
    {/if}

    {#if switching}
      <span class="loading loading-spinner loading-xs"></span>
    {/if}

    {#if error}
      <div class="tooltip tooltip-error" data-tip={error}>
        <Icon name="exclamation-circle" size="sm" className="text-error" />
      </div>
    {/if}
  </div>
{/if}
