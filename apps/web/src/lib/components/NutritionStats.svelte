<script lang="ts">
  import Icon from "./Icon.svelte"

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
      <Icon name="zap" size="sm" className="text-primary mb-1" />
      <div class="text-sm font-semibold text-primary">{calories}</div>
      <div class="text-xs text-gray-500">calories</div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Icon name="dumbbell" size="sm" className="text-secondary mb-1" />
      <div class="text-sm font-semibold text-secondary">{protein.toFixed(1)}g</div>
      <div class="text-xs text-gray-500">protein</div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Icon name="avocado" size="sm" className="text-accent mb-1" />
      <div class="text-sm font-semibold text-accent">{fat.toFixed(1)}g</div>
      <div class="text-xs text-gray-500">fat</div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Icon name="wheat" size="sm" className="text-warning mb-1" />
      <div class="text-sm font-semibold text-warning">{carbs.toFixed(1)}g</div>
      <div class="text-xs text-gray-500">carbs</div>
    </div>
  </div>
{:else}
  <!-- Original DaisyUI stats version -->
  <div
    class="stats stats-vertical lg:stats-horizontal {size === 'normal' ? 'shadow-xl' : ''} w-full {className}"
  >
    <div class="stat">
      <div class="stat-figure text-primary">
        <Icon name="zap" size="lg" />
      </div>
      <div class="stat-title">Calories</div>
      <div class="stat-value text-primary {size === 'compact' ? 'text-lg' : ''}">
        {calories}
      </div>
      {#if size === "normal"}
        <div class="stat-desc">Total</div>
      {/if}
    </div>

    <div class="stat">
      <div class="stat-figure text-secondary">
        <Icon name="dumbbell" size="lg" />
      </div>
      <div class="stat-title">Protein</div>
      <div
        class="stat-value text-secondary {size === 'compact' ? 'text-lg' : ''}"
      >
        {protein.toFixed(1)}g
      </div>
      {#if size === "normal"}
        <div class="stat-desc">Essential for muscle</div>
      {/if}
    </div>

    <div class="stat">
      <div class="stat-figure text-accent">
        <Icon name="avocado" size="lg" />
      </div>
      <div class="stat-title">Fat</div>
      <div class="stat-value text-accent {size === 'compact' ? 'text-lg' : ''}">
        {fat.toFixed(1)}g
      </div>
      {#if size === "normal"}
        <div class="stat-desc">Healthy fats</div>
      {/if}
    </div>

    <div class="stat">
      <div class="stat-figure text-warning">
        <Icon name="wheat" size="lg" />
      </div>
      <div class="stat-title">Carbs</div>
      <div class="stat-value text-warning {size === 'compact' ? 'text-lg' : ''}">
        {carbs.toFixed(1)}g
      </div>
      {#if size === "normal"}
        <div class="stat-desc">Energy source</div>
      {/if}
    </div>
  </div>
{/if}
