import { Capacitor } from '@capacitor/core';

/**
 * Capacitor-compatible microphone permission and access utilities
 */
export class CapacitorMicrophone {
  private static isNativeApp(): boolean {
    return Capacitor.isNativePlatform();
  }

  /**
   * Request microphone permissions in Capacitor
   */
  static async requestPermissions(): Promise<boolean> {
    try {
      // Debug logging
      console.log('🎤 Requesting microphone permissions...');
      console.log('🔍 User agent:', navigator.userAgent);
      console.log('🔍 Is HTTPS:', location.protocol === 'https:');
      console.log('🔍 Is Capacitor native:', this.isNativeApp());
      console.log('🔍 MediaDevices available:', !!navigator.mediaDevices);
      console.log('🔍 getUserMedia available:', !!navigator.mediaDevices?.getUserMedia);
      
      // For both native and web platforms, use getUserMedia to trigger permission request
      // On iOS, this will show the permission dialog using our Info.plist descriptions
      return await this.requestMicrophoneAccess();
    } catch (error) {
      console.error('🚨 Permission request failed:', error);
      return false;
    }
  }

  /**
   * Request actual microphone access (triggers permission dialog if needed)
   */
  private static async requestMicrophoneAccess(): Promise<boolean> {
    try {
      console.log('🎤 Attempting to access microphone...');
      
      if (typeof navigator === 'undefined' || !navigator.mediaDevices) {
        console.error('❌ MediaDevices API not available');
        throw new Error('MediaDevices API not available');
      }

      if (!navigator.mediaDevices.getUserMedia) {
        console.error('❌ getUserMedia not supported');
        throw new Error('getUserMedia not supported');
      }

      console.log('🎤 Calling getUserMedia...');
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          sampleRate: 44100,
        },
      });

      // Stop the stream immediately - we just wanted to check permission
      stream.getTracks().forEach((track: MediaStreamTrack) => track.stop());
      console.log('✅ Microphone permission granted');
      return true;
    } catch (error) {
      console.error('❌ Microphone access denied:', error);
      console.error('❌ Error name:', (error as Error)?.name || 'Unknown');
      console.error('❌ Error message:', (error as Error)?.message || 'No message');
      console.error('❌ Error stack:', (error as Error)?.stack || 'No stack');
      return false;
    }
  }

  /**
   * Check if microphone permission is already granted
   */
  static async checkPermissionStatus(): Promise<'granted' | 'denied' | 'prompt'> {
    try {
      // Simple permission check - try to access microphone
      return await this.requestMicrophoneAccess() ? 'granted' : 'denied';
    } catch (error) {
      console.error('Error checking permission status:', error);
      return 'denied';
    }
  }

  /**
   * Get microphone stream with proper Capacitor handling
   */
  static async getMicrophoneStream(constraints: MediaStreamConstraints = {
    audio: {
      echoCancellation: true,
      noiseSuppression: true,
      sampleRate: 44100,
    }
  }): Promise<MediaStream> {
    console.log('🎤 Getting microphone stream with constraints:', constraints);
    
    try {
      // Try with full constraints first
      const stream = await navigator.mediaDevices.getUserMedia(constraints);
      console.log('✅ Got microphone stream with full constraints');
      return stream;
    } catch (error) {
      console.log('⚠️ Failed with full constraints, trying basic audio only...');
      console.error('First attempt error:', error);
      
      try {
        // Fallback to basic audio only
        const basicStream = await navigator.mediaDevices.getUserMedia({ audio: true });
        console.log('✅ Got microphone stream with basic constraints');
        return basicStream;
      } catch (basicError) {
        console.error('❌ Failed with basic constraints too:', basicError);
        throw basicError;
      }
    }
  }

  /**
   * Show user-friendly error messages based on the error type
   */
  static getErrorMessage(error: unknown): string {
    const errorName = (error as Error)?.name || '';
    const errorMessage = (error as Error)?.message || '';

    if (errorName === 'NotAllowedError' || errorMessage.includes('Permission denied')) {
      return 'Microphone access denied. Please grant permission in your device settings and try again.';
    } else if (errorName === 'NotFoundError') {
      return 'No microphone found on this device.';
    } else if (errorName === 'NotSupportedError') {
      return 'Microphone access is not supported on this device.';
    } else if (errorName === 'SecurityError') {
      return 'Microphone access blocked due to security restrictions.';
    } else {
      return `Failed to access microphone: ${errorMessage || 'Unknown error'}`;
    }
  }
}
