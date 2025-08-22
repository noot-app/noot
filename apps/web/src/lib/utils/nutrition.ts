/**
 * Utility functions for nutrition data transformation
 */

/**
 * Transform a NutritionSummary object by removing the "total_" prefix from field names
 * This allows dynamic handling of nutrition data without hardcoded field mappings
 * 
 * @param summary - The NutritionSummary object from API response
 * @returns Record<string, number> with normalized field names (total_ prefix removed)
 */
export function transformNutritionSummary(summary: any): Record<string, number> {
  if (!summary || typeof summary !== 'object') {
    return {};
  }

  const transformed: Record<string, number> = {};

  // Transform all total_* fields by removing the "total_" prefix
  for (const [key, value] of Object.entries(summary)) {
    if (typeof key === 'string' && key.startsWith('total_')) {
      const normalizedKey = key.replace(/^total_/, '');
      // Only include if the value is a number (not null/undefined)
      if (typeof value === 'number') {
        transformed[normalizedKey] = value;
      }
    }
  }

  return transformed;
}

/**
 * Transform a CompleteNutrient object to match the expected format
 * This handles the case where nutrition data doesn't have the "total_" prefix
 * 
 * @param nutrition - The CompleteNutrient object from API response
 * @returns Record<string, number> with the nutrition data
 */
export function transformCompleteNutrient(nutrition: any): Record<string, number> {
  if (!nutrition || typeof nutrition !== 'object') {
    return {};
  }

  const transformed: Record<string, number> = {};

  // Copy all numeric fields directly
  for (const [key, value] of Object.entries(nutrition)) {
    if (typeof value === 'number') {
      transformed[key] = value || 0;
    }
  }

  return transformed;
}