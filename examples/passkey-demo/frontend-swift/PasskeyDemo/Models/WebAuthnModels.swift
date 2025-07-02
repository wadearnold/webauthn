import Foundation

// MARK: - Registration Models

struct RegistrationRequest: Codable {
    let username: String
    let displayName: String
}

// Backend response wrapper - matches actual API response
struct RegistrationOptionsResponse: Codable {
    let publicKey: RegistrationOptions
}

struct RegistrationOptions: Codable {
    let challenge: String
    let rp: RelyingParty
    let user: User
    let pubKeyCredParams: [PubKeyCredParam]
    let timeout: Int
    let excludeCredentials: [CredentialDescriptor]?
    let authenticatorSelection: AuthenticatorSelection
    let attestation: String
    
    struct RelyingParty: Codable {
        let id: String
        let name: String
    }
    
    struct User: Codable {
        let id: String
        let name: String
        let displayName: String
    }
    
    struct PubKeyCredParam: Codable {
        let type: String
        let alg: Int
    }
    
    struct CredentialDescriptor: Codable {
        let type: String
        let id: String
        let transports: [String]?
    }
    
    struct AuthenticatorSelection: Codable {
        let authenticatorAttachment: String?
        let residentKey: String?
        let requireResidentKey: Bool?
        let userVerification: String?
    }
}

struct RegistrationCredential: Codable {
    let id: String
    let rawId: String
    let type: String
    let response: AuthenticatorAttestationResponse
    
    struct AuthenticatorAttestationResponse: Codable {
        let attestationObject: String
        let clientDataJSON: String
        let transports: [String]?
    }
}

// MARK: - Authentication Models

struct AuthenticationRequest: Codable {
    let username: String?
}

// Backend response wrapper - matches actual API response
struct AuthenticationOptionsResponse: Codable {
    let publicKey: AuthenticationOptions
}

struct AuthenticationOptions: Codable {
    let challenge: String
    let timeout: Int
    let rpId: String?
    let allowCredentials: [CredentialDescriptor]?
    let userVerification: String?
    
    struct CredentialDescriptor: Codable {
        let type: String
        let id: String
        let transports: [String]?
    }
}

struct AuthenticationCredential: Codable {
    let id: String
    let rawId: String
    let type: String
    let response: AuthenticatorAssertionResponse
    
    struct AuthenticatorAssertionResponse: Codable {
        let authenticatorData: String
        let clientDataJSON: String
        let signature: String
        let userHandle: String?
    }
}

// MARK: - User Management Models

struct UserPasskey: Codable, Identifiable {
    let id: String
    let name: String
    let createdAt: String
    let lastUsed: String?
    let transports: [String]
    let backedUp: Bool
    let userVerified: Bool
    
    var displayName: String {
        return name.isEmpty ? "Passkey \(id.prefix(8))" : name
    }
    
    var isCloudSynced: Bool {
        return backedUp && transports.contains("hybrid")
    }
    
    var createdDate: Date? {
        ISO8601DateFormatter().date(from: createdAt)
    }
    
    var lastUsedDate: Date? {
        guard let lastUsed = lastUsed else { return nil }
        return ISO8601DateFormatter().date(from: lastUsed)
    }
}

struct AuthenticationResult: Codable {
    let message: String
    let data: AuthenticationData?
    
    // Computed property for compatibility
    var success: Bool {
        return data != nil
    }
    
    var username: String? {
        return data?.username
    }
}

struct AuthenticationData: Codable {
    let username: String
    let displayName: String
    let userId: String
}

struct RegistrationResult: Codable {
    let message: String
    let data: RegistrationData?
    
    // Computed property for compatibility
    var success: Bool {
        return data != nil
    }
    
    var username: String? {
        return data?.username
    }
}

struct RegistrationData: Codable {
    let credentialId: String
    let username: String
    let displayName: String
    let userId: String
}

// MARK: - Error Models

struct APIError: Codable, Error {
    let error: String
    let code: String?
    
    var localizedDescription: String {
        return error
    }
}

// MARK: - Response Wrapper

struct APIResponse<T: Codable>: Codable {
    let data: T?
    let error: String?
    let success: Bool
}

// MARK: - Base64URL Utilities

extension String {
    /// Convert base64URL string to Data
    func base64URLDecode() -> Data? {
        var base64 = self
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")
        
        // Add padding if needed
        let padding = base64.count % 4
        if padding > 0 {
            base64 += String(repeating: "=", count: 4 - padding)
        }
        
        return Data(base64Encoded: base64)
    }
}

extension Data {
    /// Convert Data to base64URL string
    func base64URLEncode() -> String {
        return self.base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }
}

// MARK: - WebAuthn Constants

enum WebAuthnError: Error, LocalizedError {
    case notSupported
    case userCancel
    case invalidChallenge
    case networkError(String)
    case apiError(String)
    case encodingError
    case unknown(String)
    
    var errorDescription: String? {
        switch self {
        case .notSupported:
            return "WebAuthn is not supported on this device"
        case .userCancel:
            return "User cancelled the authentication request"
        case .invalidChallenge:
            return "Invalid challenge received from server"
        case .networkError(let message):
            return "Network error: \(message)"
        case .apiError(let message):
            return "API error: \(message)"
        case .encodingError:
            return "Failed to encode/decode data"
        case .unknown(let message):
            return "Unknown error: \(message)"
        }
    }
}