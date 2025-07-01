#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "🏗️  Building React app with ngrok configuration..."

# Load ngrok URL from environment file if it exists
if [ -f "$PROJECT_DIR/.env" ]; then
    echo "📡 Loading ngrok configuration from .env"
    source "$PROJECT_DIR/.env"
    if [ -n "$NGROK_URL" ]; then
        echo "🔗 Using ngrok tunnel: $NGROK_URL"
        export VITE_NGROK_URL="$NGROK_URL"
    else
        echo "⚠️  No NGROK_URL found in .env file"
    fi
else
    echo "⚠️  No .env file found - building for localhost mode"
fi

# Build the React app
echo "📦 Running Vite build..."
npm run build

echo "✅ Build complete!"
echo ""
if [ -n "$VITE_NGROK_URL" ]; then
    echo "🌍 Built for ngrok mode: $VITE_NGROK_URL"
    echo "📍 Access via: $VITE_NGROK_URL"
else
    echo "🏠 Built for localhost mode"
    echo "📍 Access via: http://localhost:8080"
fi