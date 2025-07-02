import Foundation
import Security

// MARK: - API Configuration

struct APIConfiguration {
    /// Detect ngrok URL from bundled config file
    static var ngrokURL: String? {
        // Read from bundled config file (simple and reliable)
        if let bundledURL = getBundledNgrokURL() {
            print("🔧 Using ngrok URL from config: \(bundledURL)")
            return bundledURL
        }
        
        print("⚠️ No ngrok URL found in config file, using localhost mode")
        return nil
    }
    
    /// Read ngrok URL from bundled config file (most reliable)
    private static func getBundledNgrokURL() -> String? {
        guard let path = Bundle.main.path(forResource: "ngrok-config", ofType: "json") else {
            print("🔍 No ngrok-config.json found in app bundle")
            print("💡 To enable cross-platform mode:")
            print("   1. Copy: cp PasskeyDemo/ngrok-config.json.template PasskeyDemo/ngrok-config.json")
            print("   2. Edit ngrok-config.json with your ngrok URL")
            print("   3. Rebuild the app")
            return nil
        }
        
        guard let data = try? Data(contentsOf: URL(fileURLWithPath: path)) else {
            print("⚠️ Could not read ngrok-config.json from app bundle")
            return nil
        }
        
        do {
            if let json = try JSONSerialization.jsonObject(with: data) as? [String: Any],
               let ngrokURL = json["ngrok_url"] as? String,
               !ngrokURL.isEmpty && ngrokURL != "https://your-ngrok-url.ngrok-free.app" {
                print("✅ Found valid ngrok URL in config: \(ngrokURL)")
                return ngrokURL
            } else {
                print("⚠️ ngrok-config.json contains placeholder or empty URL")
                print("💡 Edit PasskeyDemo/ngrok-config.json with your actual ngrok URL")
            }
        } catch {
            print("⚠️ Error parsing ngrok-config.json: \(error)")
        }
        
        return nil
    }
}

class APIService: ObservableObject {
    static let shared = APIService()
    
    // ngrok-based cross-platform passkey configuration
    // Automatically detects ngrok tunnel or falls back to localhost
    
    private let baseURL: String = {
        // Check for ngrok URL in app configuration
        if let ngrokURL = APIConfiguration.ngrokURL {
            let apiURL = "\(ngrokURL)/api"
            print("🌐 iOS App configured for cross-platform mode: \(apiURL)")
            return apiURL
        }
        
        // Fallback to localhost for development
        print("🏠 iOS App using localhost mode: http://localhost:8080/api")
        return "http://localhost:8080/api"
    }()
    
    private let session: URLSession
    
    
    init() {
        let config = URLSessionConfiguration.default
        config.httpCookieAcceptPolicy = .always
        config.httpShouldSetCookies = true
        
        // For development: Create session that accepts development certificates
        // ngrok provides trusted certificates, but we keep support for localhost development
        self.session = URLSession(
            configuration: config,
            delegate: DevelopmentCertificateDelegate(),
            delegateQueue: nil
        )
    }
    
    // MARK: - Generic Request Method
    
