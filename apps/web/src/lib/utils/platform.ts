import { Capacitor } from '@capacitor/core';

/**
 * Platform detection utilities
 */
export class Platform {
  /**
   * Check if running in a Capacitor native mobile app
   */
  static isNativeApp(): boolean {
    return Capacitor.isNativePlatform();
  }

  /**
   * Check if running in a web browser
   */
  static isWeb(): boolean {
    return !Capacitor.isNativePlatform();
  }

  /**
   * Get the current platform name
   */
  static getPlatform(): string {
    return Capacitor.getPlatform();
  }

  /**
   * Check if running on iOS
   */
  static isIOS(): boolean {
    return Capacitor.getPlatform() === 'ios';
  }

  /**
   * Check if running on Android
   */
  static isAndroid(): boolean {
    return Capacitor.getPlatform() === 'android';
  }
}
