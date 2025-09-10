<script lang="ts">
  import NutritionTrendsChart from "./NutritionTrendsChart.svelte"
  import MacroCompositionChart from "./MacroCompositionChart.svelte"
  import GoalProgressChart from "./GoalProgressChart.svelte"
  import EventCorrelationChart from "./EventCorrelationChart.svelte"
  import DailyGoalProgressGrid from "./DailyGoalProgressGrid.svelte"

  export let consumptionsData: any = null
  export let goalsData: any = null
  export let eventsData: any = null
  export let dateRange: string = "7d"
</script>

<div class="space-y-4 lg:space-y-6">
  <!-- Main Nutrition Chart - Full Width -->
  <div class="card bg-base-200 shadow-sm">
    <div class="card-body p-4 lg:p-6">
      <NutritionTrendsChart consumptions={consumptionsData?.consumptions || []} />
    </div>
  </div>

  <!-- Secondary Charts - Responsive Grid -->
  <div class="grid grid-cols-1 xl:grid-cols-2 gap-4 lg:gap-6">
    <!-- Macro Composition -->
    <div class="card bg-base-200 shadow-sm">
      <div class="card-body p-4 lg:p-6">
        <MacroCompositionChart consumptions={consumptionsData?.consumptions || []} />
      </div>
    </div>

    <!-- Goal Progress -->
    <div class="card bg-base-200 shadow-sm">
      <div class="card-body p-4 lg:p-6">
        <GoalProgressChart consumptions={consumptionsData?.consumptions || []} goals={goalsData} {dateRange} />
      </div>
    </div>

    <!-- Daily Goal Progress Grid - Full Width -->
    <div class="card bg-base-200 shadow-sm xl:col-span-2">
      <div class="card-body p-4 lg:p-6">
        <DailyGoalProgressGrid consumptions={consumptionsData?.consumptions || []} goals={goalsData} {dateRange} />
      </div>
    </div>

    <!-- Event Correlation - Full Width -->
    <div class="card bg-base-200 shadow-sm xl:col-span-2">
      <div class="card-body p-4 lg:p-6">
        <EventCorrelationChart 
          consumptions={consumptionsData?.consumptions || []} 
          events={eventsData?.events || []} 
        />
      </div>
    </div>
  </div>
</div>