    private func makeRequest<T: Codable>(
        endpoint: String,
        method: HTTPMethod = .GET,
        body: Codable? = nil
    ) async throws -> T {
        guard let url = URL(string: "\(baseURL)\(endpoint)") else {
            throw WebAuthnError.networkError("Invalid URL")
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = method.rawValue
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        if let body = body {
            do {
                request.httpBody = try JSONEncoder().encode(body)
            } catch {
                throw WebAuthnError.encodingError
            }
        }
        
        do {
            let (data, response) = try await session.data(for: request)
            
            guard let httpResponse = response as? HTTPURLResponse else {
                throw WebAuthnError.networkError("Invalid response")
            }
            
            if httpResponse.statusCode >= 400 {
                // Try to decode error response
                if let errorResponse = try? JSONDecoder().decode(APIError.self, from: data) {
                    throw WebAuthnError.apiError(errorResponse.error)
                } else {
                    throw WebAuthnError.networkError("HTTP \(httpResponse.statusCode)")
                }
            }
            
            do {
                return try JSONDecoder().decode(T.self, from: data)
            } catch {
                print("Decoding error: \(error)")
                print("Response data: \(String(data: data, encoding: .utf8) ?? "nil")")
                throw WebAuthnError.encodingError
            }
        } catch let error as WebAuthnError {
            throw error
        } catch {
            throw WebAuthnError.networkError(error.localizedDescription)
        }
    }
    
    // MARK: - Registration API
    
    func beginRegistration(username: String, displayName: String) async throws -> RegistrationOptions {
        let request = RegistrationRequest(username: username, displayName: displayName)
        let response: RegistrationOptionsResponse = try await makeRequest(endpoint: "/register/begin", method: .POST, body: request)
        return response.publicKey
    }
    
    func finishRegistration(credential: RegistrationCredential) async throws -> RegistrationResult {
        return try await makeRequest(endpoint: "/register/finish", method: .POST, body: credential)
    }
    
    // MARK: - Authentication API
    
    func beginAuthentication(username: String? = nil) async throws -> AuthenticationOptions {
        let request = AuthenticationRequest(username: username)
        let response: AuthenticationOptionsResponse = try await makeRequest(endpoint: "/login/begin", method: .POST, body: request)
        return response.publicKey
    }
    
    func finishAuthentication(credential: AuthenticationCredential) async throws -> AuthenticationResult {
        return try await makeRequest(endpoint: "/login/finish", method: .POST, body: credential)
    }
    
    // MARK: - User Management API
    
    func getUserPasskeys() async throws -> [UserPasskey] {
        return try await makeRequest(endpoint: "/user/passkeys", method: .GET)
    }
    
    func deletePasskey(credentialId: String) async throws {
        let _: EmptyResponse = try await makeRequest(
            endpoint: "/user/passkeys/\(credentialId)", 
            method: .DELETE
        )
    }
    
    func logout() async throws {
        let _: EmptyResponse = try await makeRequest(endpoint: "/logout", method: .POST)
    }
    
    // MARK: - Health Check
    
    func healthCheck() async throws -> HealthResponse {
        return try await makeRequest(endpoint: "/health", method: .GET)
    }
    
    /// Get current API base URL for debugging
    func getCurrentBaseURL() -> String {
        return baseURL
    }
    
    /// Get current configuration status for debugging
    func getConfigurationStatus() -> String {
        let ngrokURL = APIConfiguration.ngrokURL
        let mode = ngrokURL != nil ? "Cross-Platform (ngrok)" : "Local Development"
        
        if let ngrokURL = ngrokURL {
            return """
            🔧 iOS App Configuration:
            Mode: \(mode)
            Base URL: \(baseURL)
            Ngrok URL: \(ngrokURL)
            Status: ✅ Ready for cross-platform passkey sharing
            """
        } else {
            return """
            🔧 iOS App Configuration:
            Mode: \(mode)
            Base URL: \(baseURL)
            Ngrok URL: Not configured
            Status: 🏠 Localhost mode - passkeys limited to this device
            """
        }
    }
}

// MARK: - HTTP Method Enum

enum HTTPMethod: String {
    case GET = "GET"
    case POST = "POST"
    case PUT = "PUT"
    case DELETE = "DELETE"
}

// MARK: - Helper Response Types

struct EmptyResponse: Codable {}

struct HealthResponse: Codable {
    let status: String
    let time: String
}

// MARK: - URL Session Extensions

extension URLSession {
    /// Helper method for debugging network requests
    func debugRequest(_ request: URLRequest) {
        print("🌐 API Request: \(request.httpMethod ?? "GET") \(request.url?.absoluteString ?? "unknown")")
        if let body = request.httpBody,
           let bodyString = String(data: body, encoding: .utf8) {
            print("📤 Request Body: \(bodyString)")
        }
        if let headers = request.allHTTPHeaderFields {
            print("📋 Headers: \(headers)")
        }
    }
}

// MARK: - Base64URL Encoding Helpers

extension Data {
    func urlSafeBase64EncodedString() -> String {
        return base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }
}

extension String {
    func urlSafeBase64Decoded() -> Data? {
        var base64 = self
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")
        
        while base64.count % 4 != 0 {
            base64 += "="
        }
        
        return Data(base64Encoded: base64)
    }
}

// MARK: - Development Certificate Delegate

class DevelopmentCertificateDelegate: NSObject, URLSessionDelegate {
    func urlSession(
        _ session: URLSession,
        didReceive challenge: URLAuthenticationChallenge,
        completionHandler: @escaping (URLSession.AuthChallengeDisposition, URLCredential?) -> Void
    ) {
        print("🔒 Certificate challenge for host: \(challenge.protectionSpace.host)")
        print("🔒 Authentication method: \(challenge.protectionSpace.authenticationMethod)")
        
        // Accept certificates for localhost development only
        // ngrok domains use trusted certificates and don't need custom handling
        let developmentHosts = ["localhost", "127.0.0.1"]
        
        if developmentHosts.contains(challenge.protectionSpace.host) {
            print("🔒 Accepting development certificate for: \(challenge.protectionSpace.host)")
            
            // Check if this is a server trust challenge
            guard challenge.protectionSpace.authenticationMethod == NSURLAuthenticationMethodServerTrust else {
                print("❌ Not a server trust challenge, method: \(challenge.protectionSpace.authenticationMethod)")
                completionHandler(.performDefaultHandling, nil)
                return
            }
            
            // Get server trust
            guard let serverTrust = challenge.protectionSpace.serverTrust else {
                print("❌ No server trust found")
                completionHandler(.performDefaultHandling, nil)
                return
            }
            
            // For development: Evaluate and accept the certificate
            // First, try to evaluate the trust normally using modern API
            var error: CFError?
            let isValid = SecTrustEvaluateWithError(serverTrust, &error)
            
            if isValid {
                print("🔒 Trust evaluation: Certificate is valid")
            } else if let error = error {
                let errorDescription = CFErrorCopyDescription(error) as String
                print("🔒 Trust evaluation failed: \(errorDescription)")
            } else {
                print("🔒 Trust evaluation failed: Unknown error")
            }
            
            // For development, accept the certificate regardless of evaluation result
            // This is safe for localhost development certificates
            let credential = URLCredential(trust: serverTrust)
            print("✅ Accepting localhost certificate for development (bypassing validation)")
            
            // Use the credential and tell iOS to accept it
            completionHandler(.useCredential, credential)
            return
        }
        
        // For all other hosts, use default handling
        print("🔒 Using default handling for host: \(challenge.protectionSpace.host)")
        completionHandler(.performDefaultHandling, nil)
    }
    
    // Additional delegate method for more granular control
    func urlSession(
        _ session: URLSession,
        task: URLSessionTask,
        didReceive challenge: URLAuthenticationChallenge,
        completionHandler: @escaping (URLSession.AuthChallengeDisposition, URLCredential?) -> Void
    ) {
        print("🔒 Task-level certificate challenge for: \(challenge.protectionSpace.host)")
        
        // Delegate to the session-level method
        self.urlSession(session, didReceive: challenge, completionHandler: completionHandler)
    }
}

