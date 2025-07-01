// Package main implements a WebAuthn passkey demo backend server.
//
// This package demonstrates best practices for implementing WebAuthn authentication
// in Go, including cross-platform passkey support, secure session management,
// and proper error handling.
//
// Key concepts demonstrated:
//   - WebAuthn Relying Party implementation
//   - Thread-safe in-memory storage
//   - Credential lifecycle management
//   - Session handling for multi-round authentication
//   - Cross-platform RPID configuration
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/go-webauthn/webauthn/webauthn"
)

// User represents a user in the WebAuthn system and implements the webauthn.User interface.
//
// The User struct stores all information necessary for WebAuthn operations:
//   - ID: Unique identifier for the user (must be consistent across sessions)
//   - Username: Human-readable username (must be unique)
//   - DisplayName: User's preferred display name (can be changed)
//   - Credentials: All registered WebAuthn credentials for this user
//   - CreatedAt: Account creation timestamp
//
// This implementation uses a UUID as the user ID to ensure uniqueness and
// prevent user enumeration attacks.
type User struct {
	ID          []byte                  `json:"id"`          // WebAuthn user ID (UUID bytes)
	Username    string                  `json:"username"`    // Unique username for login
	DisplayName string                  `json:"displayName"` // User's display name
	Credentials []webauthn.Credential   `json:"credentials"` // All registered credentials
	CreatedAt   time.Time               `json:"createdAt"`   // Account creation time
}

// WebAuthnID returns the user's unique identifier for WebAuthn operations.
//
// This ID must remain constant for the user across all sessions and devices.
// WebAuthn uses this to link credentials to users during authentication.
//
// Important: This ID should never change once assigned to a user, as it would
// invalidate all their existing credentials.
func (u User) WebAuthnID() []byte {
	return u.ID
}

// WebAuthnName returns the username for this user.
//
// This is used during credential creation and is typically displayed to the user
// when selecting credentials. It should be a human-readable identifier.
//
// Note: This can be the same as the username used for traditional login.
func (u User) WebAuthnName() string {
	return u.Username
}

