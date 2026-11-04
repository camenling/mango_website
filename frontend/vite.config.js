import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@wxperia/liquid-glass-vue': resolve(__dirname, 'src/lib/liquid-glass-vue.js'),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5175,
    strictPort: true,
    allowedHosts: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8199',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: '../backend/resource/public',
    emptyOutDir: true,
  },
})
