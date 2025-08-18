/**
 * Cross-platform audio recording hook matching the original noot web implementation
 */
import { useState, useRef, useEffect } from 'react';
import { Platform, Alert } from 'react-native';
import { Audio } from 'expo-av';

const API_URL = process.env.EXPO_PUBLIC_API_URL || 'http://localhost:3000';

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

export function useAudioRecorder() {
  const [recording, setRecording] = useState<Audio.Recording | null>(null);
  const [uploading, setUploading] = useState(false);
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
        throw new Error('Cannot reach backend server. Make sure it\'s running.');
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
        const data = await response.json();
        console.log('Upload successful:', data);
        setStatus('');
        return data;
      } else {
        const errorText = await response.text();
        console.error('Upload failed:', response.status, errorText);
        throw new Error(`HTTP error! status: ${response.status} - ${errorText}`);
      }
    } catch (error) {
      console.error('Error uploading recording:', error);
      
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
      return null;
    } finally {
      setUploading(false);
    }
  };

  return {
    recording: !!recording,
    uploading,
    status,
    startRecording,
    stopRecording,
  };
}