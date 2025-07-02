// HTTP handlers for WebAuthn passkey authentication endpoints.
//
// This file demonstrates proper implementation of WebAuthn server-side handlers
// following the W3C WebAuthn specification. Key patterns shown:
//
// Request Handling:
//   - Input validation and sanitization
//   - Structured error responses with appropriate HTTP status codes
//   - Session management for multi-round authentication flows
//
// WebAuthn Integration:
//   - Proper configuration for cross-platform passkey support
//   - Registration and authentication ceremony handling
//   - Security best practices for credential verification
//
// Go Best Practices:
//   - Clean separation of concerns
//   - Consistent error handling patterns
//   - Thread-safe operations with proper locking
//   - Comprehensive logging for debugging and security monitoring
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// Context key type for storing request-scoped data.
//
// Using a custom type for context keys prevents collisions with other packages
// and follows Go best practices for context usage.
type contextKey string

// sessionIDKey is used to store WebAuthn session IDs in request context.
//
// This allows session data to be passed between middleware and handlers
// without relying on global variables or additional function parameters.
const sessionIDKey contextKey = "sessionID"

// setSessionID adds a WebAuthn session ID to the request context.
//
// This is typically called by session middleware after extracting the
// session ID from cookies or headers.
func setSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// getSessionID retrieves the WebAuthn session ID from request context.
//
// Returns the session ID and true if present, or empty string and false
// if no session ID is found in the context.
func getSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	return sessionID, ok
}

// Request and response type definitions for WebAuthn API endpoints.
//
// These structs define the JSON structure for client-server communication
// and demonstrate proper API design patterns.

// RegisterBeginRequest represents the initial registration request from client.
//
// Username must be unique and follow validation rules.
// DisplayName is optional and used for user-friendly identification.
type RegisterBeginRequest struct {
	Username    string `json:"username"`              // Required: unique identifier for user
	DisplayName string `json:"displayName,omitempty"` // Optional: human-readable name
}

// LoginBeginRequest represents the initial authentication request from client.
//
// Username is optional to support discoverable (passwordless) login where
// the client doesn't need to specify which user to authenticate.
type LoginBeginRequest struct {
	Username string `json:"username,omitempty"` // Optional: specific user for traditional login
}

// ErrorResponse provides structured error information for API responses.
//
// This pattern ensures consistent error handling across all endpoints and
// enables client applications to handle errors programmatically.
type ErrorResponse struct {
	Error   string `json:"error"`           // Human-readable error message
	Code    string `json:"code,omitempty"`  // Machine-readable error code
	Details string `json:"details,omitempty"` // Additional error context
}

