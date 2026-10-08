import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'
import { fileURLToPath } from 'node:url'

// In dev, Vite serves the UI and forwards API + photo requests to the Go
// server (`make run`, :8080), so the browser sees one origin: no CORS.
export default defineConfig({
  plugins: [svelte()],
  // `@/...` = src/..., so imports don't turn into ../../../ chains.
  // Mirrored in tsconfig.app.json "paths".
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/images': 'http://localhost:8080',
    },
  },
})
