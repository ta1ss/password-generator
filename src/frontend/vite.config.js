import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist', // Specify the directory where built files will be placed
    assetsDir: 'static', // Specify the assets directory within 'dist'
  },
})
