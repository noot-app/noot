<script lang="ts">
  import ConsumptionCard from "./ConsumptionCard.svelte"
  import type { paths } from "$lib/api/schema"

  // Type definitions
  type ConsumptionsResponse = paths["/consumptions"]["get"]["responses"]["200"]["content"]["application/json"]
  type Consumption = ConsumptionsResponse["consumptions"][0]

  // Props
  export let consumption: Consumption
</script>

<ConsumptionCard 
  {consumption}
  variant="compact"
  clickable={false}
/>
