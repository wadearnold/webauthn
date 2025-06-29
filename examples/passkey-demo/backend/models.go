package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/go-webauthn/webauthn/webauthn"
)

// User represents a user in our system
type User struct {
	ID          []byte                  `json:"id"`
	Username    string                  `json:"username"`
	DisplayName string                  `json:"displayName"`
	Credentials []webauthn.Credential   `json:"credentials"`
	CreatedAt   time.Time               `json:"createdAt"`
}

// WebAuthn interface implementation
func (u User) WebAuthnID() []byte {
	return u.ID
}

func (u User) WebAuthnName() string {
	return u.Username
}

func (u User) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

// Session represents a WebAuthn session
type Session struct {
	UserID      []byte                  `json:"userId"`
	SessionData webauthn.SessionData    `json:"sessionData"`
	CreatedAt   time.Time               `json:"createdAt"`
}

// PasskeyInfo represents a passkey for frontend display
type PasskeyInfo struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	CreatedAt        time.Time `json:"createdAt"`
	LastUsed         time.Time `json:"lastUsed"`
	Transports       []string  `json:"transports"`
	BackedUp         bool      `json:"backedUp"`
	BackupEligible   bool      `json:"backupEligible"`
	UserVerified     bool      `json:"userVerified"`
	AttestationType  string    `json:"attestationType"`
	AuthenticatorAttachment string `json:"authenticatorAttachment"`
	SignCount        uint32    `json:"signCount"`
	AAGUID          string    `json:"aaguid"`
	// User information associated with this credential
	Username        string    `json:"username"`
	DisplayName     string    `json:"displayName"`
}

// InMemoryStore provides thread-safe in-memory storage
type InMemoryStore struct {
	users    map[string]*User    // username -> User
	userIDs  map[string]*User    // base64(userID) -> User  
	sessions map[string]*Session // sessionID -> Session
	mu       sync.RWMutex
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		users:    make(map[string]*User),
		userIDs:  make(map[string]*User),
		sessions: make(map[string]*Session),
	}
}

// CreateUser creates a new user
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

// GetUser retrieves a user by username
func (s *InMemoryStore) GetUser(username string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	return user, exists
}

// GetUserByID retrieves a user by WebAuthn user ID
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

// generatePasskeyName creates a friendly name for the passkey
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

// Custom errors
var (
	ErrUserExists         = &AppError{Code: "USER_EXISTS", Message: "User already exists"}
	ErrUserNotFound       = &AppError{Code: "USER_NOT_FOUND", Message: "User not found"}
	ErrCredentialNotFound = &AppError{Code: "CREDENTIAL_NOT_FOUND", Message: "Credential not found"}
	ErrInvalidSession     = &AppError{Code: "INVALID_SESSION", Message: "Invalid or expired session"}
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}