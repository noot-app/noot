import { writable } from 'svelte/store';
import { getStorageItem, setStorageItem } from '$lib/utils/secure-storage';

export type Units = 'grams' | 'ounces';

function createUnitsStore() {
  // Initialize with grams as default
  const { subscribe, set } = writable<Units>('grams');

  return {
    subscribe,
    set: (value: Units) => {
      // Validate the value before storing
      if (value !== 'grams' && value !== 'ounces') {
        console.warn('Invalid units value:', value);
        return;
      }
      
      if (typeof window !== 'undefined') {
        setStorageItem('noot-units', value);
      }
      set(value);
    },
    // Initialize from localStorage on client side
    init: () => {
      if (typeof window !== 'undefined') {
        const stored = getStorageItem('noot-units') as Units | null;
        if (stored && (stored === 'grams' || stored === 'ounces')) {
          set(stored);
        }
      }
    }
  };
}

export const units = createUnitsStore();

// Conversion functions
export function convertWeight(value: number, fromUnit: Units, toUnit: Units): number {
  if (fromUnit === toUnit) return value;
  
  if (fromUnit === 'grams' && toUnit === 'ounces') {
    return value * 0.035274; // grams to ounces
  }
  
  if (fromUnit === 'ounces' && toUnit === 'grams') {
    return value * 28.3495; // ounces to grams
  }
  
  return value;
}

export function formatWeight(value: number, unit: Units): string {
  const rounded = Math.round(value * 100) / 100;
  return `${rounded} ${unit === 'grams' ? 'g' : 'oz'}`;
}
