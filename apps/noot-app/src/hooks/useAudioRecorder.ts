import { Platform } from 'react-native';

export type RecordingResult = { uri: string; mimeType: string };

export function useAudioRecorder() {
  let mediaRecorder: MediaRecorder | null = null;
  let chunks: BlobPart[] = [];
  let recordingObj: any = null; // expo-av Recording on native

  async function start(): Promise<void> {
    if (Platform.OS === 'web') {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      chunks = [];
      mediaRecorder = new MediaRecorder(stream, { mimeType: 'audio/webm' });
      mediaRecorder.ondataavailable = (e) => e.data.size && chunks.push(e.data);
      mediaRecorder.start();
    } else {
      const { Audio } = await import('expo-av');
      const { status } = await Audio.requestPermissionsAsync();
      if (!status || status !== 'granted') throw new Error('Mic permission denied');
      await Audio.setAudioModeAsync({ allowsRecordingIOS: true, playsInSilentModeIOS: true });
      recordingObj = new Audio.Recording();
      await recordingObj.prepareToRecordAsync(Audio.RECORDING_OPTIONS_PRESET_HIGH_QUALITY);
      await recordingObj.startAsync();
    }
  }

  async function stop(): Promise<RecordingResult> {
    if (Platform.OS === 'web') {
      if (!mediaRecorder) throw new Error('No active recording');
      return new Promise((resolve) => {
        mediaRecorder.onstop = async () => {
          const blob = new Blob(chunks, { type: 'audio/webm' });
          const uri = URL.createObjectURL(blob);
          resolve({ uri, mimeType: 'audio/webm' });
        };
        mediaRecorder.stop();
        mediaRecorder.stream.getTracks().forEach((t) => t.stop());
        mediaRecorder = null;
      });
    } else {
      await recordingObj.stopAndUnloadAsync();
      const uri = recordingObj.getURI();
      return { uri, mimeType: 'audio/m4a' };
    }
  }

  return { start, stop };
}