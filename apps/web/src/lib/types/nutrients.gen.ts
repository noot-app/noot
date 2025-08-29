// Nutrient type definitions
// THIS FILE IS GENERATED - DO NOT EDIT MANUALLY
// Generated from nutrients.yaml

export interface CompleteNutrient {
  calories: number; // Calories (kcal)
  protein_g: number; // Protein (g)
  total_fat_g: number; // Total Fat (g)
  saturated_fat_g: number; // Saturated Fat (g)
  trans_fat_g: number; // Trans Fat (g)
  monounsaturated_fat_g: number; // Monounsaturated Fat (g)
  polyunsaturated_fat_g: number; // Polyunsaturated Fat (g)
  cholesterol_mg: number; // Cholesterol (mg)
  sodium_mg: number; // Sodium (mg)
  total_carbs_g: number; // Total Carbohydrates (g)
  dietary_fiber_g: number; // Dietary Fiber (g)
  total_sugars_g: number; // Total Sugars (g)
  added_sugars_g: number; // Added Sugars (g)
  vitamin_a_mcg: number; // Vitamin A (mcg)
  vitamin_c_mg: number; // Vitamin C (mg)
  vitamin_d_mcg: number; // Vitamin D (mcg)
  vitamin_e_mg: number; // Vitamin E (mg)
  vitamin_k_mcg: number; // Vitamin K (mcg)
  thiamine_mg: number; // Thiamine (B1) (mg)
  riboflavin_mg: number; // Riboflavin (B2) (mg)
  niacin_mg: number; // Niacin (B3) (mg)
  vitamin_b6_mg: number; // Vitamin B6 (mg)
  folate_mcg: number; // Folate (mcg)
  vitamin_b12_mcg: number; // Vitamin B12 (mcg)
  biotin_mcg: number; // Biotin (mcg)
  pantothenic_acid_mg: number; // Pantothenic Acid (mg)
  choline_mg: number; // Choline (mg)
  calcium_mg: number; // Calcium (mg)
  iron_mg: number; // Iron (mg)
  magnesium_mg: number; // Magnesium (mg)
  phosphorus_mg: number; // Phosphorus (mg)
  potassium_mg: number; // Potassium (mg)
  zinc_mg: number; // Zinc (mg)
  copper_mg: number; // Copper (mg)
  manganese_mg: number; // Manganese (mg)
  selenium_mcg: number; // Selenium (mcg)
  iodine_mcg: number; // Iodine (mcg)
  molybdenum_mcg: number; // Molybdenum (mcg)
  chromium_mcg: number; // Chromium (mcg)
  fluoride_mg: number; // Fluoride (mg)
  chloride_mg: number; // Chloride (mg)
  omega3_ala_g: number; // Omega-3 ALA (g)
  omega3_epa_g: number; // Omega-3 EPA (g)
  omega3_dha_g: number; // Omega-3 DHA (g)
  omega6_g: number; // Omega-6 (g)
  alcohol_g: number; // Alcohol (g)
  caffeine_mg: number; // Caffeine (mg)
  creatine_mg: number; // Creatine (mg)
}

export interface ItemNutrients {
  // Original serving data
  original_serving_grams?: number | null; // Serving Size
  original_calories?: number | null; // Calories
  original_protein_g?: number | null; // Protein
  original_total_fat_g?: number | null; // Total Fat
  original_saturated_fat_g?: number | null; // Saturated Fat
  original_trans_fat_g?: number | null; // Trans Fat
  original_monounsaturated_fat_g?: number | null; // Monounsaturated Fat
  original_polyunsaturated_fat_g?: number | null; // Polyunsaturated Fat
  original_cholesterol_mg?: number | null; // Cholesterol
  original_sodium_mg?: number | null; // Sodium
  original_total_carbs_g?: number | null; // Total Carbohydrates
  original_dietary_fiber_g?: number | null; // Dietary Fiber
  original_total_sugars_g?: number | null; // Total Sugars
  original_added_sugars_g?: number | null; // Added Sugars
  original_vitamin_a_mcg?: number | null; // Vitamin A
  original_vitamin_c_mg?: number | null; // Vitamin C
  original_vitamin_d_mcg?: number | null; // Vitamin D
  original_vitamin_e_mg?: number | null; // Vitamin E
  original_vitamin_k_mcg?: number | null; // Vitamin K
  original_thiamine_mg?: number | null; // Thiamine (B1)
  original_riboflavin_mg?: number | null; // Riboflavin (B2)
  original_niacin_mg?: number | null; // Niacin (B3)
  original_vitamin_b6_mg?: number | null; // Vitamin B6
  original_folate_mcg?: number | null; // Folate
  original_vitamin_b12_mcg?: number | null; // Vitamin B12
  original_biotin_mcg?: number | null; // Biotin
  original_pantothenic_acid_mg?: number | null; // Pantothenic Acid
  original_choline_mg?: number | null; // Choline
  original_calcium_mg?: number | null; // Calcium
  original_iron_mg?: number | null; // Iron
  original_magnesium_mg?: number | null; // Magnesium
  original_phosphorus_mg?: number | null; // Phosphorus
  original_potassium_mg?: number | null; // Potassium
  original_zinc_mg?: number | null; // Zinc
  original_copper_mg?: number | null; // Copper
  original_manganese_mg?: number | null; // Manganese
  original_selenium_mcg?: number | null; // Selenium
  original_iodine_mcg?: number | null; // Iodine
  original_molybdenum_mcg?: number | null; // Molybdenum
  original_chromium_mcg?: number | null; // Chromium
  original_fluoride_mg?: number | null; // Fluoride
  original_chloride_mg?: number | null; // Chloride
  original_omega3_ala_g?: number | null; // Omega-3 ALA
  original_omega3_epa_g?: number | null; // Omega-3 EPA
  original_omega3_dha_g?: number | null; // Omega-3 DHA
  original_omega6_g?: number | null; // Omega-6
  original_alcohol_g?: number | null; // Alcohol
  original_caffeine_mg?: number | null; // Caffeine
  original_creatine_mg?: number | null; // Creatine

