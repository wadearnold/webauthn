import SwiftUI

struct AuthenticationView: View {
    @ObservedObject var webAuthnService: WebAuthnService
    let onAuthenticated: () -> Void
    let onShowRegistration: () -> Void
    
    @State private var username = ""
    @State private var showingError = false
    @State private var usernameError: String?
    @State private var authenticationMode: AuthMode = .discoverable
    
    enum AuthMode {
        case discoverable
        case username
    }
    
    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Header
                VStack(spacing: 12) {
                    Image(systemName: "faceid")
                        .font(.system(size: 60))
                        .foregroundColor(.blue)
                    
                    Text("Sign In with Passkey")
                        .font(.largeTitle)
                        .fontWeight(.bold)
                    
                    Text("Use your passkey to sign in securely")
                        .font(.body)
                        .foregroundColor(.secondary)
                        .multilineTextAlignment(.center)
                }
                
                // WebAuthn Support Check
                if !webAuthnService.isWebAuthnSupported {
                    ErrorCard(
                        title: "WebAuthn Not Supported",
                        message: "This device doesn't support passkeys. Please use a device with iOS 16+ and Touch ID or Face ID."
                    )
                } else {
                    VStack(spacing: 20) {
                        // Discoverable Login (Passwordless)
                        VStack(spacing: 12) {
                            Button(action: handleDiscoverableLogin) {
                                HStack {
                                    if webAuthnService.isLoading && authenticationMode == .discoverable {
                                        ProgressView()
                                            .scaleEffect(0.8)
                                    } else {
                                        Image(systemName: "key.fill")
                                    }
                                    Text(webAuthnService.isLoading && authenticationMode == .discoverable ? 
                                         "Signing in..." : "Sign in with Passkey")
                                }
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(Color.blue)
                                .foregroundColor(.white)
                                .cornerRadius(10)
                            }
                            .disabled(webAuthnService.isLoading)
                            
                            Text("No username required - your device will show available passkeys")
                                .font(.caption)
                                .foregroundColor(.secondary)
                                .multilineTextAlignment(.center)
                        }
                        
                        // Divider
                        HStack {
                            Rectangle()
                                .frame(height: 1)
                                .foregroundColor(.gray.opacity(0.3))
                            
                            Text("or sign in with username")
                                .font(.caption)
                                .foregroundColor(.secondary)
                                .padding(.horizontal, 8)
                            
                            Rectangle()
                                .frame(height: 1)
                                .foregroundColor(.gray.opacity(0.3))
                        }
                        
                        // Username-based Login
                        VStack(spacing: 16) {
                            VStack(alignment: .leading, spacing: 8) {
                                Text("Username")
                                    .font(.headline)
                                
                                TextField("Enter your username", text: $username)
                                    .textFieldStyle(.roundedBorder)
                                    .autocapitalization(.none)
                                    .disableAutocorrection(true)
                                    .onChange(of: username) { _ in
                                        authenticationMode = .username
                                        validateUsername()
                                    }
                                
                                if let error = usernameError {
                                    Text("⚠️ \(error)")
                                        .font(.caption)
                                        .foregroundColor(.red)
                                }
                            }
                            
                            Button(action: handleUsernameLogin) {
                                HStack {
                                    if webAuthnService.isLoading && authenticationMode == .username {
                                        ProgressView()
                                            .scaleEffect(0.8)
                                    } else {
                                        Image(systemName: "person.fill")
                                    }
                                    Text(webAuthnService.isLoading && authenticationMode == .username ? 
                                         "Signing in..." : "Sign in with Username")
                                }
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(isUsernameFormValid ? Color.secondary : Color.gray)
                                .foregroundColor(.white)
                                .cornerRadius(10)
                            }
                            .disabled(!isUsernameFormValid || webAuthnService.isLoading)
                        }
                        
                        // Divider
                        HStack {
                            Rectangle()
                                .frame(height: 1)
                                .foregroundColor(.gray.opacity(0.3))
                            
                            Text("don't have a passkey?")
                                .font(.caption)
                                .foregroundColor(.secondary)
                                .padding(.horizontal, 8)
                            
                            Rectangle()
                                .frame(height: 1)
                                .foregroundColor(.gray.opacity(0.3))
                        }
                        
                        // Registration Link
                        Button("Create New Passkey") {
                            onShowRegistration()
                        }
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.secondary.opacity(0.2))
                        .foregroundColor(.blue)
                        .cornerRadius(10)
                        .disabled(webAuthnService.isLoading)
                        
                        // Demo Features Info
                        DemoFeaturesCard()
                    }
                }
            }
            .padding()
        }
        .alert("Sign In Failed", isPresented: $showingError) {
            Button("Try Again") { showingError = false }
        } message: {
            Text(webAuthnService.errorMessage ?? "Unknown error occurred")
        }
        .onAppear {
            // Clear any previous errors
            webAuthnService.errorMessage = nil
        }
    }
    
    private var isUsernameFormValid: Bool {
        return !username.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty && 
               usernameError == nil &&
               webAuthnService.isWebAuthnSupported
    }
    
    private func validateUsername() {
        let trimmed = username.trimmingCharacters(in: .whitespacesAndNewlines)
        
        if trimmed.isEmpty {
            usernameError = nil
            return
        }
        
        if trimmed.count < 3 {
            usernameError = "Username must be at least 3 characters long"
            return
        }
        
        if trimmed.count > 30 {
            usernameError = "Username must be no more than 30 characters long"
            return
        }
        
        let allowedCharacters = CharacterSet.alphanumerics.union(CharacterSet(charactersIn: "._-"))
        if trimmed.rangeOfCharacter(from: allowedCharacters.inverted) != nil {
            usernameError = "Username can only contain letters, numbers, dots, hyphens, and underscores"
            return
        }
        
        if trimmed.hasPrefix(".") || trimmed.hasPrefix("-") || trimmed.hasPrefix("_") ||
           trimmed.hasSuffix(".") || trimmed.hasSuffix("-") || trimmed.hasSuffix("_") {
            usernameError = "Username cannot start or end with dots, hyphens, or underscores"
            return
        }
        
        usernameError = nil
    }
    
    private func handleDiscoverableLogin() {
        authenticationMode = .discoverable
        Task {
            do {
                let result = try await webAuthnService.authenticate()
                if result.success {
                    onAuthenticated()
                } else {
                    showingError = true
                }
            } catch {
                showingError = true
            }
        }
    }
    
    private func handleUsernameLogin() {
        authenticationMode = .username
        Task {
            do {
                let result = try await webAuthnService.authenticate(
                    username: username.trimmingCharacters(in: .whitespacesAndNewlines)
                )
                if result.success {
                    onAuthenticated()
                } else {
                    showingError = true
                }
            } catch {
                showingError = true
            }
        }
    }
}

// MARK: - Supporting Views

struct DemoFeaturesCard: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Demo Features")
                .font(.headline)
            
            VStack(alignment: .leading, spacing: 4) {
                FeatureRow(
                    icon: "key.fill",
                    title: "Passwordless",
                    description: "Click \"Sign in with Passkey\" for true passwordless authentication"
                )
                FeatureRow(
                    icon: "person.fill",
                    title: "Username-based",
                    description: "Enter username first, then authenticate with passkey"
                )
                FeatureRow(
                    icon: "icloud.fill",
                    title: "Multi-device",
                    description: "Your passkeys work across devices when synced via iCloud"
                )
            }
        }
        .padding()
        .background(Color.blue.opacity(0.1))
        .cornerRadius(8)
    }
}

struct FeatureRow: View {
    let icon: String
    let title: String
    let description: String
    
    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            Image(systemName: icon)
                .foregroundColor(.blue)
                .frame(width: 20)
            
            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.caption)
                    .fontWeight(.semibold)
                
                Text(description)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
        }
    }
}

#Preview {
    AuthenticationView(
        webAuthnService: WebAuthnService(),
        onAuthenticated: {},
        onShowRegistration: {}
    )
}