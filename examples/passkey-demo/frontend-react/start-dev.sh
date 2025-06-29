#!/bin/bash

# Cross-Platform WebAuthn Passkey Demo - Development Server
echo ""
echo "🌐 Cross-Platform WebAuthn Passkey Demo"
echo "========================================"

# Check for HTTPS certificates
CERT_FILE="../certs/passkey-demo.local+4.pem"
if [ -f "$CERT_FILE" ]; then
    echo "🔒 HTTPS Mode: ENABLED"
    echo "🔐 React Frontend: https://passkey-demo.local:5173"
    echo "📡 Backend API: https://passkey-demo.local:8080"
    echo "✅ Full WebAuthn functionality available"
else
    echo "🔓 HTTP Mode: Active (HTTPS certificates not found)"
    echo "🔐 React Frontend: http://passkey-demo.local:5173 (⚠️  WebAuthn limited)"
    echo "📡 Backend API: http://passkey-demo.local:8080"
    echo "💡 Run '../setup-https.sh' to enable HTTPS for full WebAuthn support"
    echo "🔄 Fallback: http://localhost:5173 (WebAuthn works on localhost)"
fi

echo ""
echo "⚠️  SETUP REQUIRED:"
echo "   Add to /etc/hosts (requires sudo):"
echo "   127.0.0.1 passkey-demo.local"
echo ""
echo "🔗 For detailed setup: See README.md or HTTPS-SETUP.md"
echo "🚀 Starting Vite development server..."
echo ""

# Start Vite
npm run dev:direct