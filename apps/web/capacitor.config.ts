import type { CapacitorConfig } from '@capacitor/cli';
import { networkInterfaces } from 'os';

// Get the local IP address dynamically
function getLocalIP(): string {
  const nets = networkInterfaces();
  for (const name of Object.keys(nets)) {
    const netInfo = nets[name];
    if (netInfo) {
      for (const net of netInfo) {
        // Skip over non-IPv4 and internal addresses
        if (net.family === 'IPv4' && !net.internal) {
          console.log(`🌐 Capacitor detected local IP: ${net.address}`);
          return net.address;
        }
      }
    }
  }
  console.log('🌐 Capacitor using fallback: localhost');
  return 'localhost'; // fallback
}

// Determine the server URL
const serverUrl = process.env.CAPACITOR_SERVER_URL || `http://${getLocalIP()}:3000`;
console.log(`🚀 Capacitor server URL: ${serverUrl}`);

const config: CapacitorConfig = {
  appId: 'io.nootapp',
  appName: 'Noot',
  webDir: 'www',
  server: {
    url: serverUrl,
    cleartext: process.env.NODE_ENV !== 'production', // HTTP for dev, HTTPS for production
    allowNavigation: ['*']
  }
};

export default config;