// WebAuthnDisplayName returns the user's display name.
//
// This is shown in authenticator prompts and credential management interfaces.
// It can be more user-friendly than the username (e.g., "John Doe" vs "jdoe123").
//
// This field can be updated without affecting existing credentials.
func (u User) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnCredentials returns all credentials registered to this user.
//
// During authentication, WebAuthn will check if any of these credentials
// can be used to authenticate the user. The authenticator will only respond
// if it has access to one of these credentials.
//
// For security, only return credentials that are still valid and haven't been
// revoked or deleted.
func (u User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

// Session represents a temporary WebAuthn session during multi-round authentication.
//
// WebAuthn authentication happens in two phases:
//  1. Begin: Server generates a challenge and returns options to client
//  2. Finish: Client returns signed challenge, server verifies signature
//
// The Session stores the state between these two phases. This includes:
//   - UserID: Which user initiated the session (nil for discoverable login)
//   - SessionData: Challenge, user ID, and other verification data
//   - CreatedAt: When the session was created (for expiration)
//
// Security considerations:
//   - Sessions should expire quickly (5-15 minutes)
//   - Session IDs should be cryptographically random
//   - Sessions should be deleted after successful completion
type Session struct {
	UserID      []byte               `json:"userId"`      // User who initiated session (nil for discoverable)
	SessionData webauthn.SessionData `json:"sessionData"` // WebAuthn challenge and verification data
	CreatedAt   time.Time            `json:"createdAt"`   // Session creation time for expiration
}

// PasskeyInfo represents credential information formatted for frontend display.
//
// This struct converts the raw WebAuthn credential data into a user-friendly
// format that can be displayed in management interfaces. It includes:
//
// Security Information:
//   - ID: Unique credential identifier (base64-encoded)
//   - SignCount: Number of times credential has been used (for clone detection)
//   - AAGUID: Authenticator model identifier
//
// User Experience Information:
//   - Name: Human-readable name for the credential
//   - CreatedAt/LastUsed: Timestamps for user reference
//   - Transports: How the credential can be activated (USB, NFC, etc.)
//
// Sync and Backup Status:
//   - BackedUp: Whether credential is currently backed up to cloud
//   - BackupEligible: Whether credential can be backed up
//   - AuthenticatorAttachment: Platform (built-in) vs cross-platform (external)
//
// This information helps users understand and manage their credentials.
type PasskeyInfo struct {
	ID                      string    `json:"id"`                      // Base64-encoded credential ID
	Name                    string    `json:"name"`                    // Human-friendly credential name
	CreatedAt               time.Time `json:"createdAt"`               // When credential was created
	LastUsed                time.Time `json:"lastUsed"`                // Last authentication time
	Transports              []string  `json:"transports"`              // Available transport methods
	BackedUp                bool      `json:"backedUp"`                // Currently backed up to cloud
	BackupEligible          bool      `json:"backupEligible"`          // Can be backed up
	UserVerified            bool      `json:"userVerified"`            // Requires user verification
	AttestationType         string    `json:"attestationType"`         // Type of attestation provided
	AuthenticatorAttachment string    `json:"authenticatorAttachment"` // Platform or cross-platform
	SignCount               uint32    `json:"signCount"`               // Usage counter for clone detection
	AAGUID                  string    `json:"aaguid"`                  // Authenticator model ID (hex)
	// User information associated with this credential
	Username    string `json:"username"`    // Owner's username
	DisplayName string `json:"displayName"` // Owner's display name
}

// InMemoryStore provides thread-safe in-memory storage for the WebAuthn demo.
//
// This implementation is suitable for development and testing but should be
// replaced with persistent storage (database) for production use.
//
// Thread Safety:
// All methods use read-write locks to ensure safe concurrent access.
// - Read operations (Get*) use read locks for better performance
// - Write operations (Create*, Update*, Delete*) use write locks
//
// Storage Structure:
//   - users: Username-based lookup for login and user management
//   - userIDs: WebAuthn user ID-based lookup for discoverable login
//   - sessions: Temporary session storage with automatic expiration
//
// Design Patterns Demonstrated:
//   - Interface-based design for easy testing and database migration
//   - Proper error handling with custom error types
//   - Concurrent safety with minimal lock contention
//   - Automatic cleanup of expired resources
type InMemoryStore struct {
	users    map[string]*User    // username -> User (for traditional lookup)
	userIDs  map[string]*User    // string(userID) -> User (for WebAuthn lookup)
	sessions map[string]*Session // sessionID -> Session (temporary storage)
	mu       sync.RWMutex        // Protects all maps for concurrent access
}

// NewInMemoryStore creates a new in-memory store with initialized maps.
//
// The store is immediately ready for use and safe for concurrent access.
// All maps are initialized to prevent nil pointer panics.
//
// Example usage:
//   store := NewInMemoryStore()
//   user, err := store.CreateUser("alice", "Alice Smith")
//   if err != nil {
//       log.Fatal(err)
//   }
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		users:    make(map[string]*User),
		userIDs:  make(map[string]*User),
		sessions: make(map[string]*Session),
	}
}

// CreateUser creates a new user with the given username and display name.
//
// This method:
//  1. Checks if username is already taken
//  2. Generates a new UUID for the WebAuthn user ID
//  3. Creates the user with empty credentials list
//  4. Stores the user in both username and userID indices
//
// The username must be unique across the system. The WebAuthn user ID
// (UUID) ensures uniqueness even if usernames are reused after deletion.
//
// Returns ErrUserExists if the username is already taken.
//
// Thread-safe: Uses write lock for atomic user creation.
func (s *InMemoryStore) CreateUser(username, displayName string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; exists {
		return nil, ErrUserExists
	}

	userID := uuid.New()
	user := &User{
		ID:          userID[:],
		Username:    username,
		DisplayName: displayName,
		Credentials: []webauthn.Credential{},
		CreatedAt:   time.Now(),
	}

	s.users[username] = user
	s.userIDs[string(userID[:])] = user

	return user, nil
}

// GetUser retrieves a user by username.
//
// This is used for traditional username-based login flows and user management
// operations. Returns the user and true if found, or nil and false if not found.
//
// Thread-safe: Uses read lock for concurrent access.
func (s *InMemoryStore) GetUser(username string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	return user, exists
}

