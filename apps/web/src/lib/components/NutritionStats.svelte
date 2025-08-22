<script lang="ts">
  import Icon from './Icon.svelte';
  
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
      <Icon name="zap" size="lg" />
    </div>
    <div class="stat-title">Calories</div>
    <div class="stat-value text-primary {size === 'compact' ? 'text-lg' : ''}">{calories}</div>
    {#if size === 'normal'}
      <div class="stat-desc">Total</div>
    {/if}
  </div>

  <div class="stat">
    <div class="stat-figure text-secondary">
      <Icon name="dumbbell" size="lg" />
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
      <Icon name="avocado" size="lg" />
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
      <Icon name="wheat" size="lg" />
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
