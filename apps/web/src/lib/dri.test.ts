import { describe, it, expect } from "vitest"
import { getDefaultDRIGoals, hasGoals } from "$lib/dri"

describe("DRI fallback goals", () => {
  it("should return default male DRI goals", () => {
    const goals = getDefaultDRIGoals()
    
    expect(goals.source).toBe("dri")
    expect(goals.life_stage.sex).toBe("male")
    expect(goals.life_stage.age_bracket).toBe("19-30 y")
    expect(goals.targets.calories).toBe(2000)
    expect(goals.targets.protein_g).toBe(56)
    expect(goals.upper_limits.added_sugars_g).toBe(50)
    expect(goals.units.calories).toBe("kcal")
  })

  it("should return default female DRI goals when specified", () => {
    const goals = getDefaultDRIGoals("female")
    
    expect(goals.source).toBe("dri")
    expect(goals.life_stage.sex).toBe("female")
    expect(goals.life_stage.age_bracket).toBe("19-30 y")
    expect(goals.targets.calories).toBe(2000)
    expect(goals.targets.protein_g).toBe(46) // Lower than males
    expect(goals.targets.iron_mg).toBe(18) // Higher than males
    expect(goals.upper_limits.alcohol_g).toBe(14) // Lower than males
  })

  it("should include new omega-3 and creatine nutrients", () => {
    const maleGoals = getDefaultDRIGoals("male")
    const femaleGoals = getDefaultDRIGoals("female")

    // Check omega-3 EPA/DHA targets
    expect(maleGoals.targets.omega3_epa_g).toBe(0.25)
    expect(maleGoals.targets.omega3_dha_g).toBe(0.25)
    expect(femaleGoals.targets.omega3_epa_g).toBe(0.25)
    expect(femaleGoals.targets.omega3_dha_g).toBe(0.25)

    // Check creatine targets
    expect(maleGoals.targets.creatine_mg).toBe(3000)
    expect(femaleGoals.targets.creatine_mg).toBe(2500)

    // Check units
    expect(maleGoals.units.omega3_epa_g).toBe("g")
    expect(maleGoals.units.omega3_dha_g).toBe("g")
    expect(maleGoals.units.creatine_mg).toBe("mg")
  })

  it("should have proper nutrient mappings", () => {
    const goals = getDefaultDRIGoals()
    
    // Test key nutrients are present
    expect(goals.targets).toHaveProperty("vitamin_c_mg")
    expect(goals.targets).toHaveProperty("calcium_mg")
    expect(goals.targets).toHaveProperty("dietary_fiber_g")
    expect(goals.upper_limits).toHaveProperty("sodium_mg")
    
    // Test units are provided
    expect(goals.units.vitamin_c_mg).toBe("mg")
    expect(goals.units.calcium_mg).toBe("mg")
    expect(goals.units.vitamin_a_mcg).toBe("mcg")
  })

  it("should validate goals availability", () => {
    const goals = getDefaultDRIGoals()
    expect(hasGoals(goals)).toBe(true)
    expect(hasGoals(null)).toBe(false)
  })

  it("should include all essential nutrients", () => {
    const goals = getDefaultDRIGoals()
    
    // Essential macronutrients
    expect(goals.targets.calories).toBeGreaterThan(0)
    expect(goals.targets.protein_g).toBeGreaterThan(0)
    expect(goals.targets.total_carbs_g).toBeGreaterThan(0)
    expect(goals.targets.total_fat_g).toBeGreaterThan(0)
    expect(goals.targets.monounsaturated_fat_g).toBeGreaterThan(0)
    expect(goals.targets.polyunsaturated_fat_g).toBeGreaterThan(0)
    expect(goals.targets.omega3_ala_g).toBeGreaterThan(0)
    expect(goals.targets.omega6_g).toBeGreaterThan(0)
    
    // Key vitamins
    expect(goals.targets.vitamin_c_mg).toBeGreaterThan(0)
    expect(goals.targets.vitamin_d_mcg).toBeGreaterThan(0)
    expect(goals.targets.folate_mcg).toBeGreaterThan(0)
    
    // Key minerals  
    expect(goals.targets.calcium_mg).toBeGreaterThan(0)
    expect(goals.targets.iron_mg).toBeGreaterThan(0)
    expect(goals.targets.potassium_mg).toBeGreaterThan(0)
  })
})
