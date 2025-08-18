# Welcome to your Expo app 👋

# Noot Nutrition App

A cross-platform React Native app built with Expo for nutrition analysis through voice recording.

## Features

- 🎙️ **Voice Recording**: Record meal descriptions using your device microphone
- 🥗 **Nutrition Analysis**: Get detailed nutrition information for your meals
- 📱 **Cross-Platform**: Works on iOS, Android, and Web
- 🔄 **Real-time Processing**: Audio is processed and analyzed instantly

## Quick Start

### Option 1: Use Development Server (Recommended)

From the project root:

```bash
script/server --development  # Starts both backend + frontend
```

### Option 2: Manual Setup

1. **Start the backend server**:
   ```bash
   cd ../..  # Go to project root
   ENV=development go run ./cmd/noot
   ```

2. **Start the Expo app**:
   ```bash
   npm install  # If not already installed
   npm run web  # For web (http://localhost:8081)
   npm run ios  # For iOS simulator
   npm run android  # For Android emulator
   ```

## Configuration

The app automatically connects to the backend API at `http://localhost:3000`.

For production or different backend URLs, set the `EXPO_PUBLIC_API_URL` environment variable:

```bash
export EXPO_PUBLIC_API_URL=https://your-backend-url.com
```

## Development

This is a standard Expo app with TypeScript. Key files:

- `app/(tabs)/index.tsx` - Main nutrition recording screen
- `app.json` - Expo configuration
- `package.json` - Dependencies and scripts

## Troubleshooting

If you encounter issues:

1. **Run diagnostics**: `script/troubleshoot` (from project root)
2. **Clear Metro cache**: `npx expo start --clear`
3. **Check ports**: Backend (3000), Frontend (8081)
4. **Permissions**: Allow microphone access when prompted

## Technical Details

- **Framework**: React Native + Expo SDK ~53.0
- **Audio**: expo-av for recording
- **Navigation**: Expo Router
- **Platform**: Web, iOS, Android support
- **Backend**: Go API server with OpenAI integration

## Get started

1. Install dependencies

   ```bash
   npm install
   ```

2. Start the app

   ```bash
   npx expo start
   ```

In the output, you'll find options to open the app in a

- [development build](https://docs.expo.dev/develop/development-builds/introduction/)
- [Android emulator](https://docs.expo.dev/workflow/android-studio-emulator/)
- [iOS simulator](https://docs.expo.dev/workflow/ios-simulator/)
- [Expo Go](https://expo.dev/go), a limited sandbox for trying out app development with Expo

You can start developing by editing the files inside the **app** directory. This project uses [file-based routing](https://docs.expo.dev/router/introduction).

## Get a fresh project

When you're ready, run:

```bash
npm run reset-project
```

This command will move the starter code to the **app-example** directory and create a blank **app** directory where you can start developing.

## Learn more

To learn more about developing your project with Expo, look at the following resources:

- [Expo documentation](https://docs.expo.dev/): Learn fundamentals, or go into advanced topics with our [guides](https://docs.expo.dev/guides).
- [Learn Expo tutorial](https://docs.expo.dev/tutorial/introduction/): Follow a step-by-step tutorial where you'll create a project that runs on Android, iOS, and the web.

## Join the community

Join our community of developers creating universal apps.

- [Expo on GitHub](https://github.com/expo/expo): View our open source platform and contribute.
- [Discord community](https://chat.expo.dev): Chat with Expo users and ask questions.