// GetUserByID retrieves a user by their WebAuthn user ID.
//
// This is primarily used during discoverable (passwordless) login where
// the client provides the user ID from the credential response.
//
// The userID parameter should be the exact bytes returned by user.WebAuthnID().
//
// Thread-safe: Uses read lock for concurrent access.
func (s *InMemoryStore) GetUserByID(userID []byte) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.userIDs[string(userID)]
	return user, exists
}

// UpdateUser updates user credentials
func (s *InMemoryStore) UpdateUser(user *User) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.Username] = user
	s.userIDs[string(user.ID)] = user
}

// DeleteUserPasskey removes a specific credential from user
func (s *InMemoryStore) DeleteUserPasskey(username string, credentialID []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]
	if !exists {
		return ErrUserNotFound
	}

	// Find and remove the credential
	for i, cred := range user.Credentials {
		if string(cred.ID) == string(credentialID) {
			// Remove credential from slice
			user.Credentials = append(user.Credentials[:i], user.Credentials[i+1:]...)
			s.users[username] = user
			s.userIDs[string(user.ID)] = user
			return nil
		}
	}

	return ErrCredentialNotFound
}

// GetUserPasskeys returns passkey info for frontend
func (s *InMemoryStore) GetUserPasskeys(username string) ([]PasskeyInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Remove duplicates from user credentials first
	uniqueCredentials := removeDuplicateCredentials(user.Credentials)
	if len(uniqueCredentials) != len(user.Credentials) {
		fmt.Printf("INFO: Removed %d duplicate credentials for user %s\n", 
			len(user.Credentials)-len(uniqueCredentials), user.Username)
		// Update user with cleaned credentials
		user.Credentials = uniqueCredentials
		s.users[user.Username] = user
		s.userIDs[string(user.ID)] = user
	}

	passkeys := make([]PasskeyInfo, len(uniqueCredentials))
	for i, cred := range uniqueCredentials {
		// Convert transport enums to strings
		transports := make([]string, len(cred.Transport))
		for j, transport := range cred.Transport {
			transports[j] = string(transport)
		}

		// Convert AAGUID to hex string
		aaguidStr := ""
		if len(cred.Authenticator.AAGUID) > 0 {
			aaguidStr = fmt.Sprintf("%x", cred.Authenticator.AAGUID)
		}

		// Use individual credential creation time if available, fallback to user creation
		credCreatedAt := user.CreatedAt
		if cred.Authenticator.SignCount == 0 {
			// For demo: use user creation time + small offset for each credential
			credCreatedAt = user.CreatedAt.Add(time.Duration(i) * time.Minute)
		}

		passkeys[i] = PasskeyInfo{
			ID:                      string(cred.ID),
			Name:                    generatePasskeyName(cred),
			CreatedAt:               credCreatedAt,
			LastUsed:                time.Now().Add(-time.Duration(i)*time.Hour), // Simulate different last used times
			Transports:              transports,
			BackedUp:                cred.Flags.BackupState,
			BackupEligible:          cred.Flags.BackupEligible,
			UserVerified:            cred.Flags.UserVerified,
			AttestationType:         cred.AttestationType,
			AuthenticatorAttachment: string(cred.Authenticator.Attachment),
			SignCount:               cred.Authenticator.SignCount,
			AAGUID:                  aaguidStr,
			Username:                user.Username,
			DisplayName:             user.DisplayName,
		}
	}

	return passkeys, nil
}

// Session management
func (s *InMemoryStore) StoreSession(sessionID string, userID []byte, sessionData webauthn.SessionData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[sessionID] = &Session{
		UserID:      userID,
		SessionData: sessionData,
		CreatedAt:   time.Now(),
	}
}

func (s *InMemoryStore) GetSession(sessionID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, false
	}

	// Check if session is expired (5 minutes for demo)
	if time.Since(session.CreatedAt) > 5*time.Minute {
		delete(s.sessions, sessionID)
		return nil, false
	}

	return session, true
}

func (s *InMemoryStore) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
}

