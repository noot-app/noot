/**
 * Cross-platform audio recording hook matching the original noot web implementation
 */
import { useState, useRef, useEffect } from 'react';
import { Platform, Alert } from 'react-native';
import { Audio } from 'expo-av';

const API_URL = process.env.EXPO_PUBLIC_API_URL || 'http://localhost:3000';
const isDev = __DEV__;

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

// Development-only logging function
function devLog(...args: any[]) {
  if (isDev) {
    console.log(...args);
  }
}

function devError(...args: any[]) {
  if (isDev) {
    console.error(...args);
  }
}

function devWarn(...args: any[]) {
  if (isDev) {
    console.warn(...args);
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
  const uploadPromiseRef = useRef<Promise<any> | null>(null);
  const resolveUploadRef = useRef<((value: any) => void) | null>(null);

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
        const result = await uploadRecording(blob);
        if (resolveUploadRef.current) {
          resolveUploadRef.current(result);
          resolveUploadRef.current = null;
        }
      };

      mediaRecorderRef.current = mediaRecorder;
    } catch (error) {
      devError('Web audio setup failed:', error);
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
        devLog('Requesting permissions..');
        await Audio.requestPermissionsAsync();
        await Audio.setAudioModeAsync({
          allowsRecordingIOS: true,
          playsInSilentModeIOS: true,
        });

        devLog('Starting recording..');
        const { recording } = await Audio.Recording.createAsync(
          Audio.RecordingOptionsPresets.HIGH_QUALITY
        );
        setRecording(recording);
        devLog('Recording started');
      }
    } catch (err) {
      devError('Failed to start recording', err);
      Alert.alert('Error', 'Failed to start recording');
      setStatus('❌ Failed to start recording');
    }
  };

  const stopRecording = async () => {
    try {
      setStatus('⏳ Processing...');
      
      if (Platform.OS === 'web') {
        // Web recording - create a promise that resolves when upload completes
        return new Promise((resolve) => {
          resolveUploadRef.current = resolve;
          if (mediaRecorderRef.current && mediaRecorderRef.current.state === 'recording') {
            mediaRecorderRef.current.stop();
          } else {
            resolve(null);
          }
        });
      } else {
        // Native recording
        devLog('Stopping recording..');
        setRecording(null);
        if (recording) {
          await recording.stopAndUnloadAsync();
          await Audio.setAudioModeAsync({
            allowsRecordingIOS: false,
          });
          const uri = recording.getURI();
          devLog('Recording stopped and stored at', uri);
          
          if (uri) {
            return await uploadRecording(uri);
          }
        }
        return null;
      }
    } catch (error) {
      devError('Error stopping recording:', error);
      setStatus('❌ Error stopping recording');
      return null;
    }
  };

  const uploadRecording = async (audioData: string | Blob) => {
    try {
      devLog('Starting upload:', { 
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
        devLog('Testing backend connection...');
        const healthResponse = await fetch(`${API_URL}/api/health`, { 
          method: 'GET',
          mode: 'cors',
          headers: {
            'Accept': 'application/json',
          },
        });
        devLog('Backend health check:', healthResponse.ok, healthResponse.status);
        if (!healthResponse.ok) {
          throw new Error(`Health check failed: ${healthResponse.status}`);
        }
      } catch (healthError) {
        devError('Backend health check failed:', healthError);
        throw new Error('Cannot reach backend server. Make sure it\'s running on port 3000.');
      }

      devLog('Making upload request...');
      const response = await fetch(`${API_URL}/api/ingest`, {
        method: 'POST',
        mode: 'cors',
        body: formData,
        // Don't set Content-Type header - let the browser/fetch set it automatically for multipart/form-data
      });

      devLog('Upload response:', response.status, response.ok, response.statusText);

      if (response.ok) {
        const data = await response.json();
        devLog('Upload successful:', data);
        setStatus('');
        return data;
      } else {
        const errorText = await response.text();
        devError('Upload failed:', response.status, response.statusText, errorText);
        throw new Error(`Server error: ${response.status} ${response.statusText}`);
      }
    } catch (error) {
      devError('Error uploading recording:', error);
      
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