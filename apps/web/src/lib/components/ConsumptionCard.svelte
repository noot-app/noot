<script lang="ts">
  import { goto } from '$app/navigation'
  import NutritionStats from "./NutritionStats.svelte"
  import Label from "./Label.svelte"
  import Plus from "./icons/Plus.svelte"
  import Eye from "./icons/Eye.svelte"
  import type { paths } from "$lib/api/schema"

  // Type definitions
  type ConsumptionsResponse = paths["/consumptions"]["get"]["responses"]["200"]["content"]["application/json"]
  type Consumption = ConsumptionsResponse["consumptions"][0]

  // Props
  export let consumption: Consumption
  export let mode: 'readonly' | 'actions' | 'clickable' = 'readonly'
  export let clickable = false  // Simple clickable prop
  export let showNutrition = true
  export let showLabels = true
  export let showTimestamp = true
  export let showNote = true
  export let variant: 'default' | 'compact' | 'minimal' = 'default'
  export let className = ''
  
  // Event handlers (optional)
  export let onReLog: ((consumption: Consumption) => void) | undefined = undefined
  export let onView: ((consumption: Consumption) => void) | undefined = undefined
  export let onClick: ((consumption: Consumption) => void) | undefined = undefined

  // Internal state
  $: isClickable = clickable || mode === 'clickable'
  $: showActions = mode === 'actions'

  // Handle click
  function handleClick() {
    if (isClickable) {
      if (onClick) {
        onClick(consumption)
      } else {
        // Default navigation to consumption detail page
        goto(`/consumptions/${consumption.id}`)
      }
    }
  }

  // Handle re-log
  function handleReLog() {
    if (onReLog) {
      onReLog(consumption)
    }
  }

  // Handle view
  function handleView() {
    if (onView) {
      onView(consumption)
    }
  }

  // Card content styles based on variant
  $: contentClasses = [
    'group relative overflow-hidden transition-all duration-300 ease-out',
    'bg-white/50 backdrop-blur-sm border border-black/[0.08] rounded-2xl',
    'hover:shadow-lg hover:shadow-black/5 hover:border-black/[0.12]',
    isClickable ? 'cursor-pointer' : '',
    variant === 'default' ? 'p-6' : '',
    variant === 'compact' ? 'p-4' : '',
    variant === 'minimal' ? 'p-3' : '',
  ].filter(Boolean).join(' ')

  $: spacingClass = variant === 'minimal' ? 'space-y-1.5' : variant === 'compact' ? 'space-y-2' : 'space-y-3'
  $: innerPaddingClass = variant === 'minimal' ? 'px-2' : variant === 'compact' ? 'px-3' : 'px-4'
  $: nutritionSize = variant === 'minimal' ? 'minimal' as const : variant === 'compact' ? 'minimal' as const : 'compact' as const
</script>

<div class="card bg-base-200 {className}">
  <div class="card-body p-4">
    {#if showActions}
      <!-- Layout with actions -->
      <div class="flex items-center justify-between">
        <div class="flex-1">
          <!-- Content area -->
          {#if isClickable && consumption?.id}
            <a 
              href="/consumptions/{consumption.id}"
              class={contentClasses}
            >
              <!-- Subtle hover gradient overlay -->
              <div class="absolute inset-0 bg-gradient-to-br from-black/[0.02] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>

              <div class={spacingClass}>
                <div class={innerPaddingClass}>
                  <!-- Title/Transcript Section -->
                  {#if consumption?.title && consumption?.transcript}
                    <div class="space-y-1">
                      <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
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
                      <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
                        {consumption.title}
                      </h3>
                    </div>
                  {:else if consumption?.transcript}
                    <div>
                      <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
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
                </div>

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
                  <div class="{innerPaddingClass} flex items-center justify-between gap-3 pt-1">
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
            </a>
          {:else}
            <!-- svelte-ignore a11y-click-events-have-key-events -->
            <!-- svelte-ignore a11y-no-static-element-interactions -->
            <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
            <div 
              class={contentClasses}
              on:click={handleClick}
              role={isClickable ? 'button' : undefined}
              tabindex={isClickable ? 0 : undefined}
            >
              <!-- Subtle hover gradient overlay -->
              {#if isClickable}
                <div class="absolute inset-0 bg-gradient-to-br from-black/[0.02] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
              {/if}

              <div class={spacingClass}>
                <div class={innerPaddingClass}>
                  <!-- Title/Transcript Section -->
                  {#if consumption?.title && consumption?.transcript}
                    <div class="space-y-1">
                      <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
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
                      <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
                        {consumption.title}
                      </h3>
                    </div>
                  {:else if consumption?.transcript}
                    <div>
                      <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
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
                </div>

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
                  <div class="{innerPaddingClass} flex items-center justify-between gap-3 pt-1">
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
          {/if}
        </div>
        
        <!-- Action buttons -->
        <div class="flex flex-col gap-2 ml-4">
          <!-- Re-log button -->
          {#if onReLog}
            <button
              class="btn btn-primary"
              on:click={handleReLog}
              title="Re-log this consumption"
            >
              <Plus className="w-5 h-5" />
            </button>
          {/if}
          
          <!-- View button -->
          {#if onView}
            <button
              class="btn btn-outline"
              on:click={handleView}
              title="View full consumption details"
            >
              <Eye className="w-5 h-5" />
            </button>
          {:else if consumption?.id}
            <a
              href="/consumptions/{consumption.id}"
              class="btn btn-outline"
              title="View full consumption details"
            >
              <Eye className="w-5 h-5" />
            </a>
          {/if}
        </div>
      </div>
    {:else}
      <!-- Standalone content without actions -->
      {#if isClickable && consumption?.id}
        <a 
          href="/consumptions/{consumption.id}"
          class={contentClasses}
        >
          <!-- Subtle hover gradient overlay -->
          <div class="absolute inset-0 bg-gradient-to-br from-black/[0.02] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>

          <div class={spacingClass}>
            <div class={innerPaddingClass}>
              <!-- Title/Transcript Section -->
              {#if consumption?.title && consumption?.transcript}
                <div class="space-y-1">
                  <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
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
                  <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
                    {consumption.title}
                  </h3>
                </div>
              {:else if consumption?.transcript}
                <div>
                  <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
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
            </div>

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
              <div class="{innerPaddingClass} flex items-center justify-between gap-3 pt-1">
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
        </a>
      {:else}
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <!-- svelte-ignore a11y-no-static-element-interactions -->
        <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
        <div 
          class={contentClasses}
          on:click={handleClick}
          role={isClickable ? 'button' : undefined}
          tabindex={isClickable ? 0 : undefined}
        >
          <!-- Subtle hover gradient overlay -->
          {#if isClickable}
            <div class="absolute inset-0 bg-gradient-to-br from-black/[0.02] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
          {/if}

          <div class={spacingClass}>
            <div class={innerPaddingClass}>
              <!-- Title/Transcript Section -->
              {#if consumption?.title && consumption?.transcript}
                <div class="space-y-1">
                  <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
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
                  <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
                    {consumption.title}
                  </h3>
                </div>
              {:else if consumption?.transcript}
                <div>
                  <h3 class="font-semibold text-gray-900 leading-snug line-clamp-2 {variant === 'minimal' ? 'text-lg' : 'text-xl'}">
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
            </div>

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
              <div class="{innerPaddingClass} flex items-center justify-between gap-3 pt-1">
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
      {/if}
    {/if}
  </div>
</div>
