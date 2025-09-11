import { describe, it, expect } from 'vitest'
import {
  getNutrientValue,
  formatValue,
  isUpperLimitExceeded,
  getProgress,
  getActualProgress,
  getProgressBarClass,
  getDailyText,
  getOverageText
} from './nutrition-display'

// Mock goals data for testing
const mockGoals = {
  targets: {
    protein_g: 50,
    sodium_mg: 1500,
    calories: 2000,
    vitamin_c_mg: 90
  },
  upper_limits: {
    sodium_mg: 2000,
    trans_fat_g: 0,
    added_sugars_g: 50,
    saturated_fat_g: 20
  },
  units: {
    protein_g: 'g',
    sodium_mg: 'mg',
    calories: '',
    vitamin_c_mg: 'mg',
    trans_fat_g: 'g',
    added_sugars_g: 'g',
    saturated_fat_g: 'g'
  },
  source: 'dri' as const,
  life_stage: {
    sex: 'male' as const,
    age_bracket: '19-30 y'
  }
}

const mockNutrients = {
  protein_g: 25,
  sodium_mg: 1600,
  calories: 1800,
  vitamin_c_mg: 45,
  trans_fat_g: 0.5,
  added_sugars_g: 30,
  saturated_fat_g: 15
}

describe('getNutrientValue', () => {
  it('should return valid numeric values', () => {
    expect(getNutrientValue('protein_g', mockNutrients)).toBe(25)
    expect(getNutrientValue('sodium_mg', mockNutrients)).toBe(1600)
  })

  it('should return 0 for missing nutrients', () => {
    expect(getNutrientValue('missing_nutrient', mockNutrients)).toBe(0)
  })

  it('should return 0 for invalid values', () => {
    const invalidNutrients = {
      invalid1: NaN,
      invalid2: Infinity,
      invalid3: -Infinity,
      invalid4: null as any,
      invalid5: undefined as any,
      invalid6: 'string' as any
    }
    
    expect(getNutrientValue('invalid1', invalidNutrients)).toBe(0)
    expect(getNutrientValue('invalid2', invalidNutrients)).toBe(0)
    expect(getNutrientValue('invalid3', invalidNutrients)).toBe(0)
    expect(getNutrientValue('invalid4', invalidNutrients)).toBe(0)
    expect(getNutrientValue('invalid5', invalidNutrients)).toBe(0)
    expect(getNutrientValue('invalid6', invalidNutrients)).toBe(0)
  })

  it('should handle zero values correctly', () => {
    expect(getNutrientValue('zero_value', { zero_value: 0 })).toBe(0)
  })
})

describe('formatValue', () => {
  it('should format whole numbers correctly', () => {
    expect(formatValue(25)).toBe('25')
    expect(formatValue(100)).toBe('100')
    expect(formatValue(1500)).toBe('1500')
  })

  it('should format decimal numbers less than 1 to 1 decimal place', () => {
    expect(formatValue(0.5)).toBe('0.5')
    expect(formatValue(0.25)).toBe('0.3')
    expect(formatValue(0.12)).toBe('0.1')
  })

  it('should format decimal numbers greater than 1 to whole numbers', () => {
    expect(formatValue(25.7)).toBe('26')
    expect(formatValue(100.3)).toBe('100')
    expect(formatValue(1500.8)).toBe('1501')
  })

  it('should handle zero and invalid values', () => {
    expect(formatValue(0)).toBe('0')
    expect(formatValue(NaN)).toBe('0')
    expect(formatValue(Infinity)).toBe('0')
    expect(formatValue(-Infinity)).toBe('0')
  })
})

