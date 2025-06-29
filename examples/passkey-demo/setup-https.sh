#!/bin/bash

# WebAuthn Passkey Demo - HTTPS Setup Script
# This script installs mkcert and generates local HTTPS certificates for development

set -e

echo "🔐 WebAuthn Passkey Demo - HTTPS Setup"
echo "======================================"
echo

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ This script is designed for macOS. For other platforms, see:"
    echo "   Linux: https://github.com/FiloSottile/mkcert#linux"
    echo "   Windows: https://github.com/FiloSottile/mkcert#windows"
    exit 1
fi

# Check if Homebrew is installed
if ! command -v brew &> /dev/null; then
    echo "❌ Homebrew is required but not installed."
    echo "   Install from: https://brew.sh"
    exit 1
fi

echo "📋 Prerequisites check:"
echo "✅ macOS detected"
echo "✅ Homebrew available"
echo

# Install mkcert if not already installed
if ! command -v mkcert &> /dev/null; then
    echo "📦 Installing mkcert..."
    brew install mkcert
    echo "✅ mkcert installed"
else
    echo "✅ mkcert already installed"
fi

# Install the CA certificate if not already done
echo "🔐 Setting up Certificate Authority..."
mkcert -install
echo "✅ CA certificate installed in system trust store"
echo

# Create certificates directory
CERT_DIR="$(dirname "$0")/certs"
mkdir -p "$CERT_DIR"

echo "📜 Generating HTTPS certificates..."
cd "$CERT_DIR"

# Generate certificates for passkey-demo.local
mkcert \
    "passkey-demo.local" \
    "api.passkey-demo.local" \
    "localhost" \
    "127.0.0.1" \
    "::1"

echo "✅ Certificates generated in: $CERT_DIR"
echo

# List generated files
echo "📋 Generated certificate files:"
ls -la "$CERT_DIR"
echo

# Update hosts file if needed
echo "🌐 Checking /etc/hosts configuration..."
if ! grep -q "passkey-demo.local" /etc/hosts; then
    echo "⚠️  Domain not found in /etc/hosts. Adding now..."
    echo "# WebAuthn Passkey Demo - Cross-Platform Configuration" | sudo tee -a /etc/hosts
    echo "127.0.0.1 passkey-demo.local" | sudo tee -a /etc/hosts
    echo "127.0.0.1 api.passkey-demo.local" | sudo tee -a /etc/hosts
    echo "✅ Added passkey-demo.local to /etc/hosts"
else
    echo "✅ passkey-demo.local already configured in /etc/hosts"
fi

echo
echo "🎉 HTTPS setup complete!"
echo
echo "🚀 Next steps:"
echo "1. Start the backend: cd backend && go run ."
echo "2. Start the frontend: cd frontend-react && npm run dev"
echo "3. Access the demo at: https://passkey-demo.local:5173"
echo
echo "🔒 Security notes:"
echo "• Certificates are valid for 90 days"
echo "• Regenerate with: mkcert passkey-demo.local localhost 127.0.0.1"
echo "• Remove CA: mkcert -uninstall"
echo