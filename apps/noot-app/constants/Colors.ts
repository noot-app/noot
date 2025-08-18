/**
 * Apple-inspired design colors matching the original noot web frontend.
 * Uses sophisticated gradients, glass morphism, and proper light/dark theming.
 */

export const Colors = {
  light: {
    // Primary colors
    bgPrimary: '#ffffff',
    bgSecondary: '#f8fafc',
    textPrimary: '#1e293b',
    textSecondary: '#475569',
    textTertiary: '#64748b',
    
    // Glass morphism
    glassBg: 'rgba(255, 255, 255, 0.8)',
    glassBorder: 'rgba(30, 41, 59, 0.1)',
    
    // Gradients and effects
    bgGradient: ['#f8fafc', '#e2e8f0', '#cbd5e1', '#94a3b8'],
    buttonGradient: ['#3b82f6', '#8b5cf6'],
    buttonHoverGradient: ['#4f96ff', '#9f70fd'],
    shadowColor: 'rgba(0, 0, 0, 0.1)',
    
    // UI elements
    borderColor: 'rgba(30, 41, 59, 0.15)',
    
    // Tab navigation
    tint: '#3b82f6',
    icon: '#475569',
    tabIconDefault: '#64748b',
    tabIconSelected: '#3b82f6',
  },
  dark: {
    // Primary colors
    bgPrimary: '#0f172a',
    bgSecondary: '#1e293b',
    textPrimary: '#ffffff',
    textSecondary: '#cbd5e1',
    textTertiary: '#93c5fd',
    
    // Glass morphism
    glassBg: 'rgba(255, 255, 255, 0.08)',
    glassBorder: 'rgba(255, 255, 255, 0.15)',
    
    // Gradients and effects
    bgGradient: ['#1a1a2e', '#16213e', '#0f3460', '#533483'],
    buttonGradient: ['#3b82f6', '#8b5cf6'],
    buttonHoverGradient: ['#4f96ff', '#9f70fd'],
    shadowColor: 'rgba(0, 0, 0, 0.3)',
    
    // UI elements
    borderColor: 'rgba(255, 255, 255, 0.15)',
    
    // Tab navigation
    tint: '#ffffff',
    icon: '#cbd5e1',
    tabIconDefault: '#93c5fd',
    tabIconSelected: '#ffffff',
  },
  
  // Common colors (theme-independent)
  common: {
    // Status colors
    red: '#ef4444',
    green: '#22c55e',
    yellow: '#eab308',
    amber: '#f59e0b',
    blue: '#3b82f6',
    
    // Recording state
    recordingRed: '#ef4444',
    recordingRedDark: '#dc2626',
    
    // White for text on colored backgrounds
    white: '#ffffff',
    
    // Percentage indicator colors
    percentageLow: '#ef4444',    // red < 25%
    percentageMedium: '#f59e0b', // amber 25-50%
    percentageGood: '#eab308',   // yellow 50-75%
    percentageHigh: '#22c55e',   // green 75-100%
    percentageOver: '#3b82f6',   // blue > 100%
  }
};
