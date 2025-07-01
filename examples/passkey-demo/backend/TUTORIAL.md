# WebAuthn Passkey Demo Backend Tutorial

This tutorial walks through building a production-ready WebAuthn passkey server in Go. The code demonstrates best practices for implementing passwordless authentication with cross-platform passkey support.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [WebAuthn Concepts](#webauthn-concepts)
3. [Code Organization](#code-organization)
4. [Implementation Guide](#implementation-guide)
5. [Security Considerations](#security-considerations)
6. [Testing and Deployment](#testing-and-deployment)

## Architecture Overview

This backend implements a complete WebAuthn Relying Party (RP) server with:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Browser   │    │   iOS App       │    │  Android App    │
│                 │    │                 │    │                 │
│ Navigator.      │    │ Authentication  │    │ Credential      │
│ credentials.*   │    │ Services        │    │ Manager         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    HTTPS/WebAuthn Protocol
                                 │
         ┌───────────────────────▼───────────────────────┐
         │              Go Backend Server                │
         │                                               │
         │  ┌─────────────┐  ┌─────────────┐  ┌────────┐│
         │  │   Router    │  │ Middleware  │  │  CORS  ││
         │  └─────────────┘  └─────────────┘  └────────┘│
         │                                               │
         │  ┌─────────────┐  ┌─────────────┐  ┌────────┐│
         │  │  Handlers   │  │   Models    │  │ Store  ││
         │  └─────────────┘  └─────────────┘  └────────┘│
         │                                               │
         │  ┌─────────────────────────────────────────┐  │
         │  │     github.com/go-webauthn/webauthn    │  │
         │  └─────────────────────────────────────────┘  │
         └───────────────────────────────────────────────┘
```

### Key Components:

- **HTTP Server**: Handles cross-origin requests from web and mobile apps
- **WebAuthn Library**: Implements FIDO2/WebAuthn protocol specification
- **Session Management**: Secure temporary storage for multi-round auth flows
- **User Storage**: Persistent storage for users and their credentials
- **Middleware**: CORS, logging, session handling, and security headers

## WebAuthn Concepts

### Relying Party (RP)
Your server application that wants to authenticate users. Key properties:
- **RPID**: Domain that owns the credentials (e.g., "example.com")
- **RP Origins**: URLs that can initiate WebAuthn ceremonies
- **Display Name**: Human-readable name shown to users

### Authentication Flows

#### Registration (Credential Creation)
```
Client                    Server                   Authenticator
  │                         │                         │
  │ 1. POST /register/begin │                         │
  │────────────────────────▶│                         │
  │                         │ 2. Generate challenge   │
  │                         │    Create options       │
  │ 3. Credential options   │                         │
  │◀────────────────────────│                         │
  │                         │                         │
  │ 4. navigator.credentials.create()                 │
  │─────────────────────────────────────────────────▶│
  │                         │                     5. User consent
  │                         │                        & biometrics
  │ 6. New credential       │                         │
  │◀─────────────────────────────────────────────────│
  │                         │                         │
  │ 7. POST /register/finish                          │
  │────────────────────────▶│                         │
  │                         │ 8. Verify signature     │
  │                         │    Store credential     │
  │ 9. Success response     │                         │
  │◀────────────────────────│                         │
```

#### Authentication (Credential Assertion)
```
Client                    Server                   Authenticator
  │                         │                         │
  │ 1. POST /login/begin    │                         │
  │────────────────────────▶│                         │
  │                         │ 2. Generate challenge   │
  │                         │    Create options       │
  │ 3. Auth options         │                         │
  │◀────────────────────────│                         │
  │                         │                         │
  │ 4. navigator.credentials.get()                    │
  │─────────────────────────────────────────────────▶│
  │                         │                     5. User verification
  │                         │                        & biometrics
  │ 6. Credential assertion │                         │
  │◀─────────────────────────────────────────────────│
  │                         │                         │
  │ 7. POST /login/finish   │                         │
  │────────────────────────▶│                         │
  │                         │ 8. Verify signature     │
  │                         │    Update sign count    │
  │ 9. Auth success         │                         │
  │◀────────────────────────│                         │
```

### Credential Types

**Platform Authenticators** (Passkeys):
- Built into the device (Face ID, Touch ID, Windows Hello)
- Private keys stored in secure hardware (Secure Enclave, TPM)
- Can sync across devices via cloud keychain

**Cross-Platform Authenticators** (Security Keys):
- External devices (USB, NFC, Bluetooth)
- Portable between devices
- Usually not synced

## Code Organization

```
backend/
├── main.go              # Application entry point and server setup
├── handlers.go          # HTTP request handlers
├── models.go           # Data models and storage interface
├── middleware.go       # HTTP middleware (CORS, logging, sessions)
├── config.go           # Configuration management
├── errors.go           # Custom error types and handling
├── validation.go       # Input validation logic
└── server.go           # Server struct and routing setup
```

### Go Best Practices Demonstrated:

1. **Package Organization**: Clear separation of concerns
2. **Interface Design**: Storage abstraction for easy testing
3. **Error Handling**: Consistent error types and responses
4. **Concurrency**: Safe concurrent access with proper locking
5. **Testing**: Unit tests with mocks and table-driven tests
6. **Documentation**: Comprehensive godoc comments

## Implementation Guide

### Step 1: WebAuthn Configuration

```go
// WebAuthn configuration for cross-platform compatibility
func NewWebAuthnConfig(rpid string, origins []string) *webauthn.Config {
    return &webauthn.Config{
        // RPID must match the domain in browser URL for security
        RPID:          rpid,
        RPDisplayName: "Your App Name",
        RPOrigins:     origins,
        
        // Prefer no attestation for better compatibility
        AttestationPreference: protocol.PreferNoAttestation,
        
        // Configure for passkeys (resident credentials)
        AuthenticatorSelection: protocol.AuthenticatorSelection{
            // Platform authenticators preferred (Face ID, Touch ID)
            AuthenticatorAttachment: protocol.Platform,
            // Require resident keys for discoverable credentials
            ResidentKey: protocol.ResidentKeyRequirementRequired,
            RequireResidentKey: protocol.ResidentKeyRequired(),
            // Require user verification for security
            UserVerification: protocol.VerificationRequired,
        },
        
        // Reasonable timeouts for user interaction
        Timeouts: webauthn.TimeoutsConfig{
            Registration: webauthn.TimeoutConfig{
                Enforce: true,
                Timeout: 60 * time.Second,
            },
            Login: webauthn.TimeoutConfig{
                Enforce: true,
                Timeout: 60 * time.Second,
            },
        },
    }
}
```

### Step 2: User Model Implementation

```go
// User implements the webauthn.User interface
type User struct {
    ID          []byte                `json:"id"`
    Username    string                `json:"username"`
    DisplayName string                `json:"displayName"`
    Credentials []webauthn.Credential `json:"credentials"`
    CreatedAt   time.Time             `json:"createdAt"`
}

// WebAuthn interface methods
func (u User) WebAuthnID() []byte                      { return u.ID }
func (u User) WebAuthnName() string                   { return u.Username }
func (u User) WebAuthnDisplayName() string            { return u.DisplayName }
func (u User) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }
```

### Step 3: Registration Handler

```go
func (app *App) handleRegisterBegin(w http.ResponseWriter, r *http.Request) {
    var req RegisterBeginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        app.writeError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate input
    if err := validateUsername(req.Username); err != nil {
        app.writeError(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Get or create user
    user, err := app.store.GetOrCreateUser(req.Username, req.DisplayName)
    if err != nil {
        app.writeError(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Begin WebAuthn registration
    options, sessionData, err := app.webAuthn.BeginRegistration(
        user,
        webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
        webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
            AuthenticatorAttachment: protocol.Platform,
            ResidentKey: protocol.ResidentKeyRequirementRequired,
            RequireResidentKey: protocol.ResidentKeyRequired(),
            UserVerification: protocol.VerificationRequired,
        }),
    )
    if err != nil {
        app.writeError(w, fmt.Sprintf("Failed to begin registration: %v", err), 
                      http.StatusInternalServerError)
        return
    }

    // Store session data
    sessionID := uuid.New().String()
    app.store.StoreSession(sessionID, user.ID, *sessionData)

    // Set session cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "webauthn-session",
        Value:    sessionID,
        Path:     "/",
        HttpOnly: true,
        Secure:   true, // HTTPS only in production
        SameSite: http.SameSiteStrictMode,
        MaxAge:   300, // 5 minutes
    })

    // Return options to client
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(options)
}
```

### Step 4: Authentication Handler

```go
func (app *App) handleLoginBegin(w http.ResponseWriter, r *http.Request) {
    var req LoginBeginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        app.writeError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    var options *protocol.CredentialAssertion
    var sessionData *webauthn.SessionData
    var err error
    var userID []byte

    if req.Username != "" {
        // Traditional login with username
        user, exists := app.store.GetUser(req.Username)
        if !exists {
            app.writeError(w, "Authentication failed", http.StatusUnauthorized)
            return
        }
        
        options, sessionData, err = app.webAuthn.BeginLogin(user)
        userID = user.ID
    } else {
        // Discoverable login (passwordless)
        options, sessionData, err = app.webAuthn.BeginDiscoverableLogin()
        userID = nil // Will be determined during finish
    }

    if err != nil {
        app.writeError(w, fmt.Sprintf("Failed to begin login: %v", err), 
                      http.StatusInternalServerError)
        return
    }

    // Store session data
    sessionID := uuid.New().String()
    app.store.StoreSession(sessionID, userID, *sessionData)

    // Set session cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "webauthn-session",
        Value:    sessionID,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
        MaxAge:   300,
    })

    json.NewEncoder(w).Encode(options)
}
```

## Security Considerations

### RPID Security
- RPID must be a domain suffix of the origin
- Never use `localhost` in production
- Use consistent RPID across all platforms

### Session Management
- Use cryptographically secure session IDs
- Implement session expiration (5-15 minutes)
- Clear sessions after completion
- Use HttpOnly, Secure, SameSite cookies

### Credential Validation
- Always verify signatures in finish handlers
- Check credential exists and is valid for user
- Update sign counter to detect cloned authenticators
- Validate challenge and origin

### Rate Limiting
```go
// Example rate limiting middleware
func rateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
    // Implementation for production use
}
```

### Input Validation
```go
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,30}$`)

func validateUsername(username string) error {
    if !usernameRegex.MatchString(username) {
        return errors.New("invalid username format")
    }
    // Additional validation...
    return nil
}
```

