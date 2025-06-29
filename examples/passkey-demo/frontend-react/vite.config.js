import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import fs from 'fs'
import path from 'path'

// Check for HTTPS certificates
const certPath = path.resolve('../certs')
const certFile = path.join(certPath, 'passkey-demo.local+4.pem')
const keyFile = path.join(certPath, 'passkey-demo.local+4-key.pem')

const httpsConfig = fs.existsSync(certFile) && fs.existsSync(keyFile) ? {
  key: fs.readFileSync(keyFile),
  cert: fs.readFileSync(certFile),
} : undefined

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    host: '0.0.0.0', // Allow connections from any host
    // HTTPS configuration (if certificates are available)
    https: httpsConfig,
    // Configure for cross-platform WebAuthn compatibility
    hmr: {
      host: 'passkey-demo.local',
      protocol: httpsConfig ? 'wss' : 'ws'
    }
  },
  // Ensure proper domain handling for WebAuthn
  define: {
    __DEV_DOMAIN__: JSON.stringify('passkey-demo.local'),
    __HTTPS_ENABLED__: JSON.stringify(!!httpsConfig)
  }
})