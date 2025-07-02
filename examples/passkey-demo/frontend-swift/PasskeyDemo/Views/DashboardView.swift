import SwiftUI

struct DashboardView: View {
    @ObservedObject var webAuthnService: WebAuthnService
    let onLogout: () -> Void
    
    @State private var showingDeleteAlert = false
    @State private var passkeyToDelete: UserPasskey?
    @State private var showingError = false
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack(spacing: 24) {
                    // Welcome Header
                    VStack(spacing: 12) {
                        Image(systemName: "checkmark.shield.fill")
                            .font(.system(size: 60))
                            .foregroundColor(.green)
                        
                        Text("Welcome!")
                            .font(.largeTitle)
                            .fontWeight(.bold)
                        
                        if let username = webAuthnService.currentUser {
                            Text("Signed in as \(username)")
                                .font(.body)
                                .foregroundColor(.secondary)
                        }
                        
                        Text("You've successfully authenticated with your passkey")
                            .font(.body)
                            .foregroundColor(.secondary)
                            .multilineTextAlignment(.center)
                    }
                    
                    // User Passkeys Section
                    VStack(alignment: .leading, spacing: 16) {
                        HStack {
                            Text("Your Passkeys")
                                .font(.title2)
                                .fontWeight(.semibold)
                            
                            Spacer()
                            
                            Button(action: refreshPasskeys) {
                                Image(systemName: "arrow.clockwise")
                            }
                            .disabled(webAuthnService.isLoading)
                        }
                        
                        if webAuthnService.isLoading {
                            ProgressView("Loading passkeys...")
                                .frame(maxWidth: .infinity)
                                .padding()
                        } else if webAuthnService.userPasskeys.isEmpty {
                            EmptyPasskeysView()
                        } else {
                            LazyVStack(spacing: 12) {
                                ForEach(webAuthnService.userPasskeys) { passkey in
                                    PasskeyCard(
                                        passkey: passkey,
                                        onDelete: { deletePasskey(passkey) }
                                    )
                                }
                            }
                        }
                    }
                    
                    // Cross-Platform Info
                    CrossPlatformInfoCard()
                    
                    // Actions
                    VStack(spacing: 12) {
                        Button("Sign Out") {
                            handleLogout()
                        }
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.red)
                        .foregroundColor(.white)
                        .cornerRadius(10)
                    }
                }
                .padding()
            }
            .navigationTitle("Dashboard")
            .navigationBarTitleDisplayMode(.inline)
            .refreshable {
                await webAuthnService.loadUserPasskeys()
            }
        }
        .onAppear {
            Task {
                await webAuthnService.loadUserPasskeys()
            }
        }
        .alert("Delete Passkey", isPresented: $showingDeleteAlert) {
            Button("Cancel", role: .cancel) {
                passkeyToDelete = nil
            }
            Button("Delete", role: .destructive) {
                if let passkey = passkeyToDelete {
                    confirmDeletePasskey(passkey)
                }
            }
        } message: {
            if let passkey = passkeyToDelete {
                Text("Are you sure you want to delete the passkey \"\(passkey.displayName)\"?\n\nThis will remove it from the server, but you may need to manually delete it from your device's keychain as well.")
            }
        }
        .alert("Error", isPresented: $showingError) {
            Button("OK") { showingError = false }
        } message: {
            Text(webAuthnService.errorMessage ?? "Unknown error occurred")
        }
    }
    
    private func refreshPasskeys() {
        Task {
            await webAuthnService.loadUserPasskeys()
        }
    }
    
    private func deletePasskey(_ passkey: UserPasskey) {
        passkeyToDelete = passkey
        showingDeleteAlert = true
    }
    
    private func confirmDeletePasskey(_ passkey: UserPasskey) {
        Task {
            do {
                try await webAuthnService.deletePasskey(passkey)
            } catch {
                showingError = true
            }
        }
        passkeyToDelete = nil
    }
    
    private func handleLogout() {
        Task {
            await webAuthnService.logout()
            onLogout()
        }
    }
}

// MARK: - Supporting Views

