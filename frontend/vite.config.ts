import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    host: true, // Dockerコンテナ外からのアクセスを許可
    port: 5173,
    // バックエンドへのプロキシ設定をしておくとCORSで悩みません
    proxy: {
      '/api': 'http://localhost:8080'
    }
  }
})
