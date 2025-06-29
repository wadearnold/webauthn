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

### 3. Backend Configuration

The app is configured to connect to:
- **HTTPS**: `https://passkey-demo.local:8080/api` (preferred)
- **HTTP**: `http://passkey-demo.local:8080/api` (fallback)

Ensure your backend is running with HTTPS certificates:

```bash
cd ../backend
go run .
```

### 4. Build and Run

1. Select your target device (iOS 16+ required)
2. Click **Build and Run** (⌘+R)
3. Trust the developer certificate if prompted

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

## 📚 Resources

- [Apple AuthenticationServices Documentation](https://developer.apple.com/documentation/authenticationservices)
- [WebAuthn API Reference](https://w3c.github.io/webauthn/)
- [FIDO Alliance](https://fidoalliance.org/)
- [SwiftUI Documentation](https://developer.apple.com/documentation/swiftui)

## 📄 License

This demo follows the same license as the main WebAuthn library (BSD 3-Clause).
