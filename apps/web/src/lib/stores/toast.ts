import { writable } from "svelte/store"

export type ToastType = "success" | "error" | "info" | "warning"

export type Toast = {
  id: number
  type: ToastType
  message: string
  duration?: number
}

// Create the toast store
function createToastStore() {
  const { subscribe, update } = writable<Toast[]>([])
  let toastId = 0

  return {
    subscribe,

    // Add a new toast
    add: (type: ToastType, message: string, duration = 4000) => {
      const id = ++toastId
      const toast: Toast = { id, type, message, duration }

      update((toasts) => [toast, ...toasts])

      // Auto-remove after duration
      if (duration > 0) {
        setTimeout(() => {
          update((toasts) => toasts.filter((t) => t.id !== id))
        }, duration)
      }

      return id
    },

    // Remove a specific toast
    dismiss: (id: number) => {
      update((toasts) => toasts.filter((t) => t.id !== id))
    },

    // Clear all toasts
    clear: () => {
      update(() => [])
    },
  }
}

export const toastStore = createToastStore()

// Convenience functions for different toast types
export const toast = {
  success: (message: string, duration?: number) =>
    toastStore.add("success", message, duration),
  error: (message: string, duration?: number) =>
    toastStore.add("error", message, duration),
  info: (message: string, duration?: number) =>
    toastStore.add("info", message, duration),
  warning: (message: string, duration?: number) =>
    toastStore.add("warning", message, duration),
  dismiss: (id: number) => toastStore.dismiss(id),
  clear: () => toastStore.clear(),
}
