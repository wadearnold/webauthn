#!/bin/bash

# Cross-Platform WebAuthn Passkey Demo - Development Server
echo ""
echo "🌐 Cross-Platform WebAuthn Passkey Demo"
echo "========================================"
echo "🔐 React Frontend: http://passkey-demo.local:5173"
echo "📡 Backend API: http://passkey-demo.local:8080"
echo ""
echo "⚠️  CRITICAL SETUP REQUIRED:"
echo "   Add to /etc/hosts (requires sudo):"
echo "   127.0.0.1 passkey-demo.local"
echo ""
echo "🔗 For detailed setup: See README.md"
echo "🚀 Starting Vite development server..."
echo "   Note: Vite shows localhost URLs, but use passkey-demo.local instead"
echo ""

# Start Vite
npm run dev:direct