# WebAuthn Passkey Demo - Backend

A Go backend server implementing WebAuthn passkey authentication with **cross-platform domain configuration** for seamless passkey sharing across web, iOS, and Android platforms.

## 🚨 **IMPORTANT**: WebAuthn Security Requirements

This backend uses `passkey-demo.local` as the Relying Party ID (RPID) to enable **cross-platform passkey compatibility**. However, WebAuthn has strict security requirements:

### 🔒 **WebAuthn Security Model**

**WebAuthn requires HTTPS for custom domains** (security requirement). Only `localhost` gets special exemption.

#### ✅ **Option 1: Use localhost (Easiest)**
```
Access: http://localhost:5173
RPID: passkey-demo.local (configured in backend)
Result: ✅ Works immediately, but limited cross-platform testing
```

#### 🌐 **Option 2: Use custom domain with hosts file (Advanced)**
```
Setup: Add '127.0.0.1 passkey-demo.local' to /etc/hosts
Access: http://passkey-demo.local:5173
RPID: passkey-demo.local
Result: ⚠️ Requires HTTPS for WebAuthn to work
```

#### 🔐 **Option 3: HTTPS Setup (Production-like - Recommended)**
```bash
# Automated setup (from passkey-demo directory)
./setup-https.sh

# Manual setup:
brew install mkcert
mkcert -install
mkdir -p certs && cd certs
mkcert passkey-demo.local localhost 127.0.0.1

# Then access:
# Frontend: https://passkey-demo.local:5173
# Backend: https://passkey-demo.local:8080
# Result: ✅ Full cross-platform compatibility
```

### ⚡ Quick Setup (macOS)

**1. Add domain to your hosts file:**
```bash
sudo vim /etc/hosts
```

**2. Add these lines at the end:**
```
# WebAuthn Passkey Demo - Cross-Platform Configuration
127.0.0.1 passkey-demo.local
127.0.0.1 api.passkey-demo.local
```

**3. Save and verify:**
```bash
# Test domain resolution
ping passkey-demo.local
# Should respond from 127.0.0.1
```

**4. Start the backend:**
```bash
go run .
```

**5. Run HTTPS setup (recommended):**
```bash
cd .. && ./setup-https.sh
```

**6. Access the demo:**
- **Backend API**: https://passkey-demo.local:8080 (HTTPS) or http://passkey-demo.local:8080 (!WebAuthn limited)
- **React Frontend**: https://passkey-demo.local:5173 (after starting frontend)

## 🤔 Why Local Domain Configuration?

### WebAuthn Security Model

WebAuthn passkeys are **tied to the Relying Party ID (RPID)**, which must match across all platforms for passkey sharing to work.

#### ❌ Without Domain Configuration (Broken Cross-Platform)

```
Platform          Origin                    RPID         Passkey Scope
Web (React)       http://localhost:5173     localhost    ❌ localhost only
iOS Simulator     http://127.0.0.1:8080     127.0.0.1    ❌ IP-specific  
Android Emulator  http://10.0.2.2:8080      10.0.2.2     ❌ Different IP
```

**Result**: Each platform creates **separate, incompatible passkeys**. No cross-platform functionality.

#### ✅ With Domain Configuration (Cross-Platform Compatible)

```
Platform          Origin                              RPID                Passkey Scope
Web (React)       http://passkey-demo.local:5173      passkey-demo.local  ✅ Shared domain
iOS Simulator     http://passkey-demo.local:8080      passkey-demo.local  ✅ Same domain
Android Emulator  http://passkey-demo.local:8080      passkey-demo.local  ✅ Same domain
iOS Native App    ios-app://[TEAM].[BUNDLE]          passkey-demo.local  ✅ Associated domain
Android App       android-app://[PACKAGE]             passkey-demo.local  ✅ Associated domain
```

**Result**: **Single passkey works across all platforms** via shared keychain sync (iCloud Keychain, Google Password Manager).

### Real-World Production Scenario

In production, you'd use a real domain:

