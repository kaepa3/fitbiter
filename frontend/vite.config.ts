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
    hmr: {
      host: 'fitbit.local',
      protocol: 'wss', // HTTPS経由なので wss を指定
      clientPort: 443, // ブラウザからはNPMの443番(HTTPS)経由で見ているため
    },
    // バックエンドへのプロキシ設定をしておくとCORSで悩みません
    proxy: {
      '/api': {
        target: 'http://backend:8080',
        changeOrigin: true,
      }
    },
    allowedHosts: [
      'fitbit.local',
      'api-fitbit.local' // 一応バックエンド側の名前も入れておくと安心
    ],
  }
})
