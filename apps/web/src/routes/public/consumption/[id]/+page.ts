import type { PageLoad } from "./$types"
import { apiClient } from "$lib/api/client"

export const load: PageLoad = async ({ params }) => {
  const id = params.id

  // Use the new public consumption endpoint that doesn't require authentication
  const consumptionRes = await apiClient.GET("/public/consumption/{id}", { 
    params: { path: { id } } 
  })

  return {
    consumption: consumptionRes.data,
    error: consumptionRes.error
  }
}