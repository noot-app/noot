<script lang="ts">
  import NutritionStats from "./NutritionStats.svelte"
  import Label from "./Label.svelte"
  import type { paths } from "$lib/api/schema"

  // Type definitions
  type ConsumptionsResponse = paths["/consumptions"]["get"]["responses"]["200"]["content"]["application/json"]
  type Consumption = ConsumptionsResponse["consumptions"][0]

  // Props
  export let consumption: Consumption
  export let clickable = true
  export let showNutrition = true
  export let showLabels = true
  export let showTimestamp = true
  export let showNote = true
  export let variant: 'default' | 'compact' | 'minimal' = 'default'
  export let className = ''
  export let onclick: ((event: { consumption: Consumption }) => void) | undefined = undefined

  // Handle click
  function handleClick() {
    if (clickable && onclick) {
      onclick({ consumption })
    }
  }

  // Determine styles based on variant
  $: containerClasses = [
    'space-y-3 transition-all duration-200',
    clickable ? 'cursor-pointer hover:bg-base-200/50 focus:bg-base-200/50 focus:outline-none focus:ring-2 focus:ring-primary/20 rounded-lg p-3' : '',
    variant === 'compact' ? 'space-y-2' : '',
    variant === 'minimal' ? 'space-y-1' : '',
    className
  ].filter(Boolean).join(' ')

  $: nutritionSize = (variant === 'compact' || variant === 'minimal') ? 'compact' as const : 'normal' as const
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
<div 
  class={containerClasses}
  on:click={handleClick}
  role={clickable ? 'button' : undefined}
  tabindex={clickable ? 0 : undefined}
>
  <!-- Title/Transcript -->
  {#if consumption?.title && consumption?.transcript}
    <div>
      <p class="text-lg font-medium text-base-content line-clamp-2">
        {consumption.title}
      </p>
      {#if variant !== 'minimal'}
        <p class="text-sm text-base-content/70 line-clamp-2 mt-1">
          {consumption.transcript}
        </p>
      {/if}
    </div>
  {:else if consumption?.title}
    <div>
      <p class="text-lg font-medium text-base-content line-clamp-2">
        {consumption.title}
      </p>
    </div>
  {:else if consumption?.transcript}
    <div>
      <p class="text-lg font-medium text-base-content line-clamp-2">
        {consumption.transcript}
      </p>
    </div>
  {/if}

  <!-- Note -->
  {#if showNote && consumption?.note && variant !== 'minimal'}
    <div>
      <p class="text-xs text-base-content/60 italic line-clamp-1">
        {consumption.note}
      </p>
    </div>
  {/if}

  <!-- Nutrition Summary -->
  {#if showNutrition && consumption?.summary}
    <div class="bg-base-100 rounded-lg p-3">
      <NutritionStats
        calories={consumption.summary.totals.calories}
        protein={consumption.summary.totals.protein_g}
        carbs={consumption.summary.totals.total_carbs_g}
        fat={consumption.summary.totals.total_fat_g}
        size={nutritionSize}
        className="bg-transparent shadow-none"
      />
    </div>
  {/if}

  <!-- Labels -->
  {#if showLabels && consumption?.id && consumption?.labels && consumption.labels.length > 0}
    <div class="flex items-center gap-2 flex-wrap">
      <span class="text-xs text-base-content/60 font-medium">Labels:</span>
      {#each consumption.labels as label}
        <Label 
          name={label.name} 
          color={label.color} 
          size={variant === 'minimal' ? 'xs' : 'xs'} 
        />
      {/each}
    </div>
  {/if}

  <!-- Timestamp -->
  {#if showTimestamp && consumption?.created_at}
    <div>
      <p class="text-xs text-base-content/50">
        {new Date(consumption.created_at).toLocaleDateString()} at {new Date(consumption.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
      </p>
    </div>
  {/if}
</div>
