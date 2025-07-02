#!/bin/bash

echo "🔍 Checking iOS App Configuration..."
echo ""

CONFIG_FILE="PasskeyDemo/ngrok-config.json"
TEMPLATE_FILE="PasskeyDemo/ngrok-config.json.template"

# Check if config file exists
if [ -f "$CONFIG_FILE" ]; then
    echo "✅ Config file found: $CONFIG_FILE"
    
    # Check if it contains a valid ngrok URL
    if grep -q "https://.*\.ngrok-free\.app" "$CONFIG_FILE"; then
        NGROK_URL=$(grep -o "https://[^\"]*\.ngrok-free\.app" "$CONFIG_FILE")
        echo "✅ Valid ngrok URL found: $NGROK_URL"
        echo "🌐 iOS app will use cross-platform mode"
    elif grep -q "your-ngrok-url" "$CONFIG_FILE"; then
        echo "⚠️  Config file contains placeholder URL"
        echo "💡 Edit $CONFIG_FILE with your actual ngrok URL"
    else
        echo "⚠️  No valid ngrok URL found in config file"
        echo "💡 Edit $CONFIG_FILE with your ngrok URL"
    fi
else
    echo "❌ Config file not found: $CONFIG_FILE"
    
    if [ -f "$TEMPLATE_FILE" ]; then
        echo "💡 Template found. Run: cp $TEMPLATE_FILE $CONFIG_FILE"
        echo "   Then edit $CONFIG_FILE with your ngrok URL"
    else
        echo "❌ Template file also missing: $TEMPLATE_FILE"
        echo "💡 Create $CONFIG_FILE with:"
        echo '   {"ngrok_url": "https://your-ngrok-url.ngrok-free.app"}'
    fi
    echo "🏠 iOS app will use localhost mode"
fi

echo ""

# Check if ngrok is running and show current URL
if command -v ngrok &> /dev/null; then
    echo "🔍 Checking for active ngrok tunnel..."
    NGROK_URL=$(curl -s http://localhost:4040/api/tunnels 2>/dev/null | grep -o 'https://[^"]*\.ngrok-free\.app' | head -1)
    
    if [ ! -z "$NGROK_URL" ]; then
        echo "✅ Active ngrok tunnel found: $NGROK_URL"
        
        if [ -f "$CONFIG_FILE" ]; then
            if grep -q "$NGROK_URL" "$CONFIG_FILE"; then
                echo "✅ Config file matches active ngrok tunnel"
            else
                echo "⚠️  Config file does not match active ngrok tunnel"
                echo "💡 Update $CONFIG_FILE with current URL: $NGROK_URL"
            fi
        fi
    else
        echo "❌ No active ngrok tunnel found"
        echo "💡 Start ngrok with: cd .. && ./scripts/start-ngrok.sh"
    fi
else
    echo "ℹ️  ngrok not found in PATH (this is optional)"
fi

echo ""
echo "🎯 Configuration Summary:"
echo "- For cross-platform passkey sharing: Ensure $CONFIG_FILE contains your ngrok URL"
echo "- For local development only: Leave $CONFIG_FILE missing or with placeholder URL"
echo "- After making changes: Clean and rebuild in Xcode"