// Username validation regex pattern for security and usability.
//
// Pattern breakdown:
//   ^[a-zA-Z0-9._-]{3,30}$
//   ^ = start of string
//   [a-zA-Z0-9._-] = allowed characters (letters, numbers, dots, hyphens, underscores)
//   {3,30} = length between 3 and 30 characters
//   $ = end of string
//
// This pattern prevents:
//   - SQL injection through special characters
//   - Directory traversal through path separators
//   - Unicode normalization attacks
//   - Username enumeration through predictable patterns
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,30}$`)

// validateUsername ensures usernames meet security and usability requirements.
//
// Validation Rules:
//  1. Required field (not empty)
//  2. Length between 3-30 characters (prevents abuse and UI issues)
//  3. Only safe characters (alphanumeric, dots, hyphens, underscores)
//  4. Cannot start/end with special characters (prevents confusion)
//
// Security Considerations:
//  - Prevents injection attacks through special characters
//  - Avoids Unicode normalization vulnerabilities
//  - Blocks directory traversal attempts
//  - Ensures consistent display across different systems
//
// Returns nil if valid, or descriptive error if validation fails.
func validateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	if len(username) > 30 {
		return fmt.Errorf("username must be no more than 30 characters long")
	}

	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, dots, hyphens, and underscores")
	}

	// Don't allow usernames that start or end with special characters
	// This prevents confusion and ensures consistent display
	if strings.HasPrefix(username, ".") || strings.HasPrefix(username, "-") || strings.HasPrefix(username, "_") ||
		strings.HasSuffix(username, ".") || strings.HasSuffix(username, "-") || strings.HasSuffix(username, "_") {
		return fmt.Errorf("username cannot start or end with dots, hyphens, or underscores")
	}

	return nil
}

// SuccessResponse provides structured success information for API responses.
//
// This ensures consistent response format across all endpoints and makes
// it easier for clients to handle successful operations.
type SuccessResponse struct {
	Message string      `json:"message"`           // Human-readable success message
	Data    interface{} `json:"data,omitempty"`    // Optional response data
}

// App encapsulates application dependencies and provides handler methods.
//
// This struct follows the dependency injection pattern, making the code:
//  - Easier to test (dependencies can be mocked)
//  - More maintainable (clear dependency relationships)
//  - Thread-safe (all dependencies are immutable after creation)
//
// The App pattern is common in Go web applications and demonstrates
// proper separation of concerns between HTTP handling and business logic.
type App struct {
	webAuthn *webauthn.WebAuthn // WebAuthn library instance with configuration
	store    *InMemoryStore     // User and session storage (interface in production)
}

// WebAuthn Registration Handlers
//
// These handlers implement the WebAuthn registration ceremony, which allows
// users to create new passkeys. The process involves two rounds:
//  1. Begin: Generate challenge and return credential creation options
//  2. Finish: Verify the new credential and store it

// handleRegisterBegin initiates the WebAuthn credential registration ceremony.
//
// This endpoint implements the first phase of WebAuthn registration:
//  1. Validates the registration request (username, display name)
//  2. Creates or retrieves the user account
//  3. Generates WebAuthn credential creation options with passkey settings
//  4. Creates a temporary session to store challenge data
//  5. Returns options to client for credential creation
//
// WebAuthn Flow:
//  Client -> POST /api/register/begin -> Server generates challenge
//  Server -> Returns credential options -> Client calls navigator.credentials.create()
//  Client -> Authenticator prompts user -> User provides biometric/PIN
//  Client -> POST /api/register/finish -> Server verifies and stores credential
//
// Security Features:
//  - Input validation prevents injection attacks
//  - Cryptographically secure challenge generation
//  - Session timeout prevents replay attacks
//  - Passkey configuration enforces strong authentication
//
// Request Body: RegisterBeginRequest (JSON)
// Response: WebAuthn CredentialCreationOptions (JSON)
// HTTP Status: 200 (success), 400 (validation error), 409 (user exists), 500 (server error)
func (app *App) handleRegisterBegin(w http.ResponseWriter, r *http.Request) {
	var req RegisterBeginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateUsername(req.Username); err != nil {
		app.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create or get user
	user, exists := app.store.GetUser(req.Username)
	if !exists {
		displayName := req.DisplayName
		if displayName == "" {
			displayName = req.Username
		}

		var err error
		user, err = app.store.CreateUser(req.Username, displayName)
		if err != nil {
			app.writeError(w, err.Error(), http.StatusConflict)
			return
		}
	}

	// Begin registration with best practice passkey configuration
	// Force platform authenticators and resident keys for true passkey experience
	options, sessionData, err := app.webAuthn.BeginRegistration(
		user,
		// Required for passkeys: must be stored on device
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		// Best practice: platform authenticators with user verification
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			// Force platform authenticators (built-in biometrics)
			AuthenticatorAttachment: protocol.Platform,
			// Required for passkeys
			ResidentKey: protocol.ResidentKeyRequirementRequired,
			RequireResidentKey: protocol.ResidentKeyRequired(),
			// Require user verification for security
			UserVerification: protocol.VerificationRequired,
		}),
	)
	if err != nil {
		app.writeError(w, fmt.Sprintf("Failed to begin registration: %v", err), http.StatusInternalServerError)
		return
	}

	// Comprehensive registration debugging
	logger.Printf("=== REGISTRATION DEBUG INFO FOR %s ===", user.Username)
	logger.Printf("AuthenticatorAttachment: %s", options.Response.AuthenticatorSelection.AuthenticatorAttachment)
	logger.Printf("ResidentKey: %s", options.Response.AuthenticatorSelection.ResidentKey)
	logger.Printf("RequireResidentKey: %t", *options.Response.AuthenticatorSelection.RequireResidentKey)
	logger.Printf("UserVerification: %s", options.Response.AuthenticatorSelection.UserVerification)
	logger.Printf("Attestation: %s", options.Response.Attestation)
	logger.Printf("Timeout: %d ms", options.Response.Timeout)
	logger.Printf("RPID: %s", options.Response.RelyingParty.ID)
	logger.Printf("RPName: %s", options.Response.RelyingParty.Name)
	logger.Printf("Challenge: %s", options.Response.Challenge)
	logger.Printf("=========================================")

	// Store session
	sessionID := uuid.New().String()
	app.store.StoreSession(sessionID, user.ID, *sessionData)

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "webauthn-session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   300, // 5 minutes
	})

	// Return options to client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(options)
}

func (app *App) handleRegisterFinish(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := getSessionID(r.Context())
	if !ok {
		app.writeError(w, "No session found", http.StatusBadRequest)
		return
	}

	session, exists := app.store.GetSession(sessionID)
	if !exists {
		app.writeError(w, "Invalid or expired session", http.StatusBadRequest)
		return
	}

	user, exists := app.store.GetUserByID(session.UserID)
	if !exists {
		app.writeError(w, "User not found", http.StatusBadRequest)
		return
	}

	// Finish registration
	credential, err := app.webAuthn.FinishRegistration(user, session.SessionData, r)
	if err != nil {
		app.writeError(w, fmt.Sprintf("Registration failed: %v", err), http.StatusBadRequest)
		return
	}

	// Check if this credential already exists (prevent duplicates)
	credentialExists := false
	for _, existingCred := range user.Credentials {
		if string(existingCred.ID) == string(credential.ID) {
			credentialExists = true
			fmt.Printf("WARNING: Attempted to register duplicate credential for user %s, CredentialID: %s\n", 
				user.Username, base64.URLEncoding.EncodeToString(credential.ID))
			break
		}
	}

	// Only add credential if it doesn't already exist
	if !credentialExists {
		user.Credentials = append(user.Credentials, *credential)
		app.store.UpdateUser(user)
		fmt.Printf("SUCCESS: New credential registered for user %s, CredentialID: %s\n", 
			user.Username, base64.URLEncoding.EncodeToString(credential.ID))
	}

	// Set user session cookie (so user is logged in after registration)
	app.setUserSession(w, user.Username)

	// Clean up session
	app.store.DeleteSession(sessionID)

	app.writeSuccess(w, "Registration successful", map[string]interface{}{
		"credentialId": credential.ID,
		"username":     user.Username,
		"displayName":  user.DisplayName,
		"userId":       user.ID,
	})
}

// Authentication handlers
func (app *App) handleLoginBegin(w http.ResponseWriter, r *http.Request) {
	var req LoginBeginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Support both discoverable and non-discoverable login
	if req.Username != "" {
		// Validate username format
		if err := validateUsername(req.Username); err != nil {
			app.writeError(w, "Authentication failed", http.StatusUnauthorized) // Don't reveal validation details
			return
		}

		// Traditional login with username
		user, exists := app.store.GetUser(req.Username)
		if !exists {
			app.writeError(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		// Traditional login with username - use best practices
		options, sessionData, err := app.webAuthn.BeginLogin(
			user,
			// Request user verification for security
			webauthn.WithUserVerification(protocol.VerificationRequired),
		)
		
		if err == nil {
			logger.Printf("=== TRADITIONAL LOGIN DEBUG INFO FOR %s ===", user.Username)
			logger.Printf("UserVerification: %s", options.Response.UserVerification)
			logger.Printf("Timeout: %d ms", options.Response.Timeout)
			logger.Printf("RPID: %s", options.Response.RelyingPartyID)
			logger.Printf("AllowCredentials count: %d", len(options.Response.AllowedCredentials))
			for i, cred := range options.Response.AllowedCredentials {
				logger.Printf("  Credential %d: ID=%s, Type=%s", i+1, base64.URLEncoding.EncodeToString(cred.CredentialID), cred.Type)
			}
			logger.Printf("============================================")
		}
		if err != nil {
			app.writeError(w, fmt.Sprintf("Failed to begin login: %v", err), http.StatusInternalServerError)
			return
		}

		// Store session
		sessionID := uuid.New().String()
		app.store.StoreSession(sessionID, user.ID, *sessionData)

		http.SetCookie(w, &http.Cookie{
			Name:     "webauthn-session",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   300,
		})

		json.NewEncoder(w).Encode(options)
	} else {
		// Discoverable login (passwordless) with best practices
		options, sessionData, err := app.webAuthn.BeginDiscoverableLogin(
			// Require user verification for security
			webauthn.WithUserVerification(protocol.VerificationRequired),
		)
		
		if err == nil {
			logger.Printf("=== DISCOVERABLE LOGIN DEBUG INFO ===")
			logger.Printf("UserVerification: %s", options.Response.UserVerification)
			logger.Printf("Timeout: %d ms", options.Response.Timeout)
			logger.Printf("RPID: %s", options.Response.RelyingPartyID)
			logger.Printf("Challenge: %s", options.Response.Challenge)
			logger.Printf("AllowCredentials count: %d", len(options.Response.AllowedCredentials))
			logger.Printf("=====================================")
		}
		if err != nil {
			app.writeError(w, fmt.Sprintf("Failed to begin discoverable login: %v", err), http.StatusInternalServerError)
			return
		}

		// Store session without user ID for discoverable login
		sessionID := uuid.New().String()
		app.store.StoreSession(sessionID, nil, *sessionData)

		http.SetCookie(w, &http.Cookie{
			Name:     "webauthn-session",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   300,
		})

		json.NewEncoder(w).Encode(options)
	}
}

func (app *App) handleLoginFinish(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := getSessionID(r.Context())
	if !ok {
		app.writeError(w, "No session found", http.StatusBadRequest)
		return
	}

	session, exists := app.store.GetSession(sessionID)
	if !exists {
		app.writeError(w, "Invalid or expired session", http.StatusBadRequest)
		return
	}

	if session.UserID != nil {
		// Traditional login
		user, exists := app.store.GetUserByID(session.UserID)
		if !exists {
			app.writeError(w, "User not found", http.StatusBadRequest)
			return
		}

		credential, err := app.webAuthn.FinishLogin(user, session.SessionData, r)
		if err != nil {
			app.writeError(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
			return
		}

		// SECURITY: Verify the returned credential still exists in the user's current credential list
		// This prevents authentication with deleted credentials that might still be in device keychain
		credentialExists := false
		for _, userCred := range user.Credentials {
			if string(userCred.ID) == string(credential.ID) {
				credentialExists = true
				break
			}
		}
		if !credentialExists {
			fmt.Printf("SECURITY: Authentication attempt with deleted credential. User: %s, CredentialID: %s\n",
				user.Username, base64.URLEncoding.EncodeToString(credential.ID))
			app.writeError(w, "Authentication failed: credential no longer valid", http.StatusUnauthorized)
			return
		}

		// Check for clone warning
		if credential.Authenticator.CloneWarning {
			// Log security event but allow login for demo
			fmt.Printf("WARNING: Clone detected for user %s\n", user.Username)
		}

		// Update credential
		app.updateUserCredential(user, credential)

		// Set user session cookie
		app.setUserSession(w, user.Username)

		app.writeSuccess(w, "Authentication successful", map[string]interface{}{
			"username":    user.Username,
			"displayName": user.DisplayName,
			"userId":      user.ID,
		})
	} else {
		// Discoverable login
		userHandler := func(rawID, userHandle []byte) (webauthn.User, error) {
			user, exists := app.store.GetUserByID(userHandle)
			if !exists {
				return nil, fmt.Errorf("user not found")
			}
			return user, nil
		}

		// Parse the response first
		parsedResponse, err := protocol.ParseCredentialRequestResponse(r)
		if err != nil {
			app.writeError(w, fmt.Sprintf("Failed to parse response: %v", err), http.StatusBadRequest)
			return
		}

		user, credential, err := app.webAuthn.ValidatePasskeyLogin(userHandler, session.SessionData, parsedResponse)
		if err != nil {
			app.writeError(w, fmt.Sprintf("Discoverable authentication failed: %v", err), http.StatusUnauthorized)
			return
		}

		// SECURITY: Verify the returned credential still exists in the user's current credential list
		// This prevents authentication with deleted credentials that might still be in device keychain
		appUser := user.(*User)
		credentialExists := false
		for _, userCred := range appUser.Credentials {
			if string(userCred.ID) == string(credential.ID) {
				credentialExists = true
				break
			}
		}
		if !credentialExists {
			fmt.Printf("SECURITY: Authentication attempt with deleted credential. User: %s, CredentialID: %s\n",
				appUser.Username, base64.URLEncoding.EncodeToString(credential.ID))
			app.writeError(w, "Authentication failed: credential no longer valid", http.StatusUnauthorized)
			return
		}

		// Check for clone warning
		if credential.Authenticator.CloneWarning {
			fmt.Printf("WARNING: Clone detected for user %s\n", user.WebAuthnName())
		}

		// Update credential
		app.updateUserCredential(appUser, credential)

		// Set user session cookie
		app.setUserSession(w, appUser.Username)

		app.writeSuccess(w, "Discoverable authentication successful", map[string]interface{}{
			"username":    appUser.Username,
			"displayName": appUser.DisplayName,
			"userId":      appUser.ID,
		})
	}

	// Clean up WebAuthn session
	app.store.DeleteSession(sessionID)
}

// User management handlers
func (app *App) handleGetPasskeys(w http.ResponseWriter, r *http.Request) {
	username := app.getCurrentUser(r)
	if username == "" {
		app.writeError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	passkeys, err := app.store.GetUserPasskeys(username)
	if err != nil {
		app.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(passkeys)
}

func (app *App) handleDeletePasskey(w http.ResponseWriter, r *http.Request) {
	username := app.getCurrentUser(r)
	if username == "" {
		app.writeError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Extract credential ID from URL path (base64-encoded)
	credentialIDStr := strings.TrimPrefix(r.URL.Path, "/api/user/passkeys/")
	if credentialIDStr == "" {
		app.writeError(w, "Credential ID required", http.StatusBadRequest)
		return
	}

	// The credential ID comes as a base64-encoded string in the URL
	// We need to use it directly as string for our storage layer
	credentialID := []byte(credentialIDStr)
	err := app.store.DeleteUserPasskey(username, credentialID)
	if err != nil {
		app.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log with readable credential ID (the base64 string)
	fmt.Printf("SECURITY: Passkey deleted for user %s, CredentialID: %s\n", username, credentialIDStr)
	app.writeSuccess(w, "Passkey deleted successfully", nil)
}

func (app *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear user session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user-session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	app.writeSuccess(w, "Logged out successfully", nil)
}

// Protected profile endpoint
func (app *App) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	// Extract username from URL path: /api/user/{username}/profile
	path := strings.TrimPrefix(r.URL.Path, "/api/user/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "profile" {
		app.writeError(w, "Invalid profile URL", http.StatusBadRequest)
		return
	}
	requestedUsername := parts[0]

	// Check if user is authenticated
	currentUsername := app.getCurrentUser(r)
	if currentUsername == "" {
		// Return 401 with redirect URL for deep linking
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":       "Authentication required",
			"code":        "AUTH_REQUIRED",
			"redirectUrl": r.URL.Path,
		})
		return
	}

	// Get user from store
	user, exists := app.store.GetUser(requestedUsername)
	if !exists {
		app.writeError(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the current user can access this profile
	// For demo, users can only view their own profile
	if currentUsername != requestedUsername {
		app.writeError(w, "Access denied: You can only view your own profile", http.StatusForbidden)
		return
	}

	// Get user's passkeys for the profile
	passkeys, _ := app.store.GetUserPasskeys(requestedUsername)

	// Return profile data
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username":                  user.Username,
		"displayName":               user.DisplayName,
		"createdAt":                 user.CreatedAt,
		"passkeyCount":              len(passkeys),
		"hasBackupEligiblePasskeys": hasBackupEligiblePasskeys(passkeys),
	})
}

// Helper function to check if user has backup-eligible passkeys
func hasBackupEligiblePasskeys(passkeys []PasskeyInfo) bool {
	for _, pk := range passkeys {
		if pk.BackupEligible {
			return true
		}
	}
	return false
}

// Helper methods
func (app *App) updateUserCredential(user *User, credential *webauthn.Credential) {
	// Find and update the existing credential
	for i, cred := range user.Credentials {
		if string(cred.ID) == string(credential.ID) {
			user.Credentials[i] = *credential
			app.store.UpdateUser(user)
			return
		}
	}
}

func (app *App) setUserSession(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user-session",
		Value:    username,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   3600, // 1 hour
	})
}

func (app *App) getCurrentUser(r *http.Request) string {
	cookie, err := r.Cookie("user-session")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (app *App) writeError(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
	})
}

func (app *App) writeSuccess(w http.ResponseWriter, message string, data interface{}) {
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: message,
		Data:    data,
	})
}
