/**
 * Siri-like microphone button with wave animations matching the original design
 */
import React, { useEffect, useRef, useCallback } from 'react';
import { View, StyleSheet, Animated, Easing } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { GradientButton } from './ThemedComponents';
import { Colors } from '@/constants/Colors';

interface SiriMicrophoneProps {
  recording: boolean;
  uploading: boolean;
  onPressIn: () => void;
  onPressOut: () => void;
  disabled?: boolean;
}

export function SiriMicrophone({ 
  recording, 
  uploading, 
  onPressIn, 
  onPressOut, 
  disabled 
}: SiriMicrophoneProps) {
  const commonColors = Colors.common;
  
  // Animation refs for Siri waves
  const wave1 = useRef(new Animated.Value(12)).current;
  const wave2 = useRef(new Animated.Value(8)).current;
  const wave3 = useRef(new Animated.Value(16)).current;
  const wave4 = useRef(new Animated.Value(6)).current;
  const wave5 = useRef(new Animated.Value(10)).current;
  
  // Pulse animation
  const pulseScale = useRef(new Animated.Value(1)).current;
  const pulseOpacity = useRef(new Animated.Value(0)).current;

  const createWaveAnimation = useCallback((animValue: Animated.Value, minHeight: number, maxHeight: number, duration: number) => {
    return Animated.loop(
      Animated.sequence([
        Animated.timing(animValue, {
          toValue: maxHeight,
          duration: duration / 2,
          easing: Easing.inOut(Easing.sin),
          useNativeDriver: false,
        }),
        Animated.timing(animValue, {
          toValue: minHeight,
          duration: duration / 2,
          easing: Easing.inOut(Easing.sin),
          useNativeDriver: false,
        }),
      ])
    );
  }, []);

  useEffect(() => {
    if (recording) {
      // Start all wave animations with different timing
      createWaveAnimation(wave1, 12, 32, 1500).start();
      createWaveAnimation(wave2, 8, 24, 1800).start();
      createWaveAnimation(wave3, 16, 40, 2100).start();
      createWaveAnimation(wave4, 6, 18, 1700).start();
      createWaveAnimation(wave5, 10, 28, 1900).start();
      
      // Pulse animation
      Animated.loop(
        Animated.sequence([
          Animated.parallel([
            Animated.timing(pulseScale, {
              toValue: 1.4,
              duration: 1000,
              easing: Easing.out(Easing.ease),
              useNativeDriver: true,
            }),
            Animated.timing(pulseOpacity, {
              toValue: 0.6,
              duration: 500,
              useNativeDriver: true,
            }),
          ]),
          Animated.parallel([
            Animated.timing(pulseScale, {
              toValue: 1,
              duration: 1000,
              easing: Easing.out(Easing.ease),
              useNativeDriver: true,
            }),
            Animated.timing(pulseOpacity, {
              toValue: 0,
              duration: 500,
              useNativeDriver: true,
            }),
          ]),
        ])
      ).start();
    } else {
      // Stop animations and reset
      wave1.stopAnimation(() => wave1.setValue(12));
      wave2.stopAnimation(() => wave2.setValue(8));
      wave3.stopAnimation(() => wave3.setValue(16));
      wave4.stopAnimation(() => wave4.setValue(6));
      wave5.stopAnimation(() => wave5.setValue(10));
      pulseScale.stopAnimation(() => pulseScale.setValue(1));
      pulseOpacity.stopAnimation(() => pulseOpacity.setValue(0));
    }
  }, [recording, createWaveAnimation, wave1, wave2, wave3, wave4, wave5, pulseScale, pulseOpacity]);

  return (
    <View style={styles.container}>
      {/* Pulse rings */}
      <Animated.View 
        style={[
          styles.pulseRing,
          {
            transform: [{ scale: pulseScale }],
            opacity: pulseOpacity,
            borderColor: commonColors.white,
          }
        ]} 
      />
      <Animated.View 
        style={[
          styles.pulseRing,
          {
            transform: [{ scale: pulseScale }],
            opacity: pulseOpacity,
            borderColor: commonColors.white,
          }
        ]} 
      />
      
      {/* Main button */}
      <GradientButton
        onPressIn={onPressIn}
        onPressOut={onPressOut}
        disabled={disabled || uploading}
        recording={recording}
        style={styles.button}
      >
        {/* Siri waves (only visible when recording) */}
        {recording && (
          <View style={styles.wavesContainer}>
            <Animated.View style={[styles.wave, { height: wave1, backgroundColor: 'rgba(255, 255, 255, 0.95)' }]} />
            <Animated.View style={[styles.wave, { height: wave2, backgroundColor: 'rgba(255, 255, 255, 0.85)' }]} />
            <Animated.View style={[styles.wave, { height: wave3, backgroundColor: 'rgba(255, 255, 255, 0.75)' }]} />
            <Animated.View style={[styles.wave, { height: wave4, backgroundColor: 'rgba(255, 255, 255, 0.65)' }]} />
            <Animated.View style={[styles.wave, { height: wave5, backgroundColor: 'rgba(255, 255, 255, 0.55)' }]} />
          </View>
        )}
        
        {/* Static microphone icon (hidden when recording) */}
        {!recording && (
          <Ionicons
            name="mic"
            size={64}
            color={commonColors.white}
            style={[
              styles.micIcon,
              { opacity: recording ? 0 : 1 }
            ]}
          />
        )}
        
        {/* Upload indicator */}
        {uploading && (
          <Ionicons
            name="hourglass"
            size={64}
            color={commonColors.white}
          />
        )}
      </GradientButton>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    position: 'relative',
    width: 192,
    height: 192,
    justifyContent: 'center',
    alignItems: 'center',
  },
  button: {
    // GradientButton handles the styling
  },
  pulseRing: {
    position: 'absolute',
    width: 192,
    height: 192,
    borderRadius: 96,
    borderWidth: 2,
    borderColor: 'rgba(255, 255, 255, 0.6)',
  },
  wavesContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 4,
  },
  wave: {
    width: 3,
    borderRadius: 2,
  },
  micIcon: {
    textShadowColor: 'rgba(0, 0, 0, 0.3)',
    textShadowOffset: { width: 0, height: 2 },
    textShadowRadius: 4,
  },
});