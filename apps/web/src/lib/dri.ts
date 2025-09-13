/**
 * Default DRI (Dietary Reference Intake) values for nutrition goals
 * Based on National Academies / Institute of Medicine DRI summary tables
 * 
 * This provides fallback nutrition targets for public consumption views
 * when user authentication is not available.
 */

export interface DRIGoals {
  targets: Record<string, number>
  upper_limits: Record<string, number>
  units: Record<string, string>
  source: "dri" | "custom"
  custom_name?: string
  life_stage: {
    sex: "male" | "female" | "unspecified"
    age_bracket: string
  }
  disabled_nutrients?: string[]
}

// Default adult profile (covers most users for public view)
const DEFAULT_ADULT_MALE_19_30: DRIGoals = {
  targets: {
    // Macronutrients
    calories: 2000,
    total_carbs_g: 130,
    protein_g: 56,
    dietary_fiber_g: 38,
    total_fat_g: 65, // ~30% of 2000 kcal diet (FDA DV)
    monounsaturated_fat_g: 20, // Estimated based on dietary guidelines
    polyunsaturated_fat_g: 15, // Estimated based on dietary guidelines
    omega3_ala_g: 1.6, // DRI for adult males
    omega3_epa_g: 0.25, // Combined EPA+DHA recommendation, typical split (converted to grams)
    omega3_dha_g: 0.25, // Combined EPA+DHA recommendation, typical split (converted to grams)
    omega6_g: 17, // DRI for adult males
    creatine_mg: 3000, // Typical daily synthesis/turnover for adult males (converted to mg)
    
    // Vitamins
    vitamin_a_mcg: 900,
    vitamin_c_mg: 90,
    vitamin_d_mcg: 15,
    vitamin_e_mg: 15,
    vitamin_k_mcg: 120,
    thiamine_mg: 1.2,
    riboflavin_mg: 1.3,
    niacin_mg: 16,
    vitamin_b6_mg: 1.3,
    folate_mcg: 400,
    vitamin_b12_mcg: 2.4,
    pantothenic_acid_mg: 5,
    biotin_mcg: 30,
    choline_mg: 550,
    
    // Minerals
    calcium_mg: 1000,
    iron_mg: 8,
    magnesium_mg: 400,
    phosphorus_mg: 700,
    potassium_mg: 3400,
    sodium_mg: 1500,
    zinc_mg: 11,
    copper_mg: 0.9,
    manganese_mg: 2.3,
    selenium_mcg: 55,
    iodine_mcg: 150,
    molybdenum_mcg: 45,
    chromium_mcg: 35,
    fluoride_mg: 4,
    chloride_mg: 2300,
  },
  
  upper_limits: {
    // Nutrients to limit
    added_sugars_g: 50, // FDA DV (10% of 2000 kcal diet)
    saturated_fat_g: 20, // FDA DV
    trans_fat_g: 0, // Minimize
    cholesterol_mg: 300, // FDA DV
    sodium_mg: 2300, // FDA DV
    alcohol_g: 28, // Moderate consumption guideline (2 drinks/day for men)
    caffeine_mg: 400, // FDA guidance
  },
  
  units: {
    calories: "kcal",
    total_carbs_g: "g",
    protein_g: "g",
    dietary_fiber_g: "g",
    total_fat_g: "g",
    monounsaturated_fat_g: "g",
    polyunsaturated_fat_g: "g",
    omega3_ala_g: "g",
    omega3_epa_g: "g",
    omega3_dha_g: "g",
    omega6_g: "g",
    creatine_mg: "mg",
    saturated_fat_g: "g",
    trans_fat_g: "g",
    cholesterol_mg: "mg",
    added_sugars_g: "g",
    sodium_mg: "mg",
    potassium_mg: "mg",
    calcium_mg: "mg",
    iron_mg: "mg",
    magnesium_mg: "mg",
    phosphorus_mg: "mg",
    zinc_mg: "mg",
    copper_mg: "mg",
    manganese_mg: "mg",
    selenium_mcg: "mcg",
    iodine_mcg: "mcg",
    molybdenum_mcg: "mcg",
    chromium_mcg: "mcg",
    fluoride_mg: "mg",
    chloride_mg: "mg",
    vitamin_a_mcg: "mcg",
    vitamin_c_mg: "mg",
    vitamin_d_mcg: "mcg",
    vitamin_e_mg: "mg",
    vitamin_k_mcg: "mcg",
    thiamine_mg: "mg",
    riboflavin_mg: "mg",
    niacin_mg: "mg",
    vitamin_b6_mg: "mg",
    folate_mcg: "mcg",
    vitamin_b12_mcg: "mcg",
    pantothenic_acid_mg: "mg",
    biotin_mcg: "mcg",
    choline_mg: "mg",
    alcohol_g: "g",
    caffeine_mg: "mg",
  },
  
  source: "dri" as const,
  life_stage: {
    sex: "male" as const,
    age_bracket: "19-30 y"
  }
}

