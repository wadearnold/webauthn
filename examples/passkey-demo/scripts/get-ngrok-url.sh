#!/bin/bash

echo "========================================="
echo "Getting current ngrok tunnel URL"
echo "========================================="
echo ""

# Check if ngrok is running
if ! pgrep -f ngrok > /dev/null; then
    echo "❌ ngrok is not running"
    echo "   Start ngrok: ./scripts/start-ngrok.sh"
    exit 1
fi

# Get URL from ngrok API
if curl -s http://localhost:4040/api/tunnels &> /dev/null; then
    NGROK_URL=$(curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')
    
    if [ "$NGROK_URL" != "null" ] && [ -n "$NGROK_URL" ]; then
        echo "🌍 Current ngrok URL: $NGROK_URL"
        echo ""
        echo "🔄 Export environment variable:"
        echo "   export NGROK_URL=$NGROK_URL"
        echo ""
        echo "📋 Copy this URL for frontend configuration:"
        echo "   $NGROK_URL"
    else
        echo "❌ Failed to get ngrok URL"
        exit 1
    fi
else
    echo "❌ ngrok API not accessible"
    echo "   Make sure ngrok is running on port 4040"
    exit 1
fi