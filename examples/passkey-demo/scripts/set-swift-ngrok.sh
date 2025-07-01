#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================="
echo "Setting ngrok URL for Swift frontend"
echo "========================================="
echo ""

# Load ngrok URL from environment file
if [ -f "$PROJECT_DIR/.env" ]; then
    source "$PROJECT_DIR/.env"
    
    if [ -n "$NGROK_URL" ]; then
        # Update Info.plist with ngrok URL for build-time injection
        PLIST_PATH="$PROJECT_DIR/frontend-swift/PasskeyDemo/Info.plist"
        
        if [ -f "$PLIST_PATH" ]; then
            # Use PlistBuddy to set the ngrok URL
            /usr/libexec/PlistBuddy -c "Add :NGROK_URL string '$NGROK_URL'" "$PLIST_PATH" 2>/dev/null || \
            /usr/libexec/PlistBuddy -c "Set :NGROK_URL '$NGROK_URL'" "$PLIST_PATH"
            
            echo "✅ Swift app configured with ngrok URL: $NGROK_URL"
            echo ""
            echo "📱 The Swift app will now use:"
            echo "   API Base URL: $NGROK_URL/api"
            echo ""
            echo "🔄 Rebuild the Swift app to apply changes"
        else
            echo "❌ Info.plist not found at: $PLIST_PATH"
            exit 1
        fi
    else
        echo "❌ NGROK_URL not found in .env file"
        exit 1
    fi
else
    echo "❌ .env file not found. Start ngrok first:"
    echo "   ./scripts/start-ngrok.sh"
    exit 1
fi

echo ""
echo "💡 Next steps:"
echo "   1. Rebuild the Swift app in Xcode"
echo "   2. The app will automatically use the ngrok tunnel"
echo ""