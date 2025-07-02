#!/bin/bash

echo "🍎 Setting up Apple App Site Association (AASA)"
echo "=============================================="

TEAM_ID=$1

if [ -z "$TEAM_ID" ]; then
    echo "❌ Team ID required!"
    echo ""
    echo "Usage: ./setup-aasa.sh YOUR_TEAM_ID"
    echo ""
    echo "💡 Find your Team ID:"
    echo "1. Run: cd ../frontend-swift && ./setup-domain.sh"
    echo "2. Or check Xcode: PasskeyDemo target → Signing & Capabilities"
    echo "3. Look for Team ID (e.g., 927BS6LD3W)"
    exit 1
fi

# Validate Team ID format (alphanumeric, 10 characters)
if [[ ! "$TEAM_ID" =~ ^[A-Z0-9]{10}$ ]]; then
    echo "⚠️  Warning: Team ID format looks unusual: $TEAM_ID"
    echo "💡 Expected format: 10 characters, letters and numbers (e.g., 927BS6LD3W)"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Create directory
mkdir -p static/.well-known

# Create AASA file
cat > static/.well-known/apple-app-site-association << EOF
{
  "webcredentials": {
    "apps": ["$TEAM_ID.com.passkey.demo.ios"]
  }
}
EOF

echo "✅ AASA file created successfully!"
echo "📁 Location: static/.well-known/apple-app-site-association"
echo "🔗 App identifier: $TEAM_ID.com.passkey.demo.ios"
echo ""

# Check if ngrok is running
if [ -f "../.env" ]; then
    NGROK_URL=$(grep "export NGROK_URL" ../.env | cut -d'=' -f2)
    if [ ! -z "$NGROK_URL" ]; then
        echo "🌐 Your ngrok domain: $NGROK_URL"
        echo ""
        echo "🚀 Next steps:"
        echo "1. Start backend: source ../.env && ./passkey-backend"
        echo "2. Test AASA: curl $NGROK_URL/.well-known/apple-app-site-association"
        echo "3. Build iOS app on device and test passkey registration"
        echo ""
        echo "🔍 Expected AASA response:"
        cat static/.well-known/apple-app-site-association
    fi
else
    echo "⚠️  No .env file found. Run ./scripts/start-ngrok.sh first."
fi