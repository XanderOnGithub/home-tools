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
    // API_URL lets a second server (e.g. test data on another port) be used
    // without touching the one `make run` starts.
    proxy: {
      '/api': process.env.API_URL ?? 'http://localhost:8080',
      '/images': process.env.API_URL ?? 'http://localhost:8080',
    },
  },
})
