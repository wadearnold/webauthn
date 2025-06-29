import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    host: '0.0.0.0', // Allow connections from any host
    // Configure for cross-platform WebAuthn compatibility
    hmr: {
      host: 'passkey-demo.local'
    }
  },
  // Ensure proper domain handling for WebAuthn
  define: {
    __DEV_DOMAIN__: JSON.stringify('passkey-demo.local')
  }
})