describe('isUpperLimitExceeded', () => {
  it('should return true when upper limit is exceeded', () => {
    expect(isUpperLimitExceeded('sodium_mg', mockGoals, { sodium_mg: 2500 })).toBe(true)
    expect(isUpperLimitExceeded('added_sugars_g', mockGoals, { added_sugars_g: 60 })).toBe(true)
  })

  it('should return false when upper limit is not exceeded', () => {
    expect(isUpperLimitExceeded('sodium_mg', mockGoals, { sodium_mg: 1800 })).toBe(false)
    expect(isUpperLimitExceeded('added_sugars_g', mockGoals, { added_sugars_g: 30 })).toBe(false)
  })

  it('should handle zero limits (like trans fat)', () => {
    expect(isUpperLimitExceeded('trans_fat_g', mockGoals, { trans_fat_g: 0 })).toBe(false)
    expect(isUpperLimitExceeded('trans_fat_g', mockGoals, { trans_fat_g: 0.1 })).toBe(true)
  })

  it('should return false when no upper limit exists', () => {
    expect(isUpperLimitExceeded('protein_g', mockGoals, { protein_g: 100 })).toBe(false)
  })

  it('should return false when goals is null', () => {
    expect(isUpperLimitExceeded('sodium_mg', null, { sodium_mg: 2500 })).toBe(false)
  })
})

describe('getProgress', () => {
  it('should calculate target-based progress correctly', () => {
    // protein: 25g / 50g target = 50%
    expect(getProgress('protein_g', mockGoals, mockNutrients, false)).toBe(50)
    
    // calories: 1800 / 2000 target = 90%
    expect(getProgress('calories', mockGoals, mockNutrients, false)).toBe(90)
  })

  it('should cap progress at 100% for display', () => {
    const highNutrients = { protein_g: 75 } // 75/50 = 150%, should cap at 100%
    expect(getProgress('protein_g', mockGoals, highNutrients, false)).toBe(100)
  })

  it('should use upper limit logic for restricted nutrients when showLimitsOnly is true', () => {
    // sodium: 1600mg / 2000mg upper limit = 80%
    expect(getProgress('sodium_mg', mockGoals, mockNutrients, true)).toBe(80)
  })

  it('should handle zero limits correctly', () => {
    expect(getProgress('trans_fat_g', mockGoals, { trans_fat_g: 0 }, true)).toBe(0)
    expect(getProgress('trans_fat_g', mockGoals, { trans_fat_g: 0.5 }, true)).toBe(100)
  })

  it('should return 0 when no goals available', () => {
    expect(getProgress('protein_g', null, mockNutrients, false)).toBe(0)
  })

  it('should return 0 for nutrients with no targets or limits', () => {
    expect(getProgress('unknown_nutrient', mockGoals, mockNutrients, false)).toBe(0)
  })
})

describe('getActualProgress', () => {
  it('should calculate actual progress without capping', () => {
    // protein: 25g / 50g target = 50%
    expect(getActualProgress('protein_g', mockGoals, mockNutrients)).toBe(50)
    
    // Test over 100%
    const highNutrients = { protein_g: 75 } // 75/50 = 150%
    expect(getActualProgress('protein_g', mockGoals, highNutrients)).toBe(150)
  })

  it('should prioritize targets over upper limits', () => {
    // sodium has both target (1500mg) and upper limit (2000mg)
    // Should use target: 1600mg / 1500mg = 106.67%
    expect(getActualProgress('sodium_mg', mockGoals, mockNutrients)).toBeCloseTo(106.67, 1)
  })

  it('should fall back to upper limits when no target exists', () => {
    // saturated_fat_g has only upper limit (20g), no target
    expect(getActualProgress('saturated_fat_g', mockGoals, mockNutrients)).toBe(75) // 15/20 = 75%
  })

  it('should handle zero limits with high percentage', () => {
    expect(getActualProgress('trans_fat_g', mockGoals, { trans_fat_g: 0.5 })).toBe(200)
    expect(getActualProgress('trans_fat_g', mockGoals, { trans_fat_g: 0 })).toBe(0)
  })

  it('should return 0 when no goals available', () => {
    expect(getActualProgress('protein_g', null, mockNutrients)).toBe(0)
  })
})

