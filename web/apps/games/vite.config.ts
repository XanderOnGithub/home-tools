import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'
import { fileURLToPath } from 'node:url'

// In dev, Vite serves the UI (:5174, so fitness can run on :5173 at the
// same time) and forwards API calls to the Go server, started with
// `make run TOOL=games`. One origin for the browser: no CORS.
export default defineConfig({
  plugins: [svelte()],
  // `@/...` = src/..., mirrored in tsconfig.app.json "paths".
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    proxy: {
      '/api': process.env.API_URL ?? 'http://localhost:8080',
    },
  },
})
