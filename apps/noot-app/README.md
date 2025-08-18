# noot Expo App 🍎📱

React Native + Expo app that replicates the noot web UI for iOS, Android, and Web platforms using the existing Go backend.

## Features

- **Cross-platform:** Single codebase runs on iOS, Android, and Web
- **Audio Recording:** Native recording (m4a) on iOS/Android, WebM on Web
- **Same Backend:** Uses existing `/api/ingest` endpoint
- **Same UI Flow:** Big mic button → record → upload → render summary + items + transcript

## Quick Start

1. **Set up environment:**
   ```bash
   cp .env.example .env
   # Edit .env to set EXPO_PUBLIC_API_URL
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Start the noot backend in dev mode:**
   ```bash
   cd ../..  # Go to repo root
   ENV=dev go run ./cmd/noot
   ```

4. **Run the app:**
   ```bash
   # Web (http://localhost:19006)
   npm run web
   
   # iOS simulator
   npm run ios
   
   # Android emulator 
   npm run android
   ```

## Environment Configuration

### For Web Development
```bash
EXPO_PUBLIC_API_URL=http://localhost:3000
```

### For iOS/Android Development
If your noot backend is running on your development machine, you'll need your local IP:

```bash
# Find your local IP (macOS/Linux)
ifconfig | grep "inet " | grep -v 127.0.0.1

# Then set the API URL
EXPO_PUBLIC_API_URL=http://192.168.1.100:3000
```

## Development Notes

- **CORS:** Backend automatically enables CORS for localhost:19006 and 127.0.0.1:19006 when running in development mode (`ENV=dev` or `DEBUG=true`)
- **Audio Permissions:** App will request microphone permissions on first recording attempt
- **File Formats:** Uploads m4a (iOS/Android) or webm (Web) to the backend, which handles both formats

## Troubleshooting

**CORS Errors:**
- Make sure backend is running with `ENV=dev` or `DEBUG=true`
- Check that EXPO_PUBLIC_API_URL matches your backend address

**Recording Issues:**
- Grant microphone permissions when prompted
- On Web: Ensure you're using HTTPS or localhost (required for getUserMedia)
- On iOS: Check iOS Simulator → Device → Settings → Privacy → Microphone

**Network Issues on Mobile:**
- Ensure your device and computer are on the same WiFi network
- Use your computer's IP address, not localhost, for EXPO_PUBLIC_API_URL
- Test the API URL in your mobile browser first to ensure connectivity