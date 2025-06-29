import Foundation

class APIService: ObservableObject {
    static let shared = APIService()
    
    // Cross-platform passkey compatibility configuration
    // IMPORTANT: The backend MUST use passkey-demo.local as RPID regardless of how we connect
    // This temporary workaround resolves iOS Simulator DNS issues while maintaining
    // the same RPID for true cross-platform passkey sharing
    
    private let baseURL: String = {
        #if targetEnvironment(simulator)
        // iOS Simulator workaround: Use HTTP localhost to avoid certificate mismatch
        // Backend still uses passkey-demo.local RPID for cross-platform compatibility
        return "http://localhost:8080/api"
        #else
        // Physical device: Use the proper HTTPS domain
        return "https://passkey-demo.local:8080/api"
        #endif
    }()
    
    private let session: URLSession
    
    init() {
        let config = URLSessionConfiguration.default
        config.httpCookieAcceptPolicy = .always
        config.httpShouldSetCookies = true
        self.session = URLSession(configuration: config)
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
        return try await makeRequest(endpoint: "/register/begin", method: .POST, body: request)
    }
    
    func finishRegistration(credential: RegistrationCredential) async throws -> RegistrationResult {
        return try await makeRequest(endpoint: "/register/finish", method: .POST, body: credential)
    }
    
    // MARK: - Authentication API
    
    func beginAuthentication(username: String? = nil) async throws -> AuthenticationOptions {
        let request = AuthenticationRequest(username: username)
        return try await makeRequest(endpoint: "/login/begin", method: .POST, body: request)
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