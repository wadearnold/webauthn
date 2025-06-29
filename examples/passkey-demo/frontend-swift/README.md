# WebAuthn Passkey Demo - iOS Swift Frontend

A native iOS Swift app demonstrating WebAuthn passkey authentication that works seamlessly with the shared keychain across Apple devices.

## 🎯 Demo Goals

Showcase that passkeys created on any device (web, iOS, Android) can be used to authenticate across platforms when using shared keychains (iCloud Keychain, Google Password Manager, etc.).

## 🛠 Implementation Plan

### Core Features to Implement

- **Passkey Registration**: Create new passkeys using iOS WebAuthn APIs
- **Passwordless Authentication**: Sign in using Face ID/Touch ID
- **Cross-Platform Sync**: Demonstrate passkeys work across devices
- **Passkey Management**: View and delete user's passkeys
- **Deep Link Authentication**: Handle authentication from external links

### Technical Stack

- **Language**: Swift
- **Framework**: SwiftUI or UIKit
- **WebAuthn**: ASAuthorizationWebBrowserPlatformPublicKeyCredential (iOS 16+)
- **Biometrics**: Local Authentication framework
- **Networking**: URLSession for backend communication
- **Deployment Target**: iOS 16.0+ (required for WebAuthn)

### Project Structure

```
PasskeyDemoiOS/
├── PasskeyDemoiOS.xcodeproj
├── PasskeyDemoiOS/
│   ├── App/
│   │   ├── PasskeyDemoiOSApp.swift
│   │   └── ContentView.swift
│   ├── Views/
│   │   ├── RegisterView.swift
│   │   ├── LoginView.swift
│   │   ├── DashboardView.swift
│   │   └── ProfileView.swift
│   ├── Services/
│   │   ├── WebAuthnService.swift
│   │   ├── APIService.swift
│   │   └── KeychainService.swift
│   ├── Models/
│   │   ├── User.swift
│   │   ├── PasskeyInfo.swift
│   │   └── APIResponse.swift
│   └── Utils/
│       ├── Base64URL.swift
│       └── BiometricsHelper.swift
├── README.md
└── Package.swift (if using SPM)
```

### Key Implementation Components

#### 1. WebAuthn Service (`WebAuthnService.swift`)

```swift
import AuthenticationServices
import Foundation

class WebAuthnService: NSObject, ObservableObject {
    
    func createPasskey(challenge: Data, userID: Data, userName: String, displayName: String) async throws -> ASAuthorizationPlatformPublicKeyCredentialRegistration {
        
        let provider = ASAuthorizationPlatformPublicKeyCredentialProvider(relyingPartyIdentifier: "passkey-demo.local")
        
        let request = provider.createCredentialRegistrationRequest(
            challenge: challenge,
            name: userName,
            userID: userID
        )
        
        request.displayName = displayName
        request.userVerificationPreference = .required
        
        let controller = ASAuthorizationController(authorizationRequests: [request])
        controller.delegate = self
        controller.presentationContextProvider = self
        
        return try await withCheckedThrowingContinuation { continuation in
            // Handle response in delegate methods
        }
    }
    
    func authenticateWithPasskey(challenge: Data) async throws -> ASAuthorizationPlatformPublicKeyCredentialAssertion {
        // Implementation for authentication
    }
}
```

#### 2. API Service (`APIService.swift`)

```swift
import Foundation

class APIService {
    private let baseURL = "http://passkey-demo.local:8080"
    
    func registerBegin(username: String, displayName: String) async throws -> RegistrationOptions {
        // Call /api/register/begin
    }
    
    func registerFinish(credential: ASAuthorizationPlatformPublicKeyCredentialRegistration) async throws -> AuthResponse {
        // Call /api/register/finish
    }
    
    func loginBegin(username: String? = nil) async throws -> AuthenticationOptions {
        // Call /api/login/begin
    }
    
    func loginFinish(assertion: ASAuthorizationPlatformPublicKeyCredentialAssertion) async throws -> AuthResponse {
        // Call /api/login/finish
    }
    
    func getUserPasskeys() async throws -> [PasskeyInfo] {
        // Call /api/user/passkeys
    }
    
    func deletePasskey(credentialId: String) async throws {
        // Call DELETE /api/user/passkeys/{id}
    }
}
```

