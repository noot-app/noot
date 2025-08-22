<script lang="ts">
  import { apiClient } from "$lib/api/client";
  import { PUBLIC_APP_NAME } from "$env/static/public";
  import { onMount } from "svelte";
  import NutritionStats from "$lib/components/NutritionStats.svelte";
  import Goals from "$lib/components/Goals.svelte";
  import { transformNutritionSummary } from "$lib/utils/nutrition";

  let currentView: 'today' | 'week' = 'today';
  let isLoading = false;
  let error = "";
  let summaryData: any = null;

  onMount(() => {
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

  // Extract current nutrition values for goals comparison
  $: currentNutrition = summaryData?.summary ? transformNutritionSummary(summaryData.summary) : undefined;
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
        <NutritionStats 
          calories={summaryData.summary.total_calories || 0}
          protein={summaryData.summary.total_protein_g || 0}
          carbs={summaryData.summary.total_carbs_g || 0}
          fat={summaryData.summary.total_fat_g || 0}
          size="normal"
        />

        <!-- Nutrition Goals -->
        <Goals {currentNutrition} />

        <!-- Daily Breakdown (for week view) -->
        {#if currentView === 'week' && summaryData.summary.daily_breakdown && summaryData.summary.daily_breakdown.length > 0}
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
                    {#each summaryData.summary.daily_breakdown as day}
                      <tr>
                        <td>{new Date(day.date).toLocaleDateString()}</td>
                        <td>{day.calories}</td>
                        <td>{day.protein_g?.toFixed(1) || 0}g</td>
                        <td>{day.total_carbs_g?.toFixed(1) || 0}g</td>
                        <td>{day.total_fat_g?.toFixed(1) || 0}g</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        {/if}

        <!-- Today's Details (for today view) -->
        {#if currentView === 'today' && summaryData.summary.daily_breakdown && summaryData.summary.daily_breakdown.length > 0}
          <div class="card bg-base-200 shadow-xl">
            <div class="card-body">
              <h2 class="card-title mb-4">Today's Summary</h2>
              <div class="stats stats-vertical lg:stats-horizontal shadow w-full">
                <div class="stat">
                  <div class="stat-title">Meals Logged</div>
                  <div class="stat-value text-lg">{summaryData.summary.consumption_count}</div>
                  <div class="stat-desc">Today</div>
                </div>
                <div class="stat">
                  <div class="stat-title">Avg per Meal</div>
                  <div class="stat-value text-lg">{Math.round(summaryData.summary.total_calories / summaryData.summary.consumption_count)}</div>
                  <div class="stat-desc">Calories</div>
                </div>
                <div class="stat">
                  <div class="stat-title">Fiber</div>
                  <div class="stat-value text-lg">{(summaryData.summary.total_fiber_g || 0).toFixed(1)}g</div>
                  <div class="stat-desc">Total</div>
                </div>
                <div class="stat">
                  <div class="stat-title">Sodium</div>
                  <div class="stat-value text-lg">{summaryData.summary.total_sodium_mg || 0} mg</div>
                  <div class="stat-desc">Total</div>
                </div>
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
        {#if (!summaryData.summary.total_calories || summaryData.summary.total_calories === 0) && !summaryData.recent_consumptions?.length}
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
