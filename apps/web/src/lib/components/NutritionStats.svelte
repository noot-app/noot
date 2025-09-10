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
  <!-- Minimal version for cards -->
  <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 {className}">
    <div class="flex flex-col items-center text-center">
      <Zap className="w-8 h-8 text-honey mb-1" />
      <div class="text-xxl font-bold text-honey">{calories}</div>
      <div class="text-xs text-gray-500">calories</div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Dumbbell className="w-8 h-8 text-secondary mb-1" />
      <div class="text-lg font-bold text-secondary">{protein.toFixed(1)}g</div>
      <div class="text-xs text-gray-500">protein</div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Avocado className="w-8 h-8 text-accent mb-1" />
      <div class="text-lg font-bold text-accent">{fat.toFixed(1)}g</div>
      <div class="text-xs text-gray-500">fat</div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Wheat className="w-8 h-8 text-warning mb-1" />
      <div class="text-lg font-bold text-warning">{carbs.toFixed(1)}g</div>
      <div class="text-xs text-gray-500">carbs</div>
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
