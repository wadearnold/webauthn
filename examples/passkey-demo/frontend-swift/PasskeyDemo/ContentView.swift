import SwiftUI

struct ContentView: View {
    @StateObject private var webAuthnService = WebAuthnService()
    @State private var currentView: AppView = .authentication
    
    enum AppView {
        case authentication
        case registration
        case dashboard
    }
    
    var body: some View {
        NavigationView {
            Group {
                switch currentView {
                case .authentication:
                    AuthenticationView(
                        webAuthnService: webAuthnService,
                        onAuthenticated: { 
                            currentView = .dashboard
                        },
                        onShowRegistration: {
                            currentView = .registration
                        }
                    )
                case .registration:
                    RegistrationView(
                        webAuthnService: webAuthnService,
                        onRegistered: {
                            currentView = .dashboard
                        },
                        onBackToLogin: {
                            currentView = .authentication
                        }
                    )
                case .dashboard:
                    DashboardView(
                        webAuthnService: webAuthnService,
                        onLogout: {
                            currentView = .authentication
                        }
                    )
                }
            }
            .navigationTitle("WebAuthn Passkey Demo")
            .navigationBarTitleDisplayMode(.inline)
        }
        .navigationViewStyle(StackNavigationViewStyle())
    }
}

#Preview {
    ContentView()
}