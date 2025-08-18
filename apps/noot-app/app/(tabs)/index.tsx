import React, { useState, useRef, useEffect } from 'react';
import { View, Text, Pressable, Alert, StyleSheet, ActivityIndicator, ScrollView, Platform } from 'react-native';
import { Audio } from 'expo-av';

const API_URL = 'http://localhost:3000';

// Web-specific types and interfaces
interface MediaRecorderOptions {
  mimeType: string;
}

declare global {
  interface Window {
    MediaRecorder: {
      new (stream: MediaStream, options?: MediaRecorderOptions): MediaRecorder;
      isTypeSupported(mimeType: string): boolean;
    };
  }
}

export default function HomeScreen() {
  const [recording, setRecording] = useState<Audio.Recording | null>(null);
  const [uploading, setUploading] = useState(false);
  const [nutritionData, setNutritionData] = useState<any>(null);
  const [status, setStatus] = useState<string>('');
  
  // Web-specific state
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const chunksRef = useRef<Blob[]>([]);

  useEffect(() => {
    const setupWebAudioWrapper = async () => {
      if (Platform.OS === 'web') {
        await setupWebAudio();
      }
    };
    
    setupWebAudioWrapper();
    
    return () => {
      // Cleanup on unmount
      if (Platform.OS === 'web' && streamRef.current) {
        streamRef.current.getTracks().forEach(track => track.stop());
      }
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const setupWebAudio = async () => {
    if (Platform.OS !== 'web') return;
    
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ 
        audio: {
          channelCount: 1,
          sampleRate: 44100,
        } 
      });
      
      streamRef.current = stream;

      const mediaRecorder = new window.MediaRecorder(stream, {
        mimeType: 'audio/webm;codecs=opus'
      });

      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          chunksRef.current.push(event.data);
        }
      };

      mediaRecorder.onstop = async () => {
        const blob = new Blob(chunksRef.current, { type: 'audio/webm' });
        chunksRef.current = [];
        await uploadRecording(blob);
      };

      mediaRecorderRef.current = mediaRecorder;
    } catch (error) {
      console.error('Web audio setup failed:', error);
      setStatus('❌ Microphone access denied. Please allow microphone access and refresh.');
    }
  };

  const startRecording = async () => {
    try {
      setStatus('🎤 Recording... Release to stop');
      setNutritionData(null);

      if (Platform.OS === 'web') {
        // Web recording using MediaRecorder
        if (!mediaRecorderRef.current) {
          await setupWebAudio();
        }
        
        if (mediaRecorderRef.current && mediaRecorderRef.current.state !== 'recording') {
          chunksRef.current = [];
          mediaRecorderRef.current.start();
        }
      } else {
        // Native recording using expo-av
        console.log('Requesting permissions..');
        await Audio.requestPermissionsAsync();
        await Audio.setAudioModeAsync({
          allowsRecordingIOS: true,
          playsInSilentModeIOS: true,
        });

        console.log('Starting recording..');
        const { recording } = await Audio.Recording.createAsync(
          Audio.RecordingOptionsPresets.HIGH_QUALITY
        );
        setRecording(recording);
        console.log('Recording started');
      }
    } catch (err) {
      console.error('Failed to start recording', err);
      Alert.alert('Error', 'Failed to start recording');
      setStatus('❌ Failed to start recording');
    }
  };

  const stopRecording = async () => {
    try {
      setStatus('⏳ Processing...');
      
      if (Platform.OS === 'web') {
        // Web recording
        if (mediaRecorderRef.current && mediaRecorderRef.current.state === 'recording') {
          mediaRecorderRef.current.stop();
        }
      } else {
        // Native recording
        console.log('Stopping recording..');
        setRecording(null);
        if (recording) {
          await recording.stopAndUnloadAsync();
          await Audio.setAudioModeAsync({
            allowsRecordingIOS: false,
          });
          const uri = recording.getURI();
          console.log('Recording stopped and stored at', uri);
          
          if (uri) {
            await uploadRecording(uri);
          }
        }
      }
    } catch (error) {
      console.error('Error stopping recording:', error);
      setStatus('❌ Error stopping recording');
    }
  };

  const uploadRecording = async (audioData: string | Blob) => {
    try {
      console.log('Starting upload:', { 
        platform: Platform.OS,
        apiUrl: API_URL,
        audioType: audioData instanceof Blob ? `Blob(${audioData.size})` : 'URI'
      });
      
      setUploading(true);
      setStatus('🔄 Processing...');

      const formData = new FormData();
      
      if (Platform.OS === 'web') {
        // Web: audioData is a Blob
        formData.append('audio', audioData as Blob, 'recording.webm');
      } else {
        // Native: audioData is a file URI string
        formData.append('audio', {
          uri: audioData as string,
          type: 'audio/m4a',
          name: 'recording.m4a',
        } as any);
      }

      // Test backend connectivity first
      try {
        console.log('Testing backend connection...');
        const healthResponse = await fetch(`${API_URL}/api/health`, { 
          method: 'GET',
          mode: 'cors'
        });
        console.log('Backend health check:', healthResponse.ok, healthResponse.status);
      } catch (healthError) {
        console.error('Backend health check failed:', healthError);
        throw new Error('Cannot reach backend server. Make sure it\'s running on port 3000.');
      }

      console.log('Making upload request...');
      const response = await fetch(`${API_URL}/api/ingest`, {
        method: 'POST',
        mode: 'cors',
        body: formData,
        // Don't set Content-Type header - let the browser/fetch set it automatically
      });

      console.log('Upload response:', response.status, response.ok);

      if (response.ok) {
        try {
          const data = await response.json();
          console.log('Upload successful:', data);
          
          // Debug: Log the actual response structure
          console.log('Response structure:', {
            hasTranscript: !!data.transcript,
            hasSummary: !!data.summary,
            hasItems: !!data.items,
            summaryKeys: data.summary ? Object.keys(data.summary) : null,
            itemsLength: data.items ? data.items.length : 0,
            totalsKeys: data.summary?.totals ? Object.keys(data.summary.totals) : null,
            firstItem: data.items?.[0] ? {
              hasItemField: !!data.items[0].item,
              itemName: data.items[0].item?.name || data.items[0].name,
              hasNutrients: !!data.items[0].item?.nutrients
            } : null
          });
          
          setNutritionData(data);
          setStatus('');
        } catch (jsonError) {
          console.error('JSON parsing error:', jsonError);
          // Can't read response text after json() fails
          throw new Error('Failed to parse JSON response');
        }
      } else {
        const errorText = await response.text();
        console.error('Upload failed:', response.status, errorText);
        throw new Error(`HTTP error! status: ${response.status} - ${errorText}`);
      }
    } catch (error) {
      console.error('Error uploading recording:', error);
      console.error('Error details:', {
        message: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : null,
        type: typeof error,
        name: error instanceof Error ? error.name : 'Unknown'
      });
      
      let errorMessage = 'Failed to process recording. Please try again.';
      if (error instanceof Error) {
        if (error.message.includes('Cannot reach backend')) {
          errorMessage = 'Backend server not reachable. Check if it\'s running.';
        } else if (error.name === 'TypeError' && error.message.includes('fetch')) {
          errorMessage = 'Network error. Check your connection and backend server.';
        }
      }
      
      Alert.alert('Error', errorMessage);
      setStatus('❌ Error processing audio. Please try again.');
    } finally {
      setUploading(false);
    }
  };

  const formatNumber = (num: number | undefined): string => {
    if (typeof num !== 'number' || isNaN(num)) return '0';
    return num % 1 === 0 ? num.toString() : num.toFixed(1);
  };

  const getPercentageColor = (percentage: number): string => {
    if (percentage < 25) return '#ef4444'; // red
    if (percentage < 50) return '#f59e0b'; // amber
    if (percentage < 75) return '#eab308'; // yellow
    if (percentage < 100) return '#22c55e'; // green
    return '#3b82f6'; // blue (over 100%)
  };

  const renderNutritionData = () => {
    if (!nutritionData) return null;

    try {
      const { transcript, summary, items } = nutritionData;
      
      return (
        <ScrollView style={styles.resultsContainer} showsVerticalScrollIndicator={false}>
          {/* Transcript Section */}
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>What you said:</Text>
            <View style={styles.transcriptContainer}>
              <Text style={styles.transcript}>&ldquo;{transcript || 'No transcript available'}&rdquo;</Text>
            </View>
          </View>
          
          {/* Summary Section */}
          {summary && summary.totals && (
            <View style={styles.section}>
              <Text style={styles.sectionTitle}>⚡ Nutrition Summary</Text>
              {renderMacronutrients(summary)}
              {renderSugarBreakdown(summary)}
              {renderFiberAndMinerals(summary)}
            </View>
          )}

          {/* Individual Items Section */}
          {items && items.length > 0 && (
            <View style={styles.section}>
              <Text style={styles.sectionTitle}>🍽️ Food Items Detected</Text>
              {items.map((itemData: any, index: number) => {
                try {
                  const item = itemData.item; // Extract the actual item from ItemWithNutrition wrapper
                  const nutrients = item?.nutrients;
                  
                  return (
                    <View key={index} style={styles.foodItem}>
                      <Text style={styles.itemName}>{item?.name || 'Unknown item'}</Text>
                      {nutrients ? (
                        <View style={styles.itemNutrients}>
                          <Text style={styles.itemNutrientText}>
                            {formatNumber(nutrients.calories)} cal • {formatNumber(nutrients.protein_g)}g protein
                          </Text>
                          {nutrients.total_carbs_g && (
                            <Text style={styles.itemNutrientText}>
                              {formatNumber(nutrients.total_carbs_g)}g carbs • {formatNumber(nutrients.total_fat_g)}g fat
                            </Text>
                          )}
                        </View>
                      ) : (
                        itemData.note && (
                          <Text style={styles.itemNote}>{itemData.note}</Text>
                        )
                      )}
                    </View>
                  );
                } catch (itemError) {
                  console.error('Error rendering item:', itemError, itemData);
                  return (
                    <View key={index} style={styles.foodItem}>
                      <Text style={styles.itemName}>Error rendering item</Text>
                    </View>
                  );
                }
              })}
            </View>
          )}
        </ScrollView>
      );
    } catch (error) {
      console.error('Error rendering nutrition data:', error, nutritionData);
      return (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Error displaying results</Text>
          <Text style={styles.itemNote}>Please try recording again</Text>
        </View>
      );
    }
  };

  const renderMacronutrients = (summary: any) => {
    try {
      if (!summary.totals) return null;
      
      const totals = summary.totals;
      const percentDaily = summary.percent_of_daily || {};
      
      const macros = [
        { 
          label: 'Calories', 
          value: formatNumber(totals.calories), 
          unit: 'kcal', 
          percent: Math.min(percentDaily.calories || 0, 100)
        },
        { 
          label: 'Protein', 
          value: formatNumber(totals.protein_g), 
          unit: 'g', 
          percent: Math.min(percentDaily.protein || 0, 100)
        },
        { 
          label: 'Carbs', 
          value: formatNumber(totals.total_carbs_g), 
          unit: 'g', 
          percent: Math.min(percentDaily.total_carbs || 0, 100)
        },
        { 
          label: 'Fat', 
          value: formatNumber(totals.total_fat_g), 
          unit: 'g', 
          percent: Math.min(percentDaily.total_fat || 0, 100)
        }
      ];

      return (
        <View style={styles.nutritionGrid}>
          {macros.map((macro, index) => (
            <View key={index} style={styles.nutritionItem}>
              <View style={[styles.percentageIndicator, { backgroundColor: getPercentageColor(macro.percent) }]} />
              <Text style={styles.nutritionLabel}>{macro.label}</Text>
              <Text style={styles.nutritionValue}>{macro.value}{macro.unit}</Text>
              <Text style={styles.nutritionPercent}>{Math.round(macro.percent)}% daily</Text>
            </View>
          ))}
        </View>
      );
    } catch (error) {
      console.error('Error rendering macronutrients:', error, summary);
      return null;
    }
  };

  const renderSugarBreakdown = (summary: any) => {
    try {
      if (!summary.totals) return null;
      
      const totals = summary.totals;
      const totalSugarsG = totals.total_sugars_g || 0;
      const addedSugarsG = totals.added_sugars_g || 0;
      const naturalSugarsG = Math.max(0, totalSugarsG - addedSugarsG);
      
      if (totalSugarsG === 0) return null;
      
      const addedSugarLimit = 50;
      const addedSugarPercent = Math.min((addedSugarsG / addedSugarLimit) * 100, 100);
      
      return (
        <View style={styles.sugarSection}>
          <Text style={styles.nutritionSubtitle}>🍯 Sugar Breakdown</Text>
          <View style={styles.sugarGrid}>
            {addedSugarsG > 0 && (
              <View style={styles.nutritionItem}>
                <View style={[styles.percentageIndicator, { 
                  backgroundColor: addedSugarPercent > 80 ? '#ef4444' : getPercentageColor(addedSugarPercent) 
                }]} />
                <Text style={styles.nutritionLabel}>Added Sugar</Text>
                <Text style={styles.nutritionValue}>{formatNumber(addedSugarsG)}g</Text>
                <Text style={styles.nutritionPercent}>of {addedSugarLimit}g limit</Text>
              </View>
            )}
            {naturalSugarsG > 0 && (
              <View style={styles.nutritionItem}>
                <View style={[styles.percentageIndicator, { backgroundColor: '#22c55e' }]} />
                <Text style={styles.nutritionLabel}>Natural Sugar</Text>
                <Text style={styles.nutritionValue}>{formatNumber(naturalSugarsG)}g</Text>
                <Text style={styles.nutritionPercent}>from fruits & dairy</Text>
              </View>
            )}
          </View>
        </View>
      );
    } catch (error) {
      console.error('Error rendering sugar breakdown:', error, summary);
      return null;
    }
  };

  const renderFiberAndMinerals = (summary: any) => {
    try {
      if (!summary.totals) return null;
      
      const totals = summary.totals;
      const percentDaily = summary.percent_of_daily || {};
      
      const nutrients = [
        { 
          label: 'Fiber', 
          value: formatNumber(totals.dietary_fiber_g), 
          unit: 'g', 
          percent: Math.min(percentDaily.dietary_fiber || 0, 100)
        },
        { 
          label: 'Sodium', 
          value: formatNumber(totals.sodium_mg), 
          unit: 'mg', 
          percent: Math.min(percentDaily.sodium || 0, 100)
        },
        { 
          label: 'Calcium', 
          value: formatNumber(totals.calcium_mg), 
          unit: 'mg', 
          percent: Math.min(percentDaily.calcium || 0, 100)
        },
        { 
          label: 'Iron', 
          value: formatNumber(totals.iron_mg), 
          unit: 'mg', 
          percent: Math.min(percentDaily.iron || 0, 100)
        }
      ].filter(nutrient => parseFloat(nutrient.value) > 0);

      if (nutrients.length === 0) return null;

      return (
        <View style={styles.mineralsSection}>
          <Text style={styles.nutritionSubtitle}>💊 Vitamins & Minerals</Text>
          <View style={styles.nutritionGrid}>
            {nutrients.map((nutrient, index) => (
              <View key={index} style={styles.nutritionItem}>
                <View style={[styles.percentageIndicator, { backgroundColor: getPercentageColor(nutrient.percent) }]} />
                <Text style={styles.nutritionLabel}>{nutrient.label}</Text>
                <Text style={styles.nutritionValue}>{nutrient.value}{nutrient.unit}</Text>
                <Text style={styles.nutritionPercent}>{Math.round(nutrient.percent)}% daily</Text>
              </View>
            ))}
          </View>
        </View>
      );
    } catch (error) {
      console.error('Error rendering fiber and minerals:', error, summary);
      return null;
    }
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>🥗 Noot Nutrition</Text>
        <Text style={styles.subtitle}>Record your meal to get nutrition info</Text>
        {status ? <Text style={styles.statusText}>{status}</Text> : null}
      </View>

      <View style={styles.recordingContainer}>
        <Pressable
          style={[styles.recordButton, recording && styles.recordButtonActive]}
          onPressIn={startRecording}
          onPressOut={stopRecording}
          disabled={uploading}
        >
          {uploading ? (
            <ActivityIndicator size="large" color="#fff" />
          ) : (
            <>
              <Text style={styles.recordButtonText}>
                {recording ? '🎤 Recording...' : '🎙️ Hold to Record'}
              </Text>
              {recording && <Text style={styles.recordingHint}>Release to stop</Text>}
            </>
          )}
        </Pressable>
      </View>

      {renderNutritionData()}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8fafc',
    padding: 20,
  },
  header: {
    alignItems: 'center',
    marginTop: 40,
    marginBottom: 40,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#1e293b',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#64748b',
    textAlign: 'center',
    marginBottom: 8,
  },
  statusText: {
    fontSize: 14,
    color: '#6366f1',
    textAlign: 'center',
    marginTop: 8,
  },
  recordingContainer: {
    alignItems: 'center',
    marginBottom: 30,
  },
  recordButton: {
    backgroundColor: '#3b82f6',
    width: 200,
    height: 200,
    borderRadius: 100,
    justifyContent: 'center',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 4,
    },
    shadowOpacity: 0.3,
    shadowRadius: 4.65,
    elevation: 8,
  },
  recordButtonActive: {
    backgroundColor: '#ef4444',
    transform: [{ scale: 1.1 }],
  },
  recordButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
    textAlign: 'center',
  },
  recordingHint: {
    color: '#fff',
    fontSize: 14,
    marginTop: 8,
    textAlign: 'center',
  },
  resultsContainer: {
    flex: 1,
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 20,
    marginTop: 20,
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 3.84,
    elevation: 5,
  },
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#1e293b',
    marginBottom: 16,
  },
  transcriptContainer: {
    backgroundColor: '#f1f5f9',
    borderRadius: 12,
    padding: 16,
    borderLeftWidth: 4,
    borderLeftColor: '#3b82f6',
  },
  transcript: {
    fontSize: 16,
    color: '#475569',
    fontStyle: 'italic',
    lineHeight: 24,
  },
  nutritionGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
    marginBottom: 16,
  },
  nutritionItem: {
    backgroundColor: '#f8fafc',
    borderRadius: 12,
    padding: 16,
    flex: 1,
    minWidth: '45%',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#e2e8f0',
  },
  percentageIndicator: {
    width: 4,
    height: 20,
    borderRadius: 2,
    marginBottom: 8,
  },
  nutritionLabel: {
    fontSize: 14,
    fontWeight: '600',
    color: '#64748b',
    marginBottom: 4,
    textAlign: 'center',
  },
  nutritionValue: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1e293b',
    marginBottom: 2,
    textAlign: 'center',
  },
  nutritionPercent: {
    fontSize: 12,
    color: '#64748b',
    textAlign: 'center',
  },
  nutritionSubtitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1e293b',
    marginBottom: 12,
  },
  sugarSection: {
    marginBottom: 20,
  },
  sugarGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
  },
  mineralsSection: {
    marginBottom: 20,
  },
  foodItem: {
    backgroundColor: '#f8fafc',
    padding: 16,
    borderRadius: 12,
    marginBottom: 12,
    borderLeftWidth: 4,
    borderLeftColor: '#10b981',
  },
  itemName: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#1e293b',
    marginBottom: 8,
  },
  itemNutrients: {
    gap: 4,
  },
  itemNutrientText: {
    fontSize: 14,
    color: '#64748b',
  },
  itemDetails: {
    fontSize: 14,
    color: '#64748b',
  },
  itemNote: {
    fontSize: 14,
    color: '#f59e0b',
    fontStyle: 'italic',
    marginTop: 4,
  },
});
