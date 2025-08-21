<script lang="ts">
  import { onMount } from "svelte";
  import { apiClient } from "$lib/api/client";

  // Define types based on actual API responses
  interface GoalSet {
    name: string;
    created_at: string;
    updated_at: string;
  }

  interface GoalSetsResponse {
    goal_sets: GoalSet[];
    active_goal_name: string;
    user: {
      id: string;
      email: string;
      subject: string;
      provider: string;
      subscription_tier: string;
      created_at: string;
    };
  }

  export let onGoalChanged: () => void = () => {}; // Callback when goal changes

  let goalSets: GoalSet[] = [];
  let activeGoalName = "";
  let loading = true;
  let switching = false;
  let error = "";

  onMount(async () => {
    await loadGoalSets();
  });

  async function loadGoalSets() {
    try {
      loading = true;
      const response = await fetch('/api/v1/goals/sets', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: GoalSetsResponse = await response.json();
      goalSets = data.goal_sets;
      activeGoalName = data.active_goal_name || "";
    } catch (err) {
      error = `Failed to load goal sets: ${err}`;
      console.error("Goal sets error:", err);
    } finally {
      loading = false;
    }
  }

  async function switchGoal(goalName: string) {
    if (switching || goalName === activeGoalName) return;

    try {
      switching = true;
      const response = await fetch('/api/v1/goals/active', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name: goalName })
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      activeGoalName = goalName;
      // Notify parent component that the goal has changed
      onGoalChanged();
    } catch (err) {
      error = `Failed to switch goal: ${err}`;
      console.error("Switch goal error:", err);
    } finally {
      switching = false;
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
        <svg class="w-4 h-4 text-error" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </div>
    {/if}
  </div>
{/if}
