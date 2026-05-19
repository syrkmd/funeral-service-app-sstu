import { defineConfig } from 'vite'
import path from 'path'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },

  assetsInclude: ['**/*.svg', '**/*.csv'],
  server: {
    proxy: {
      '/api': 'http://localhost:8090',
      '/metrics': 'http://localhost:8090',
      '/swagger': 'http://localhost:8090',
      '/healthz': 'http://localhost:8090',
    },
  },
})
