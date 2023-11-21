import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import commonjs from '@rollup/plugin-commonjs';
// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), commonjs({
    // Exclude ES modules from being transformed to CommonJS
    exclude: 'node_modules/**'
  })],
  // If you need to resolve .js files as well
  resolve: {
    extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json']
  },
  build: {
    outDir: 'dist', // Specify the directory where built files will be placed
    assetsDir: 'static', // Specify the assets directory within 'dist'
  },
})
