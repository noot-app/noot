<script lang="ts">
  import { apiClient } from "$lib/api/client";
  import { units, convertWeight, formatWeight, type Units } from "$lib/stores/units";
  import { PUBLIC_APP_NAME } from "$env/static/public";
  import { onMount } from "svelte";

  let currentView: 'today' | 'week' = 'today';
  let isLoading = false;
  let error = "";
  let summaryData: any = null;

  $: currentUnits = $units;

  onMount(() => {
    units.init();
    loadSummary();
  });

  async function loadSummary() {
    isLoading = true;
    error = "";
    
    try {
      let startDate: string, endDate: string;

      if (currentView === 'today') {
        // Get today in local timezone
        const today = new Date();
        const startOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 0, 0, 0);
        const endOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 23, 59, 59, 999);
        
        startDate = startOfDay.toISOString();
        endDate = endOfDay.toISOString();
      } else {
        // Get week (7 days back from today) in local timezone
        const today = new Date();
        const weekAgo = new Date(today.getTime() - (6 * 24 * 60 * 60 * 1000)); // 6 days ago + today = 7 days
        
        const startOfWeek = new Date(weekAgo.getFullYear(), weekAgo.getMonth(), weekAgo.getDate(), 0, 0, 0);
        const endOfToday = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 23, 59, 59, 999);
        
        startDate = startOfWeek.toISOString();
        endDate = endOfToday.toISOString();
      }

      const response = await apiClient.GET('/nutrition-summary', {
        params: {
          query: {
            start: startDate,
            end: endDate
          }
        }
      });

      if (response.error) {
        throw new Error(`API Error: ${response.error}`);
      }

      summaryData = response.data;
    } catch (err) {
      error = `Error loading summary: ${err}`;
      console.error("Summary error:", err);
    } finally {
      isLoading = false;
    }
  }

  function switchView(view: 'today' | 'week') {
    if (currentView !== view) {
      currentView = view;
      loadSummary();
    }
  }

  function toggleUnits() {
    units.set(currentUnits === 'grams' ? 'ounces' : 'grams');
  }

  function convertNutrientValue(value: number): number {
    return convertWeight(value, 'grams', currentUnits);
  }
</script>

