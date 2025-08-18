import React, { useState } from 'react';
import { View, Pressable, Text, ActivityIndicator, Platform, ScrollView } from 'react-native';
import { useAudioRecorder } from '@/hooks/useAudioRecorder';

const API_URL = process.env.EXPO_PUBLIC_API_URL || 'http://localhost:3000';

export default function Home() {
  const { start, stop } = useAudioRecorder();
  const [recording, setRecording] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [status, setStatus] = useState<string>('');

  async function onPress() {
    if (!recording) {
      await start();
      setRecording(true);
      setStatus('Recording...');
    } else {
      const audio = await stop();
      setRecording(false);
      setStatus('Processing...');
      setUploading(true);
      try {
        const form = new FormData();
        form.append('audio', {
          // IMPORTANT: field name must be "audio" to match ingestHandler
          uri: audio.uri,
          name: `meal.${audio.mimeType.includes('webm') ? 'webm' : 'm4a'}`,
          type: audio.mimeType,
        } as any);
        const resp = await fetch(`${API_URL}/api/ingest`, {
          method: 'POST',
          body: form,
        });
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const data = await resp.json();
        setResult(data);
        setStatus('');
      } catch (e: any) {
        setStatus(`Error: ${e.message}`);
      } finally {
        setUploading(false);
      }
    }
  }

  return (
    <ScrollView contentContainerStyle={{ flexGrow: 1 }}>
      <View style={{ flex: 1, alignItems: 'center', justifyContent: 'center', padding: 24, gap: 16 }}>
        <Pressable
          onPress={onPress}
          style={{ width: 160, height: 160, borderRadius: 80, backgroundColor: recording ? '#ef4444' : '#3b82f6', alignItems: 'center', justifyContent: 'center' }}
        >
          {uploading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={{ color: 'white', fontSize: 18 }}>{recording ? 'Stop' : 'Tap to Record'}</Text>
          )}
        </Pressable>

        {status ? <Text style={{ color: '#64748b' }}>{status}</Text> : null}

        {result && (
          <View style={{ width: '100%', maxWidth: 900, gap: 16 }}>
            <View style={{ padding: 16, borderRadius: 12, backgroundColor: Platform.OS === 'web' ? 'rgba(0,0,0,0.05)' : '#11182710' }}>
              <Text style={{ fontSize: 20, fontWeight: '700', marginBottom: 8 }}>Nutrition Summary</Text>
              <Text>Calories: {Math.round(result.summary?.totals?.calories || 0)} kcal</Text>
              <Text>Protein: {Math.round(result.summary?.totals?.protein_g || 0)} g</Text>
              <Text>Carbs: {Math.round(result.summary?.totals?.total_carbs_g || 0)} g</Text>
              <Text>Fat: {Math.round(result.summary?.totals?.total_fat_g || 0)} g</Text>
            </View>

            <View style={{ padding: 16, borderRadius: 12, backgroundColor: Platform.OS === 'web' ? 'rgba(0,0,0,0.05)' : '#11182710' }}>
              <Text style={{ fontSize: 20, fontWeight: '700', marginBottom: 8 }}>Items</Text>
              {Array.isArray(result.items) && result.items.length > 0 ? (
                result.items.map((it: any, idx: number) => (
                  <View key={idx} style={{ marginBottom: 12 }}>
                    <Text style={{ fontWeight: '600' }}>{it.item?.name}</Text>
                    {it.item?.quantity ? (
                      <Text style={{ color: '#6b7280' }}>
                        {it.item.quantity} {it.item.unit || ''}
                      </Text>
                    ) : null}
                    {it.item?.nutrients ? (
                      <Text>
                        {Math.round(it.item.nutrients.calories)} kcal, {Math.round(it.item.nutrients.protein_g)}g protein, {Math.round(it.item.nutrients.total_carbs_g)}g carbs, {Math.round(it.item.nutrients.total_fat_g)}g fat
                      </Text>
                    ) : (
                      <Text style={{ color: '#6b7280' }}>Nutrition data unavailable</Text>
                    )}
                  </View>
                ))
              ) : (
                <Text style={{ color: '#6b7280' }}>No items</Text>
              )}
            </View>

            <View style={{ padding: 16, borderRadius: 12, backgroundColor: Platform.OS === 'web' ? 'rgba(0,0,0,0.05)' : '#11182710' }}>
              <Text style={{ fontSize: 16, fontWeight: '600', marginBottom: 8 }}>What you said</Text>
              <Text style={{ fontStyle: 'italic' }}>{result.transcript || ''}</Text>
            </View>
          </View>
        )}
      </View>
    </ScrollView>
  );
}