struct PasskeyCard: View {
    let passkey: UserPasskey
    let onDelete: () -> Void
    
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    HStack {
                        Image(systemName: passkey.isCloudSynced ? "icloud.fill" : "key.fill")
                            .foregroundColor(passkey.isCloudSynced ? .blue : .secondary)
                        
                        Text(passkey.displayName)
                            .font(.headline)
                    }
                    
                    Text("Created: \(formatDate(passkey.createdDate))")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    
                    if let lastUsed = passkey.lastUsedDate {
                        Text("Last used: \(formatDate(lastUsed))")
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }
                }
                
                Spacer()
                
                Button(action: onDelete) {
                    Image(systemName: "trash")
                        .foregroundColor(.red)
                }
            }
            
            // Passkey Details
            VStack(alignment: .leading, spacing: 4) {
                DetailRow(
                    icon: "checkmark.shield",
                    text: passkey.userVerified ? "User verified" : "No user verification",
                    color: passkey.userVerified ? .green : .orange
                )
                
                DetailRow(
                    icon: passkey.isCloudSynced ? "icloud" : "device",
                    text: passkey.isCloudSynced ? "Synced via iCloud" : "Device only",
                    color: passkey.isCloudSynced ? .blue : .secondary
                )
                
                if !passkey.transports.isEmpty {
                    DetailRow(
                        icon: "antenna.radiowaves.left.and.right",
                        text: "Transports: \(passkey.transports.joined(separator: ", "))",
                        color: .secondary
                    )
                }
            }
        }
        .padding()
        .background(Color(.systemGray6))
        .cornerRadius(12)
    }
    
    private func formatDate(_ date: Date?) -> String {
        guard let date = date else { return "Unknown" }
        
        let formatter = RelativeDateTimeFormatter()
        formatter.unitsStyle = .abbreviated
        return formatter.localizedString(for: date, relativeTo: Date())
    }
}

struct DetailRow: View {
    let icon: String
    let text: String
    let color: Color
    
    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: icon)
                .foregroundColor(color)
                .frame(width: 16)
            
            Text(text)
                .font(.caption)
                .foregroundColor(.secondary)
        }
    }
}

struct EmptyPasskeysView: View {
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "key.slash")
                .font(.system(size: 40))
                .foregroundColor(.secondary)
            
            Text("No passkeys found")
                .font(.headline)
                .foregroundColor(.secondary)
            
            Text("This might happen if you haven't registered any passkeys yet, or if there was an error loading them.")
                .font(.caption)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
        }
        .padding()
        .background(Color(.systemGray6))
        .cornerRadius(12)
    }
}

struct CrossPlatformInfoCard: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Cross-Platform Passkeys")
                .font(.headline)
            
            VStack(alignment: .leading, spacing: 8) {
                InfoBullet(
                    icon: "iphone",
                    text: "Your passkeys work across all your Apple devices via iCloud Keychain"
                )
                InfoBullet(
                    icon: "globe",
                    text: "Use the same passkeys on the web at \(getCurrentWebURL())"
                )
                InfoBullet(
                    icon: "androidlogo",
                    text: "Coming soon: Android app with Google Password Manager sync"
                )
            }
            
            Text("Passkey Management")
                .font(.headline)
                .padding(.top, 8)
            
            VStack(alignment: .leading, spacing: 8) {
                InfoBullet(
                    icon: "gear",
                    text: "To remove passkeys from your device, go to Settings → Passwords"
                )
                InfoBullet(
                    icon: "trash",
                    text: "Deleting here only removes the server record"
                )
            }
        }
        .padding()
        .background(Color.blue.opacity(0.1))
        .cornerRadius(12)
    }
    
    private func getCurrentWebURL() -> String {
        // Get the current ngrok URL from APIConfiguration
        if let ngrokURL = APIConfiguration.ngrokURL {
            return ngrokURL
        } else {
            return "http://localhost:8080"
        }
    }
}

struct InfoBullet: View {
    let icon: String
    let text: String
    
    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            Image(systemName: icon)
                .foregroundColor(.blue)
                .frame(width: 20)
            
            Text(text)
                .font(.caption)
                .foregroundColor(.secondary)
        }
    }
}

#Preview {
    DashboardView(
        webAuthnService: {
            let service = WebAuthnService()
            service.currentUser = "testuser"
            return service
        }(),
        onLogout: {}
    )
}