```go
config := &webauthn.Config{
    RPID: "yourdomain.com",
    RPOrigins: []string{
        "https://yourdomain.com",                    // Web app
        "https://app.yourdomain.com",                // Mobile web
        "ios-app://TEAMID.com.yourapp.passkey",     // iOS app  
        "android-app://com.yourapp.passkey",        // Android app
    },
}
```

**This enables**:
- **Web users** create passkey → **Mobile app users** can use the same passkey
- **iPhone users** create passkey → **Android users** can use it (via Google sync)
- **Seamless experience** across all your company's apps and websites

## 🛠 Backend Configuration

### WebAuthn Setup

```go
config := &webauthn.Config{
    RPDisplayName: "WebAuthn Passkey Demo",
    RPID:          "passkey-demo.local",  // Critical for cross-platform
    RPOrigins: []string{
        "http://passkey-demo.local:5173",  // React frontend
        "http://passkey-demo.local:3000",  // Alternative React port
        "http://passkey-demo.local:8080",  // API server (for mobile)
        "capacitor://passkey-demo.local",  // Capacitor hybrid apps
        "ionic://passkey-demo.local",     // Ionic hybrid apps
    },
    AuthenticatorSelection: protocol.AuthenticatorSelection{
        AuthenticatorAttachment: protocol.Platform,     // Force biometrics
        ResidentKey: protocol.ResidentKeyRequirementPreferred,
        UserVerification: protocol.VerificationPreferred,
    },
}
```

### Key Configuration Choices

- **RPID**: `passkey-demo.local` (enables cross-platform passkey sharing)
- **Platform Authenticators**: Forces built-in biometrics (Touch ID, Face ID, fingerprint)
- **Resident Keys**: Required for discoverable credentials (passwordless login)
- **User Verification**: Preferred (allows fallback if biometrics unavailable)

## 🔄 API Endpoints

All endpoints support **JSON** requests and responses with **CORS** enabled for cross-platform access.

### Registration Flow
```http
POST /api/register/begin
Content-Type: application/json

{
    "username": "testuser",
    "displayName": "Test User"
}
```

```http
POST /api/register/finish
Content-Type: application/json

{
    "id": "credential-id",
    "rawId": "base64-encoded-raw-id",
    "type": "public-key",
    "response": {
        "attestationObject": "base64-encoded-attestation",
        "clientDataJSON": "base64-encoded-client-data",
        "transports": ["internal", "hybrid"]
    }
}
```

### Authentication Flow
```http
POST /api/login/begin
Content-Type: application/json

{
    "username": "testuser"  // Optional for discoverable login
}
```

```http
POST /api/login/finish
Content-Type: application/json

{
    "id": "credential-id",
    "rawId": "base64-encoded-raw-id", 
    "type": "public-key",
    "response": {
        "authenticatorData": "base64-encoded-auth-data",
        "clientDataJSON": "base64-encoded-client-data",
        "signature": "base64-encoded-signature",
        "userHandle": "base64-encoded-user-handle"
    }
}
```

### Passkey Management
```http
GET /api/user/passkeys
Cookie: user-session=username

Response: [
    {
        "id": "credential-id",
        "name": "Device Name",
        "createdAt": "2024-01-01T00:00:00Z",
        "lastUsed": "2024-01-01T00:00:00Z",
        "transports": ["internal", "hybrid"],
        "backedUp": true,
        "userVerified": true
    }
]
```

```http
DELETE /api/user/passkeys/{credential-id}
Cookie: user-session=username
```

## 🧪 Cross-Platform Testing

### Testing Matrix

| Create Platform | Authenticate Platform | Expected Result |
|----------------|----------------------|-----------------|
| React Web | iOS Simulator | ✅ Should work |
| iOS Simulator | React Web | ✅ Should work |
| React Web | Android Emulator | ✅ Should work |
| iOS Device | Android Device | ✅ Should work (via Google PM) |
| Mac Safari | iPhone Safari | ✅ Should work (via iCloud) |

### Testing Steps

