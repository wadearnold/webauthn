import SwiftUI

@main
struct PasskeyDemoApp: App {
    init() {
        // Log configuration at startup
        logAppConfiguration()
    }
    
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
    
    private func logAppConfiguration() {
        print("🚀 PasskeyDemo iOS App Starting...")
        print(APIService.shared.getConfigurationStatus())
        print("📱 Ready for cross-platform passkey authentication")
    }
}