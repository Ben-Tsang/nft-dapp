import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {   // 主要是这个
    alias: {
      '@': '/src', // 将 @ 映射到 /src 目录
    },
  },
  // 让外部访问
  server: {
    host: '0.0.0.0',
    port: 5173
  }
})
