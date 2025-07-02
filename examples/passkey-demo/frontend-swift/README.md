# WebAuthn Passkey Demo - iOS (Swift)

A native iOS implementation of the WebAuthn passkey demo using SwiftUI and AuthenticationServices framework. This app demonstrates cross-platform passkey functionality with iCloud Keychain sync.

## 🌟 Features

- **Native iOS WebAuthn**: Uses `ASAuthorizationPlatformPublicKeyCredentialProvider` for platform authenticators
- **iCloud Keychain Sync**: Passkeys automatically sync across all your Apple devices
- **Face ID/Touch ID**: Native biometric authentication integrated with iOS
- **SwiftUI Interface**: Modern, declarative UI with iOS design patterns
- **Cross-Platform Compatible**: Same backend as web and Android implementations
- **Comprehensive Error Handling**: User-friendly error messages and fallbacks
- **Real-time Updates**: Live status updates and passkey management

## 🏗️ Architecture

### Core Components

#### 📱 **Views** (`Views/`)
- **ContentView.swift**: Main app navigation and state management
- **AuthenticationView.swift**: Login screen with discoverable and username-based auth
- **RegistrationView.swift**: User registration with passkey creation
- **DashboardView.swift**: Passkey management and user dashboard

#### 🔧 **Services** (`Services/`)
- **WebAuthnService.swift**: Core WebAuthn logic using AuthenticationServices
- **APIService.swift**: HTTP client for backend communication

#### 📊 **Models** (`Models/`)
- **WebAuthnModels.swift**: Data structures matching backend API contracts

### WebAuthn Integration

The app integrates with iOS's native WebAuthn implementation:

```swift
// Platform authenticator for biometric authentication
let platformProvider = ASAuthorizationPlatformPublicKeyCredentialProvider(
    relyingPartyIdentifier: "passkey-demo.local"
)

// Registration request
let registrationRequest = platformProvider.createCredentialRegistrationRequest(
    challenge: challengeData,
    name: username,
    userID: userIdData
)

// Authentication request  
let assertionRequest = platformProvider.createCredentialAssertionRequest(
    challenge: challengeData
)
```

## 🚀 Setup Instructions

### Prerequisites

- **Xcode 15.0+** (for iOS 16+ deployment target)
- **iOS 16.0+** device or simulator
- **Face ID/Touch ID** enabled device (for full functionality)
- **Backend server** running (see `../backend/README.md`)

### 1. Open in Xcode

```bash
cd examples/passkey-demo/frontend-swift
open PasskeyDemo.xcodeproj
```

### 2. Configure Development Team

1. Select the **PasskeyDemo** project in the navigator
2. Under **Signing & Capabilities**, set your **Development Team**
3. Update **Bundle Identifier** if needed (e.g., `com.yourteam.passkey.demo`)

### 3. Development Mode Configuration

Choose between local development or cross-platform testing:

#### 🏠 Local Development Mode (Fast Iteration)
```bash
# Start backend on localhost
cd ../backend && go run .

# iOS app automatically uses localhost mode
# No additional configuration needed
```

#### 🌐 Cross-Platform Mode (Recommended for Testing)

**Option 1: Manual Configuration**  
```bash
# 1. Start ngrok tunnel
cd .. && ./scripts/start-ngrok.sh

# 2. Copy template and edit with your ngrok URL:
cp PasskeyDemo/ngrok-config.json.template PasskeyDemo/ngrok-config.json
# Edit ngrok-config.json and replace with your actual ngrok URL from step 1

# 3. Start backend with ngrok
cd ../backend && source ../.env && go run .

# 4. Build and run iOS app in Xcode
```

**Option 2: Automatic Configuration**  
```bash
# 1. Start ngrok tunnel
cd .. && ./scripts/start-ngrok.sh

# 2. Configure iOS app automatically
./set-ngrok-url.sh

# 3. Start backend with ngrok
cd ../backend && source ../.env && go run .

# 4. Build and run iOS app in Xcode
```

**Verify Configuration**  
```bash
# Check if your iOS app is properly configured
./check-config.sh
```

**✅ Cross-Platform Benefits:**
- Same RPID across web, iOS, and Android
- Passkey sharing between Safari and iOS app
- True production-like testing environment

### 4. Build and Run

1. Select your target device (iOS 16+ required)
2. **For Simulator**: Enable Face ID via **Features** → **Face ID** → **Enrolled**
3. Click **Build and Run** (⌘+R)
4. Trust the developer certificate if prompted

## 📱 Usage Guide

### Registration Flow

1. **Launch App**: Opens to authentication screen
2. **Create New Passkey**: Tap to navigate to registration
3. **Enter Username**: 3-30 characters, letters/numbers/dots/hyphens/underscores
4. **Optional Display Name**: Your full name (optional)
5. **Create Passkey**: Triggers Face ID/Touch ID prompt
6. **Biometric Auth**: Complete Face ID/Touch ID authentication
7. **Success**: Automatically navigates to dashboard

