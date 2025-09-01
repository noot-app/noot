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
    "import.meta.env.VITE_TAG": JSON.stringify(process.env.VITE_TAG || 'dev'),
  },
  test: {
    include: ["src/**/*.{test,spec}.{js,ts}"],
    globals: true, /// allows to skip import of test functions like `describe`, `it`, `expect`, etc.
  },
})
