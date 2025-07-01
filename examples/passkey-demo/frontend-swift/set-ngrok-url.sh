#!/bin/bash

# Script to configure the iOS app with the current ngrok URL
# This should be run after starting the ngrok tunnel

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$PROJECT_DIR/.env"

if [ ! -f "$ENV_FILE" ]; then
    echo "❌ .env file not found at $ENV_FILE"
    echo "Please run the start-ngrok.sh script first to create the tunnel"
    exit 1
fi

# Source the environment file to get the ngrok URL
source "$ENV_FILE"

if [ -z "$NGROK_URL" ]; then
    echo "❌ NGROK_URL not found in .env file"
    echo "Please run the start-ngrok.sh script first to create the tunnel"
    exit 1
fi

echo "🔧 Configuring iOS app with ngrok URL: $NGROK_URL"

# Set the ngrok URL in iOS app using defaults for both device and simulator
# This will be picked up by the APIConfiguration.ngrokURL

# For device/production bundle ID
defaults write com.passkey.demo.ios NGROK_URL "$NGROK_URL"

# For common simulator bundle IDs (Xcode may use different bundle IDs in simulator)
defaults write FAKETEAMID.com.passkey.demo.ios NGROK_URL "$NGROK_URL"

# Also set in global defaults as fallback
defaults write -g NGROK_URL "$NGROK_URL"

echo "✅ iOS app configured with ngrok URL for device and simulator"
echo "🏃‍♂️ Now build and run the iOS app to use the ngrok tunnel"
echo ""
echo "Note: If running in Xcode, you may need to clean and rebuild the project"
echo "      for the new configuration to take effect."