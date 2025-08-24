<script lang="ts">
  import type { components } from "$lib/api/schema";
  import NutrientCategoryDisplay from "./NutrientCategoryDisplay.svelte";

  type CompleteNutrient = components["schemas"]["CompleteNutrient"];

  export let nutrients: CompleteNutrient;

  // Convert CompleteNutrient to the format expected by NutrientCategoryDisplay
  $: nutrientData = nutrients ? Object.fromEntries(
    Object.entries(nutrients).map(([key, value]) => [key, typeof value === 'number' ? value : 0])
  ) : {};
</script>

<NutrientCategoryDisplay 
  nutrients={nutrientData} 
  showProgress={true}
  showGoals={true}
  isExpandable={true}
  title="View Nutrient Profile"
/>