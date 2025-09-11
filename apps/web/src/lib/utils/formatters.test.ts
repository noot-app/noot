import { describe, it, expect } from 'vitest'
import {
  formatNumber,
  formatGrams,
  formatMilligrams,
  formatMicrograms,
  formatCalories,
  formatPercentage,
  formatDate,
  formatDateTime,
  formatNutrientName
} from './formatters'

describe('Formatter Utilities', () => {
  describe('formatNumber', () => {
    it('should format numbers to specified decimal places', () => {
      expect(formatNumber(123.456, 2)).toBe('123.46')
      expect(formatNumber(123.456, 1)).toBe('123.5')
      expect(formatNumber(123.456, 0)).toBe('123')
    })

    it('should handle null and undefined values', () => {
      expect(formatNumber(null)).toBe('0')
      expect(formatNumber(undefined)).toBe('0')
    })

    it('should handle invalid numbers', () => {
      expect(formatNumber(NaN)).toBe('0')
      expect(formatNumber(Infinity)).toBe('0')
      expect(formatNumber(-Infinity)).toBe('0')
    })

    it('should handle negative numbers', () => {
      expect(formatNumber(-123.456, 2)).toBe('-123.46')
      expect(formatNumber(-0.1, 1)).toBe('-0.1')
    })

    it('should handle zero values', () => {
      expect(formatNumber(0, 2)).toBe('0.00')
      expect(formatNumber(0)).toBe('0.0')
    })

    it('should use default decimal places', () => {
      expect(formatNumber(123.456)).toBe('123.5') // Default is 1
    })

    it('should handle very large numbers', () => {
      expect(formatNumber(999999.999, 0)).toBe('1000000')
      expect(formatNumber(1234567.89, 2)).toBe('1234567.89')
    })

    it('should handle very small numbers', () => {
      expect(formatNumber(0.00001, 5)).toBe('0.00001')
      expect(formatNumber(0.00001, 3)).toBe('0.000')
    })
  })

  describe('formatGrams', () => {
    it('should format numbers with gram unit', () => {
      expect(formatGrams(123.456)).toBe('123.5g')
      expect(formatGrams(0)).toBe('0.0g')
      expect(formatGrams(50.25, 2)).toBe('50.25g')
    })

    it('should handle null values', () => {
      expect(formatGrams(null)).toBe('0g')
      expect(formatGrams(undefined)).toBe('0g')
    })

    it('should respect decimal places', () => {
      expect(formatGrams(123.456, 0)).toBe('123g')
      expect(formatGrams(123.456, 3)).toBe('123.456g')
    })
  })

  describe('formatMilligrams', () => {
    it('should format numbers with milligram unit', () => {
      expect(formatMilligrams(123.456)).toBe('123.5mg')
      expect(formatMilligrams(0)).toBe('0.0mg')
      expect(formatMilligrams(50.25, 2)).toBe('50.25mg')
    })

    it('should handle null values', () => {
      expect(formatMilligrams(null)).toBe('0mg')
      expect(formatMilligrams(undefined)).toBe('0mg')
    })

    it('should respect decimal places', () => {
      expect(formatMilligrams(123.456, 0)).toBe('123mg')
      expect(formatMilligrams(123.456, 3)).toBe('123.456mg')
    })
  })

  describe('formatMicrograms', () => {
    it('should format numbers with microgram unit', () => {
      expect(formatMicrograms(123.456)).toBe('123.5mcg')
      expect(formatMicrograms(0)).toBe('0.0mcg')
      expect(formatMicrograms(50.25, 2)).toBe('50.25mcg')
    })

    it('should handle null values', () => {
      expect(formatMicrograms(null)).toBe('0mcg')
      expect(formatMicrograms(undefined)).toBe('0mcg')
    })

    it('should respect decimal places', () => {
      expect(formatMicrograms(123.456, 0)).toBe('123mcg')
      expect(formatMicrograms(123.456, 3)).toBe('123.456mcg')
    })
  })

  describe('formatCalories', () => {
    it('should format calories as whole numbers', () => {
      expect(formatCalories(123.456)).toBe('123')
      expect(formatCalories(123.9)).toBe('124')
      expect(formatCalories(0)).toBe('0')
    })

    it('should handle null values', () => {
      expect(formatCalories(null)).toBe('0')
      expect(formatCalories(undefined)).toBe('0')
    })

    it('should always round to whole numbers', () => {
      expect(formatCalories(0.1)).toBe('0')
      expect(formatCalories(0.5)).toBe('1')
      expect(formatCalories(0.9)).toBe('1')
    })
  })

  describe('formatPercentage', () => {
    it('should format numbers with percentage unit', () => {
      expect(formatPercentage(123.456)).toBe('123.5%')
      expect(formatPercentage(0)).toBe('0.0%')
      expect(formatPercentage(50.25, 2)).toBe('50.25%')
    })

    it('should handle null values', () => {
      expect(formatPercentage(null)).toBe('0%')
      expect(formatPercentage(undefined)).toBe('0%')
    })

    it('should respect decimal places', () => {
      expect(formatPercentage(123.456, 0)).toBe('123%')
      expect(formatPercentage(123.456, 3)).toBe('123.456%')
    })

    it('should handle typical percentage values', () => {
      expect(formatPercentage(100)).toBe('100.0%')
      expect(formatPercentage(33.333, 1)).toBe('33.3%')
      expect(formatPercentage(66.667, 1)).toBe('66.7%')
    })
  })

  describe('formatDate', () => {
    it('should format Date objects', () => {
      const date = new Date('2023-12-25T10:30:00Z')
      const result = formatDate(date)
      // Result will vary by locale, but should be a valid date string
      expect(result).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}|\d{4}-\d{2}-\d{2}|\d{1,2}\.\d{1,2}\.\d{4}/)
    })

    it('should format date strings', () => {
      const result = formatDate('2023-12-25T10:30:00Z')
      expect(result).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}|\d{4}-\d{2}-\d{2}|\d{1,2}\.\d{1,2}\.\d{4}/)
    })

    it('should handle ISO date strings', () => {
      const result = formatDate('2023-01-15')
      expect(result).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}|\d{4}-\d{2}-\d{2}|\d{1,2}\.\d{1,2}\.\d{4}/)
    })

    it('should be consistent between Date and string inputs', () => {
      const dateStr = '2023-06-15T12:00:00Z'
      const dateObj = new Date(dateStr)
      
      const resultFromString = formatDate(dateStr)
      const resultFromDate = formatDate(dateObj)
      
      expect(resultFromString).toBe(resultFromDate)
    })
  })

  describe('formatDateTime', () => {
    it('should format Date objects with time', () => {
      const date = new Date('2023-12-25T10:30:00Z')
      const result = formatDateTime(date)
      // Should include both date and time components
      expect(result).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}.*\d{1,2}:\d{2}|\d{4}-\d{2}-\d{2}.*\d{1,2}:\d{2}/)
    })

    it('should format datetime strings', () => {
      const result = formatDateTime('2023-12-25T10:30:00Z')
      expect(result).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}.*\d{1,2}:\d{2}|\d{4}-\d{2}-\d{2}.*\d{1,2}:\d{2}/)
    })

    it('should include more information than formatDate', () => {
      const dateStr = '2023-06-15T14:30:00Z'
      const dateResult = formatDate(dateStr)
      const dateTimeResult = formatDateTime(dateStr)
      
      expect(dateTimeResult.length).toBeGreaterThan(dateResult.length)
      expect(dateTimeResult).toContain(':') // Should contain time separator
    })
  })

  describe('formatNutrientName', () => {
    it('should capitalize words and replace underscores', () => {
      expect(formatNutrientName('total_protein')).toBe('Total Protein')
      expect(formatNutrientName('vitamin_c')).toBe('Vitamin C')
      expect(formatNutrientName('dietary_fiber')).toBe('Dietary Fiber')
    })

    it('should remove unit suffixes', () => {
      expect(formatNutrientName('calcium_mg')).toBe('Calcium')
      expect(formatNutrientName('vitamin_a_mcg')).toBe('Vitamin A')
      expect(formatNutrientName('protein_g')).toBe('Protein')
    })

    it('should handle single words', () => {
      expect(formatNutrientName('calories')).toBe('Calories')
      expect(formatNutrientName('sodium')).toBe('Sodium')
    })

    it('should handle already formatted names', () => {
      expect(formatNutrientName('Vitamin D')).toBe('Vitamin D')
      expect(formatNutrientName('Total Fat')).toBe('Total Fat')
    })

    it('should handle complex nutrient names', () => {
      expect(formatNutrientName('polyunsaturated_fat_g')).toBe('Polyunsaturated Fat')
      expect(formatNutrientName('pantothenic_acid_mg')).toBe('Pantothenic Acid')
      expect(formatNutrientName('vitamin_b12_mcg')).toBe('Vitamin B12')
    })

    it('should handle edge cases', () => {
      expect(formatNutrientName('')).toBe('')
      expect(formatNutrientName('_')).toBe(' ')
      expect(formatNutrientName('a_b_c_d')).toBe('A B C D')
    })

    it('should handle names with numbers', () => {
      expect(formatNutrientName('vitamin_b6_mg')).toBe('Vitamin B6')
      expect(formatNutrientName('omega3_epa_g')).toBe('Omega3 Epa')
    })

    it('should preserve case for already capitalized parts', () => {
      expect(formatNutrientName('vitamin_B12_mcg')).toBe('Vitamin B12')
      expect(formatNutrientName('omega_3_ALA_g')).toBe('Omega 3 ALA')
    })

    it('should handle multiple consecutive underscores', () => {
      expect(formatNutrientName('total__protein__g')).toBe('Total  Protein ')
    })
  })

  describe('Integration scenarios', () => {
    it('should format nutrition data consistently', () => {
      const nutritionData = {
        calories: 250.7,
        protein: 15.234,
        fat: 8.9,
        sodium: 456.78,
        vitaminC: 12.345
      }

      expect(formatCalories(nutritionData.calories)).toBe('251')
      expect(formatGrams(nutritionData.protein)).toBe('15.2g')
      expect(formatGrams(nutritionData.fat)).toBe('8.9g')
      expect(formatMilligrams(nutritionData.sodium)).toBe('456.8mg')
      expect(formatMilligrams(nutritionData.vitaminC)).toBe('12.3mg')
    })

    it('should handle zero values across all formatters', () => {
      expect(formatCalories(0)).toBe('0')
      expect(formatGrams(0)).toBe('0.0g')
      expect(formatMilligrams(0)).toBe('0.0mg')
      expect(formatMicrograms(0)).toBe('0.0mcg')
      expect(formatPercentage(0)).toBe('0.0%')
    })

    it('should handle null values across all formatters', () => {
      expect(formatCalories(null)).toBe('0')
      expect(formatGrams(null)).toBe('0g')
      expect(formatMilligrams(null)).toBe('0mg')
      expect(formatMicrograms(null)).toBe('0mcg')
      expect(formatPercentage(null)).toBe('0%')
    })

    it('should maintain consistency in precision', () => {
      const value = 123.456789
      
      expect(formatGrams(value, 2)).toBe('123.46g')
      expect(formatMilligrams(value, 2)).toBe('123.46mg')
      expect(formatMicrograms(value, 2)).toBe('123.46mcg')
      expect(formatPercentage(value, 2)).toBe('123.46%')
    })
  })
})