## Testing and Deployment

### Unit Testing
```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        username string
        valid    bool
    }{
        {"validuser", true},
        {"test123", true},
        {"ab", false},        // too short
        {"invalid user", false}, // space not allowed
    }
    
    for _, tt := range tests {
        err := validateUsername(tt.username)
        if tt.valid && err != nil {
            t.Errorf("Expected %s to be valid, got error: %v", tt.username, err)
        }
        if !tt.valid && err == nil {
            t.Errorf("Expected %s to be invalid, got no error", tt.username)
        }
    }
}
```

### Production Deployment
- Use HTTPS everywhere
- Implement proper logging and monitoring
- Use persistent storage (PostgreSQL, MongoDB)
- Add rate limiting and DDoS protection
- Monitor sign counter anomalies
- Implement backup and recovery

### Environment Configuration
```go
type Config struct {
    Port       string
    RPID       string
    RPOrigins  []string
    DBHost     string
    DBPassword string
    LogLevel   string
}

func LoadConfig() (*Config, error) {
    // Load from environment variables or config file
}
```

## Advanced Features

### Attestation Verification
```go
// For high-security applications, verify attestation
config.AttestationPreference = protocol.PreferDirectAttestation
```

### Metadata Service Integration
```go
// Use FIDO MDS to verify authenticator models
import "github.com/go-webauthn/webauthn/metadata"
```

### Enterprise Features
- Authenticator policy enforcement
- Audit logging for compliance
- Integration with SSO systems
- Multi-tenant support

## Conclusion

This backend demonstrates a production-ready WebAuthn implementation following Go best practices. Key takeaways:

1. **Security First**: Proper validation, secure sessions, HTTPS enforcement
2. **Cross-Platform**: Consistent RPID enables passkey sharing
3. **User Experience**: Support both traditional and discoverable login
4. **Maintainability**: Clean code structure, comprehensive tests
5. **Scalability**: Interface-based design for easy database integration

For production use, replace the in-memory store with a persistent database and add appropriate monitoring, logging, and security measures.