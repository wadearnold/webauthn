# WebAuthn Passkey Demo - Go Backend Tutorial

A comprehensive Go backend implementing WebAuthn passkey authentication with cross-platform support. This tutorial demonstrates production-ready patterns, security best practices, and educational code organization for developers learning WebAuthn.

## 🎯 Learning Objectives

This backend serves as both a working demo and educational resource, demonstrating:

- **WebAuthn Implementation**: Complete Relying Party server with W3C specification compliance
- **Go Best Practices**: Idiomatic Go code with proper documentation and error handling
- **Security Patterns**: Input validation, session management, and credential lifecycle
- **Cross-Platform Architecture**: RPID configuration for seamless passkey sharing
- **API Design**: RESTful endpoints with structured responses and comprehensive error handling
- **Production Readiness**: Patterns that scale from demo to production deployment

## 🏗️ Architecture Overview

```
                              WebAuthn Backend Architecture
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                    HTTP Server                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │   Middleware    │  │    Handlers     │  │    Storage      │  │   Config    │ │
│  │                 │  │                 │  │                 │  │             │ │
│  │ • CORS          │  │ • Registration  │  │ • InMemoryStore │  │ • WebAuthn  │ │
│  │ • Logging       │  │ • Authentication│  │ • Session Mgmt  │  │ • RPID      │ │
│  │ • Session       │  │ • User Mgmt     │  │ • User Data     │  │ • Origins   │ │
│  │ • JSON          │  │ • Profile       │  │ • Credentials   │  │ • Timeouts  │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
                                         │
                     ┌───────────────────┼───────────────────┐
                     │                   │                   │
                     ▼                   ▼                   ▼
              ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
              │ Web Browser │   │   iOS App   │   │Android App  │
              │             │   │             │   │             │
              │ React SPA   │   │ SwiftUI     │   │ Compose     │
              │ JavaScript  │   │ AuthSvc     │   │ CredMgr     │
              └─────────────┘   └─────────────┘   └─────────────┘
```

## 🔧 Code Organization

The backend follows Go best practices with clear separation of concerns:

```
backend/
├── main.go              # Application entry point and server setup
├── handlers.go          # HTTP request handlers with WebAuthn ceremony logic
├── models.go           # Data models, storage interface, and business logic
├── middleware.go       # HTTP middleware (CORS, logging, sessions)
├── TUTORIAL.md         # Comprehensive WebAuthn implementation guide
└── README.md           # This file - getting started and development guide
```

### Key Design Patterns

**Dependency Injection**
```go
type App struct {
    webAuthn *webauthn.WebAuthn // Configured WebAuthn instance
    store    *InMemoryStore     // Storage abstraction (interface-ready)
}
```

**Structured Error Handling**
```go
type AppError struct {
    Code    string `json:"code"`    // Machine-readable
    Message string `json:"message"` // Human-readable
}
```

**Thread-Safe Storage**
```go
type InMemoryStore struct {
    users    map[string]*User
    sessions map[string]*Session
    mu       sync.RWMutex        // Protects concurrent access
}
```

## 🚀 Quick Start

### Prerequisites
- **Go 1.21+** for modern language features
- **ngrok account** (free tier sufficient) for cross-platform testing
- **Modern browser** with WebAuthn support

### Development Modes

#### 1. Local Development (Fast Iteration)
```bash
# Build the backend
go build -o passkey-backend .

# Start in localhost mode
./passkey-backend -localhost
# RPID: localhost, Origin: http://localhost:*
```

**Use for**: UI development, API testing, rapid prototyping

#### 2. Cross-Platform Mode (Production-like)
```bash
# Terminal 1: Start ngrok tunnel
cd ..
./scripts/start-ngrok.sh

# Terminal 2: Start backend with ngrok configuration
cd backend
go build -o passkey-backend .
source ../.env && ./passkey-backend
# RPID: your-tunnel.ngrok.io, Origin: https://your-tunnel.ngrok.io
```

**Use for**: iOS/Android testing, cross-platform passkey validation, production simulation

#### Command Line Options

- `-localhost`: Force localhost mode (ignores NGROK_URL environment variable)
- `-h`: Show help and available flags

