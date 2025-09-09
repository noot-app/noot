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
  <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 {className}">
    <div class="flex flex-col items-center text-center">
      <Icon name="zap" size="md" className="text-honey mb-2" />
      <div class="text-sm font-semibold text-honey">{calories}</div>
      <div class="text-xs text-gray-500">calories</div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Icon name="dumbbell" size="md" className="text-secondary mb-2" />
      <div class="text-sm font-semibold text-secondary">{protein.toFixed(1)}g</div>
      <div class="text-xs text-gray-500">protein</div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Icon name="avocado" size="md" className="text-accent mb-2" />
      <div class="text-sm font-semibold text-accent">{fat.toFixed(1)}g</div>
      <div class="text-xs text-gray-500">fat</div>
    </div>

    <div class="flex flex-col items-center text-center">
      <Icon name="wheat" size="md" className="text-warning mb-2" />
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
      <div class="stat-figure text-honey">
        <Icon name="zap" size="lg" />
      </div>
      <div class="stat-title">Calories</div>
      <div class="stat-value text-honey {size === 'compact' ? 'text-lg' : ''}">
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
