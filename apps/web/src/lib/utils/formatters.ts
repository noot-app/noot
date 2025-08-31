/**
 * Formatting utilities for consistent display across the application
 */

/**
 * Format a number to a specified number of decimal places
 * Returns "0" for invalid/null values
 */
export function formatNumber(
  value: number | null | undefined,
  decimals: number = 1,
): string {
  if (value == null || !isFinite(value)) return "0"
  return value.toFixed(decimals)
}

/**
 * Format a nutrition value with grams unit
 */
export function formatGrams(
  value: number | null | undefined,
  decimals: number = 1,
): string {
  return `${formatNumber(value, decimals)}g`
}

/**
 * Format a nutrition value with milligrams unit
 */
export function formatMilligrams(
  value: number | null | undefined,
  decimals: number = 1,
): string {
  return `${formatNumber(value, decimals)}mg`
}

/**
 * Format a nutrition value with micrograms unit
 */
export function formatMicrograms(
  value: number | null | undefined,
  decimals: number = 1,
): string {
  return `${formatNumber(value, decimals)}mcg`
}

/**
 * Format calories (no unit, rounded to whole number)
 */
export function formatCalories(value: number | null | undefined): string {
  return formatNumber(value, 0)
}

/**
 * Format a percentage value
 */
export function formatPercentage(
  value: number | null | undefined,
  decimals: number = 1,
): string {
  return `${formatNumber(value, decimals)}%`
}

/**
 * Format date for display
 */
export function formatDate(date: Date | string): string {
  const d = typeof date === "string" ? new Date(date) : date
  return d.toLocaleDateString()
}

/**
 * Format datetime for display
 */
export function formatDateTime(date: Date | string): string {
  const d = typeof date === "string" ? new Date(date) : date
  return d.toLocaleString()
}

/**
 * Format a nutrient name for display (capitalize and replace underscores)
 */
export function formatNutrientName(name: string): string {
  return (
    name
      .replace(/_/g, " ")
      .replace(/\b\w/g, (l) => l.toUpperCase())
      // Remove unit suffixes since they're shown separately
      .replace(/ Mcg$/, "")
      .replace(/ Mg$/, "")
      .replace(/ G$/, "")
  )
}
