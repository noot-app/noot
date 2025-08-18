/**
 * Main noot nutrition app screen matching the original web frontend design
 * with Apple-inspired aesthetics, animated microphone, and comprehensive theming
 */
import React, { useState } from 'react';
import { View, StyleSheet, StatusBar } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { ThemedBackground, ThemedText } from '@/components/ThemedComponents';
import { MicrophoneButton } from '@/components/MicrophoneButton';
import { ThemeToggle } from '@/components/ThemeToggle';
import { NutritionResults } from '@/components/NutritionResults';
import { useAudioRecorder } from '@/hooks/useAudioRecorder';
import { useColorScheme, toggleTheme } from '@/hooks/useColorScheme';

export default function HomeScreen() {
  const insets = useSafeAreaInsets();
  const colorScheme = useColorScheme();
  
  const [nutritionData, setNutritionData] = useState<any>(null);
  const { recording, uploading, status, startRecording, stopRecording } = useAudioRecorder();

  const handleStartRecording = async () => {
    setNutritionData(null); // Clear previous results
    await startRecording();
  };

  const handleStopRecording = async () => {
    const result = await stopRecording();
    if (result) {
      setNutritionData(result);
    }
  };

  const handleThemeToggle = () => {
    toggleTheme();
  };

  return (
    <ThemedBackground>
      <StatusBar 
        barStyle={colorScheme === 'dark' ? 'light-content' : 'dark-content'} 
        backgroundColor="transparent"
        translucent
      />
      
      {/* Theme Toggle Button */}
      <ThemeToggle 
        style={{ top: insets.top + 16 }} 
        onToggle={handleThemeToggle} 
      />

      <View style={[styles.container, { paddingTop: insets.top }]}>
        {/* Main Content Area */}
        <View style={styles.mainContent}>
          {/* Status Message */}
          {status && (
            <View style={styles.statusContainer}>
              <ThemedText type="tertiary" style={styles.statusText}>
                {status}
              </ThemedText>
            </View>
          )}

          {/* Microphone Button */}
          <View style={styles.microphoneContainer}>
            <MicrophoneButton
              recording={recording}
              uploading={uploading}
              onPressIn={handleStartRecording}
              onPressOut={handleStopRecording}
              disabled={uploading}
            />
          </View>

          {/* Instruction Text (only when no data) */}
          {!nutritionData && !status && (
            <View style={styles.instructionContainer}>
              <ThemedText style={styles.instructionTitle}>
                🍎 noot
              </ThemedText>
              <ThemedText type="secondary" style={styles.instructionSubtitle}>
                Press and hold the microphone to record your meal
              </ThemedText>
            </View>
          )}
        </View>

        {/* Results Section */}
        {nutritionData && (
          <View style={styles.resultsSection}>
            <NutritionResults data={nutritionData} />
          </View>
        )}
      </View>
    </ThemedBackground>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    minHeight: '100%',
  },
  mainContent: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 32,
    minHeight: '100%', // Ensure full viewport height
  },
  statusContainer: {
    position: 'absolute',
    top: '30%',
    alignItems: 'center',
    paddingHorizontal: 32,
    zIndex: 10,
  },
  statusText: {
    fontSize: 18,
    fontWeight: '500',
    textAlign: 'center',
    letterSpacing: -0.01,
  },
  microphoneContainer: {
    alignItems: 'center',
    justifyContent: 'center',
  },
  instructionContainer: {
    alignItems: 'center',
    marginTop: 32,
    paddingHorizontal: 32,
  },
  instructionTitle: {
    fontSize: 48,
    fontWeight: '700',
    marginBottom: 16,
    letterSpacing: -0.02,
    textAlign: 'center',
  },
  instructionSubtitle: {
    fontSize: 18,
    textAlign: 'center',
    lineHeight: 24,
    letterSpacing: -0.01,
    maxWidth: 400,
  },
  resultsSection: {
    flex: 1,
    width: '100%',
    paddingHorizontal: 32,
    paddingTop: 32,
    paddingBottom: 32,
    alignItems: 'center',
    minHeight: 0, // Allow shrinking
  },
});
