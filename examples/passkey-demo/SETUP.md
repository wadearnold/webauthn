# WebAuthn Passkey Demo - Complete Setup Guide

🎯 **Goal**: Get the cross-platform passkey demo running with your own ngrok domain for testing iOS, Android, and web passkey sharing.

## Prerequisites

- **Go 1.21+**
- **Node.js 18+** 
- **Xcode 15+** with Apple Developer account
- **ngrok account** (free tier works)
- **iOS device** (recommended - simulator has limitations)

## Quick Start (5 Minutes)

### 1. Setup ngrok Domain

```bash
# Clone and navigate to demo
git clone <repository>
cd examples/passkey-demo

# Start ngrok tunnel (generates .env with your domain)
./scripts/start-ngrok.sh
```

This creates `.env` with your unique ngrok URL:
```bash
export NGROK_URL=https://abc123-your-tunnel.ngrok-free.app
export NGROK_PID=12345
```

### 2. Configure iOS App

```bash
cd frontend-swift
./setup-domain.sh
```

This script will:
1. Extract your ngrok domain from `.env`
2. Update `ngrok-config.json` with your domain
3. Update `PasskeyDemo.entitlements` with your domain
4. Show your Apple Developer Team ID

### 3. Update AASA File

```bash
cd ../backend
./setup-aasa.sh YOUR_TEAM_ID
```

Replace `YOUR_TEAM_ID` with the Team ID shown in step 2.

### 4. Build and Run

```bash
# Terminal 1: Start backend
cd backend
source ../.env && go build -o passkey-backend . && ./passkey-backend

# Terminal 2: Build React (optional)
cd frontend-react && npm run build

# Xcode: Build and run iOS app on device
```

### 5. Test Cross-Platform

1. **Register on iOS**: Create a passkey in the iOS app
2. **Login on web**: Visit your ngrok URL in Safari
3. **Same passkey works**: The iOS passkey appears in web login!

## Detailed Setup

### iOS Configuration Script

The `setup-domain.sh` script automates iOS configuration:

```bash
#!/bin/bash
# Read ngrok URL from .env
NGROK_URL=$(grep NGROK_URL ../.env | cut -d'=' -f2)
DOMAIN=$(echo $NGROK_URL | sed 's/https:\/\///')

# Update iOS configuration
echo "🔧 Configuring iOS app for domain: $DOMAIN"

# Update ngrok-config.json
cat > PasskeyDemo/ngrok-config.json << EOF
{
  "ngrok_url": "$NGROK_URL"
}
EOF

# Update entitlements
sed -i '' "s/67e9-76-154-22-254\.ngrok-free\.app/$DOMAIN/g" PasskeyDemo/PasskeyDemo.entitlements

echo "✅ iOS app configured for $DOMAIN"
echo "📱 Your Apple Developer Team ID: $(grep DEVELOPMENT_TEAM PasskeyDemo.xcodeproj/project.pbxproj | head -1 | sed 's/.*= //;s/;//')"
```

### Backend AASA Script

The `setup-aasa.sh` script configures the Apple App Site Association:

```bash
#!/bin/bash
TEAM_ID=$1

if [ -z "$TEAM_ID" ]; then
    echo "Usage: ./setup-aasa.sh YOUR_TEAM_ID"
    echo "Find your Team ID in Xcode or from the iOS setup script"
    exit 1
fi

mkdir -p static/.well-known

cat > static/.well-known/apple-app-site-association << EOF
{
  "webcredentials": {
    "apps": ["$TEAM_ID.com.passkey.demo.ios"]
  }
}
EOF

echo "✅ AASA file configured for Team ID: $TEAM_ID"
echo "🔗 File location: static/.well-known/apple-app-site-association"
```

## Manual Configuration (If Scripts Don't Work)

### iOS App Configuration

1. **Get your ngrok domain** from `.env`:
   ```bash
   cat .env
   # export NGROK_URL=https://abc123.ngrok-free.app
   ```

2. **Update ngrok-config.json**:
   ```json
   {
     "ngrok_url": "https://abc123.ngrok-free.app"
   }
   ```

3. **Update entitlements** in `PasskeyDemo.entitlements`:
   ```xml
   <array>
       <string>webcredentials:abc123.ngrok-free.app</string>
       <string>webcredentials:localhost?mode=developer</string>
       <string>applinks:abc123.ngrok-free.app</string>
   </array>
   ```

4. **Get your Team ID** from Xcode:
   - Open `PasskeyDemo.xcodeproj`
   - Select PasskeyDemo target → Signing & Capabilities
   - Note your Team ID (e.g., `927BS6LD3W`)

### Backend AASA Configuration

1. **Create AASA file**:
   ```bash
   mkdir -p backend/static/.well-known
   ```

2. **Add your Team ID** to `backend/static/.well-known/apple-app-site-association`:
   ```json
   {
     "webcredentials": {
       "apps": ["YOUR_TEAM_ID.com.passkey.demo.ios"]
     }
   }
   ```

### Verification

1. **Test AASA accessibility**:
   ```bash
   curl https://your-domain.ngrok-free.app/.well-known/apple-app-site-association
   ```

2. **Expected response**:
   ```json
   {
     "webcredentials": {
       "apps": ["YOUR_TEAM_ID.com.passkey.demo.ios"]
     }
   }
   ```

## Development Workflow

### Ngrok Domain Changes

Each time you restart ngrok, you get a new domain:

```bash
# Stop current backend
# Restart ngrok
./scripts/start-ngrok.sh

# Reconfigure iOS app
cd frontend-swift && ./setup-domain.sh

# Restart backend
cd ../backend && source ../.env && ./passkey-backend
```

### Testing on Real Device

**Important**: Always test on a real iOS device, not simulator. Associated domains work more reliably on physical devices.

1. Connect iPhone to computer
2. Select device in Xcode (not simulator)
3. Build and run
4. iOS will fetch AASA during installation

## Troubleshooting

### AASA Not Loading

```bash
# Check if AASA is accessible
curl -v https://your-domain.ngrok-free.app/.well-known/apple-app-site-association

# Should return 200 with JSON content
```

### Entitlements Not Working

```bash
# Verify entitlements in built app
cd frontend-swift
./verify-entitlements.sh
```

### Domain Mismatch

Check all three places use the same domain:
1. `.env` file (backend)
2. `ngrok-config.json` (iOS app)
3. `PasskeyDemo.entitlements` (iOS app)

### Registration Fails

Common causes:
- RPID mismatch (check backend logs)
- AASA file not accessible
- Testing on simulator instead of device
- Team ID incorrect in AASA file

## Production Deployment

For production, replace ngrok with your actual domain:

1. **Use your domain** instead of ngrok
2. **Valid SSL certificate** (Let's Encrypt, etc.)
3. **Update all configurations** with production domain
4. **Deploy AASA file** to `/.well-known/apple-app-site-association`

## Need Help?

1. **Check console logs** in Xcode and Console.app
2. **Verify backend logs** show correct RPID
3. **Test AASA accessibility** via curl
4. **Use real device** not simulator

---

🎉 **Success**: When working, you can register a passkey on iOS and use it to login on the web browser - true cross-platform authentication!