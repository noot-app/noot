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
    'group relative overflow-hidden transition-all duration-300 ease-out',
    'bg-white/50 backdrop-blur-sm border border-black/[0.08] rounded-2xl',
    'hover:shadow-lg hover:shadow-black/5 hover:border-black/[0.12]',
    clickable ? 'cursor-pointer' : '',
    variant === 'default' ? 'p-6' : '',
    variant === 'compact' ? 'p-4' : '',
    variant === 'minimal' ? 'p-3' : '',
    className
  ].filter(Boolean).join(' ')

  $: spacingClass = variant === 'minimal' ? 'space-y-2' : variant === 'compact' ? 'space-y-3' : 'space-y-4'
  $: nutritionSize = variant === 'minimal' ? 'minimal' as const : 'compact' as const
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
  <!-- Subtle hover gradient overlay -->
  {#if clickable}
    <div class="absolute inset-0 bg-gradient-to-br from-black/[0.02] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
  {/if}

  <div class={spacingClass}>
    <!-- Title/Transcript Section -->
    {#if consumption?.title && consumption?.transcript}
      <div class="space-y-1">
        <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 text-{variant === 'minimal' ? 'sm' : 'base'}">
          {consumption.title}
        </h3>
        {#if variant !== 'minimal'}
          <p class="text-gray-600 text-sm leading-relaxed line-clamp-2">
            {consumption.transcript}
          </p>
        {/if}
      </div>
    {:else if consumption?.title}
      <div>
        <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 text-{variant === 'minimal' ? 'sm' : 'base'}">
          {consumption.title}
        </h3>
      </div>
    {:else if consumption?.transcript}
      <div>
        <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 text-{variant === 'minimal' ? 'sm' : 'base'}">
          {consumption.transcript}
        </h3>
      </div>
    {/if}

    <!-- Note Section -->
    {#if showNote && consumption?.note && variant !== 'minimal'}
      <div class="px-3 py-2 bg-gray-50/80 rounded-lg border border-gray-100">
        <p class="text-xs text-gray-600 italic leading-relaxed line-clamp-2">
          {consumption.note}
        </p>
      </div>
    {/if}

    <!-- Nutrition Section -->
    {#if showNutrition && consumption?.summary}
      <div class="pt-2 border-t border-gray-100">
        <NutritionStats
          calories={consumption.summary.totals.calories}
          protein={consumption.summary.totals.protein_g}
          carbs={consumption.summary.totals.total_carbs_g}
          fat={consumption.summary.totals.total_fat_g}
          size={nutritionSize}
        />
      </div>
    {/if}

    <!-- Labels and Timestamp Section -->
    {#if (showLabels && consumption?.labels && consumption.labels.length > 0) || (showTimestamp && consumption?.created_at)}
      <div class="flex items-center justify-between gap-3 pt-1">
        <!-- Labels -->
        {#if showLabels && consumption?.labels && consumption.labels.length > 0}
          <div class="flex items-center gap-1.5 flex-wrap flex-1 min-w-0">
            {#each consumption.labels.slice(0, variant === 'minimal' ? 2 : 4) as label}
              <Label 
                name={label.name} 
                color={label.color} 
                size="xs"
              />
            {/each}
            {#if consumption.labels.length > (variant === 'minimal' ? 2 : 4)}
              <span class="text-xs text-gray-400 font-medium">
                +{consumption.labels.length - (variant === 'minimal' ? 2 : 4)}
              </span>
            {/if}
          </div>
        {/if}

        <!-- Timestamp -->
        {#if showTimestamp && consumption?.created_at}
          <div class="flex-shrink-0">
            <time class="text-xs text-gray-400 font-medium tabular-nums">
              {new Date(consumption.created_at).toLocaleDateString('en-US', { 
                month: 'short', 
                day: 'numeric',
                ...(variant !== 'minimal' && { hour: '2-digit', minute: '2-digit' })
              })}
            </time>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