1. **Create passkey on one platform**:
   ```bash
   # Example: Register user on React web
   curl -X POST http://passkey-demo.local:8080/api/register/begin \
        -H "Content-Type: application/json" \
        -d '{"username":"testuser","displayName":"Test User"}'
   ```

2. **Authenticate from different platform**:
   ```bash
   # Example: Login from mobile app
   curl -X POST http://passkey-demo.local:8080/api/login/begin \
        -H "Content-Type: application/json" \
        -d '{"username":"testuser"}'
   ```

3. **Verify passkey in dashboard**:
   - Check passkey appears across all platforms
   - Verify transport types and backup status
   - Test deletion synchronization

## 🚀 Running the Backend

### Prerequisites
- **Go 1.21+**
- **Domain configuration** (see setup above)
- **HTTPS certificate** (for production)

### Development Mode
```bash
# Clone and navigate
git clone <repository>
cd examples/passkey-demo/backend

# Install dependencies
go mod download

# Run server
go run .

# Server starts on :8080
# API: http://passkey-demo.local:8080
```

### Production Mode
```bash
# Build binary
go build -o passkey-demo-backend .

# Run with environment variables
RPID=yourdomain.com \
RP_ORIGINS=https://yourdomain.com,https://app.yourdomain.com \
./passkey-demo-backend
```

## 🔧 Configuration Options

### Environment Variables
```bash
# WebAuthn Configuration
RPID=passkey-demo.local                    # Relying Party ID
RP_DISPLAY_NAME="WebAuthn Passkey Demo"    # Display name
RP_ORIGINS=http://passkey-demo.local:5173  # Allowed origins (comma-separated)

# Server Configuration  
PORT=8080                                  # Server port
LOG_LEVEL=info                            # Logging level
```

### Security Considerations

- **RPID Validation**: Server validates all requests against configured RPID
- **Origin Checking**: CORS middleware enforces allowed origins
- **Session Security**: HTTP-only cookies with SameSite protection
- **Credential Validation**: Prevents authentication with deleted credentials
- **Clone Detection**: Warns about potentially compromised authenticators

## 🐛 Troubleshooting

### Common Issues

**1. "Failed to create credential" errors**
```bash
# Check domain resolution
nslookup passkey-demo.local
# Should return 127.0.0.1

# Verify hosts file
cat /etc/hosts | grep passkey-demo
```

**2. CORS errors in browser console**
```bash
# Check server logs for origin mismatches
# Verify frontend is accessing passkey-demo.local, not localhost
```

**3. Passkeys not syncing across platforms**
```bash
# Verify same RPID is used across all platforms
# Check keychain sync settings (iCloud, Google)
# Ensure platforms are using same domain
```

**4. Biometric authentication fails**
- Ensure device has biometrics enabled
- Check if Touch ID/Face ID is configured
- Verify browser permissions for security keys

### Debug Logging

The server provides comprehensive debug output:

```
=== REGISTRATION DEBUG INFO FOR testuser ===
AuthenticatorAttachment: platform
ResidentKey: required
RequireResidentKey: true
UserVerification: required
RPID: passkey-demo.local
RPName: WebAuthn Passkey Demo
=========================================
```

**Enable verbose logging:**
```bash
go run . -v
```

## 📚 Additional Resources

- [WebAuthn Specification](https://w3c.github.io/webauthn/)
- [FIDO Alliance](https://fidoalliance.org/)
- [go-webauthn Library Documentation](https://github.com/go-webauthn/webauthn)
- [MDN WebAuthn API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Authentication_API)

## 🔒 Security Notes

This is a **demonstration server** with in-memory storage. For production:

- Use **persistent database** storage
- Implement **rate limiting** and **CAPTCHA**
- Add **audit logging** for all authentication events
- Use **HTTPS** with valid certificates
- Implement **backup recovery** mechanisms
- Add **enterprise attestation** validation

---

**⚠️ Remember**: The local domain configuration is **essential** for cross-platform passkey functionality. Without it, each platform will create separate, incompatible passkeys!