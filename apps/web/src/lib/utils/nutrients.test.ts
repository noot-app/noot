import { describe, it, expect } from 'vitest'
import { RESTRICTED_NUTRIENTS, isRestrictedNutrient, type RestrictedNutrient, filterDisabledNutrients, filterDisabledNutrientKeys } from './nutrients'

describe('nutrients utility', () => {
  describe('RESTRICTED_NUTRIENTS', () => {
    it('should contain all expected restricted nutrients', () => {
      const expected = [
        'added_sugars_g',
        'saturated_fat_g',
        'trans_fat_g',
        'cholesterol_mg',
        'alcohol_g',
        'caffeine_mg'
      ]
      
      expect(RESTRICTED_NUTRIENTS).toEqual(expected)
      expect(RESTRICTED_NUTRIENTS).toHaveLength(6)
    })

    it('should be a const array', () => {
      // TypeScript prevents mutation at compile time with 'as const'
      // This test just verifies the array is properly defined
      expect(Array.isArray(RESTRICTED_NUTRIENTS)).toBe(true)
      expect(RESTRICTED_NUTRIENTS.length).toBeGreaterThan(0)
    })
  })

  describe('isRestrictedNutrient', () => {
    it('should return true for restricted nutrients', () => {
      expect(isRestrictedNutrient('added_sugars_g')).toBe(true)
      expect(isRestrictedNutrient('saturated_fat_g')).toBe(true)
      expect(isRestrictedNutrient('trans_fat_g')).toBe(true)
      expect(isRestrictedNutrient('cholesterol_mg')).toBe(true)
      expect(isRestrictedNutrient('alcohol_g')).toBe(true)
      expect(isRestrictedNutrient('caffeine_mg')).toBe(true)
    })

    it('should return false for non-restricted nutrients', () => {
      expect(isRestrictedNutrient('sodium_mg')).toBe(false)
      expect(isRestrictedNutrient('calories')).toBe(false)
      expect(isRestrictedNutrient('protein_g')).toBe(false)
      expect(isRestrictedNutrient('vitamin_c_mg')).toBe(false)
      expect(isRestrictedNutrient('invalid_nutrient')).toBe(false)
    })

    it('should handle empty strings and undefined gracefully', () => {
      expect(isRestrictedNutrient('')).toBe(false)
      expect(isRestrictedNutrient('undefined')).toBe(false)
    })
  })

  describe('RestrictedNutrient type', () => {
    it('should accept valid restricted nutrient strings', () => {
      // This tests compile-time type checking
      const validNutrient: RestrictedNutrient = 'added_sugars_g'
      expect(validNutrient).toBe('added_sugars_g')
    })
  })

  describe('filterDisabledNutrients', () => {
    it('should return original object when no disabled nutrients', () => {
      const nutrients = { protein_g: 50, vitamin_c_mg: 100, selenium_mcg: 20 }
      expect(filterDisabledNutrients(nutrients, [])).toEqual(nutrients)
      expect(filterDisabledNutrients(nutrients, undefined)).toEqual(nutrients)
    })

    it('should filter out disabled nutrients', () => {
      const nutrients = { protein_g: 50, vitamin_c_mg: 100, selenium_mcg: 20 }
      const disabled = ['selenium_mcg']
      const result = filterDisabledNutrients(nutrients, disabled)
      
      expect(result).toEqual({ protein_g: 50, vitamin_c_mg: 100 })
      expect(result).not.toHaveProperty('selenium_mcg')
    })

    it('should filter out multiple disabled nutrients', () => {
      const nutrients = { protein_g: 50, vitamin_c_mg: 100, selenium_mcg: 20, molybdenum_mcg: 10 }
      const disabled = ['selenium_mcg', 'molybdenum_mcg']
      const result = filterDisabledNutrients(nutrients, disabled)
      
      expect(result).toEqual({ protein_g: 50, vitamin_c_mg: 100 })
    })

    it('should handle non-existent disabled nutrients gracefully', () => {
      const nutrients = { protein_g: 50, vitamin_c_mg: 100 }
      const disabled = ['nonexistent_nutrient']
      const result = filterDisabledNutrients(nutrients, disabled)
      
      expect(result).toEqual(nutrients)
    })
  })

  describe('filterDisabledNutrientKeys', () => {
    it('should return original array when no disabled nutrients', () => {
      const keys = ['protein_g', 'vitamin_c_mg', 'selenium_mcg']
      expect(filterDisabledNutrientKeys(keys, [])).toEqual(keys)
      expect(filterDisabledNutrientKeys(keys, undefined)).toEqual(keys)
    })

    it('should filter out disabled nutrient keys', () => {
      const keys = ['protein_g', 'vitamin_c_mg', 'selenium_mcg']
      const disabled = ['selenium_mcg']
      const result = filterDisabledNutrientKeys(keys, disabled)
      
      expect(result).toEqual(['protein_g', 'vitamin_c_mg'])
    })

    it('should filters out multiple disabled nutrient keys', () => {
      const keys = ['protein_g', 'vitamin_c_mg', 'selenium_mcg', 'molybdenum_mcg']
      const disabled = ['selenium_mcg', 'molybdenum_mcg']
      const result = filterDisabledNutrientKeys(keys, disabled)
      
      expect(result).toEqual(['protein_g', 'vitamin_c_mg'])
    })
  })
})
