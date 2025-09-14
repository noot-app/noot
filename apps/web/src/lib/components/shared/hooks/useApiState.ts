import { writable } from 'svelte/store'

export interface ApiState<T> {
  data: T | null
  loading: boolean
  error: string
}

export function createApiState<T>(initialData: T | null = null) {
  const { subscribe, set, update } = writable<ApiState<T>>({
    data: initialData,
    loading: false,
    error: ''
  })

  return {
    subscribe,
    setLoading: (loading: boolean) => update(state => ({ ...state, loading })),
    setError: (error: string) => update(state => ({ ...state, error, loading: false })),
    setData: (data: T) => update(state => ({ ...state, data, error: '', loading: false })),
    reset: () => set({ data: initialData, loading: false, error: '' }),
    
    // Helper for common async operations
    async execute<R>(promise: Promise<R>, onSuccess?: (result: R) => T): Promise<R | null> {
      try {
        update(state => ({ ...state, loading: true, error: '' }))
        const result = await promise
        
        if (onSuccess) {
          const data = onSuccess(result)
          update(state => ({ ...state, data, loading: false, error: '' }))
        } else {
          update(state => ({ ...state, loading: false, error: '' }))
        }
        
        return result
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'An error occurred'
        update(state => ({ ...state, loading: false, error: errorMessage }))
        return null
      }
    }
  }
}

// Hook for component usage
export function useApiState<T>(initialData: T | null = null) {
  return createApiState<T>(initialData)
}