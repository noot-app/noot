// Common API Response types
export interface ApiResponse<T> {
  data?: T
  error?: {
    error: string
    code?: string
  }
}

// Consumption types
export interface Consumption {
  id: string
  user_id: string
  transcript: string
  title?: string | null
  note?: string | null
  created_at: string
  labels?: Label[]
  items?: ConsumptionItem[]
  summary?: NutritionSummary
  is_public: boolean
}

export interface ConsumptionItem {
  item: {
    name: string
    grams: number
    user_quantity?: number | null
    user_unit?: string | null
    brand?: string | null
    note?: string | null
    labels?: Label[]
    nutrients?: NutritionData
    ingredients?: OFFIngredient[]
    url?: string | null
  }
  note?: string
}

export interface OFFIngredient {
  id?: string
  text?: string
  percent_estimate?: number | null
  percent_max?: number | null
  percent_min?: number | null
}

export interface Label {
  id: string
  name: string
  description?: string | null
  color: string
  created_at: string
  updated_at: string
}

export interface NutritionData {
  [key: string]: number | undefined
  calories?: number
  protein_g?: number
  total_fat_g?: number
  total_carbs_g?: number
  // Add other common nutrients as needed
}

export interface NutritionSummary {
  totals: NutritionData
}

// Goal types - this matches the API Goals schema which is an object, not an array
export interface Goals {
  targets: { [key: string]: number }
  upper_limits: { [key: string]: number }
  units: { [key: string]: string }
  source: "dri" | "custom"
  custom_name?: string
  disabled_nutrients?: string[]
  life_stage: LifeStage
}

export interface LifeStage {
  sex: "male" | "female" | "unspecified"
  age_bracket: string
}

// Individual goal item (for backwards compatibility if needed)
export interface Goal {
  id: string
  name: string
  target_value: number
  unit: string
  nutrient_key: string
  created_at: string
}

// Event types
export interface Event {
  id: string
  name: string
  date: string
  type: string
}

// Chart data types
export interface ChartDataPoint {
  date: string
  value: number
  label?: string
}

export interface ChartSeries {
  name: string
  data: ChartDataPoint[]
  color?: string
}

// Form validation
export interface ValidationError {
  field: string
  message: string
}

// Loading states
export interface LoadingState {
  isLoading: boolean
  error?: string
  data?: unknown
}