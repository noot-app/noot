import { describe, it, expect } from 'vitest';
import { transformNutritionSummary, transformCompleteNutrient } from './nutrition';

describe('transformNutritionSummary', () => {
  it('should transform total_ prefixed fields correctly', () => {
    const mockSummary = {
      total_calories: 2000,
      total_protein_g: 150.5,
      total_carbs_g: 250,
      total_fat_g: 75,
      total_fiber_g: 30,
      total_sodium_mg: 2300,
      total_vitamin_c_mg: 90,
      consumption_count: 3, // should be ignored
      user_id: 'abc123' // should be ignored
    };

    const result = transformNutritionSummary(mockSummary);

    expect(result).toEqual({
      calories: 2000,
      protein_g: 150.5,
      carbs_g: 250,
      fat_g: 75,
      fiber_g: 30,
      sodium_mg: 2300,
      vitamin_c_mg: 90
    });
  });

  it('should handle null and undefined values', () => {
    const mockSummary = {
      total_calories: null,
      total_protein_g: undefined,
      total_carbs_g: 0,
      total_fat_g: 75.5
    };

    const result = transformNutritionSummary(mockSummary);

    expect(result).toEqual({
      carbs_g: 0,
      fat_g: 75.5
    });
  });

  it('should return empty object for invalid input', () => {
    expect(transformNutritionSummary(null)).toEqual({});
    expect(transformNutritionSummary(undefined)).toEqual({});
    expect(transformNutritionSummary('string')).toEqual({});
  });

  it('should ignore non-total fields', () => {
    const mockSummary = {
      total_calories: 2000,
      calories: 1500, // should be ignored since it doesn't start with total_
      protein_g: 100, // should be ignored
      total_protein_g: 150,
      consumption_count: 5 // should be ignored
    };

    const result = transformNutritionSummary(mockSummary);

    expect(result).toEqual({
      calories: 2000,
      protein_g: 150
    });
  });
});

describe('transformCompleteNutrient', () => {
  it('should transform all numeric fields', () => {
    const mockNutrition = {
      calories: 300,
      protein_g: 25.5,
      total_fat_g: 15,
      vitamin_c_mg: 20,
      name: 'test food', // should be ignored
      brand: 'test brand' // should be ignored
    };

    const result = transformCompleteNutrient(mockNutrition);

    expect(result).toEqual({
      calories: 300,
      protein_g: 25.5,
      total_fat_g: 15,
      vitamin_c_mg: 20
    });
  });

  it('should handle null and undefined values', () => {
    const mockNutrition = {
      calories: null,
      protein_g: undefined,
      total_fat_g: 0,
      vitamin_c_mg: 20.5
    };

    const result = transformCompleteNutrient(mockNutrition);

    expect(result).toEqual({
      total_fat_g: 0,
      vitamin_c_mg: 20.5
    });
  });

  it('should return empty object for invalid input', () => {
    expect(transformCompleteNutrient(null)).toEqual({});
    expect(transformCompleteNutrient(undefined)).toEqual({});
    expect(transformCompleteNutrient('string')).toEqual({});
  });
});