import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
  ],
  server: {
    host: true, // Dockerコンテナ外からのアクセスを許可
    port: 5173,
    // バックエンドへのプロキシ設定をしておくとCORSで悩みません
    proxy: {
      '/api': 'http://localhost:8080'
    }
  }
})
