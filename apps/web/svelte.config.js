import adapter from "@sveltejs/adapter-cloudflare"
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte"
import { mdsvex } from "mdsvex"

const isProduction = (process.env.NODE_ENV || 'production') !== 'development'

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
    adapter: adapter(),
    csp: {
      mode: 'auto',
      directives: {
        'default-src': ['self', 'https://*.nootapp.io', 'https://uygqcgnmlzmuwkpsmixs.supabase.co'],
        'script-src': ['self', 'https://*.nootapp.io', 'https://uygqcgnmlzmuwkpsmixs.supabase.co', ...(isProduction ? [] : ['unsafe-inline', 'unsafe-eval'])],
        'style-src': ['self', 'unsafe-inline', 'https://*.nootapp.io'],
        'img-src': ['self', 'data:', 'https://*.nootapp.io', 'https://uygqcgnmlzmuwkpsmixs.supabase.co'],
        'font-src': ['self', 'https://*.nootapp.io'],
        'connect-src': ['self', 'https://*.nootapp.io', 'https://uygqcgnmlzmuwkpsmixs.supabase.co', ...(!isProduction ? ['http://localhost:*', 'http://127.0.0.1:*', 'http://192.168.1.180:*'] : [])],
        'form-action': ['self', 'https://*.nootapp.io', 'https://uygqcgnmlzmuwkpsmixs.supabase.co'],
        'frame-ancestors': ['none'],
        'base-uri': ['self'],
        'object-src': ['none'],
        ...(isProduction ? { 'upgrade-insecure-requests': true } : {})
      }
    },
    csrf: {
      trustedOrigins: ['https://uygqcgnmlzmuwkpsmixs.supabase.co']
    }
  },
}

export default config
