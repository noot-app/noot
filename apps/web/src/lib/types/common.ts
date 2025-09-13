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
  title: string
  note?: string
  created_at: string
  updated_at: string
  labels?: Label[]
  items?: ConsumptionItem[]
  nutrition?: NutritionData
}

export interface ConsumptionItem {
  id: string
  name: string
  quantity: number
  unit: string
  nutrition?: NutritionData
}

export interface Label {
  id: string
  name: string
  color?: string
}

export interface NutritionData {
  [key: string]: number | undefined
}

// Goal types
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