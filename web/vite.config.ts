import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      // Any request starting with /api will go to your Go backend
      '/api': {
        target: 'http://localhost:8080', // <-- Change this to your Go Port!
        changeOrigin: true,
        cookieDomainRewrite: "localhost"
      },
    },
  },
})
