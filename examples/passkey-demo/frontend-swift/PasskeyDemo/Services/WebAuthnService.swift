import Foundation
import AuthenticationServices
import SwiftUI
import LocalAuthentication

@MainActor
class WebAuthnService: NSObject, ObservableObject {
    @Published var isLoading = false
    @Published var errorMessage: String?
    @Published var currentUser: String?
    @Published var userPasskeys: [UserPasskey] = []
    
    private let apiService = APIService.shared
    
    // MARK: - Registration
    
    func register(username: String, displayName: String) async throws -> RegistrationResult {
        isLoading = true
        errorMessage = nil
        
        defer { isLoading = false }
        
        do {
            // Step 1: Get registration options from server
            print("🚀 Starting registration for username: \(username)")
            let options = try await apiService.beginRegistration(
                username: username, 
                displayName: displayName
            )
            
            print("📋 Registration options received")
            print("Challenge: \(options.challenge.prefix(20))...")
            print("RPID: \(options.rp.id)")
            print("User ID: \(options.user.id)")
            
            // Step 2: Create platform authenticator request
            let platformProvider = ASAuthorizationPlatformPublicKeyCredentialProvider(
                relyingPartyIdentifier: options.rp.id
            )
            
            guard let challengeData = options.challenge.base64URLDecode() else {
                throw WebAuthnError.invalidChallenge
            }
            
            guard let userIdData = options.user.id.base64URLDecode() else {
                throw WebAuthnError.invalidChallenge
            }
            
            let registrationRequest = platformProvider.createCredentialRegistrationRequest(
                challenge: challengeData,
                name: options.user.name,
                userID: userIdData
            )
            
            // Configure authenticator selection
            registrationRequest.displayName = options.user.displayName
            registrationRequest.userVerificationPreference = .preferred
            
            print("🔐 Created platform authenticator request")
            
            // Step 3: Prompt user for biometric authentication
            let authController = ASAuthorizationController(authorizationRequests: [registrationRequest])
            authController.delegate = self
            authController.presentationContextProvider = self
            
            let credential = try await withCheckedThrowingContinuation { continuation in
                self.registrationContinuation = continuation
                authController.performRequests()
            }
            
            print("✅ User completed biometric authentication")
            
            // Step 4: Send credential to server
            let result = try await apiService.finishRegistration(credential: credential)
            
            if result.success {
                currentUser = result.username
                print("🎉 Registration successful for user: \(result.username ?? "unknown")")
                await loadUserPasskeys()
            }
            
            return result
            
        } catch let error as WebAuthnError {
            errorMessage = error.localizedDescription
            print("❌ Registration failed: \(error.localizedDescription)")
            throw error
        } catch {
            let webAuthnError = WebAuthnError.unknown(error.localizedDescription)
            errorMessage = webAuthnError.localizedDescription
            print("❌ Registration failed: \(error.localizedDescription)")
            throw webAuthnError
        }
    }
    
    // MARK: - Authentication
    
    func authenticate(username: String? = nil) async throws -> AuthenticationResult {
        isLoading = true
        errorMessage = nil
        
        defer { isLoading = false }
        
        do {
            // Step 1: Get authentication options from server
            print("🚀 Starting authentication for username: \(username ?? "discoverable")")
            let options = try await apiService.beginAuthentication(username: username)
            
            print("📋 Authentication options received")
            print("Challenge: \(options.challenge.prefix(20))...")
            print("RPID: \(options.rpId ?? "nil")")
            
            // Step 2: Create platform authenticator request
            let platformProvider = ASAuthorizationPlatformPublicKeyCredentialProvider(
                relyingPartyIdentifier: options.rpId ?? "passkey-demo.local"
            )
            
            guard let challengeData = options.challenge.base64URLDecode() else {
                throw WebAuthnError.invalidChallenge
            }
            
            let assertionRequest = platformProvider.createCredentialAssertionRequest(challenge: challengeData)
            
            // Configure user verification
            assertionRequest.userVerificationPreference = .preferred
            
            // Set allowed credentials if provided (username-based auth)
            if let allowedCredentials = options.allowCredentials {
                assertionRequest.allowedCredentials = allowedCredentials.compactMap { cred in
                    guard let credentialIdData = cred.id.base64URLDecode() else { return nil }
                    return ASAuthorizationPlatformPublicKeyCredentialDescriptor(credentialID: credentialIdData)
                }
                print("🔑 Username-based auth with \(allowedCredentials.count) allowed credentials")
            } else {
                print("🔍 Discoverable credential authentication")
            }
            
            // Step 3: Prompt user for biometric authentication
            let authController = ASAuthorizationController(authorizationRequests: [assertionRequest])
            authController.delegate = self
            authController.presentationContextProvider = self
            
            let credential = try await withCheckedThrowingContinuation { continuation in
                self.authenticationContinuation = continuation
                authController.performRequests()
            }
            
            print("✅ User completed biometric authentication")
            
            // Step 4: Send credential to server
            let result = try await apiService.finishAuthentication(credential: credential)
            
            if result.success {
                currentUser = result.username
                print("🎉 Authentication successful for user: \(result.username ?? "unknown")")
                await loadUserPasskeys()
            }
            
            return result
            
        } catch let error as WebAuthnError {
            errorMessage = error.localizedDescription
            print("❌ Authentication failed: \(error.localizedDescription)")
            throw error
        } catch {
            let webAuthnError = WebAuthnError.unknown(error.localizedDescription)
            errorMessage = webAuthnError.localizedDescription
            print("❌ Authentication failed: \(error.localizedDescription)")
            throw webAuthnError
        }
    }
    
