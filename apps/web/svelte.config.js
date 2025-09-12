import cloudflareAdapter from "@sveltejs/adapter-cloudflare"
import staticAdapter from "@sveltejs/adapter-static"
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte"
import { mdsvex } from "mdsvex"

// Determine which adapter to use based on environment variable
const isCapacitor = process.env.CAPACITOR === "true"

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
    adapter: isCapacitor 
      ? staticAdapter({
          pages: 'build',
          assets: 'build',
          fallback: 'index.html',
          precompress: false,
          strict: true
        })
      : cloudflareAdapter(),
  },
}

export default config