// CleanupExpiredSessions removes old sessions (would run periodically in production)
func (s *InMemoryStore) CleanupExpiredSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for sessionID, session := range s.sessions {
		if now.Sub(session.CreatedAt) > 5*time.Minute {
			delete(s.sessions, sessionID)
		}
	}
}

// removeDuplicateCredentials removes duplicate credentials based on credential ID
func removeDuplicateCredentials(credentials []webauthn.Credential) []webauthn.Credential {
	seen := make(map[string]bool)
	var unique []webauthn.Credential
	
	for _, cred := range credentials {
		credID := string(cred.ID)
		if !seen[credID] {
			seen[credID] = true
			unique = append(unique, cred)
		}
	}
	
	return unique
}

// generatePasskeyName creates a human-friendly name for a WebAuthn credential.
//
// This function analyzes the credential's properties to generate descriptive names
// that help users identify their passkeys in management interfaces.
//
// Naming Strategy:
//  1. Platform authenticators (built-in): "Platform Passkey" or "Synced Platform Passkey"
//  2. Transport-based detection: USB, NFC, Bluetooth, or Hybrid
//  3. Backup state consideration: "Synced" for cloud-backed credentials
//  4. Fallback: Generic "Security Key" for unknown types
//
// In production applications, consider:
//  - AAGUID-based device detection for specific device names
//  - User-defined custom names
//  - Localization for international users
//  - Device type detection (iPhone, Android, Windows, etc.)
//
// Returns a user-friendly string describing the credential type and capabilities.
func generatePasskeyName(cred webauthn.Credential) string {
	// In a real app, you might detect device type based on AAGUID
	// or let users name their passkeys
	
	// Consider attachment type first
	attachment := string(cred.Authenticator.Attachment)
	if attachment == "platform" {
		if cred.Flags.BackupState {
			return "Synced Platform Passkey"
		}
		return "Platform Passkey"
	}
	
	// For cross-platform authenticators, use transport info
	if len(cred.Transport) > 0 {
		switch cred.Transport[0] {
		case "internal":
			return "Device Passkey"
		case "usb":
			return "USB Security Key"
		case "nfc":
			return "NFC Security Key"
		case "ble":
			return "Bluetooth Security Key"
		case "hybrid":
			if cred.Flags.BackupState {
				return "Synced Phone/Tablet"
			}
			return "Phone/Tablet Passkey"
		}
	}
	
	// Fallback based on backup state
	if cred.Flags.BackupEligible {
		if cred.Flags.BackupState {
			return "Synced Passkey"
		}
		return "Backup-Eligible Passkey"
	}
	
	return "Security Key"
}

// Application-specific errors with structured error codes.
//
// These errors provide both human-readable messages and machine-readable codes
// for proper error handling in client applications.
//
// Error Design Pattern:
//  - Code: Machine-readable identifier for programmatic handling
//  - Message: Human-readable description for logging and debugging
//  - JSON serializable for consistent API error responses
//
// Usage:
//   if err == ErrUserExists {
//       return http.StatusConflict
//   }
var (
	ErrUserExists         = &AppError{Code: "USER_EXISTS", Message: "User already exists"}
	ErrUserNotFound       = &AppError{Code: "USER_NOT_FOUND", Message: "User not found"}
	ErrCredentialNotFound = &AppError{Code: "CREDENTIAL_NOT_FOUND", Message: "Credential not found"}
	ErrInvalidSession     = &AppError{Code: "INVALID_SESSION", Message: "Invalid or expired session"}
)

// AppError represents a structured application error with both code and message.
//
// This design allows for:
//  - Consistent error responses across the API
//  - Client-side error handling based on error codes
//  - Human-readable messages for debugging
//  - Easy localization by translating based on codes
//
// Implements the standard error interface for compatibility with Go error handling.
type AppError struct {
	Code    string `json:"code"`    // Machine-readable error code
	Message string `json:"message"` // Human-readable error message
}

// Error implements the error interface, returning the human-readable message.
//
// This allows AppError to be used anywhere a standard Go error is expected
// while preserving the additional structure for API responses.
func (e *AppError) Error() string {
	return e.Message
}