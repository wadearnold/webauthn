#!/bin/bash

echo "🔧 Setting up iOS app for your ngrok domain"
echo "==========================================="

# Check if .env exists
if [ ! -f "../.env" ]; then
    echo "❌ No .env file found!"
    echo "💡 Run ./scripts/start-ngrok.sh first to generate ngrok domain"
    exit 1
fi

# Read ngrok URL from .env
NGROK_URL=$(grep "export NGROK_URL" ../.env | cut -d'=' -f2)
if [ -z "$NGROK_URL" ]; then
    echo "❌ No NGROK_URL found in .env file"
    exit 1
fi

# Extract domain (remove https://)
DOMAIN=$(echo $NGROK_URL | sed 's/https:\/\///')

echo "📡 Found ngrok domain: $DOMAIN"

# Update ngrok-config.json
echo "📝 Updating ngrok-config.json..."
cat > PasskeyDemo/ngrok-config.json << EOF
{
  "ngrok_url": "$NGROK_URL"
}
EOF

# Update entitlements - replace any existing ngrok domain
echo "🔐 Updating entitlements..."
# First, try to find existing ngrok domain pattern and replace it
if grep -q "ngrok-free\.app\|ngrok\.io" PasskeyDemo/PasskeyDemo.entitlements; then
    # Replace existing ngrok domain
    sed -i '' "s/[a-zA-Z0-9-]*\.ngrok-free\.app/$DOMAIN/g" PasskeyDemo/PasskeyDemo.entitlements
    sed -i '' "s/[a-zA-Z0-9-]*\.ngrok\.io/$DOMAIN/g" PasskeyDemo/PasskeyDemo.entitlements
else
    echo "⚠️  No existing ngrok domain found in entitlements"
    echo "💡 You may need to manually add: webcredentials:$DOMAIN"
fi

# Get Team ID from project
TEAM_ID=$(grep "DEVELOPMENT_TEAM" PasskeyDemo.xcodeproj/project.pbxproj | head -1 | sed 's/.*= //;s/;//' | tr -d ' ')

echo ""
echo "✅ iOS app configured successfully!"
echo "📱 Ngrok URL: $NGROK_URL"
echo "🌐 Domain: $DOMAIN"
echo "👥 Team ID: $TEAM_ID"
echo ""
echo "🚀 Next steps:"
echo "1. Run: cd ../backend && ./setup-aasa.sh $TEAM_ID"
echo "2. Start backend: source ../.env && ./passkey-backend"
echo "3. Build and run iOS app on device (not simulator)"
echo ""
echo "📋 Verification:"
echo "- Check entitlements: ./verify-entitlements.sh"
echo "- Test AASA: curl $NGROK_URL/.well-known/apple-app-site-association"