describe('getProgressBarClass', () => {
  it('should return progress-error when upper limit is exceeded in meal contribution mode', () => {
    const result = getProgressBarClass('sodium_mg', 80, mockGoals, true, false, { sodium_mg: 2500 })
    expect(result).toBe('progress-error')
  })

  it('should return progress-lighter for restricted nutrients with upper limits in meal contribution mode', () => {
    // Use saturated fat which has a non-zero limit (20g) and test with a value below that limit
    const result = getProgressBarClass('saturated_fat_g', 50, mockGoals, true, false, { saturated_fat_g: 10 })
    expect(result).toBe('progress-lighter')
  })

  it('should return progress-accent for non-restricted nutrients in meal contribution mode', () => {
    const result = getProgressBarClass('protein_g', 50, mockGoals, true, false, mockNutrients)
    expect(result).toBe('progress-accent')
  })

  it('should handle standard target-based coloring', () => {
    expect(getProgressBarClass('protein_g', 90, mockGoals, false, false, mockNutrients)).toBe('progress-success')
    expect(getProgressBarClass('protein_g', 60, mockGoals, false, false, mockNutrients)).toBe('progress-warning')
    expect(getProgressBarClass('protein_g', 30, mockGoals, false, false, mockNutrients)).toBe('progress-primary')
  })

  it('should return progress-primary when no goals available', () => {
    expect(getProgressBarClass('protein_g', 50, null, false, false, mockNutrients)).toBe('progress-primary')
  })

  it('should handle showLimitsOnly mode with error state', () => {
    expect(getProgressBarClass('trans_fat_g', 100, mockGoals, false, true, mockNutrients)).toBe('progress-error')
    expect(getProgressBarClass('trans_fat_g', 80, mockGoals, false, true, mockNutrients)).toBe('progress-lighter')
  })
})

describe('getDailyText', () => {
  it('should return target text when target exists', () => {
    expect(getDailyText('protein_g', mockGoals)).toBe('50 g')
    expect(getDailyText('calories', mockGoals)).toBe('2000 ')
  })

  it('should return limit text when only upper limit exists', () => {
    expect(getDailyText('trans_fat_g', mockGoals)).toBe('0 g limit')
    expect(getDailyText('added_sugars_g', mockGoals)).toBe('50 g limit')
  })

  it('should prioritize target over upper limit when both exist', () => {
    // sodium has both target (1500mg) and upper limit (2000mg)
    expect(getDailyText('sodium_mg', mockGoals)).toBe('1500 mg')
  })

  it('should return empty string when no goals available', () => {
    expect(getDailyText('protein_g', null)).toBe('')
  })

  it('should return empty string when neither target nor limit exists', () => {
    expect(getDailyText('unknown_nutrient', mockGoals)).toBe('')
  })
})

describe('getOverageText', () => {
  it('should return warning when upper limit is exceeded', () => {
    const result = getOverageText('sodium_mg', 2200, mockGoals)
    expect(result).toBe('⚠️ 10% over upper limit') // 2200/2000 = 110%, so 10% over
  })

  it('should return simple warning for zero limits', () => {
    expect(getOverageText('trans_fat_g', 0.5, mockGoals)).toBe('⚠️')
  })

  it('should show target overage when target exceeded but not upper limit', () => {
    const result = getOverageText('protein_g', 60, mockGoals)
    expect(result).toBe('+20% over target') // 60/50 = 120%, so 20% over
  })

  it('should distinguish between upper limit and regular limit in text', () => {
    // For nutrients with both target and upper limit
    const resultWithTarget = getOverageText('sodium_mg', 2200, mockGoals)
    expect(resultWithTarget).toContain('upper limit')
    
    // For nutrients with only upper limit
    const resultWithoutTarget = getOverageText('added_sugars_g', 60, mockGoals)
    expect(resultWithoutTarget).toContain('⚠️')
    expect(resultWithoutTarget).toContain('limit')
  })

  it('should return empty string when no overage', () => {
    expect(getOverageText('protein_g', 30, mockGoals)).toBe('')
    expect(getOverageText('sodium_mg', 1200, mockGoals)).toBe('')
  })

  it('should return empty string when no goals available', () => {
    expect(getOverageText('protein_g', 100, null)).toBe('')
  })

  it('should handle infinite values gracefully', () => {
    const goalsWithZero = { ...mockGoals, upper_limits: { ...mockGoals.upper_limits, test_nutrient: 0 } }
    expect(getOverageText('test_nutrient', 5, goalsWithZero)).toBe('⚠️')
  })
})
