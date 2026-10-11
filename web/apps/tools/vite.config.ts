import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'
import { fileURLToPath } from 'node:url'

// In dev, Vite serves the page on :5176 (fitness, games and discord use
// :5173–:5175). The page has no API, so nothing is proxied.
export default defineConfig({
  plugins: [svelte()],
  // `@/...` = src/..., mirrored in tsconfig.app.json "paths".
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
})
