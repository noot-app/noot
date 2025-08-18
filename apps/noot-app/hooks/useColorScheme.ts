/**
 * Enhanced color scheme hook with theme persistence matching the original noot design
 */
import { useState, useEffect } from 'react';
import { useColorScheme as useSystemColorScheme, Platform } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';

const THEME_KEY = 'noot-theme';

type ColorScheme = 'light' | 'dark' | null;

let themeListeners: Set<(theme: ColorScheme) => void> = new Set();
let currentTheme: ColorScheme = null;

export function useColorScheme(): NonNullable<ColorScheme> {
  const systemColorScheme = useSystemColorScheme();
  const [theme, setTheme] = useState<ColorScheme>(currentTheme || systemColorScheme || 'dark');

  useEffect(() => {
    loadTheme();
    
    const listener = (newTheme: ColorScheme) => setTheme(newTheme);
    themeListeners.add(listener);
    
    return () => {
      themeListeners.delete(listener);
    };
  }, []);

  const loadTheme = async () => {
    try {
      const savedTheme = await AsyncStorage.getItem(THEME_KEY);
      const themeToUse = (savedTheme as ColorScheme) || systemColorScheme || 'dark';
      currentTheme = themeToUse;
      setTheme(themeToUse);
      
      // Notify all listeners
      themeListeners.forEach(listener => listener(themeToUse));
    } catch (error) {
      console.warn('Failed to load theme:', error);
      const fallback = systemColorScheme || 'dark';
      currentTheme = fallback;
      setTheme(fallback);
    }
  };

  return theme || 'dark';
}

export const toggleTheme = async () => {
  try {
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    currentTheme = newTheme;
    
    await AsyncStorage.setItem(THEME_KEY, newTheme);
    
    // Notify all listeners
    themeListeners.forEach(listener => listener(newTheme));
  } catch (error) {
    console.warn('Failed to save theme:', error);
  }
};
