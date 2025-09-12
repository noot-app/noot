import adapterCloudflare from "@sveltejs/adapter-cloudflare"
import adapterStatic from "@sveltejs/adapter-static"
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte"
import { mdsvex } from "mdsvex"

// Select adapter based on environment variable
const mode = process.env.ADAPTER || 'cloudflare'
const adapters = {
  cloudflare: adapterCloudflare(),
  static: adapterStatic({ 
    fallback: 'index.html' // SPA fallback for client-side routing
  })
}

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
    // Conditional adapter selection for SSR (Cloudflare) vs SPA (Capacitor)
    adapter: adapters[mode],
  },
}

export default config