Examples:
```bash
# Force localhost mode even if NGROK_URL is set
./passkey-backend -localhost

# Use environment configuration (default)
./passkey-backend

# With environment variable
NGROK_URL=https://abc123.ngrok.io ./passkey-backend
```

### First Run

1. **Clone and setup**:
   ```bash
   git clone <repository>
   cd examples/passkey-demo/backend
   go mod download
   ```

2. **Start server**:
   ```bash
   go run .
   ```

3. **Test API**:
   ```bash
   curl http://localhost:8080/api/health
   # {"status":"ok","time":"2024-..."}
   ```

## 📚 WebAuthn Implementation Guide

### 1. Configuration

The WebAuthn configuration determines how credentials work across platforms:

```go
config := &webauthn.Config{
    RPDisplayName: "WebAuthn Passkey Demo",
    RPID:          rpid, // "localhost" or ngrok domain
    RPOrigins: []string{
        ngrokURL,                // Primary for cross-platform
        "http://localhost:5173", // Fallback for development
    },
    AttestationPreference: protocol.PreferNoAttestation,
    AuthenticatorSelection: protocol.AuthenticatorSelection{
        AuthenticatorAttachment: protocol.Platform,    // Built-in biometrics
        ResidentKey: protocol.ResidentKeyRequirementPreferred,
        UserVerification: protocol.VerificationPreferred,
    },
}
```

**Key Decisions**:
- **RPID**: Domain that owns the credentials (critical for cross-platform)
- **Platform Authenticators**: Forces Face ID, Touch ID, Windows Hello
- **Resident Keys**: Enables discoverable (passwordless) login
- **User Verification**: Requires biometric or PIN confirmation

### 2. Registration Flow

WebAuthn registration happens in two phases:

**Phase 1: Begin Registration**
```go
options, sessionData, err := app.webAuthn.BeginRegistration(
    user,
    webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
    webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
        AuthenticatorAttachment: protocol.Platform,
        ResidentKey: protocol.ResidentKeyRequirementRequired,
        UserVerification: protocol.VerificationRequired,
    }),
)
```

**Phase 2: Finish Registration**
```go
credential, err := app.webAuthn.FinishRegistration(user, sessionData, r)
if err != nil {
    return fmt.Errorf("registration failed: %w", err)
}
// Store credential for future authentication
user.Credentials = append(user.Credentials, *credential)
```

### 3. Authentication Flow

**Traditional Login (with username)**:
```go
options, sessionData, err := app.webAuthn.BeginLogin(user)
// Client provides username, server returns challenge for user's credentials
```

**Discoverable Login (passwordless)**:
```go
options, sessionData, err := app.webAuthn.BeginDiscoverableLogin()
// No username needed, authenticator returns user info with credential
```

## 🔒 Security Implementation

### Input Validation
```go
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,30}$`)

func validateUsername(username string) error {
    if !usernameRegex.MatchString(username) {
        return fmt.Errorf("invalid username format")
    }
    // Additional validation...
}
```

### Session Management
```go
// Cryptographically secure session IDs
sessionID := uuid.New().String()

// HTTP-only cookies with security attributes
http.SetCookie(w, &http.Cookie{
    Name:     "webauthn-session",
    Value:    sessionID,
    HttpOnly: true,
    Secure:   true, // HTTPS only in production
    SameSite: http.SameSiteStrictMode,
    MaxAge:   300, // 5 minute expiration
})
```

### Credential Verification
```go
// Verify credential still exists (prevents deleted credential attacks)
credentialExists := false
for _, userCred := range user.Credentials {
    if string(userCred.ID) == string(credential.ID) {
        credentialExists = true
        break
    }
}
if !credentialExists {
    return fmt.Errorf("credential no longer valid")
}
```

## 🛠️ API Reference

### Registration Endpoints

**POST /api/register/begin**
```json
Request:
{
    "username": "alice",
    "displayName": "Alice Smith"
}

Response:
{
    "publicKey": {
        "challenge": "base64-challenge",
        "rp": { "name": "Demo", "id": "localhost" },
        "user": { "name": "alice", "displayName": "Alice Smith" },
        "authenticatorSelection": {
            "authenticatorAttachment": "platform",
            "residentKey": "required",
            "userVerification": "required"
        }
    }
}
```

**POST /api/register/finish**
- Accepts WebAuthn credential response
- Verifies attestation and stores credential
- Returns success with user session

### Authentication Endpoints

**POST /api/login/begin**
```json
// Traditional login
{ "username": "alice" }

