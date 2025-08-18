import 'dotenv/config';
import type { ExpoConfig } from '@expo/config';

const API_URL = process.env.EXPO_PUBLIC_API_URL || 'http://localhost:3000';

const config: ExpoConfig = {
  name: 'noot-app',
  slug: 'noot-app',
  scheme: 'noot',
  platforms: ['ios', 'android', 'web'],
  web: { bundler: 'metro' },
  extra: { apiUrl: API_URL },
  experiments: { typedRoutes: true }
};
export default config;