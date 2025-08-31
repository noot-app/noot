<script lang="ts">
  import Goals from "$lib/components/Goals.svelte"
  import NutritionStats from "$lib/components/NutritionStats.svelte"
  import type { PageData } from "./$types"
  import { onMount } from "svelte"

  export let data: PageData

  let useDri = false
  // Default: if auto source isn't custom, show DRI; otherwise user's goals
  $: useDri = data?.goalsAuto?.source !== "custom"

  // Compute per-consumption totals shape for Goals component
  function totalsFromConsumption(c: any): Record<string, number> {
    const t = c?.summary?.totals || {}
    return { ...t }
  }

  $: currentNutrition = totalsFromConsumption(data.consumption)
</script>

<svelte:head>
  <title>Meal Details</title>
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-6 max-w-4xl">
    <div class="mb-4">
      <a class="btn btn-ghost btn-sm" href="/summary">← Back</a>
    </div>

    <div class="flex items-start justify-between gap-4 mb-4">
      <div>
        <h1 class="text-2xl font-bold">Meal Details</h1>
        {#if data.consumption?.created_at}
          <p class="text-sm text-base-content/60">{new Date(data.consumption.created_at).toLocaleString()}</p>
        {/if}
        {#if data.consumption?.transcript}
          <p class="mt-2">"{data.consumption.transcript}"</p>
        {/if}
      </div>
      <div class="btn-group">
        <button class="btn btn-sm {useDri ? '' : 'btn-active'}" on:click={() => (useDri = false)}>
          Your Goals
        </button>
        <button class="btn btn-sm {useDri ? 'btn-active' : ''}" on:click={() => (useDri = true)}>
          DRI
        </button>
      </div>
    </div>

    <NutritionStats
      calories={data.consumption?.summary?.totals?.calories || 0}
      protein={data.consumption?.summary?.totals?.protein_g || 0}
      carbs={data.consumption?.summary?.totals?.total_carbs_g || 0}
      fat={data.consumption?.summary?.totals?.total_fat_g || 0}
      size="normal"
    />

    <div class="mt-6">
      <Goals
        currentNutrition={currentNutrition}
        showMealContribution={true}
        isSharedView={useDri}
        title="How this meal contributes"
        preloadGoalsAuto={data.goalsAuto}
        preloadGoalsDri={data.goalsDri}
        source={useDri ? "dri" : "auto"}
      />
    </div>

    {#if data.consumption?.items?.length}
      <div class="card bg-base-200 shadow mt-6">
        <div class="card-body">
          <h2 class="card-title mb-2">Items in this meal</h2>
          <div class="space-y-2">
            {#each data.consumption.items as item}
              <div class="flex justify-between items-center text-sm">
                <div class="flex-1">
                  <span class="font-medium">{item.item?.name}</span>
                  {#if item.item?.brand}
                    <span class="text-base-content/60">({item.item.brand})</span>
                  {/if}
                  {#if item.item?.grams !== undefined}
                    <span class="text-base-content/60"> - {item.item.grams}g</span>
                  {/if}
                </div>
                <div class="text-right">
                  <div class="font-medium">{Math.round(item.item?.nutrients?.calories || 0)} cal</div>
                  <div class="text-xs text-base-content/60">
                    P: {(item.item?.nutrients?.protein_g || 0).toFixed(1)}g |
                    C: {(item.item?.nutrients?.total_carbs_g || 0).toFixed(1)}g |
                    F: {(item.item?.nutrients?.total_fat_g || 0).toFixed(1)}g
                  </div>
                </div>
              </div>
            {/each}
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>