    // MARK: - User Management
    
    func loadUserPasskeys() async {
        do {
            userPasskeys = try await apiService.getUserPasskeys()
            print("📊 Loaded \(userPasskeys.count) passkeys for user")
        } catch {
            print("❌ Failed to load user passkeys: \(error.localizedDescription)")
            errorMessage = "Failed to load passkeys: \(error.localizedDescription)"
        }
    }
    
    func deletePasskey(_ passkey: UserPasskey) async throws {
        do {
            try await apiService.deletePasskey(credentialId: passkey.id)
            await loadUserPasskeys() // Refresh the list
            print("🗑️ Deleted passkey: \(passkey.id)")
        } catch {
            print("❌ Failed to delete passkey: \(error.localizedDescription)")
            throw error
        }
    }
    
    func logout() async {
        do {
            try await apiService.logout()
            currentUser = nil
            userPasskeys = []
            print("👋 User logged out")
        } catch {
            print("❌ Logout failed: \(error.localizedDescription)")
            // Even if logout fails on server, clear local state
            currentUser = nil
            userPasskeys = []
        }
    }
    
    // MARK: - Capability Check
    
    var isWebAuthnSupported: Bool {
        // WebAuthn with platform authenticators requires iOS 16.0+
        guard #available(iOS 16.0, *) else {
            return false
        }
        
        // Check if the device has biometric authentication or device passcode
        let context = LAContext()
        var error: NSError?
        
        // This checks if the device can use biometrics (Face ID/Touch ID) or device passcode
        let canEvaluatePolicy = context.canEvaluatePolicy(.deviceOwnerAuthentication, error: &error)
        
        return canEvaluatePolicy
    }
    
    // MARK: - Private Properties for Async/Await Bridge
    
    private var registrationContinuation: CheckedContinuation<RegistrationCredential, Error>?
    private var authenticationContinuation: CheckedContinuation<AuthenticationCredential, Error>?
}

// MARK: - ASAuthorizationControllerDelegate

extension WebAuthnService: ASAuthorizationControllerDelegate {
    func authorizationController(controller: ASAuthorizationController, didCompleteWithAuthorization authorization: ASAuthorization) {
        
        if let platformCredential = authorization.credential as? ASAuthorizationPlatformPublicKeyCredentialRegistration {
            // Handle registration
            handleRegistrationCredential(platformCredential)
        } else if let platformCredential = authorization.credential as? ASAuthorizationPlatformPublicKeyCredentialAssertion {
            // Handle authentication
            handleAuthenticationCredential(platformCredential)
        }
    }
    
    func authorizationController(controller: ASAuthorizationController, didCompleteWithError error: Error) {
        print("❌ Authorization failed: \(error.localizedDescription)")
        
        let webAuthnError: WebAuthnError
        
        if let authError = error as? ASAuthorizationError {
            switch authError.code {
            case .canceled:
                webAuthnError = .userCancel
            case .unknown:
                webAuthnError = .unknown(authError.localizedDescription)
            case .invalidResponse:
                webAuthnError = .apiError("Invalid response from authenticator")
            case .notHandled:
                webAuthnError = .notSupported
            case .failed:
                webAuthnError = .unknown("Authentication failed")
            case .notInteractive:
                webAuthnError = .unknown("Authentication requires user interaction")
            @unknown default:
                webAuthnError = .unknown(authError.localizedDescription)
            }
        } else {
            webAuthnError = .unknown(error.localizedDescription)
        }
        
        registrationContinuation?.resume(throwing: webAuthnError)
        authenticationContinuation?.resume(throwing: webAuthnError)
        
        registrationContinuation = nil
        authenticationContinuation = nil
    }
    
    private func handleRegistrationCredential(_ credential: ASAuthorizationPlatformPublicKeyCredentialRegistration) {
        let registrationCredential = RegistrationCredential(
            id: credential.credentialID.base64URLEncode(),
            rawId: credential.credentialID.base64URLEncode(),
            type: "public-key",
            response: RegistrationCredential.AuthenticatorAttestationResponse(
                attestationObject: credential.rawAttestationObject?.base64URLEncode() ?? "",
                clientDataJSON: credential.rawClientDataJSON.base64URLEncode(),
                transports: ["internal", "hybrid"] // iOS typically supports internal and hybrid
            )
        )
        
        registrationContinuation?.resume(returning: registrationCredential)
        registrationContinuation = nil
    }
    
    private func handleAuthenticationCredential(_ credential: ASAuthorizationPlatformPublicKeyCredentialAssertion) {
        let authenticationCredential = AuthenticationCredential(
            id: credential.credentialID.base64URLEncode(),
            rawId: credential.credentialID.base64URLEncode(),
            type: "public-key",
            response: AuthenticationCredential.AuthenticatorAssertionResponse(
                authenticatorData: credential.rawAuthenticatorData.base64URLEncode(),
                clientDataJSON: credential.rawClientDataJSON.base64URLEncode(),
                signature: credential.signature.base64URLEncode(),
                userHandle: credential.userID?.base64URLEncode()
            )
        )
        
        authenticationContinuation?.resume(returning: authenticationCredential)
        authenticationContinuation = nil
    }
}

// MARK: - ASAuthorizationControllerPresentationContextProviding

extension WebAuthnService: ASAuthorizationControllerPresentationContextProviding {
    func presentationAnchor(for controller: ASAuthorizationController) -> ASPresentationAnchor {
        // Return the key window for presenting the authentication UI
        guard let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
              let window = windowScene.windows.first else {
            return ASPresentationAnchor()
        }
        return window
    }
}