### Authentication Flows

#### Passwordless (Recommended)
1. **Sign in with Passkey**: No username required
2. **iOS Passkey Selector**: iOS shows available passkeys
3. **Biometric Auth**: Complete Face ID/Touch ID
4. **Dashboard Access**: Logged in successfully

#### Username-based
1. **Enter Username**: Type your registered username
2. **Sign in with Username**: Trigger authentication
3. **Biometric Auth**: Complete Face ID/Touch ID with specific passkey
4. **Dashboard Access**: Logged in successfully

### Passkey Management

- **View All Passkeys**: See all registered passkeys with details
- **iCloud Sync Status**: Check if passkeys are synced across devices
- **Creation/Usage Dates**: Track when passkeys were created and last used
- **Delete Passkeys**: Remove passkeys from server (device removal separate)

## 🧪 Testing

### Cross-Platform Testing

1. **Register on iOS**: Create passkey on iPhone/iPad
2. **Authenticate on Web**: Use same passkey at `https://passkey-demo.local:5173`
3. **Sync Test**: Wait for iCloud sync, test on another Apple device
4. **Multi-Device**: Verify passkey works across iPhone, iPad, Mac, Web

### Device Testing

**Physical Device (Recommended):**
- Full Face ID/Touch ID functionality
- Complete iCloud Keychain sync
- Real-world user experience

**iOS Simulator:**
- Limited biometric simulation
- No actual keychain sync
- Good for UI/UX testing

**⚠️ REQUIRED: Enable Face ID in Simulator**
1. In Xcode Simulator menu: **Features** → **Face ID** → **Enrolled**
2. Without this, you'll see QR codes instead of biometric prompts
3. Once enrolled, use **Face ID** → **Matching/Non-matching** to test

## 🔒 Security Features

### iOS Integration

- **Platform Authenticators**: Uses iOS's built-in Face ID/Touch ID
- **Secure Enclave**: Private keys stored in device's Secure Enclave
- **iCloud Keychain**: End-to-end encrypted sync across Apple devices
- **User Verification**: Biometric authentication for all operations

### WebAuthn Compliance

- **FIDO2/WebAuthn Standard**: Full compliance with W3C WebAuthn specification
- **Attestation**: Supports platform attestation formats
- **User Verification**: Required user verification for all credentials
- **Resident Keys**: All passkeys are discoverable credentials

## 🐛 Troubleshooting

### Error: "Application not associated with domain"
```
Error Domain=com.apple.AuthenticationServices.AuthorizationError Code=1004
Application with identifier FAKETEAMID.com.passkey.demo.ios is not associated 
with domain 67e9-76-154-22-254.ngrok-free.app
```

**Cause**: iOS app is using localhost RPID but trying to access ngrok passkeys

**Solution**:
```bash
# 1. Configure iOS app with current ngrok URL (choose one)
./set-ngrok-url.sh
# OR manually edit PasskeyDemo/ngrok-config.json

# 2. Clean and rebuild in Xcode
# Product → Clean Build Folder  
# Product → Build and Run

# 3. Verify configuration in Xcode debug console:
# Should see: "🔧 Using ngrok URL from config: https://..."
```

### Passkeys Created in Safari Don't Work in iOS App

**Cause**: Different RPIDs being used between Safari and iOS app

**Solution**: Ensure both use the same ngrok domain:
1. ✅ **Safari**: Uses ngrok URL automatically  
2. ✅ **iOS App**: Configure with `./set-ngrok-url.sh` or manually edit config file
3. ✅ **Backend**: Uses ngrok RPID when `NGROK_URL` environment variable is set

### iOS App Uses Localhost Instead of ngrok

**Quick Fix**: Edit `PasskeyDemo/ngrok-config.json`:
```json
{"ngrok_url": "https://your-actual-ngrok-url.ngrok-free.app"}
```

**Verify Configuration**: Run `./check-config.sh` to check your setup

**Check in Xcode Console**: Look for these startup messages:
```
🚀 PasskeyDemo iOS App Starting...
🔧 iOS App Configuration:
Mode: Cross-Platform (ngrok)
Base URL: https://67e9-76-154-22-254.ngrok-free.app/api  
Ngrok URL: https://67e9-76-154-22-254.ngrok-free.app
Status: ✅ Ready for cross-platform passkey sharing
📱 Ready for cross-platform passkey authentication
```

## 📚 Resources

- [Apple AuthenticationServices Documentation](https://developer.apple.com/documentation/authenticationservices)
- [WebAuthn API Reference](https://w3c.github.io/webauthn/)
- [FIDO Alliance](https://fidoalliance.org/)
- [SwiftUI Documentation](https://developer.apple.com/documentation/swiftui)

## 📄 License

This demo follows the same license as the main WebAuthn library (BSD 3-Clause).
