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

// Determine cleartext setting
const isProduction = process.env.NODE_ENV === 'production';
const allowCleartext = !isProduction;
const isHttpsUrl = serverUrl.startsWith('https://');

// Security validation - fail fast in production with insecure config
if (isProduction && allowCleartext) {
  console.error('🚨 CRITICAL SECURITY ERROR: NODE_ENV=production but cleartext is ENABLED!');
  console.error('   This is a MAJOR security risk - HTTP traffic allowed in production!');
  console.error('   Set NODE_ENV=development for local dev or use HTTPS URLs for production.');
  process.exit(1);
}

if (isProduction && !isHttpsUrl) {
  console.error('🚨 CRITICAL SECURITY ERROR: NODE_ENV=production but server URL is NOT HTTPS!');
  console.error(`   Current URL: ${serverUrl}`);
  console.error('   Production builds should ONLY use HTTPS URLs for security.');
  process.exit(1);
}

const config: CapacitorConfig = {
  appId: 'io.nootapp',
  appName: 'Noot',
  webDir: 'www',
  server: {
    url: serverUrl,
    cleartext: allowCleartext,
    allowNavigation: ['*']
  }
};

export default config;