#### 3. Dashboard View (`DashboardView.swift`)

```swift
import SwiftUI

struct DashboardView: View {
    @StateObject private var apiService = APIService()
    @State private var passkeys: [PasskeyInfo] = []
    @State private var isLoading = false
    @State private var errorMessage: String?
    
    var body: some View {
        NavigationView {
            VStack {
                // User info header
                // Passkey list
                // Delete functionality
                // Cross-platform sync status
            }
            .navigationTitle("Your Passkeys")
            .task {
                await loadPasskeys()
            }
        }
    }
}
```

### Development Steps

1. **Setup Project**
   ```bash
   # Create new iOS project in Xcode
   # Set deployment target to iOS 16.0+
   # Add required capabilities and permissions
   ```

2. **Implement Core Services**
   - WebAuthn registration and authentication
   - API communication with backend
   - Base64URL encoding/decoding utilities

3. **Build UI Components**
   - Registration flow with username validation
   - Biometric authentication prompts
   - Passkey management dashboard
   - Error handling and loading states

4. **Testing Scenarios**
   - Create passkey on iOS → Use on web
   - Create passkey on web → Use on iOS  
   - Create passkey on Android → Use on iOS (via Google sync)
   - Delete passkey and verify removal across platforms

### Required iOS Capabilities

```xml
<!-- Info.plist -->
<key>NSFaceIDUsageDescription</key>
<string>This app uses Face ID for secure passkey authentication</string>

<!-- Entitlements -->
<key>com.apple.developer.web-browser</key>
<true/>
```

### Cross-Platform Sync Testing

The iOS app should demonstrate:

1. **iCloud Keychain Sync**: Passkeys created on iOS appear on Mac/iPad
2. **Google Password Manager**: If configured, passkeys sync with Android
3. **WebAuthn Compatibility**: Passkeys work seamlessly with web browsers
4. **Universal Links**: Deep link authentication from other platforms

### Demo Flow

1. **Registration on iOS** → Verify appears in web dashboard
2. **Authentication on web** → Use passkey created on iOS
3. **Cross-device testing** → Same passkey works on multiple Apple devices
4. **Management consistency** → Delete from iOS, verify removal everywhere

## 📱 Getting Started

```bash
# Prerequisites
# - Xcode 15.0+
# - iOS 16.0+ device or simulator
# - Backend server running on localhost:8080

# Steps
1. Open Xcode
2. Create new iOS project named "PasskeyDemoiOS"
3. Set deployment target to iOS 16.0
4. Implement WebAuthn integration
5. Test on physical device (required for biometrics)
```

## 🔄 Integration with Existing Demo

This iOS frontend will use the **same backend API** as the React frontend, demonstrating true cross-platform passkey compatibility.

### Backend API Endpoints (Shared)
- `POST /api/register/begin` - Start passkey registration
- `POST /api/register/finish` - Complete passkey registration  
- `POST /api/login/begin` - Start authentication
- `POST /api/login/finish` - Complete authentication
- `GET /api/user/passkeys` - Get user's passkeys
- `DELETE /api/user/passkeys/{id}` - Delete specific passkey

### Cross-Platform Testing Matrix

| Create Platform | Authenticate Platform | Expected Result |
|----------------|----------------------|-----------------|
| iOS App | React Web | ✅ Should work |
| React Web | iOS App | ✅ Should work |
| iOS App | Android App | ✅ Should work (via Google sync) |
| Android App | iOS App | ✅ Should work (via Google sync) |

## 🚀 Future Enhancements

- **Universal Links**: Deep link authentication from web/other apps
- **App Clips**: Lightweight authentication experiences
- **Shortcuts Integration**: Siri shortcuts for quick auth
- **Apple Watch**: Companion app for wrist-based authentication
- **Enterprise Features**: Managed app configuration for corporate use

---

**Status**: 📋 **Implementation Planned** - Ready for development