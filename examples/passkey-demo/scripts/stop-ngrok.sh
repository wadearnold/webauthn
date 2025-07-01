#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================="
echo "Stopping ngrok tunnel"
echo "========================================="
echo ""

# Kill ngrok processes
if pkill -f ngrok; then
    echo "✅ ngrok processes terminated"
else
    echo "ℹ️  No ngrok processes found"
fi

# Clean up environment file
if [ -f "$PROJECT_DIR/.env" ]; then
    rm "$PROJECT_DIR/.env"
    echo "✅ Environment file cleaned up"
fi

# Clean up log file
if [ -f "/tmp/ngrok.log" ]; then
    rm "/tmp/ngrok.log"
    echo "✅ Log file cleaned up"
fi

echo ""
echo "🛑 ngrok tunnel stopped"
echo ""