const DEFAULT_ADULT_FEMALE_19_30: DRIGoals = {
  targets: {
    // Macronutrients (adjusted for females)
    calories: 2000, // Estimated average for adult females 19-30
    total_carbs_g: 130,
    protein_g: 46, // Lower than males
    dietary_fiber_g: 25, // Lower than males
    total_fat_g: 65, // ~30% of 2000 kcal diet (FDA DV)
    monounsaturated_fat_g: 20, // Estimated based on dietary guidelines
    polyunsaturated_fat_g: 15, // Estimated based on dietary guidelines
    omega3_ala_g: 1.1, // DRI for adult females (lower than males)
    omega3_epa_g: 0.25, // Combined EPA+DHA recommendation, typical split (converted to grams)
    omega3_dha_g: 0.25, // Combined EPA+DHA recommendation, typical split (converted to grams)
    omega6_g: 12, // DRI for adult females (lower than males)
    creatine_mg: 2500, // Typical daily synthesis/turnover for adult females (converted to mg)
    
    // Vitamins (most same as males, some different)
    vitamin_a_mcg: 700, // Lower than males
    vitamin_c_mg: 75, // Lower than males
    vitamin_d_mcg: 15,
    vitamin_e_mg: 15,
    vitamin_k_mcg: 90, // Lower than males
    thiamine_mg: 1.1, // Lower than males
    riboflavin_mg: 1.1, // Lower than males
    niacin_mg: 14, // Lower than males
    vitamin_b6_mg: 1.3,
    folate_mcg: 400,
    vitamin_b12_mcg: 2.4,
    pantothenic_acid_mg: 5,
    biotin_mcg: 30,
    choline_mg: 425, // Lower than males
    
    // Minerals (several different from males)
    calcium_mg: 1000,
    iron_mg: 18, // Much higher than males (menstruation)
    magnesium_mg: 310, // Lower than males
    phosphorus_mg: 700,
    potassium_mg: 2600, // Lower than males
    sodium_mg: 1500,
    zinc_mg: 8, // Lower than males
    copper_mg: 0.9,
    manganese_mg: 1.8, // Lower than males
    selenium_mcg: 55,
    iodine_mcg: 150,
    molybdenum_mcg: 45,
    chromium_mcg: 25, // Lower than males
    fluoride_mg: 3, // Lower than males
    chloride_mg: 2300,
  },
  
  upper_limits: {
    // Same limits as males for most nutrients
    added_sugars_g: 50,
    saturated_fat_g: 20,
    trans_fat_g: 0,
    cholesterol_mg: 300,
    sodium_mg: 2300,
    alcohol_g: 14, // Lower than males (1 drink/day for women)
    caffeine_mg: 400,
  },
  
  units: {
    // Same units as males
    ...DEFAULT_ADULT_MALE_19_30.units
  },
  
  source: "dri" as const,
  life_stage: {
    sex: "female" as const,  
    age_bracket: "19-30 y"
  }
}

/**
 * Get default DRI goals for public/unauthenticated users
 * Uses a generic adult profile since we don't have user demographic data
 * 
 * @param preferredSex Optional sex preference, defaults to male
 * @returns DRI goals object compatible with Goals component
 */
export function getDefaultDRIGoals(preferredSex?: 'male' | 'female'): DRIGoals {
  // Default to male profile if not specified
  // This is arbitrary but matches backend default behavior
  if (preferredSex === 'female') {
    return { ...DEFAULT_ADULT_FEMALE_19_30 }
  }
  
  return { ...DEFAULT_ADULT_MALE_19_30 }
}

/**
 * Check if goals are available (for backwards compatibility)
 */
export function hasGoals(goals: DRIGoals | null): goals is DRIGoals {
  return goals !== null && goals.targets && Object.keys(goals.targets).length > 0
}
