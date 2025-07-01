#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================="
echo "Starting ngrok tunnel for Passkey Demo"
echo "========================================="
echo ""

# Check if ngrok is installed
if ! command -v ngrok &> /dev/null; then
    echo "❌ ngrok is not installed. Installing via Homebrew..."
    if command -v brew &> /dev/null; then
        brew install ngrok/ngrok/ngrok
    else
        echo "❌ Homebrew not found. Please install ngrok manually:"
        echo "   https://ngrok.com/download"
        exit 1
    fi
fi

# Check if ngrok is authenticated
if ! ngrok config check &> /dev/null; then
    echo "⚠️  ngrok is not authenticated. Please run:"
    echo "   ngrok config add-authtoken YOUR_AUTH_TOKEN"
    echo ""
    echo "Get your auth token from: https://dashboard.ngrok.com/get-started/your-authtoken"
    echo ""
    read -p "Continue without auth token? (tunnel will be temporary) [y/N] " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Kill any existing ngrok processes
pkill -f ngrok || true

# Start ngrok tunnel
echo "🚀 Starting ngrok tunnel on port 8080..."
ngrok http 8080 --log=stdout > /tmp/ngrok.log 2>&1 &
NGROK_PID=$!

# Wait for ngrok to start
echo "⏳ Waiting for ngrok to initialize..."
sleep 3

# Get the public URL
NGROK_URL=""
for i in {1..10}; do
    if curl -s http://localhost:4040/api/tunnels &> /dev/null; then
        NGROK_URL=$(curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')
        break
    fi
    echo "   Attempt $i/10..."
    sleep 1
done

if [ -z "$NGROK_URL" ] || [ "$NGROK_URL" == "null" ]; then
    echo "❌ Failed to get ngrok URL. Check the logs:"
    cat /tmp/ngrok.log
    exit 1
fi

# Save URL to environment file
echo "export NGROK_URL=$NGROK_URL" > "$PROJECT_DIR/.env"
echo "export NGROK_PID=$NGROK_PID" >> "$PROJECT_DIR/.env"

echo ""
echo "✅ ngrok tunnel started successfully!"
echo ""
echo "🌍 Public URL: $NGROK_URL"
echo "📍 Local URL:  http://localhost:8080"
echo "🔧 ngrok UI:   http://localhost:4040"
echo ""
echo "📝 URL saved to .env file"
echo "🔄 Load environment variables:"
echo "   source .env"
echo ""
echo "🚀 Start the backend server:"
echo "   cd backend"
echo "   source ../.env && go run ."
echo ""
echo "🛑 To stop ngrok:"
echo "   ./scripts/stop-ngrok.sh"
echo ""