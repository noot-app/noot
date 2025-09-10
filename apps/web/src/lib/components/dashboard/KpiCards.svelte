<script lang="ts">
  import Zap from "$lib/components/icons/Zap.svelte"
  import Dumbbell from "$lib/components/icons/Dumbbell.svelte"
  import Cube from "$lib/components/icons/Cube.svelte"
  import CalendarIcon from "$lib/components/icons/calendar-days.svelte"

  export let consumptionsData: any = null
  export let eventsData: any = null
  export let dateRange: string = "7d"
</script>

<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 lg:gap-6">
  <!-- Total Calories -->
  <div class="card bg-base-200 shadow-sm">
    <div class="card-body p-4 lg:p-6">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-xs lg:text-sm font-medium text-base-content-lighter">
            Total Calories
          </h3>
          <p class="text-xl lg:text-2xl font-bold text-base-content mt-1">
            {consumptionsData?.consumptions
              .reduce((total: number, c: any) => total + (c.summary?.totals?.calories || 0), 0)
              .toLocaleString() || 0}
          </p>
        </div>
        <div class="text-honey">
          <Zap className="w-6 h-6 lg:w-8 lg:h-8" />
        </div>
      </div>
    </div>
  </div>

  <!-- Average Daily Protein -->
  <div class="card bg-base-200 shadow-sm">
    <div class="card-body p-4 lg:p-6">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-xs lg:text-sm font-medium text-base-content-lighter">
            Avg Daily Protein
          </h3>
          <p class="text-xl lg:text-2xl font-bold text-base-content mt-1">
            {consumptionsData ? Math.round(
              consumptionsData.consumptions
                .reduce((total: number, c: any) => total + (c.summary?.totals?.protein_g || 0), 0) /
              Math.max(parseInt(dateRange.replace('d', '')), 1)
            ) : 0}g
          </p>
        </div>
        <div class="text-secondary">
          <Dumbbell className="w-6 h-6 lg:w-8 lg:h-8" />
        </div>
      </div>
    </div>
  </div>

  <!-- Total Consumptions -->
  <div class="card bg-base-200 shadow-sm">
    <div class="card-body p-4 lg:p-6">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-xs lg:text-sm font-medium text-base-content-lighter">
            Total Consumptions
          </h3>
          <p class="text-xl lg:text-2xl font-bold text-base-content mt-1">
            {consumptionsData?.consumptions.length || 0}
          </p>
        </div>
        <div class="text-secondary">
          <Cube className="w-6 h-6 lg:w-8 lg:h-8" />
        </div>
      </div>
    </div>
  </div>

  <!-- Events Tracked -->
  <div class="card bg-base-200 shadow-sm">
    <div class="card-body p-4 lg:p-6">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-xs lg:text-sm font-medium text-base-content-lighter">
            Events Tracked
          </h3>
          <p class="text-xl lg:text-2xl font-bold text-base-content mt-1">
            {eventsData?.events?.length || 0}
          </p>
        </div>
        <div class="text-info">
          <CalendarIcon className="w-6 h-6 lg:w-8 lg:h-8" />
        </div>
      </div>
    </div>
  </div>
</div>
