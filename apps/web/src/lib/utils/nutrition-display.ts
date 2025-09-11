import type { paths } from "$lib/api/schema"
import { isRestrictedNutrient } from "$lib/utils/nutrients"

type GoalsResponse =
  paths["/goals"]["get"]["responses"]["200"]["content"]["application/json"]
type Goals = GoalsResponse["goals"]

/**
 * Safely extracts and validates a nutrient value from the nutrients object
 */
export function getNutrientValue(key: string, nutrients: Record<string, number>): number {
  const value = nutrients[key]
  return typeof value === "number" && isFinite(value) ? value : 0
}

/**
 * Formats a numeric value for display
 */
export function formatValue(value: number): string {
  if (!isFinite(value) || value === 0) return "0"
  if (value < 1) return value.toFixed(1)
  return value.toFixed(0)
}

/**
 * Checks if a nutrient's upper limit is exceeded
 */
export function isUpperLimitExceeded(
  key: string, 
  goals: Goals | null, 
  nutrients: Record<string, number>
): boolean {
  if (!goals?.upper_limits) return false
  if (goals.upper_limits[key] === undefined) return false
  
  const current = getNutrientValue(key, nutrients)
  const limit = goals.upper_limits[key]
  
  if (limit === 0) return current > 0
  return current > limit
}

/**
 * Calculates progress percentage for display in progress bars (capped at 100%)
 */
export function getProgress(
  key: string,
  goals: Goals | null,
  nutrients: Record<string, number>,
  showLimitsOnly: boolean
): number {
  if (!goals) return 0

  const current = getNutrientValue(key, nutrients)
  const hasTarget = goals?.targets?.[key] !== undefined
  const hasUpperLimit = goals?.upper_limits?.[key] !== undefined
  
  // Nutrients that should always be in the "minimize these" section
  const isRestricted = isRestrictedNutrient(key)

  // For nutrients in showLimitsOnly mode or restricted nutrients, use upper limit logic
  if (showLimitsOnly || (hasUpperLimit && isRestricted)) {
    if (hasUpperLimit) {
      const limit = goals.upper_limits[key]
      if (limit === 0) {
        // For zero limits (like trans fat), any amount is over
        return current > 0 ? 100 : 0
      }
      // For upper limits, "progress" is how close to the limit
      const progress = (current / limit) * 100
      return isFinite(progress) ? Math.min(progress, 100) : 0
    }
  }

  // For regular nutrients (including those with custom upper limits like sodium),
  // prefer target-based progress if available
  if (hasTarget) {
    const progress = (current / goals.targets[key]) * 100
    return isFinite(progress) ? Math.min(progress, 100) : 0
  }

  // Fallback: if no target but has upper limit (shouldn't normally happen in regular sections)
  if (hasUpperLimit) {
    const limit = goals.upper_limits[key]
    if (limit === 0) return current > 0 ? 100 : 0
    const progress = (current / limit) * 100
    return isFinite(progress) ? Math.min(progress, 100) : 0
  }

  return 0
}

/**
 * Calculates actual progress percentage (not capped, for display purposes)
 */
export function getActualProgress(
  key: string,
  goals: Goals | null,
  nutrients: Record<string, number>
): number {
  if (!goals) return 0

  const current = getNutrientValue(key, nutrients)

  // Prioritize targets over upper limits for percentage calculation
  if (goals?.targets?.[key] !== undefined) {
    const progress = (current / goals.targets[key]) * 100
    return isFinite(progress) ? progress : 0
  }

  // Fall back to upper limit only if no target exists
  if (goals?.upper_limits?.[key] !== undefined) {
    const limit = goals.upper_limits[key]
    if (limit === 0) {
      return current > 0 ? 200 : 0 // Show high percentage for any trans fat
    }
    const progress = (current / limit) * 100
    return isFinite(progress) ? progress : 0
  }

  return 0
}

/**
 * Determines the CSS class for progress bar styling
 */
export function getProgressBarClass(
  key: string,
  progress: number,
  goals: Goals | null,
  showMealContribution: boolean,
  showLimitsOnly: boolean,
  nutrients: Record<string, number>
): string {
  if (!goals) return "progress-primary"
  
  const hasUpperLimit = goals?.upper_limits?.[key] !== undefined
  const isRestricted = isRestrictedNutrient(key)

  if (showMealContribution) {
    // Handle upper limit exceeded case first (applies to all nutrients with limits)
    if (hasUpperLimit && isUpperLimitExceeded(key, goals, nutrients)) {
      return "progress-error"
    }

    // For restricted nutrients with upper limits that aren't exceeded, use lighter color
    if (hasUpperLimit && isRestricted) {
      return "progress-lighter"
    }

    // All other cases (no upper limit, or upper limit not exceeded for non-restricted nutrients)
    return "progress-accent"
  }

  // For nutrients that have upper limits but are NOT restricted nutrients
  // (like sodium with custom upper limit), check if limit is exceeded
  if (hasUpperLimit && !isRestricted) {
    if (isUpperLimitExceeded(key, goals, nutrients)) return "progress-error"  // Red when exceeded
    // Otherwise fall through to normal target-based coloring
  }

  // For nutrients in the minimize section (or showLimitsOnly mode)
  if ((hasUpperLimit && isRestricted) || showLimitsOnly) {
    // For upper limits: lighter gray until hitting the limit, then red
    if (progress >= 100) return "progress-error"
    return "progress-lighter"
  }
  
  // Standard target-based coloring for regular nutrients
  if (progress >= 80) return "progress-success"
  if (progress >= 50) return "progress-warning"
  return "progress-primary"
}

/**
 * Gets the daily target/limit text for display
 */
export function getDailyText(key: string, goals: Goals | null): string {
  if (!goals) return ""

  const target = goals.targets?.[key]
  const limit = goals.upper_limits?.[key]

  if (target !== undefined) {
    return `${formatValue(target)} ${goals.units[key] || ""}`
  } else if (limit !== undefined) {
    return `${formatValue(limit)} ${goals.units[key] || ""} limit`
  }

  return ""
}

/**
 * Gets overage warning text when limits/targets are exceeded
 */
export function getOverageText(
  key: string,
  current: number,
  goals: Goals | null
): string {
  if (!goals) return ""

  const hasTarget = goals?.targets?.[key] !== undefined
  const hasUpperLimit = goals?.upper_limits?.[key] !== undefined

  // Check if upper limit is exceeded (this triggers the warning)
  if (hasUpperLimit) {
    const limit = goals.upper_limits[key]
    if (limit === 0 && current > 0) {
      return "⚠️"
    }
    if (current > limit) {
      const overage = (current / limit - 1) * 100
      // Clarify this is about upper limit, especially when there's also a target
      const limitType = hasTarget ? "upper limit" : "limit"
      return isFinite(overage)
        ? `⚠️ ${overage.toFixed(0)}% over ${limitType}`
        : `⚠️ Over ${limitType}`
    }
  }

  // For nutrients that exceed their target but not their upper limit
  if (hasTarget) {
    const actualProgress = (current / goals.targets[key]) * 100
    if (isFinite(actualProgress) && actualProgress > 100) {
      const overage = actualProgress - 100
      return isFinite(overage) ? `+${overage.toFixed(0)}% over target` : "+Over target"
    }
  }

  return ""
}
