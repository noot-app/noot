import { sveltekit } from "@sveltejs/kit/vite"
import { defineConfig } from "vitest/config"
import { buildAndCacheSearchIndex } from "./src/lib/build_index"

export default defineConfig({
  plugins: [
    sveltekit(),
    {
      name: "vite-build-search-index",
      writeBundle: {
        order: "post",
        sequential: false,
        handler: async () => {
          console.log("Building search index...")
          await buildAndCacheSearchIndex()
        },
      },
    },
  ],
  define: {
    // Inject build-time environment variables for version info
    "import.meta.env.VITE_COMMIT_SHA": JSON.stringify(process.env.VITE_COMMIT_SHA || 'unknown'),
    "import.meta.env.VITE_BUILD_TIME": JSON.stringify(process.env.VITE_BUILD_TIME || 'unknown'),
    "import.meta.env.VITE_TAG": JSON.stringify(process.env.VITE_TAG || process.env.VITE_COMMIT_SHA || 'unknown'),
  },
  test: {
    include: ["src/**/*.{test,spec}.{js,ts}"],
    globals: true, /// allows to skip import of test functions like `describe`, `it`, `expect`, etc.
    environment: 'jsdom', // Enable DOM support for Svelte component testing
    setupFiles: ['src/test-setup.ts'], // Add test setup file
    reporters: [['default', { summary: false }]], // Use clean reporter without summary for less noise
    logHeapUsage: false, // Disable heap usage logging
    onConsoleLog: () => false, // Suppress console logs in tests
  },
})
