import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'com.noot.app',
  appName: 'Noot',
  webDir: 'build', // matches SvelteKit static output directory
  bundledWebRuntime: false
};

export default config;
