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

# Create the ngrok config file with proper formatting
CONFIG_FILE="$(dirname "${BASH_SOURCE[0]}")/PasskeyDemo/ngrok-config.json"
cat > "$CONFIG_FILE" << EOF
{
  "ngrok_url": "$NGROK_URL"
}
EOF

echo "✅ Created ngrok config file: $CONFIG_FILE"
echo "✅ iOS app configured for cross-platform mode"
echo "🏃‍♂️ Now build and run the iOS app to use the ngrok tunnel"
echo ""
echo "Note: If running in Xcode, you may need to clean and rebuild the project"
echo "      for the new configuration to take effect."