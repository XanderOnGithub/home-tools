import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'
import { fileURLToPath } from 'node:url'

// In dev, Vite serves the UI (:5175, so fitness and games can run on
// :5173 and :5174 at the same time) and forwards API calls to the Go
// server, started with `make run TOOL=discord`. One origin for the
// browser: no CORS.
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
