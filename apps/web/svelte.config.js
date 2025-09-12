import cloudflareAdapter from "@sveltejs/adapter-cloudflare"
import staticAdapter from "@sveltejs/adapter-static"
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte"
import { mdsvex } from "mdsvex"

// Determine which adapter to use based on environment variable
const useStatic = process.env.ADAPTER === "static"

console.log(`Using ${useStatic ? "static" : "cloudflare"} adapter`)

/** @type {import('@sveltejs/kit').Config} */
const config = {
  extensions: [".svelte", ".md"],
  // Consult https://kit.svelte.dev/docs/integrations#preprocessors
  // for more information about preprocessors
  preprocess: [
    vitePreprocess(),
    mdsvex({
      extensions: [".md"],
    }),
  ],

  kit: {
    // Use static adapter for Capacitor builds, Cloudflare adapter for web
    adapter: useStatic
      ? staticAdapter({
          pages: "dist",
          assets: "dist",
          fallback: "index.html", // Enable SPA mode for Capacitor
          precompress: false,
          strict: false, // Allow dynamic routes to be handled client-side
        })
      : cloudflareAdapter(),
  },
}

export default config
