import type { PageLoad } from "./$types"
import { apiClient } from "$lib/api/client"

export const load: PageLoad = async ({ params }) => {
  const id = params.id

  const [consumptionRes, goalsAutoRes, goalsDriRes] = await Promise.all([
    apiClient.GET("/consumption/{id}", { params: { path: { id } } }),
    apiClient.GET("/goals", { params: { query: { source: "auto" } } }),
    apiClient.GET("/goals", { params: { query: { source: "dri" } } }),
  ])

  return {
    consumption: consumptionRes.data,
    goalsAuto: goalsAutoRes.data?.goals ?? null,
    goalsDri: goalsDriRes.data?.goals ?? null,
  }
}
