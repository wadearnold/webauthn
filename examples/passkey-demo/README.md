# WebAuthn Passkey Demo - Multi-Platform

A complete demonstration of passwordless authentication using WebAuthn passkeys with **cross-platform synchronization**. Features a Go backend with React, iOS (Swift), and Android (Kotlin) frontends to showcase passkey compatibility across all major platforms.

## 🌟 Features

- **True Passwordless Authentication**: No passwords required, just biometrics or device PINs
- **Cross-Platform Compatibility**: Same passkeys work across web, iOS, and Android
- **Discoverable Credentials**: Users can sign in without entering a username
- **Shared Keychain Sync**: Passkeys sync via iCloud Keychain and Google Password Manager
- **Multi-Platform Frontends**: React web app, native iOS (Swift), and Android (Kotlin) apps
- **Passkey Management**: View and delete registered passkeys across all platforms
- **Deep Link Authentication**: Protected routes with automatic authentication redirect
- **Clone Detection**: Security warnings for potentially compromised authenticators
- **Comprehensive Debugging**: Extensive logging for troubleshooting authentication issues

## 🏗️ Architecture

### Backend (Go)
- **WebAuthn Library**: Uses `github.com/go-webauthn/webauthn` for protocol implementation
- **In-Memory Storage**: Thread-safe storage for users, credentials, and sessions
- **RESTful API**: Clean REST endpoints shared across all frontend platforms
- **CORS Support**: Configured for cross-origin requests from multiple frontends
- **Comprehensive Logging**: Detailed debugging for WebAuthn operations

### Frontend Platforms

#### 🌐 Web (React 19) - `frontend-react/`
- **Modern React**: Uses React 19 with hooks and concurrent features
- **WebAuthn API**: Direct browser WebAuthn API integration
- **Responsive Design**: Works on desktop and mobile browsers
- **Comprehensive Debugging**: Extensive console logging for troubleshooting

#### 📱 iOS (Swift) - `frontend-swift/` (Planned)
- **Native iOS**: SwiftUI with WebAuthn platform APIs
- **iCloud Keychain**: Automatic sync across Apple devices
- **Face ID/Touch ID**: Native biometric authentication
- **Universal Links**: Deep link authentication support

#### 🤖 Android (Kotlin) - `frontend-kotlin/` (Planned)
- **Jetpack Compose**: Modern Android UI framework
- **Credential Manager**: Android WebAuthn API integration
- **Google Password Manager**: Cross-platform passkey sync
- **Biometric Authentication**: Fingerprint and face unlock support

## 🚨 **CRITICAL**: Domain Setup Required

**This demo uses `passkey-demo.local` for cross-platform WebAuthn compatibility. You MUST configure local domain resolution first:**

```bash
# Add to your hosts file (macOS/Linux)
sudo vim /etc/hosts

# Add this line at the end:
127.0.0.1 passkey-demo.local
```

**Why this matters**: WebAuthn passkeys are tied to the Relying Party ID (RPID). Using a consistent domain across all platforms enables **true cross-platform passkey sharing**.

## 🚀 Quick Start

### Prerequisites
- Go 1.22 or later
- Node.js 18 or later  
- A modern browser with WebAuthn support (Chrome, Firefox, Safari, Edge)
- **Domain configuration** (see above)

### 1. Start the Backend

```bash
cd examples/passkey-demo/backend
go mod tidy
go run .
```

**Backend will be available at**: http://passkey-demo.local:8080

### 2. Start the React Frontend

```bash
cd examples/passkey-demo/frontend-react
npm install
npm run dev
```

**Frontend will be available at**: http://passkey-demo.local:5173

⚠️ **Important**: Always use `passkey-demo.local`, not `localhost`, for proper cross-platform functionality.

### 3. Open Your Browser

Navigate to `http://localhost:5173` and start testing!

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

### Multi-device Testing
1. Register on one device
2. If your passkey is backed up, try signing in on another device
3. Test the cross-device authentication experience

