/**
 * Restricted nutrients are those that should generally be minimized in the diet.
 * These nutrients appear in the "Upper Limits (Minimize These)" section
 * rather than in their regular category sections.
 */
export const RESTRICTED_NUTRIENTS = [
  'added_sugars_g',
  'saturated_fat_g', 
  'trans_fat_g',
  'cholesterol_mg',
  'alcohol_g',
  'caffeine_mg'
] as const

/**
 * Type for restricted nutrient keys
 */
export type RestrictedNutrient = typeof RESTRICTED_NUTRIENTS[number]

/**
 * Check if a nutrient key is a restricted nutrient
 */
export function isRestrictedNutrient(nutrientKey: string): boolean {
  return RESTRICTED_NUTRIENTS.includes(nutrientKey as RestrictedNutrient)
}

/**
 * Filter nutrients object to exclude disabled nutrients
 */
export function filterDisabledNutrients<T extends Record<string, any>>(
  nutrients: T, 
  disabledNutrients: string[] = []
): T {
  if (!disabledNutrients || disabledNutrients.length === 0) {
    return nutrients
  }

  const disabledSet = new Set(disabledNutrients)
  const filtered = {} as T
  
  for (const [key, value] of Object.entries(nutrients)) {
    if (!disabledSet.has(key)) {
      filtered[key as keyof T] = value
    }
  }
  
  return filtered
}

/**
 * Filter array of nutrient keys to exclude disabled nutrients
 */
export function filterDisabledNutrientKeys(
  nutrientKeys: string[], 
  disabledNutrients: string[] = []
): string[] {
  if (!disabledNutrients || disabledNutrients.length === 0) {
    return nutrientKeys
  }

  const disabledSet = new Set(disabledNutrients)
  return nutrientKeys.filter(key => !disabledSet.has(key))
}
