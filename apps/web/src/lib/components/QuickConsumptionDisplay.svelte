<script lang="ts">
  import NutritionStats from "./NutritionStats.svelte"
  import Label from "./Label.svelte"
  import type { paths } from "$lib/api/schema"

  // Type definitions
  type ConsumptionsResponse = paths["/consumptions"]["get"]["responses"]["200"]["content"]["application/json"]
  type Consumption = ConsumptionsResponse["consumptions"][0]

  // Props
  export let consumption: Consumption
</script>

<div class="space-y-3">
  <!-- Transcript/Summary -->
  {#if consumption?.transcript}
    <div>
      <p class="text-lg font-medium text-base-content line-clamp-2">
        {consumption.transcript}
      </p>
    </div>
  {/if}

  <!-- Note -->
  {#if consumption?.note}
    <div>
      <p class="text-xs text-base-content/60 italic line-clamp-1">
        {consumption.note}
      </p>
    </div>
  {/if}

  <!-- Simple Nutrition Summary -->
  {#if consumption?.summary}
    <div class="bg-base-100 rounded-lg p-3">
      <NutritionStats
        calories={consumption.summary.totals.calories}
        protein={consumption.summary.totals.protein_g}
        carbs={consumption.summary.totals.total_carbs_g}
        fat={consumption.summary.totals.total_fat_g}
        size="compact"
        className="bg-transparent shadow-none"
      />
    </div>
  {/if}

  <!-- Labels (read-only, inline) -->
  {#if consumption?.id && consumption?.labels && consumption.labels.length > 0}
    <div class="flex items-center gap-2 flex-wrap">
      <span class="text-xs text-base-content/60 font-medium">Labels:</span>
      {#each consumption.labels as label}
        <Label name={label.name} color={label.color} size="xs" />
      {/each}
    </div>
  {/if}

  <!-- Timestamp -->
  {#if consumption?.created_at}
    <div>
      <p class="text-xs text-base-content/50">
        {new Date(consumption.created_at).toLocaleDateString()} at {new Date(consumption.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
      </p>
    </div>
  {/if}
</div>
