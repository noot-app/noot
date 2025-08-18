/**
 * Themed components that match the original noot web frontend design
 * with glass morphism, proper shadows, and comprehensive theming
 */
import React from 'react';
import { View, Text, Pressable, StyleSheet, ViewStyle, TextStyle } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { BlurView } from 'expo-blur';
import { Colors } from '@/constants/Colors';
import { useColorScheme } from '@/hooks/useColorScheme';

interface ThemedViewProps {
  style?: ViewStyle;
  children: React.ReactNode;
  lightColor?: string;
  darkColor?: string;
  glass?: boolean;
}

interface ThemedTextProps {
  style?: TextStyle;
  children: React.ReactNode;
  type?: 'primary' | 'secondary' | 'tertiary';
  lightColor?: string;
  darkColor?: string;
}

interface GlassCardProps {
  style?: ViewStyle;
  children: React.ReactNode;
  blur?: 'light' | 'dark';
}

interface GradientButtonProps {
  style?: ViewStyle;
  onPress?: () => void;
  onPressIn?: () => void;
  onPressOut?: () => void;
  children: React.ReactNode;
  disabled?: boolean;
  recording?: boolean;
}

export function ThemedView({ style, children, lightColor, darkColor, glass, ...props }: ThemedViewProps) {
  const colorScheme = useColorScheme();
  const colors = Colors[colorScheme ?? 'light'];
  
  const backgroundColor = lightColor || darkColor 
    ? (colorScheme === 'light' ? lightColor : darkColor)
    : glass 
      ? colors.glassBg 
      : colors.bgPrimary;

  return (
    <View 
      style={[
        { backgroundColor }, 
        glass && styles.glassEffect,
        style
      ]} 
      {...props}
    >
      {children}
    </View>
  );
}

export function ThemedText({ 
  style, 
  children, 
  type = 'primary', 
  lightColor, 
  darkColor, 
  ...props 
}: ThemedTextProps) {
  const colorScheme = useColorScheme();
  const colors = Colors[colorScheme ?? 'light'];
  
  const getColor = () => {
    if (lightColor || darkColor) {
      return colorScheme === 'light' ? lightColor : darkColor;
    }
    
    switch (type) {
      case 'secondary':
        return colors.textSecondary;
      case 'tertiary':
        return colors.textTertiary;
      default:
        return colors.textPrimary;
    }
  };

  return (
    <Text 
      style={[
        { color: getColor() },
        styles.defaultText,
        style
      ]} 
      {...props}
    >
      {children}
    </Text>
  );
}

export function GlassCard({ style, children, blur }: GlassCardProps) {
  const colorScheme = useColorScheme();
  const colors = Colors[colorScheme ?? 'light'];
  
  return (
    <BlurView
      intensity={20}
      tint={blur || colorScheme || 'default'}
      style={[
        styles.glassCard,
        { 
          borderColor: colors.glassBorder,
          backgroundColor: colors.glassBg 
        },
        style
      ]}
    >
      {children}
    </BlurView>
  );
}

export function GradientButton({ 
  style, 
  onPress, 
  onPressIn,
  onPressOut,
  children, 
  disabled, 
  recording,
  ...props 
}: GradientButtonProps) {
  const colorScheme = useColorScheme();
  const colors = Colors[colorScheme ?? 'light'];
  const commonColors = Colors.common;
  
  const gradientColors = recording 
    ? [commonColors.recordingRed, commonColors.recordingRedDark]
    : colors.buttonGradient;

  return (
    <Pressable 
      onPress={onPress}
      onPressIn={onPressIn}
      onPressOut={onPressOut}
      disabled={disabled}
      style={({ pressed }) => [
        styles.gradientButton,
        pressed && styles.gradientButtonPressed,
        recording && styles.gradientButtonRecording,
        style
      ]}
      {...props}
    >
      <LinearGradient
        colors={gradientColors}
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 1 }}
        style={styles.gradientBackground}
      >
        {children}
      </LinearGradient>
    </Pressable>
  );
}

export function ThemedBackground({ children }: { children: React.ReactNode }) {
  const colorScheme = useColorScheme();
  const colors = Colors[colorScheme ?? 'light'];
  
  return (
    <LinearGradient
      colors={colors.bgGradient}
      start={{ x: 0, y: 0 }}
      end={{ x: 1, y: 1 }}
      style={styles.backgroundGradient}
    >
      {children}
    </LinearGradient>
  );
}

const styles = StyleSheet.create({
  defaultText: {
    fontFamily: '-apple-system', // SF Pro equivalent
    fontWeight: '400',
    letterSpacing: -0.01,
  },
  glassEffect: {
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  glassCard: {
    borderRadius: 16,
    padding: 32,
    marginBottom: 24,
    borderWidth: 1,
    overflow: 'hidden',
  },
  gradientButton: {
    borderRadius: 96, // Half of 192px button size
    shadowOffset: {
      width: 0,
      height: 20,
    },
    shadowOpacity: 0.1,
    shadowRadius: 40,
    elevation: 8,
  },
  gradientButtonPressed: {
    transform: [{ scale: 0.95 }],
  },
  gradientButtonRecording: {
    transform: [{ scale: 1.1 }],
  },
  gradientBackground: {
    width: 192,
    height: 192,
    borderRadius: 96,
    justifyContent: 'center',
    alignItems: 'center',
    shadowOffset: {
      width: 0,
      height: 4,
    },
    shadowOpacity: 0.2,
    shadowRadius: 12,
  },
  backgroundGradient: {
    flex: 1,
  },
});