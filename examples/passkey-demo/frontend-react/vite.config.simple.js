import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Simple configuration for localhost/ngrok development
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    host: true, // Listen on all addresses
    https: false, // Use HTTP for simplicity with ngrok
  },
  // Pass ngrok URL from environment if available
  define: {
    'import.meta.env.VITE_NGROK_URL': JSON.stringify(process.env.NGROK_URL || '')
  }
})