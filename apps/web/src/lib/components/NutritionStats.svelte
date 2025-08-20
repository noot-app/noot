<script lang="ts">
  export let calories: number;
  export let protein: number;
  export let carbs: number;
  export let fat: number;
  export let size: 'normal' | 'compact' = 'normal';
  export let className: string = '';
</script>

<div class="stats stats-vertical lg:stats-horizontal shadow-xl w-full {className}">
  <div class="stat">
    <div class="stat-figure text-primary">
      <!-- Zap icon for calories -->
      <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
        <path d="M13 2L3 14h7l-1 8 10-12h-7l1-8z"/>
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
      <!-- dumb bell icon for protein -->
      <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
        <path d="M 1 9 C 1 8.448 1.448 8 2 8 L 2.5 8 C 3.052 8 3.5 8.448 3.5 9 L 3.5 15 C 3.5 15.552 3.052 16 2.5 16 L 2 16 C 1.448 16 1 15.552 1 15 Z M 4 7 C 4 6.448 4.448 6 5 6 L 6 6 C 6.552 6 7 6.448 7 7 L 7 17 C 7 17.552 6.552 18 6 18 L 5 18 C 4.448 18 4 17.552 4 17 Z M 17 7 C 17 6.448 17.448 6 18 6 L 19 6 C 19.552 6 20 6.448 20 7 L 20 17 C 20 17.552 19.552 18 19 18 L 18 18 C 17.448 18 17 17.552 17 17 Z M 20.5 9 C 20.5 8.448 20.948 8 21.5 8 L 22 8 C 22.552 8 23 8.448 23 9 L 23 15 C 23 15.552 22.552 16 22 16 L 21.5 16 C 20.948 16 20.5 15.552 20.5 15 Z M 7 11 L 17 11 L 17 13 L 7 13 Z"/>
      </svg>
    </div>
    <div class="stat-title">Protein</div>
    <div class="stat-value text-secondary {size === 'compact' ? 'text-lg' : ''}">
      {protein.toFixed(1)}g
    </div>
    {#if size === 'normal'}
      <div class="stat-desc">Essential for muscle</div>
    {/if}
  </div>

  <div class="stat">
    <div class="stat-figure text-accent">
      <!-- Avocado/Oil drop icon for fat -->
      <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
        <path d="M 12 2 C 9 2 6 8 6 14 C 6 19 9 22 12 22 C 15 22 18 19 18 14 C 18 8 15 2 12 2 Z M 12 12 C 13.657 12 15 13.343 15 15 C 15 16.657 13.657 18 12 18 C 10.343 18 9 16.657 9 15 C 9 13.343 10.343 12 12 12 Z"/>
      </svg>
    </div>
    <div class="stat-title">Fat</div>
    <div class="stat-value text-accent {size === 'compact' ? 'text-lg' : ''}">
      {fat.toFixed(1)}g
    </div>
    {#if size === 'normal'}
      <div class="stat-desc">Healthy fats</div>
    {/if}
  </div>

  <div class="stat">
    <div class="stat-figure text-warning">
      <!-- Wheat/Grain icon for carbs -->
      <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
        <path d="M 12 3 C 8.686 3 6 5.686 6 9 L 6 18 C 6 19.657 7.343 21 9 21 L 15 21 C 16.657 21 18 19.657 18 18 L 18 9 C 18 5.686 15.314 3 12 3 Z M 9 8 L 15 8 C 15.552 8 16 8.448 16 9 C 16 9.552 15.552 10 15 10 L 9 10 C 8.448 10 8 9.552 8 9 C 8 8.448 8.448 8 9 8 Z M 9 12 L 15 12 C 15.552 12 16 12.448 16 13 C 16 13.552 15.552 14 15 14 L 9 14 C 8.448 14 8 13.552 8 13 C 8 12.448 8.448 12 9 12 Z"/>
      </svg>
    </div>
    <div class="stat-title">Carbs</div>
    <div class="stat-value text-warning {size === 'compact' ? 'text-lg' : ''}">
      {carbs.toFixed(1)}g
    </div>
    {#if size === 'normal'}
      <div class="stat-desc">Energy source</div>
    {/if}
  </div>
</div>
