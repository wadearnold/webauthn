# WebAuthn Passkey Demo - Cross-Platform

A complete demonstration of passwordless authentication using WebAuthn passkeys with **cross-platform synchronization**. Features a Go backend with React web and iOS Swift frontends, showcasing passkey compatibility across platforms using ngrok tunneling.

## 🌟 Features

- **True Passwordless Authentication**: No passwords required, just biometrics or device PINs
- **Cross-Platform Compatibility**: Same passkeys work across web and iOS
- **Discoverable Credentials**: Users can sign in without entering a username
- **Shared Keychain Sync**: Passkeys sync via iCloud Keychain and browser password managers
- **Multi-Platform Frontends**: React web app and native iOS (Swift) app
- **Passkey Management**: View and delete registered passkeys across all platforms
- **ngrok Integration**: Public HTTPS URLs for iOS domain association compatibility
- **Comprehensive Debugging**: Extensive logging for troubleshooting authentication issues

## ⚠️ Critical WebAuthn Concept

**Must Read:** [WebAuthn RPID and Origin Security Guide](./WEBAUTHN-RPID-GUIDE.md)

WebAuthn requires the browser origin to match the RPID domain. This is a security feature, not a bug. The guide explains how to handle this during development.

## 🏗️ Architecture

### Backend (Go)
- **WebAuthn Library**: Uses `github.com/go-webauthn/webauthn` for protocol implementation
- **In-Memory Storage**: Thread-safe storage for users, credentials, and sessions
- **RESTful API**: Clean REST endpoints shared across all frontend platforms
- **CORS Support**: Configured for cross-origin requests and ngrok domains
- **HTTP Server**: Runs on port 8080, ngrok provides HTTPS termination

### Frontend Platforms

#### 🌐 Web (React 19) - `frontend-react/`
- **Modern React**: Uses React 19 with hooks and Vite development server
- **WebAuthn API**: Direct browser WebAuthn API integration
- **Environment-based Configuration**: Automatically detects ngrok URL from environment
- **Responsive Design**: Works on desktop and mobile browsers

#### 📱 iOS (Swift) - `frontend-swift/`
- **Native iOS**: SwiftUI with WebAuthn platform APIs
- **Automatic Configuration**: Detects ngrok URL from multiple sources
- **iCloud Keychain**: Automatic sync across Apple devices
- **Face ID/Touch ID**: Native biometric authentication

## 🚀 Quick Start

### Prerequisites
- Go 1.22 or later
- Node.js 18 or later  
- ngrok account (free tier works fine)
- A modern browser with WebAuthn support
- iOS Simulator or device (for iOS testing)

### 1. Install and Configure ngrok

```bash
# Install ngrok (if not already installed)
brew install ngrok/ngrok/ngrok

# Sign up at https://ngrok.com and get your auth token
ngrok config add-authtoken YOUR_AUTH_TOKEN

# Or continue without auth token (tunnel will be temporary)
```

### 2. Start ngrok Tunnel (One Time)

```bash
cd examples/passkey-demo
./scripts/start-ngrok.sh
```

This will:
- Start an ngrok tunnel on port 8080
- Save the tunnel URL to `.env` file
- Display the public URL and startup instructions

**Note**: The ngrok tunnel stays running - you don't need to restart it when restarting the backend.

### 3. Start the Backend

```bash
cd backend
source ../.env && go run .
```

To restart the backend (for development), just stop with `Ctrl+C` and run again:
```bash
source ../.env && go run .
```

**Backend will be available at**: 
- Local: http://localhost:8080
- Public: https://your-tunnel.ngrok.io

### 4. Start the React Frontend

```bash
cd frontend-react
npm install
npm run dev
```

**Frontend will be available at**: http://localhost:5173

The React app will automatically use the ngrok URL from the `.env` file for API calls.

### 5. Configure iOS Frontend (Optional)

```bash
# Set ngrok URL for Swift app
./scripts/set-swift-ngrok.sh

# Then rebuild the app in Xcode
cd frontend-swift
open PasskeyDemo.xcodeproj
```

### 6. Test the Demo

⚠️ **IMPORTANT WebAuthn Security Restriction:**

WebAuthn requires that the browser origin matches the RPID domain. You have two development modes:

#### Option A: Local Development Mode (Quick Testing)
Use this mode for:
- Rapid development and testing
- Working on a single platform (web only)
- Quick prototyping without cross-platform needs

```bash
# Start backend WITHOUT ngrok URL
cd backend
go run .  # RPID will be "localhost"

# Access frontend at:
http://localhost:5173
```