### Security Features
1. Try to register the same username multiple times
2. Test the clone detection by manipulating authenticator data (advanced)
3. Verify session timeout behavior

### Error Handling
1. Try to authenticate without registering
2. Cancel authentication prompts
3. Test with WebAuthn-unsupported browsers

## 🛠️ Development

### Backend Development
The backend is structured as follows:
- `main.go` - Server setup and routing
- `handlers.go` - HTTP request handlers
- `models.go` - Data models and in-memory storage
- `middleware.go` - CORS and session middleware

### Frontend Development
The frontend uses modern React patterns:
- `hooks/useWebAuthn.js` - WebAuthn API integration
- `services/api.js` - Backend API client
- `components/` - Reusable React components

### Configuration
Backend configuration in `main.go`:
```go
config := &webauthn.Config{
    RPDisplayName: "WebAuthn Passkey Demo",
    RPID:          "localhost",
    RPOrigins:     []string{"http://localhost:5173"},
    // ... other settings
}
```

## 🔒 Security Considerations

### Demo vs Production
This demo uses simplified security for ease of development:
- **In-memory storage**: Data is lost on restart
- **HTTP (not HTTPS)**: Only acceptable for localhost development
- **Simplified sessions**: Production should use secure session storage
- **No rate limiting**: Production should implement proper rate limiting

### Important: Passkey Deletion Behavior
⚠️ **WebAuthn Limitation**: When you delete a passkey in this demo, it only removes the server's record. The passkey remains in your device's keychain and may still appear during authentication prompts. This is by design for security reasons - only users can manage their device keychains.

**To clean up test passkeys from your device, search for your server name (e.g., "localhost" for local development):**
- **Mac**: System Settings → Passwords → Website & App Passwords → Search for server name
- **iPhone/iPad**: Settings → Passwords → Search for server name
- **Android**: Settings → Passwords & accounts → Google → Passkeys
- **Windows**: Settings → Accounts → Sign-in options → Security keys
- **Chrome**: Settings → Autofill and passwords → Password Manager → Passkeys

💡 **Tip**: The delete confirmation dialog will show you the exact server name to search for in your device settings.

### Production Recommendations
1. **Use HTTPS**: WebAuthn requires secure contexts in production
2. **Secure session storage**: Use encrypted cookies or server-side sessions
3. **Database storage**: Persist users and credentials in a database
4. **Rate limiting**: Implement authentication attempt limits
5. **Monitoring**: Log security events and failed attempts
6. **MDS validation**: Consider using FIDO Metadata Service for attestation validation

## 🐛 Troubleshooting

### Common Issues

#### "WebAuthn is not supported"
- Use a modern browser (Chrome 67+, Firefox 60+, Safari 14+, Edge 18+)
- Ensure you're accessing via `http://localhost` (not 127.0.0.1)
- Check if WebAuthn is disabled in browser settings

#### CORS Errors
- Ensure backend is running on port 8080
- Ensure frontend is running on port 5173
- Check that CORS origins match in `main.go`

#### "No authenticator found" during registration
- Enable biometrics or set up a PIN on your device
- For testing, you can use a USB security key
- Some browsers require user gesture before WebAuthn calls

#### Session errors
- Sessions expire after 5 minutes by default
- Clear browser cookies if you encounter stale sessions
- Restart the backend to clear all sessions

### Browser Compatibility
- **Chrome/Edge**: Full support including platform authenticators
- **Firefox**: Good support, may prompt for specific authenticator
- **Safari**: iOS 14+ and macOS Big Sur+ for full passkey support
- **Mobile browsers**: Generally good support on modern devices

## 📚 Learn More

- [WebAuthn Specification](https://www.w3.org/TR/webauthn/)
- [FIDO Alliance](https://fidoalliance.org/)
- [WebAuthn Guide](https://webauthn.guide/)
- [Passkeys.dev](https://passkeys.dev/)

## 🤝 Contributing

This demo is part of the WebAuthn library examples. To contribute:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

This demo follows the same license as the main WebAuthn library (BSD 3-Clause).