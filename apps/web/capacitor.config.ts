import { CapacitorConfig } from '@capacitor/cli';

const liveReload = process.env.LIVE_RELOAD === "true";

const config: CapacitorConfig = {
  appId: 'com.noot.app',
  appName: 'Noot',
  webDir: 'build',
  server: {
    androidScheme: 'https',
    url: liveReload ? 'http://localhost:3000' : undefined,
    cleartext: liveReload ? true : undefined,
  },
  backgroundColor: '#ffffff',
  ios: {
    allowsLinkPreview: false,
  }
};

export default config;