**Limitations of Local Mode:**
- ❌ Passkeys only work on localhost
- ❌ Cannot test cross-platform sharing (iOS/Android)
- ❌ Cannot test with real devices
- ✅ Fast iteration for web-only development

#### Option B: Cross-Platform Mode with ngrok (Production-like)
Use this mode for:
- Testing passkey sharing across platforms
- iOS/Android app testing
- Production-like environment testing
- Testing on real devices

```bash
# 1. Start ngrok tunnel (if not already running)
./scripts/start-ngrok.sh

# 2. Build React app with ngrok URL
cd frontend-react
npm run build:ngrok

# 3. Start backend with ngrok URL
cd ../backend
source ../.env && go run .

# 4. Access through ngrok URL:
https://your-tunnel.ngrok.io
```

**Benefits of ngrok Mode:**
- ✅ Same passkeys work across web, iOS, and Android
- ✅ Test on real devices using public HTTPS URL
- ✅ Production-like security model
- ✅ Proper domain association for mobile apps
- ❌ Requires accessing via ngrok URL (not localhost)

## 📱 Usage Guide

### Registration Flow
1. Enter a username (display name is optional)
2. Click "Create Passkey"
3. Follow your browser's prompts to create a passkey
4. You'll be automatically signed in after successful registration

### Authentication Flows

#### Passwordless (Recommended)
1. Click "Sign in with Passkey"
2. Your browser will show available passkeys
3. Authenticate with biometrics/PIN
4. No username required!

#### Username-based
1. Enter your username
2. Click "Sign in with Username"
3. Authenticate with your passkey when prompted

### Passkey Management
- View all your registered passkeys
- See which passkeys are backed up to the cloud
- Delete passkeys you no longer need

## 🔧 API Endpoints

### Registration
- `POST /api/register/begin` - Start passkey registration
- `POST /api/register/finish` - Complete passkey registration

### Authentication
- `POST /api/login/begin` - Start authentication (supports both discoverable and username-based)
- `POST /api/login/finish` - Complete authentication

### User Management
- `GET /api/user/passkeys` - List user's passkeys
- `DELETE /api/user/passkeys/{id}` - Delete a specific passkey
- `POST /api/logout` - Sign out

### Utility
- `GET /api/health` - Health check

## 🧪 Testing Scenarios

### Basic Flow
1. Register a new user with a passkey
2. Sign out
3. Sign back in using the passwordless flow

### Cross-Platform Testing
1. Register a passkey in the web app
2. If using iOS app, configure it with the same ngrok URL
3. Try signing in with the same passkey on iOS
4. Test the cross-platform authentication experience

### Multi-device Testing
1. Register on one device
2. If your passkey is backed up, try signing in on another device
3. Test the cross-device authentication experience

## 🛠️ Development

### Choosing Your Development Mode

**When to use Local Development Mode (localhost):**
- You're only working on the web frontend
- You need fast iteration without building React
- You're debugging backend logic
- You don't need cross-platform testing

**When to use Cross-Platform Mode (ngrok):**
- You're testing iOS or Android apps
- You need to verify passkey sharing works
- You're testing on real devices
- You're preparing for production deployment

### Development Workflows

#### Local Development Workflow (Web Only)
```bash
# Terminal 1: Start backend in local mode
cd backend
go run .  # No ngrok URL needed

# Terminal 2: Start React dev server
cd frontend-react
npm run dev

# Access at: http://localhost:5173
```

#### Cross-Platform Development Workflow
```bash
# Terminal 1: Start ngrok (once per session)
./scripts/start-ngrok.sh

# Terminal 2: Build and serve through backend
cd frontend-react
npm run build:ngrok  # Build with ngrok URL
cd ../backend
source ../.env && go run .  # Uses ngrok URL

# Access at: https://your-tunnel.ngrok.io
```

#### Hybrid Development Workflow (Recommended)
```bash
# Start with local mode for rapid development
cd backend
go run .

# When ready to test cross-platform:
# 1. Stop backend (Ctrl+C)
# 2. Start ngrok if needed: ./scripts/start-ngrok.sh
# 3. Build React: cd frontend-react && npm run build:ngrok
# 4. Restart backend with ngrok: cd ../backend && source ../.env && go run .
```

**ngrok stays running** throughout your development session - no need to restart it!

### ngrok Management

```bash
# Start ngrok tunnel (once per development session)
./scripts/start-ngrok.sh

# Get current tunnel URL
./scripts/get-ngrok-url.sh

# Stop ngrok tunnel (when done developing)
./scripts/stop-ngrok.sh
```

