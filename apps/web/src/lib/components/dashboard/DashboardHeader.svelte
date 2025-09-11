<script lang="ts">
  import SquaresTwoByTwo from "$lib/components/icons/SquaresTwoByTwo.svelte"
  
  export let dateRange: string = "7d"
  export let onDateRangeChange: () => void

  // Date range options
  const dateRangeOptions = [
    { value: "1d", label: "Today" },
    { value: "7d", label: "7 Days" },
    { value: "30d", label: "30 Days" },
  ]

  function handleChange() {
    onDateRangeChange()
  }
</script>

<div class="flex flex-col lg:flex-row lg:items-center lg:justify-between mb-8">
  <div>
    <h1 class="text-3xl font-bold text-base-content flex items-center gap-3">
      <SquaresTwoByTwo className="w-8 h-8" />
      Nutrition Dashboard
    </h1>
    <p class="text-base-content-lighter mt-2">
      Track your nutrition trends, goals, and insights
    </p>
  </div>
  
  <!-- Date Range Selector -->
  <div class="mt-4 lg:mt-0">
    <div class="flex items-center gap-2">
      <label class="text-sm font-medium text-base-content" for="date-range">
        Time Period:
      </label>
      <select
        id="date-range"
        bind:value={dateRange}
        on:change={handleChange}
        class="select select-bordered select-sm"
      >
        {#each dateRangeOptions as option}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
    </div>
  </div>
</div>