// Discoverable login (passwordless)
{}
```

**POST /api/login/finish**
- Accepts WebAuthn assertion response
- Verifies signature and updates sign count
- Returns success with user session

### User Management

**GET /api/user/passkeys**
- Returns list of user's credentials with metadata
- Includes sync status, creation dates, transport methods

**DELETE /api/user/passkeys/{id}**
- Removes credential from server storage
- Note: Doesn't remove from device keychain

## 🧪 Testing and Development

### Unit Testing Pattern
```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        username string
        valid    bool
    }{
        {"validuser", true},
        {"ab", false}, // too short
        {"invalid user", false}, // space not allowed
    }
    
    for _, tt := range tests {
        err := validateUsername(tt.username)
        if tt.valid && err != nil {
            t.Errorf("Expected %s to be valid: %v", tt.username, err)
        }
    }
}
```

### Integration Testing
```bash
# Test registration flow
curl -X POST http://localhost:8080/api/register/begin \
     -H "Content-Type: application/json" \
     -d '{"username":"test","displayName":"Test User"}'

# Test health endpoint
curl http://localhost:8080/api/health
```

### Cross-Platform Testing

1. **Register on web** → **Authenticate on mobile**
2. **Create on iOS** → **Use on Android** (via Google Password Manager)
3. **Test sync** across devices with same Apple ID/Google account

## 🚀 Production Deployment

### Environment Configuration
```go
type Config struct {
    RPID         string   `env:"RPID" default:"localhost"`
    RPOrigins    []string `env:"RP_ORIGINS" separator:","`
    Port         string   `env:"PORT" default:"8080"`
    DatabaseURL  string   `env:"DATABASE_URL"`
    RedisURL     string   `env:"REDIS_URL"`
}
```

### Database Migration
```go
// Replace InMemoryStore with database implementation
type PostgresStore struct {
    db *sql.DB
}

func (s *PostgresStore) CreateUser(username, displayName string) (*User, error) {
    // Database implementation
}
```

### Security Hardening
- **HTTPS enforcement** with valid certificates
- **Rate limiting** on authentication endpoints
- **Audit logging** for all credential operations
- **CORS restrictions** to trusted origins only
- **Session encryption** for sensitive data

### Monitoring and Observability
```go
// Example metrics collection
func (app *App) handleRegisterBegin(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    defer func() {
        metrics.RecordDuration("registration_begin", time.Since(start))
    }()
    
    // Handler implementation...
}
```

## 📚 Learning Resources

### In This Repository
- **[TUTORIAL.md](./TUTORIAL.md)**: Comprehensive WebAuthn implementation guide
- **[models.go](./models.go)**: Data structures with extensive documentation
- **[handlers.go](./handlers.go)**: HTTP handlers with security explanations

### External Resources
- [WebAuthn Specification](https://w3c.github.io/webauthn/)
- [go-webauthn Library](https://github.com/go-webauthn/webauthn)
- [FIDO Alliance](https://fidoalliance.org/)
- [Passkeys.dev](https://passkeys.dev/)

## 🤝 Contributing

This backend is designed to be educational. When contributing:

1. **Maintain Documentation**: Add comprehensive comments for educational value
2. **Follow Go Conventions**: Use idiomatic Go patterns and naming
3. **Security First**: Ensure all changes maintain security best practices
4. **Test Coverage**: Include tests for new functionality
5. **Real-World Relevance**: Keep examples production-applicable

## 🔐 Security Disclaimer

This is a **demonstration backend** with in-memory storage. For production use:

- ✅ **Use persistent database** (PostgreSQL, MongoDB)
- ✅ **Implement rate limiting** and DDoS protection
- ✅ **Add comprehensive logging** and monitoring
- ✅ **Use HTTPS** with valid certificates
- ✅ **Implement backup/recovery** mechanisms
- ✅ **Add enterprise attestation** verification

---

**🎓 Educational Goal**: This backend demonstrates how to build a production-ready WebAuthn server in Go while teaching best practices for security, concurrency, and API design.