### Backend Development
The backend is structured as follows:
- `main.go` - Server setup and routing with ngrok support
- `handlers.go` - HTTP request handlers
- `models.go` - Data models and in-memory storage
- `middleware.go` - CORS middleware with ngrok domain support

### Frontend Development

#### React Frontend
- Uses `VITE_NGROK_URL` environment variable for API base URL
- Falls back to localhost if ngrok URL not available
- Modern React patterns with hooks

#### iOS Frontend
- `APIConfiguration` struct detects ngrok URL from multiple sources
- Supports Info.plist configuration for build-time injection
- Runtime configuration via UserDefaults

### Configuration

Backend automatically uses ngrok URL from environment:
```go
ngrokURL := os.Getenv("NGROK_URL")
// CORS configured to accept ngrok domains
```

React frontend uses environment variable:
```javascript
const ngrokUrl = import.meta.env.VITE_NGROK_URL;
```

iOS frontend detects configuration automatically:
```swift
// Checks Info.plist, environment, and UserDefaults
APIConfiguration.ngrokURL
```

## 🔒 Security Considerations

### Demo vs Production
This demo uses simplified security for ease of development:
- **In-memory storage**: Data is lost on restart
- **ngrok tunneling**: Convenient for development, not recommended for production
- **Simplified sessions**: Production should use secure session storage
- **No rate limiting**: Production should implement proper rate limiting

### ngrok Security
- ngrok provides trusted HTTPS certificates
- Tunnel URLs are publicly accessible (temporary for free tier)
- Use ngrok auth tokens for additional security features
- Consider ngrok paid plans for production-like testing

### Production Recommendations
1. **Use proper HTTPS**: Deploy with real SSL certificates
2. **Secure session storage**: Use encrypted cookies or server-side sessions
3. **Database storage**: Persist users and credentials in a database
4. **Rate limiting**: Implement authentication attempt limits
5. **Monitoring**: Log security events and failed attempts
6. **Domain validation**: Use your own domains with proper DNS

## 🐛 Troubleshooting

### ngrok Issues
#### "ngrok is not installed"
```bash
brew install ngrok/ngrok/ngrok
```

#### "ngrok is not authenticated"
```bash
ngrok config add-authtoken YOUR_AUTH_TOKEN
# Get token from: https://dashboard.ngrok.com/get-started/your-authtoken
```

#### "Failed to get ngrok URL"
- Check if ngrok is running: `./scripts/get-ngrok-url.sh`
- Restart ngrok: `./scripts/stop-ngrok.sh && ./scripts/start-ngrok.sh`

### WebAuthn Issues
#### "WebAuthn is not supported"
- Use a modern browser (Chrome 67+, Firefox 60+, Safari 14+, Edge 18+)
- Ensure you're accessing via HTTPS (ngrok provides this)

#### iOS Domain Association Issues
- Ensure iOS app is configured with correct ngrok URL
- Run `./scripts/set-swift-ngrok.sh` to update configuration
- Rebuild the iOS app in Xcode after configuration changes

#### CORS Errors
- Backend automatically accepts ngrok domains
- Ensure backend is running and accessible via ngrok URL
- Check that frontend is using correct API base URL

### Session Issues
- Sessions expire after 5 minutes by default
- Clear browser cookies if you encounter stale sessions
- Restart the backend to clear all sessions

## 🛑 Cleanup

### Stop All Services
```bash
# Stop ngrok tunnel
./scripts/stop-ngrok.sh

# Stop backend (Ctrl+C in terminal)
# Stop React frontend (Ctrl+C in terminal)
```

### Clean Up Passkeys
⚠️ **Important**: Deleting passkeys in the demo only removes server records. To clean up test passkeys from your devices:

- **Mac**: System Settings → Passwords → Search for your ngrok domain
- **iPhone/iPad**: Settings → Passwords → Search for your ngrok domain  
- **Chrome**: Settings → Autofill and passwords → Passkeys

## 📚 Learn More

- [WebAuthn Specification](https://www.w3.org/TR/webauthn/)
- [ngrok Documentation](https://ngrok.com/docs)
- [Passkeys.dev](https://passkeys.dev/)
- [FIDO Alliance](https://fidoalliance.org/)

## 🤝 Contributing

This demo is part of the WebAuthn library examples. To contribute:
1. Fork the repository
2. Create a feature branch
3. Test with the ngrok setup
4. Submit a pull request

## 📄 License

This demo follows the same license as the main WebAuthn library.