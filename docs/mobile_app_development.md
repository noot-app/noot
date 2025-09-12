# Mobile App Development with Capacitor

This document outlines how to build and deploy the Noot SvelteKit web application as a mobile app (Android/iOS) using Capacitor.

## Overview

The Noot web application supports dual build modes:
- **SSR Mode**: Server-side rendering for Cloudflare Pages deployment (default)
- **Static Mode**: Static SPA build for mobile app packaging with Capacitor

## Prerequisites

### For Android Development
- [Android Studio](https://developer.android.com/studio)
- Android SDK (API level 21 or higher)
- Java Development Kit (JDK 8 or higher)

### For iOS Development (macOS only)
- [Xcode](https://developer.apple.com/xcode/) 12.0 or higher
- iOS Simulator or physical iOS device
- Apple Developer Account (for device testing and App Store deployment)

## Build Commands

### Web Build (SSR - Default)
```bash
cd apps/web
npm run build  # Uses Cloudflare adapter for SSR
```

### Static Build (Mobile)
```bash
cd apps/web
npm run build:static  # Uses static adapter for SPA
```

### Mobile Build & Copy
```bash
cd apps/web
npm run build:mobile  # Builds static and copies to mobile platforms
```

### Platform-Specific Commands
```bash
cd apps/web
npm run mobile:android  # Build and open Android Studio
npm run mobile:ios      # Build and open Xcode (macOS only)
```

## Development Workflow

### 1. Initial Setup
The Capacitor project is already initialized with:
- App ID: `com.noot.app`
- App Name: `Noot`
- Web Directory: `build` (matches SvelteKit static output)

### 2. Development Cycle
1. Make changes to your SvelteKit app
2. Test in the browser: `npm run dev`
3. Build for mobile: `npm run build:mobile`
4. Test on device/simulator using Android Studio or Xcode

### 3. Testing Mobile Features
The mobile app includes all web functionality with native device access:
- Camera/microphone for voice recording
- File system access
- Push notifications (when implemented)
- Native UI components

## Architecture Considerations

### SSR vs Static Compatibility

#### Server Hooks (`hooks.server.ts`)
- **SSR Mode**: Fully functional with Supabase SSR authentication
- **Static Mode**: Server hooks are ignored; client-side auth is used instead

#### Layout Server Load (`+layout.server.ts`)
- **SSR Mode**: Provides server-side session data
- **Static Mode**: Returns default values; client handles auth state

#### API Client
The API client (`src/lib/api/client.ts`) works in both modes:
- Uses environment variables for endpoint configuration
- Handles authentication tokens client-side
- Falls back to production API when env vars are missing

#### Supabase Integration
- **SSR Mode**: Uses `@supabase/ssr` for server-client cookie handling
- **Static Mode**: Uses `@supabase/supabase-js` browser client
- Authentication state persists in localStorage/cookies

## Environment Variables

The mobile app uses the same environment variables as the web app:
- `PUBLIC_API_BASE_URL`: Backend API endpoint
- `PUBLIC_SUPABASE_URL`: Supabase project URL
- `PUBLIC_SUPABASE_ANON_KEY`: Supabase anonymous key

Set these in your `.env` file for local development, or configure them in your mobile app build process.

## Platform-Specific Notes

### Android
- Minimum SDK: API level 21 (Android 5.0)
- Target SDK: Latest stable Android version
- Permissions: Camera, microphone, internet access

### iOS
- Minimum iOS version: 13.0
- Required capabilities: Camera, microphone, internet access
- App Store compliance: Follows iOS privacy guidelines

## Native Plugin Integration

Capacitor provides access to native device APIs through plugins:

### Core Plugins (Pre-installed)
- `@capacitor/camera`: Camera and photo library access
- `@capacitor/filesystem`: File system operations
- `@capacitor/device`: Device information
- `@capacitor/app`: App lifecycle events

### Adding Additional Plugins
```bash
cd apps/web
npm install @capacitor/[plugin-name]
npx cap sync
```

## Troubleshooting

### Build Issues
- **Static build fails**: Check for SSR-specific code in components
- **Missing dependencies**: Run `npm install` in the web directory
- **Capacitor sync fails**: Ensure platforms are added with `npx cap add android` or `npx cap add ios`

### Runtime Issues
- **API calls fail**: Verify environment variables are set correctly
- **Auth not working**: Check that Supabase client configuration is correct for browser environment
- **Routes not loading**: Ensure SPA fallback is configured (`fallback: 'index.html'`)

### Platform Issues
- **Android Studio won't open**: Check that Android SDK is properly installed
- **iOS build fails**: Verify Xcode command line tools are installed
- **App crashes on device**: Check device logs for JavaScript errors

## Deployment

### Android
1. Build signed APK in Android Studio
2. Upload to Google Play Console
3. Follow Google Play Store guidelines

### iOS
1. Archive build in Xcode
2. Upload to App Store Connect
3. Submit for App Store review

## Performance Optimization

### For Mobile Apps
- Enable tree-shaking in build process
- Optimize images and assets for mobile devices
- Use lazy loading for routes and components
- Consider offline functionality with service workers

### Bundle Size Analysis
```bash
cd apps/web
ADAPTER=static npm run build -- --analyze
```

## Security Considerations

- API keys and secrets are handled client-side in mobile apps
- Use Supabase Row Level Security (RLS) for data protection
- Implement proper authentication token refresh
- Follow platform-specific security guidelines

## Known Limitations

1. **Server-side features**: SSR-specific functionality is not available in mobile builds
2. **File uploads**: May require native plugin for advanced file handling
3. **Background processing**: Limited compared to native app capabilities
4. **Performance**: Web-based apps may be slower than fully native apps

## Future Enhancements

- Push notification support
- Offline mode with data synchronization
- Native UI components for better performance
- Biometric authentication integration
- Advanced camera features for meal photo capture