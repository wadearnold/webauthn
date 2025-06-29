import SwiftUI

struct RegistrationView: View {
    @ObservedObject var webAuthnService: WebAuthnService
    let onRegistered: () -> Void
    let onBackToLogin: () -> Void
    
    @State private var username = ""
    @State private var displayName = ""
    @State private var showingError = false
    @State private var usernameError: String?
    
    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Header
                VStack(spacing: 12) {
                    Image(systemName: "touchid")
                        .font(.system(size: 60))
                        .foregroundColor(.blue)
                    
                    Text("Create Your Passkey")
                        .font(.largeTitle)
                        .fontWeight(.bold)
                    
                    Text("Register with a username to create your first passkey")
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
                        // Demo Notice
                        DemoNoticeCard()
                        
                        // Registration Form
                        VStack(spacing: 16) {
                            // Username Field
                            VStack(alignment: .leading, spacing: 8) {
                                Text("Username *")
                                    .font(.headline)
                                
                                TextField("3-30 chars: letters, numbers, . _ -", text: $username)
                                    .textFieldStyle(.roundedBorder)
                                    .autocapitalization(.none)
                                    .disableAutocorrection(true)
                                    .onChange(of: username) { _ in
                                        validateUsername()
                                    }
                                
                                if let error = usernameError {
                                    Text("⚠️ \(error)")
                                        .font(.caption)
                                        .foregroundColor(.red)
                                } else if !username.isEmpty && usernameError == nil {
                                    Text("✅ Username looks good!")
                                        .font(.caption)
                                        .foregroundColor(.green)
                                }
                            }
                            
                            // Display Name Field
                            VStack(alignment: .leading, spacing: 8) {
                                Text("Display Name")
                                    .font(.headline)
                                
                                TextField("Your full name (optional)", text: $displayName)
                                    .textFieldStyle(.roundedBorder)
                            }
                        }
                        
                        // Create Passkey Button
                        Button(action: handleRegistration) {
                            HStack {
                                if webAuthnService.isLoading {
                                    ProgressView()
                                        .scaleEffect(0.8)
                                } else {
                                    Image(systemName: "person.crop.circle.badge.plus")
                                }
                                Text(webAuthnService.isLoading ? "Creating Passkey..." : "Create Passkey")
                            }
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(isFormValid ? Color.blue : Color.gray)
                            .foregroundColor(.white)
                            .cornerRadius(10)
                        }
                        .disabled(!isFormValid || webAuthnService.isLoading)
                        
                        // Back to Login
                        Button("Already have a passkey? Sign In") {
                            onBackToLogin()
                        }
                        .foregroundColor(.blue)
                        
                        // Registration Info
                        RegistrationInfoCard()
                    }
                }
            }
            .padding()
        }
        .alert("Registration Failed", isPresented: $showingError) {
            Button("Try Again") { showingError = false }
        } message: {
            Text(webAuthnService.errorMessage ?? "Unknown error occurred")
        }
        .onAppear {
            // Clear any previous errors
            webAuthnService.errorMessage = nil
        }
    }
    
    private var isFormValid: Bool {
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
    
    private func handleRegistration() {
        Task {
            do {
                let result = try await webAuthnService.register(
                    username: username.trimmingCharacters(in: .whitespacesAndNewlines),
                    displayName: displayName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? 
                        username.trimmingCharacters(in: .whitespacesAndNewlines) : 
                        displayName.trimmingCharacters(in: .whitespacesAndNewlines)
                )
                
                if result.success {
                    onRegistered()
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

struct DemoNoticeCard: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Demo Note")
                .font(.headline)
                .foregroundColor(.orange)
            
            Text("This is a demonstration of WebAuthn passkeys. Your credentials are stored only in memory and will be lost when the server restarts.")
                .font(.caption)
                .foregroundColor(.secondary)
        }
        .padding()
        .background(Color.orange.opacity(0.1))
        .cornerRadius(8)
    }
}

struct RegistrationInfoCard: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Username Requirements")
                .font(.headline)
            
            VStack(alignment: .leading, spacing: 4) {
                InfoRow(text: "3-30 characters long")
                InfoRow(text: "Letters, numbers, dots (.), hyphens (-), and underscores (_) only")
                InfoRow(text: "Cannot start or end with dots, hyphens, or underscores")
                InfoRow(text: "Examples: john.doe, user_123, test-user")
            }
            
            Text("What happens next?")
                .font(.headline)
                .padding(.top, 8)
            
            VStack(alignment: .leading, spacing: 4) {
                InfoRow(text: "Your device will prompt you to create a passkey")
                InfoRow(text: "Use Face ID, Touch ID, or device passcode")
                InfoRow(text: "Your passkey will be saved for future logins")
                InfoRow(text: "Passkeys sync via iCloud Keychain across your devices")
            }
        }
        .padding()
        .background(Color.blue.opacity(0.1))
        .cornerRadius(8)
    }
}

struct InfoRow: View {
    let text: String
    
    var body: some View {
        HStack(alignment: .top, spacing: 8) {
            Text("•")
                .foregroundColor(.blue)
            Text(text)
                .font(.caption)
                .foregroundColor(.secondary)
        }
    }
}

struct ErrorCard: View {
    let title: String
    let message: String
    
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.headline)
                .foregroundColor(.red)
            
            Text(message)
                .font(.body)
                .foregroundColor(.secondary)
        }
        .padding()
        .background(Color.red.opacity(0.1))
        .cornerRadius(8)
    }
}

#Preview {
    RegistrationView(
        webAuthnService: WebAuthnService(),
        onRegistered: {},
        onBackToLogin: {}
    )
}