<svelte:head>
  <title>Summary - {PUBLIC_APP_NAME}</title>
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-8 max-w-6xl">
    <!-- Header -->
    <div class="text-center mb-8">
      <h1 class="text-3xl font-bold text-primary mb-4">Nutrition Summary</h1>
      <p class="text-base-content/70">Track your nutrition intake over time</p>
    </div>

    <!-- Navigation -->
    <div class="flex justify-center mb-8">
      <div class="btn-group">
        <a href="/" class="btn btn-ghost">
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          Home
        </a>
        <a href="/record" class="btn btn-primary">Record</a>
        <button class="btn btn-ghost" on:click={toggleUnits}>
          Units: {currentUnits === 'grams' ? 'Grams' : 'Ounces'}
        </button>
      </div>
    </div>

    <!-- View Toggle -->
    <div class="flex justify-center mb-8">
      <div class="btn-group">
        <button 
          class="btn {currentView === 'today' ? 'btn-primary' : 'btn-ghost'}"
          on:click={() => switchView('today')}
        >
          Today
        </button>
        <button 
          class="btn {currentView === 'week' ? 'btn-primary' : 'btn-ghost'}"
          on:click={() => switchView('week')}
        >
          This Week
        </button>
      </div>
    </div>

    <!-- Loading -->
    {#if isLoading}
      <div class="flex justify-center items-center py-12">
        <span class="loading loading-spinner loading-lg"></span>
        <span class="ml-4 text-lg">Loading summary...</span>
      </div>
    {/if}

    <!-- Error -->
    {#if error}
      <div class="alert alert-error mb-6">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{error}</span>
        <button class="btn btn-sm" on:click={loadSummary}>Retry</button>
      </div>
    {/if}

    <!-- Summary Data -->
    {#if summaryData && !isLoading}
      <div class="space-y-8">
        <!-- Overview Stats -->
        <div class="stats stats-vertical lg:stats-horizontal shadow-xl w-full">
          <div class="stat">
            <div class="stat-figure text-primary">
              <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div class="stat-title">Total Calories</div>
            <div class="stat-value text-primary">{summaryData.summary.total_calories || 0}</div>
            <div class="stat-desc">{currentView === 'today' ? 'Today' : 'This week'}</div>
          </div>

          <div class="stat">
            <div class="stat-figure text-secondary">
              <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
              </svg>
            </div>
            <div class="stat-title">Protein</div>
            <div class="stat-value text-secondary">
              {formatWeight(convertNutrientValue(summaryData.summary.total_protein_g || 0), currentUnits)}
            </div>
            <div class="stat-desc">Essential for muscle</div>
          </div>

          <div class="stat">
            <div class="stat-figure text-accent">
              <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 4V2a1 1 0 011-1h8a1 1 0 011 1v2h4a1 1 0 011 1v2a1 1 0 01-1 1h-1v12a2 2 0 01-2 2H6a2 2 0 01-2-2V8H3a1 1 0 01-1-1V5a1 1 0 011-1h4z" />
              </svg>
            </div>
            <div class="stat-title">Carbohydrates</div>
            <div class="stat-value text-accent">
              {formatWeight(convertNutrientValue(summaryData.summary.total_carbs_g || 0), currentUnits)}
            </div>
            <div class="stat-desc">Energy source</div>
          </div>

          <div class="stat">
            <div class="stat-figure text-warning">
              <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
              </svg>
            </div>
            <div class="stat-title">Fat</div>
            <div class="stat-value text-warning">
              {formatWeight(convertNutrientValue(summaryData.summary.total_fat_g || 0), currentUnits)}
            </div>
            <div class="stat-desc">Healthy fats</div>
          </div>
        </div>

        <!-- Daily Breakdown (for week view) -->
        {#if currentView === 'week' && summaryData.daily_breakdown && summaryData.daily_breakdown.length > 0}
          <div class="card bg-base-200 shadow-xl">
            <div class="card-body">
              <h2 class="card-title mb-4">Daily Breakdown</h2>
              <div class="overflow-x-auto">
                <table class="table table-zebra">
                  <thead>
                    <tr>
                      <th>Date</th>
                      <th>Calories</th>
                      <th>Protein</th>
                      <th>Carbs</th>
                      <th>Fat</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each summaryData.daily_breakdown as day}
                      <tr>
                        <td>{new Date(day.date).toLocaleDateString()}</td>
                        <td>{day.calories}</td>
                        <td>{formatWeight(convertNutrientValue(day.protein), currentUnits)}</td>
                        <td>{formatWeight(convertNutrientValue(day.carbs), currentUnits)}</td>
                        <td>{formatWeight(convertNutrientValue(day.fat), currentUnits)}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        {/if}

        <!-- Recent Consumptions -->
        {#if summaryData.recent_consumptions && summaryData.recent_consumptions.length > 0}
          <div class="card bg-base-200 shadow-xl">
            <div class="card-body">
              <h2 class="card-title mb-4">Recent Meals</h2>
              <div class="space-y-3">
                {#each summaryData.recent_consumptions as consumption}
                  <div class="card bg-base-100 shadow">
                    <div class="card-body p-4">
                      <div class="flex justify-between items-start">
                        <div class="flex-1">
                          <p class="font-medium mb-1">"{consumption.transcript}"</p>
                          <p class="text-sm text-base-content/60">
                            {new Date(consumption.created_at).toLocaleString()}
                          </p>
                        </div>
                        {#if consumption.total_calories}
                          <div class="badge badge-primary">{consumption.total_calories} cal</div>
                        {/if}
                      </div>
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          </div>
        {/if}

        <!-- No Data State -->
        {#if (!summaryData.total_calories || summaryData.total_calories === 0) && !summaryData.recent_consumptions?.length}
          <div class="text-center py-12">
            <div class="text-6xl mb-4">🍽️</div>
            <h3 class="text-2xl font-bold mb-2">No data for {currentView === 'today' ? 'today' : 'this week'}</h3>
            <p class="text-base-content/70 mb-6">Start logging your meals to see your nutrition summary</p>
            <a href="/record" class="btn btn-primary">
              Record Your First Meal
            </a>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
