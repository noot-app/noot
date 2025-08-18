/**
 * Theme toggle button matching the original noot web frontend design
 */
import React from 'react';
import { Pressable, StyleSheet } from 'react-native';
import { BlurView } from 'expo-blur';
import { Ionicons } from '@expo/vector-icons';
import { Colors } from '@/constants/Colors';
import { useColorScheme } from '@/hooks/useColorScheme';

interface ThemeToggleProps {
  style?: any;
  onToggle?: (theme: 'light' | 'dark') => void;
}

export function ThemeToggle({ style, onToggle }: ThemeToggleProps) {
  const colorScheme = useColorScheme();
  const colors = Colors[colorScheme ?? 'light'];

  const handleToggle = () => {
    const newTheme = colorScheme === 'dark' ? 'light' : 'dark';
    onToggle?.(newTheme);
  };

  return (
    <Pressable 
      style={[styles.container, style]} 
      onPress={handleToggle}
    >
      <BlurView
        intensity={20}
        tint={colorScheme || 'default'}
        style={[
          styles.blurContainer,
          { borderColor: colors.glassBorder }
        ]}
      >
        <Ionicons
          name={colorScheme === 'dark' ? 'sunny' : 'moon'}
          size={20}
          color={colors.textPrimary}
        />
      </BlurView>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: 32,
    right: 32,
    zIndex: 1000,
  },
  blurContainer: {
    width: 48,
    height: 48,
    borderRadius: 24,
    borderWidth: 1,
    justifyContent: 'center',
    alignItems: 'center',
    overflow: 'hidden',
  },
});