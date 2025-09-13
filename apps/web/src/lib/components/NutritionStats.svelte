<script lang="ts">
  import Zap from "./icons/Zap.svelte"
  import Dumbbell from "./icons/Dumbbell.svelte"
  import Avocado from "./icons/Avocado.svelte"
  import Wheat from "./icons/Wheat.svelte"

  export let calories: number
  export let protein: number
  export let carbs: number
  export let fat: number
  export let size: "normal" | "compact" | "minimal" = "normal"
  export let className: string = ""
</script>

{#if size === "minimal"}
  <!-- Minimal version for cards - responsive layout -->
  <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 {className}">
    <!-- Mobile: compact layout, Desktop: original layout -->
    <div class="flex flex-col items-center text-center">
      <Zap className="w-5 h-5 sm:w-8 sm:h-8 text-honey mb-0.5 sm:mb-1" />
      <div class="text-sm sm:text-xxl font-bold text-honey">{calories}</div>
      <div class="text-[10px] sm:text-xs text-gray-500 leading-none sm:leading-normal">
        <span class="sm:hidden">cal</span>
        <span class="hidden sm:inline">calories</span>
      </div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Dumbbell className="w-5 h-5 sm:w-8 sm:h-8 text-secondary mb-0.5 sm:mb-1" />
      <div class="text-sm sm:text-lg font-bold text-secondary">{protein.toFixed(1)}g</div>
      <div class="text-[10px] sm:text-xs text-gray-500 leading-none sm:leading-normal">protein</div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Avocado className="w-5 h-5 sm:w-8 sm:h-8 text-accent mb-0.5 sm:mb-1" />
      <div class="text-sm sm:text-lg font-bold text-accent">{fat.toFixed(1)}g</div>
      <div class="text-[10px] sm:text-xs text-gray-500 leading-none sm:leading-normal">fat</div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Wheat className="w-5 h-5 sm:w-8 sm:h-8 text-warning mb-0.5 sm:mb-1" />
      <div class="text-sm sm:text-lg font-bold text-warning">{carbs.toFixed(1)}g</div>
      <div class="text-[10px] sm:text-xs text-gray-500 leading-none sm:leading-normal">carbs</div>
    </div>
  </div>
{:else}
  <!-- Original DaisyUI stats version -->
  <div
    class="stats stats-vertical lg:stats-horizontal {size === 'normal' ? 'shadow-xl' : ''} w-full {className}"
  >
    <div class="stat">
      <div class="stat-figure text-honey">
        <Zap className="w-8 h-8" />
      </div>
      <div class="stat-title">Calories</div>
      <div class="stat-value text-honey {size === 'compact' ? 'text-xxl font-bold' : ''}">
        {calories}
      </div>
      {#if size === "normal"}
        <div class="stat-desc">Total</div>
      {/if}
    </div>

    <div class="stat">
      <div class="stat-figure text-secondary">
        <Dumbbell className="w-8 h-8" />
      </div>
      <div class="stat-title">Protein</div>
      <div
        class="stat-value text-secondary {size === 'compact' ? 'text-xxl font-bold' : ''}"
      >
        {protein.toFixed(1)}g
      </div>
      {#if size === "normal"}
        <div class="stat-desc">Essential for muscle</div>
      {/if}
    </div>

    <div class="stat">
      <div class="stat-figure text-accent">
        <Avocado className="w-8 h-8" />
      </div>
      <div class="stat-title">Fat</div>
      <div class="stat-value text-accent {size === 'compact' ? 'text-xxl font-bold' : ''}">
        {fat.toFixed(1)}g
      </div>
      {#if size === "normal"}
        <div class="stat-desc">Healthy fats</div>
      {/if}
    </div>

    <div class="stat">
      <div class="stat-figure text-warning">
        <Wheat className="w-8 h-8" />
      </div>
      <div class="stat-title">Carbs</div>
      <div class="stat-value text-warning {size === 'compact' ? 'text-xxl font-bold' : ''}">
        {carbs.toFixed(1)}g
      </div>
      {#if size === "normal"}
        <div class="stat-desc">Energy source</div>
      {/if}
    </div>
  </div>
{/if}
