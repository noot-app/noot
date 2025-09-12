import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'io.nootapp',
  appName: 'Noot',
  webDir: 'www',
  server: {
    url: 'http://192.168.1.180:3000', // Your Mac's IP address
    cleartext: true, // Allow HTTP for local dev
    allowNavigation: ['*']
  }
};

export default config;
