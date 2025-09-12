# Mobile Development with Capacitor

This document explains how to develop and build the Noot mobile app using Capacitor.

## Overview

The Noot SvelteKit web application can be built for both web deployment (Cloudflare Pages with SSR) and mobile deployment (iOS/Android via Capacitor) using a dual adapter setup.

## Architecture

- **Web Build**: Uses `@sveltejs/adapter-cloudflare` for SSR deployment to Cloudflare Pages
- **Mobile Build**: Uses `@sveltejs/adapter-static` to create a SPA bundle that Capacitor packages into native mobile apps

## Build Commands

### Web Development & Deployment
```bash
# Regular development (web)
npm run dev

# Build for web/Cloudflare (SSR)
npm run build

# Deploy to Cloudflare
npm run deploy
```

### Mobile Development & Deployment
```bash
# Build for mobile (static SPA)
npm run build:mobile

# iOS commands
npm run ios:build    # Build static bundle and sync to iOS
npm run ios:run      # Build, sync, and run in iOS simulator
npm run ios:dev      # Live reload development for iOS

# Android commands
npm run android:build    # Build static bundle and sync to Android
npm run android:run      # Build, sync, and run in Android emulator
npm run android:dev      # Live reload development for Android
```

## Environment Variables

The mobile build requires the same environment variables as the web build:

- `PUBLIC_SUPABASE_URL` - Supabase project URL
- `PUBLIC_SUPABASE_ANON_KEY` - Supabase anonymous key
- `PUBLIC_API_BASE_URL` - Backend API base URL

## Adapter Selection

The adapter is automatically selected based on the `CAPACITOR` environment variable:

- `CAPACITOR=true` → Uses static adapter for mobile builds
- `CAPACITOR` not set → Uses Cloudflare adapter for web builds

## Key Differences Between Web and Mobile Builds

### SSR vs SPA
- **Web**: Server-side rendering with `hooks.server.ts` and `+layout.server.ts`
- **Mobile**: Client-side only, server hooks are ignored

### Authentication
- **Web**: Uses Supabase SSR with cookies handled server-side
- **Mobile**: Uses standard Supabase client with localStorage/session storage

### API Calls
- **Web**: Can make server-side API calls during SSR
- **Mobile**: All API calls happen client-side

## Development Workflow

1. **Initial Setup**: Run `npm run bootstrap` to install dependencies
2. **Test Web Build**: Run `npm run build && npm run test_run` to verify web functionality
3. **Test Mobile Build**: Run `npm run build:mobile` to verify static build works
4. **Sync Capacitor**: Run `npx cap sync` to copy built files to native projects
5. **Run on Device**: Use `npm run ios:run` or `npm run android:run`

## Platform Requirements

### iOS Development
- macOS with Xcode installed
- CocoaPods installed (`gem install cocoapods`)
- iOS Simulator or physical iOS device

### Android Development
- Android Studio installed
- Android SDK configured
- Android emulator or physical Android device

## File Structure

```
apps/web/
├── capacitor.config.ts     # Capacitor configuration
├── ios/                    # iOS native project
├── android/                # Android native project
├── build/                  # Static build output (mobile)
└── .svelte-kit/output/     # SSR build output (web)
```

## Troubleshooting

### Build Issues
- Ensure all environment variables are set
- Check that both adapters are installed: `@sveltejs/adapter-cloudflare` and `@sveltejs/adapter-static`
- Verify the `CAPACITOR` environment variable is set correctly

### Mobile-Specific Issues
- Server hooks (`hooks.server.ts`) are ignored in mobile builds
- Ensure all API calls work client-side without server-side context
- Check that authentication flows work without SSR

### Platform Issues
- **iOS**: Ensure Xcode and CocoaPods are installed
- **Android**: Ensure Android Studio and SDK are configured
- Run `npx cap doctor` to check platform setup

## References

- [Capacitor Documentation](https://capacitorjs.com/docs)
- [SvelteKit Adapters](https://kit.svelte.dev/docs/adapters)
- [Supabase Auth for Mobile](https://supabase.com/docs/guides/auth/auth-clients)