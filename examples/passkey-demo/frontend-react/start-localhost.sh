#!/bin/bash

echo ""
echo "🌐 WebAuthn Passkey Demo - Localhost Mode"
echo "=========================================="
echo "📍 Frontend: http://localhost:5173"
echo "📍 Backend: http://localhost:8080"
echo ""

# Check for ngrok URL in environment
if [ ! -z "$NGROK_URL" ]; then
    echo "🔗 ngrok detected: $NGROK_URL"
    echo "   Frontend will use ngrok backend API"
else
    echo "💡 For cross-platform testing with ngrok:"
    echo "   1. Run: cd .. && ./start-ngrok.sh"
    echo "   2. Run: source ../.env && npm run dev:localhost"
fi

echo ""
echo "🚀 Starting Vite on localhost..."
echo ""

# Use the simple config
vite --config vite.config.simple.js