# WebAuthn Library Usage Guide

*A comprehensive guide based on code analysis of the github.com/go-webauthn/webauthn library*

## Table of Contents
1. [Library Overview](#library-overview)
2. [Core Concepts](#core-concepts)
3. [Quick Start](#quick-start)
4. [Registration Flow](#registration-flow)
5. [Authentication Flow](#authentication-flow)
6. [Advanced Usage](#advanced-usage)
7. [Security Considerations](#security-considerations)
8. [Troubleshooting](#troubleshooting)

## Library Overview

This WebAuthn library provides a complete implementation of the W3C WebAuthn specification in Go. It handles the complex cryptographic operations, protocol validation, and security checks required for passwordless authentication.

### Key Features
- **Full WebAuthn Level 3 Support**: Complete implementation of the latest specification
- **Multiple Attestation Formats**: TPM, U2F, Apple, Android, Packed, and more
- **Passkey Support**: Discoverable credentials for true passwordless flows
- **Clone Detection**: Prevents authenticator duplication attacks
- **Metadata Validation**: FIDO Alliance MDS integration for authenticator verification
- **Flexible Configuration**: Extensive customization options for different use cases

### Architecture
- **High-Level API** (`webauthn` package): Simple interfaces for registration and authentication
- **Protocol Layer** (`protocol` package): Low-level WebAuthn specification implementation
- **Metadata Layer** (`metadata` package): Authenticator metadata and validation

## Core Concepts

### Users and Credentials
Every user must implement the `User` interface:

```go
type User interface {
    WebAuthnID() []byte              // Unique, stable identifier (max 64 bytes)
    WebAuthnName() string            // Username for authentication ceremonies
    WebAuthnDisplayName() string     // Human-readable display name
    WebAuthnCredentials() []Credential // User's registered credentials
}
```

### Sessions
WebAuthn requires storing session data between the "Begin" and "Finish" operations:

```go
type SessionData struct {
    Challenge            string                         // Cryptographic challenge
    RelyingPartyID       string                        // RP ID for validation
    UserID               []byte                        // User identifier
    AllowedCredentialIDs [][]byte                      // For authentication
    Expires              time.Time                     // Session timeout
    UserVerification     UserVerificationRequirement   // UV requirement
    Extensions           AuthenticationExtensions      // WebAuthn extensions
}
```

### Configuration
The library is configured through the `Config` struct:

```go
config := &webauthn.Config{
    RPDisplayName: "My Application",
    RPID:          "myapp.com",                    // Domain (no protocol/port)
    RPOrigins:     []string{"https://myapp.com"}, // Full origin URLs
    
    // Optional configurations
    AttestationPreference: protocol.PreferNoAttestation,
    AuthenticatorSelection: protocol.AuthenticatorSelection{
        UserVerification: protocol.VerificationPreferred,
    },
    Timeouts: webauthn.TimeoutsConfig{
        Registration: webauthn.TimeoutConfig{
            Enforce: true,
            Timeout: 60 * time.Second,
        },
    },
}
```

## Quick Start

### 1. Initialize WebAuthn

```go
package main

import (
    "github.com/go-webauthn/webauthn/webauthn"
    "github.com/go-webauthn/webauthn/protocol"
)

func main() {
    config := &webauthn.Config{
        RPDisplayName: "My App",
        RPID:          "localhost",
        RPOrigins:     []string{"http://localhost:3000"},
    }
    
    webAuthn, err := webauthn.New(config)
    if err != nil {
        panic(err)
    }
}
```

### 2. Implement User Interface

```go
type AppUser struct {
    ID          []byte                  `json:"id"`
    Name        string                  `json:"name"`
    DisplayName string                  `json:"display_name"`
    Credentials []webauthn.Credential   `json:"credentials"`
}

func (u AppUser) WebAuthnID() []byte                      { return u.ID }
func (u AppUser) WebAuthnName() string                    { return u.Name }
func (u AppUser) WebAuthnDisplayName() string             { return u.DisplayName }
func (u AppUser) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }
```

### 3. Session Storage

```go
// Simple in-memory session store (use Redis/database in production)
var sessions = make(map[string]webauthn.SessionData)

func storeSession(sessionID string, data webauthn.SessionData) {
    sessions[sessionID] = data
}

func getSession(sessionID string) (webauthn.SessionData, bool) {
    data, ok := sessions[sessionID]
    return data, ok
}
```

## Registration Flow

### Basic Registration

```go
func beginRegistration(w http.ResponseWriter, r *http.Request) {
    // 1. Get user (create if new)
    user := getCurrentUser(r) // Your user retrieval logic
    
    // 2. Begin registration
    credentialCreation, sessionData, err := webAuthn.BeginRegistration(user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 3. Store session
    sessionID := generateSessionID()
    storeSession(sessionID, *sessionData)
    
    // 4. Set session cookie and return options
    http.SetCookie(w, &http.Cookie{
        Name:     "webauthn-session",
        Value:    sessionID,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    })
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(credentialCreation)
}

func finishRegistration(w http.ResponseWriter, r *http.Request) {
    // 1. Get session
    sessionCookie, err := r.Cookie("webauthn-session")
    if err != nil {
        http.Error(w, "No session", http.StatusBadRequest)
        return
    }
    
    sessionData, ok := getSession(sessionCookie.Value)
    if !ok {
        http.Error(w, "Invalid session", http.StatusBadRequest)
        return
    }
    
    // 2. Get user
    user := getCurrentUser(r)
    
    // 3. Finish registration
    credential, err := webAuthn.FinishRegistration(user, sessionData, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // 4. Save credential to user
    user.Credentials = append(user.Credentials, *credential)
    saveUser(user) // Your persistence logic
    
    // 5. Clean up session
    delete(sessions, sessionCookie.Value)
    
    w.WriteHeader(http.StatusOK)
}
```

### Advanced Registration Options

```go
// Require platform authenticator with user verification
options, session, err := webAuthn.BeginRegistration(
    user,
    webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
        AuthenticatorAttachment: protocol.Platform,
        UserVerification:       protocol.VerificationRequired,
    }),
    webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
    webauthn.WithConveyancePreference(protocol.PreferDirectAttestation),
)

// Custom credential parameters (algorithms)
credParams := []protocol.CredentialParameter{
    {Type: protocol.PublicKeyCredentialType, Algorithm: webauthncose.AlgEdDSA},
    {Type: protocol.PublicKeyCredentialType, Algorithm: webauthncose.AlgES256},
}

options, session, err := webAuthn.BeginRegistration(
    user,
    webauthn.WithCredentialParameters(credParams),
)

// Exclude existing credentials to prevent duplicates
excludeList := []protocol.CredentialDescriptor{}
for _, cred := range user.WebAuthnCredentials() {
    excludeList = append(excludeList, cred.Descriptor())
}

options, session, err := webAuthn.BeginRegistration(
    user,
    webauthn.WithExclusions(excludeList),
)
```

## Authentication Flow

### Standard Authentication (Username First)

```go
func beginLogin(w http.ResponseWriter, r *http.Request) {
    // 1. Get user by username
    username := r.FormValue("username")
    user, err := getUserByUsername(username)
    if err != nil {
        // Don't reveal if user exists
        http.Error(w, "Authentication failed", http.StatusUnauthorized)
        return
    }
    
    // 2. Begin login
    credentialAssertion, sessionData, err := webAuthn.BeginLogin(user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 3. Store session
    sessionID := generateSessionID()
    storeSession(sessionID, *sessionData)
    
    // 4. Return assertion options
    http.SetCookie(w, &http.Cookie{
        Name:     "webauthn-session",
        Value:    sessionID,
        HttpOnly: true,
    })
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(credentialAssertion)
}

func finishLogin(w http.ResponseWriter, r *http.Request) {
    // 1. Get session
    sessionCookie, err := r.Cookie("webauthn-session")
    if err != nil {
        http.Error(w, "No session", http.StatusBadRequest)
        return
    }
    
    sessionData, ok := getSession(sessionCookie.Value)
    if !ok {
        http.Error(w, "Invalid session", http.StatusBadRequest)
        return
    }
    
    // 2. Get user
    user := getCurrentUser(r)
    
    // 3. Finish login
    credential, err := webAuthn.FinishLogin(user, sessionData, r)
    if err != nil {
        http.Error(w, "Authentication failed", http.StatusUnauthorized)
        return
    }
    
    // 4. Check for cloned authenticator
    if credential.Authenticator.CloneWarning {
        // SECURITY ALERT: Possible cloned authenticator!
        logSecurityEvent(user.ID, "clone_warning")
        // Consider blocking authentication or requiring additional verification
    }
    
    // 5. Update credential in database
    updateUserCredential(user.ID, credential)
    
    // 6. Create user session
    createUserSession(w, user)
    
    w.WriteHeader(http.StatusOK)
}
```

### Discoverable Login (Passkeys)

```go
func beginDiscoverableLogin(w http.ResponseWriter, r *http.Request) {
    // 1. Begin discoverable login (no username needed)
    credentialAssertion, sessionData, err := webAuthn.BeginDiscoverableLogin(
        webauthn.WithUserVerification(protocol.VerificationRequired),
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 2. Store session
    sessionID := generateSessionID()
    storeSession(sessionID, *sessionData)
    
    http.SetCookie(w, &http.Cookie{
        Name: "webauthn-session",
        Value: sessionID,
    })
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(credentialAssertion)
}

func finishDiscoverableLogin(w http.ResponseWriter, r *http.Request) {
    // 1. Get session
    sessionCookie, err := r.Cookie("webauthn-session")
    if err != nil {
        http.Error(w, "No session", http.StatusBadRequest)
        return
    }
    
    sessionData, ok := getSession(sessionCookie.Value)
    if !ok {
        http.Error(w, "Invalid session", http.StatusBadRequest)
        return
    }
    
    // 2. Define user lookup handler
    userHandler := func(rawID, userHandle []byte) (webauthn.User, error) {
        // rawID = credential ID
        // userHandle = user ID from authenticator
        return getUserByWebAuthnID(userHandle)
    }
    
    // 3. Validate discoverable login
    user, credential, err := webAuthn.ValidateDiscoverableLogin(
        userHandler, sessionData, r,
    )
    if err != nil {
        http.Error(w, "Authentication failed", http.StatusUnauthorized)
        return
    }
    
    // 4. Create user session
    createUserSession(w, user)
    
    w.WriteHeader(http.StatusOK)
}
```

## Advanced Usage

### Enterprise Attestation Validation

```go
// Configure with MDS provider for attestation validation
import "github.com/go-webauthn/webauthn/metadata"

mdsProvider := memory.New()
// or: mdsProvider := cached.New(...)

config := &webauthn.Config{
    RPDisplayName: "Enterprise App",
    RPID:          "company.com",
    RPOrigins:     []string{"https://company.com"},
    MDS:           mdsProvider,
}

// Require direct attestation
options, session, err := webAuthn.BeginRegistration(
    user,
    webauthn.WithConveyancePreference(protocol.PreferDirectAttestation),
)
```

### Custom Extensions

```go
// Request credential properties extension
extensions := map[string]any{
    "credProps": true,
    "largeBlob": map[string]any{
        "write": []byte("custom data"),
    },
}

options, session, err := webAuthn.BeginRegistration(
    user,
    webauthn.WithRegistrationExtensions(extensions),
)

// For authentication
options, session, err := webAuthn.BeginLogin(
    user,
    webauthn.WithAssertionExtensions(extensions),
)
```

### FIDO U2F Backward Compatibility

```go
// Support legacy U2F credentials
options, session, err := webAuthn.BeginLogin(
    user,
    webauthn.WithAppIdExtension("https://legacy-app.com"),
)
```

### Multi-Origin Support

```go
config := &webauthn.Config{
    RPID: "example.com",
    RPOrigins: []string{
        "https://example.com",
        "https://app.example.com",
        "https://secure.example.com",
    },
    // For embedded scenarios
    RPTopOrigins: []string{
        "https://partner.com",
    },
}
```

## Security Considerations

### Critical Security Practices

1. **Secure Session Storage**
   - Never store sessions in cookies or client-side storage
   - Use server-side storage with appropriate expiration
   - Encrypt session data if stored in databases

2. **Clone Detection**
   ```go
   if credential.Authenticator.CloneWarning {
       // CRITICAL: Handle cloned authenticator
       // Options:
       // - Block authentication
       // - Require additional verification
       // - Flag account for review
       return errors.New("security violation detected")
   }
   ```

3. **User Handle Validation**
   ```go
   // Ensure user handle matches expected user
   if !bytes.Equal(userHandle, user.WebAuthnID()) {
       return errors.New("user handle mismatch")
   }
   ```

4. **Origin Validation**
   - Configure RPOrigins exactly (including protocol and port)
   - Never use wildcards in production
   - Validate RPTopOrigins carefully for iframe scenarios

5. **Error Handling**
   ```go
   // Don't reveal user existence in responses
   if err != nil {
       log.Printf("Auth error for user %s: %v", userID, err)
       http.Error(w, "Authentication failed", http.StatusUnauthorized)
       return
   }
   ```

### Backup State Monitoring

```go
// Monitor credential backup status
if credential.Flags.BackupEligible && !credential.Flags.BackupState {
    // Credential can be backed up but isn't yet
    // Consider prompting user to enable backup
}

if credential.Flags.BackupState {
    // Credential is backed up (e.g., to cloud keychain)
    // May want to log this for audit purposes
}
```

## Troubleshooting

### Common Issues

1. **Invalid Origin Errors**
   - Ensure RPOrigins exactly match the requesting origin
   - Include protocol (https://) and port if non-standard
   - Check for typos in domain names

2. **Session Expired/Invalid**
   - Implement proper session cleanup
   - Check session storage implementation
   - Verify session IDs are properly generated

3. **User Verification Failures**
   - Check if authenticator supports required UV level
   - Consider downgrading to `VerificationPreferred`
   - Ensure biometric/PIN is set up on device

4. **Attestation Validation Errors**
   - Verify MDS configuration if using attestation
   - Check certificate chains and trust anchors
   - Consider using `PreferNoAttestation` for testing

5. **Clone Detection Warnings**
   - Investigate potential security breach
   - Check if authenticator was restored from backup
   - Consider re-registration if appropriate

### Debug Tips

```go
// Enable detailed logging
import "log"

if err != nil {
    // Log full error details server-side only
    log.Printf("WebAuthn error: %+v", err)
    
    // Return generic error to client
    http.Error(w, "Authentication failed", http.StatusBadRequest)
}
```

### Testing Considerations

- Use `http://localhost` for local development (allowed by spec)
- Test with multiple authenticator types (platform, roaming)
- Verify backup/restore scenarios
- Test timeout handling
- Validate cross-origin scenarios if applicable

This comprehensive guide should help you understand and implement WebAuthn authentication in your Go applications using this library.