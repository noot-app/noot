<script lang="ts">
  import { formatWeight, convertWeight, type Units } from "$lib/stores/units";

  export let calories: number;
  export let protein: number;
  export let carbs: number;
  export let fat: number;
  export let units: Units = 'grams';
  export let size: 'normal' | 'compact' = 'normal';
  export let className: string = '';

  function convertNutrientValue(value: number): number {
    return convertWeight(value, 'grams', units);
  }
</script>

<div class="stats stats-vertical lg:stats-horizontal shadow-xl w-full {className}">
  <div class="stat">
    <div class="stat-figure text-primary">
      <!-- Fire/Star icon for calories -->
      <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
        <path d="M12 2L10.09 8.26L4 7L7.91 13.74L8 21L12 18L16 21L16.09 13.74L20 7L13.91 8.26L12 2Z"/>
      </svg>
    </div>
    <div class="stat-title">Calories</div>
    <div class="stat-value text-primary {size === 'compact' ? 'text-lg' : ''}">{calories}</div>
    {#if size === 'normal'}
      <div class="stat-desc">Total</div>
    {/if}
  </div>

  <div class="stat">
    <div class="stat-figure text-secondary">
      <!-- Flexing arm icon for protein -->
      <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
        <path d="M12 2C13.1 2 14 2.9 14 4C14 5.1 13.1 6 12 6C10.9 6 10 5.1 10 4C10 2.9 10.9 2 12 2M21 9V7L15 9.5V12.5L21 15V13L19 12.2V9.8L21 9M3 9V7L9 9.5V12.5L3 15V13L5 12.2V9.8L3 9M15 18.5C15 19.9 13.9 21 12.5 21S10 19.9 10 18.5V16H15V18.5M8 16V18.5C8 20.43 9.57 22 11.5 22S15 20.43 15 18.5V16C15 14.9 14.1 14 13 14H11C9.9 14 9 14.9 9 16H8Z"/>
      </svg>
    </div>
    <div class="stat-title">Protein</div>
    <div class="stat-value text-secondary {size === 'compact' ? 'text-lg' : ''}">
      {formatWeight(convertNutrientValue(protein), units)}
    </div>
    {#if size === 'normal'}
      <div class="stat-desc">Essential for muscle</div>
    {/if}
  </div>

  <div class="stat">
    <div class="stat-figure text-accent">
      <!-- Wheat/Grain icon for carbs -->
      <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
        <path d="M12 2L11.09 8.26L7 7L8.91 13.74L12 14.74L15.09 13.74L17 7L12.91 8.26L12 2M12 16.74L15.09 17.74L14.09 22L12 21L9.91 22L8.91 17.74L12 16.74Z"/>
      </svg>
    </div>
    <div class="stat-title">Carbs</div>
    <div class="stat-value text-accent {size === 'compact' ? 'text-lg' : ''}">
      {formatWeight(convertNutrientValue(carbs), units)}
    </div>
    {#if size === 'normal'}
      <div class="stat-desc">Energy source</div>
    {/if}
  </div>

  <div class="stat">
    <div class="stat-figure text-warning">
      <!-- Avocado/Oil drop icon for fat -->
      <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
        <path d="M12 2C8 2 5 5 5 9C5 14.25 12 22 12 22S19 14.25 19 9C19 5 16 2 12 2M12 7C13.1 7 14 7.9 14 9S13.1 11 12 11S10 10.1 10 9S10.9 7 12 7Z"/>
      </svg>
    </div>
    <div class="stat-title">Fat</div>
    <div class="stat-value text-warning {size === 'compact' ? 'text-lg' : ''}">
      {formatWeight(convertNutrientValue(fat), units)}
    </div>
    {#if size === 'normal'}
      <div class="stat-desc">Healthy fats</div>
    {/if}
  </div>
</div>