  // Per 100g normalized data
  calories_per_100g: number; // Calories per 100g
  protein_g_per_100g: number; // Protein per 100g
  total_fat_g_per_100g: number; // Total Fat per 100g
  saturated_fat_g_per_100g: number; // Saturated Fat per 100g
  trans_fat_g_per_100g: number; // Trans Fat per 100g
  monounsaturated_fat_g_per_100g: number; // Monounsaturated Fat per 100g
  polyunsaturated_fat_g_per_100g: number; // Polyunsaturated Fat per 100g
  cholesterol_mg_per_100g: number; // Cholesterol per 100g
  sodium_mg_per_100g: number; // Sodium per 100g
  total_carbs_g_per_100g: number; // Total Carbohydrates per 100g
  dietary_fiber_g_per_100g: number; // Dietary Fiber per 100g
  total_sugars_g_per_100g: number; // Total Sugars per 100g
  added_sugars_g_per_100g: number; // Added Sugars per 100g
  vitamin_a_mcg_per_100g: number; // Vitamin A per 100g
  vitamin_c_mg_per_100g: number; // Vitamin C per 100g
  vitamin_d_mcg_per_100g: number; // Vitamin D per 100g
  vitamin_e_mg_per_100g: number; // Vitamin E per 100g
  vitamin_k_mcg_per_100g: number; // Vitamin K per 100g
  thiamine_mg_per_100g: number; // Thiamine (B1) per 100g
  riboflavin_mg_per_100g: number; // Riboflavin (B2) per 100g
  niacin_mg_per_100g: number; // Niacin (B3) per 100g
  vitamin_b6_mg_per_100g: number; // Vitamin B6 per 100g
  folate_mcg_per_100g: number; // Folate per 100g
  vitamin_b12_mcg_per_100g: number; // Vitamin B12 per 100g
  biotin_mcg_per_100g: number; // Biotin per 100g
  pantothenic_acid_mg_per_100g: number; // Pantothenic Acid per 100g
  choline_mg_per_100g: number; // Choline per 100g
  calcium_mg_per_100g: number; // Calcium per 100g
  iron_mg_per_100g: number; // Iron per 100g
  magnesium_mg_per_100g: number; // Magnesium per 100g
  phosphorus_mg_per_100g: number; // Phosphorus per 100g
  potassium_mg_per_100g: number; // Potassium per 100g
  zinc_mg_per_100g: number; // Zinc per 100g
  copper_mg_per_100g: number; // Copper per 100g
  manganese_mg_per_100g: number; // Manganese per 100g
  selenium_mcg_per_100g: number; // Selenium per 100g
  iodine_mcg_per_100g: number; // Iodine per 100g
  molybdenum_mcg_per_100g: number; // Molybdenum per 100g
  chromium_mcg_per_100g: number; // Chromium per 100g
  fluoride_mg_per_100g: number; // Fluoride per 100g
  chloride_mg_per_100g: number; // Chloride per 100g
  omega3_ala_g_per_100g: number; // Omega-3 ALA per 100g
  omega3_epa_g_per_100g: number; // Omega-3 EPA per 100g
  omega3_dha_g_per_100g: number; // Omega-3 DHA per 100g
  omega6_g_per_100g: number; // Omega-6 per 100g
  alcohol_g_per_100g: number; // Alcohol per 100g
  caffeine_mg_per_100g: number; // Caffeine per 100g
  creatine_mg_per_100g: number; // Creatine per 100g
}
