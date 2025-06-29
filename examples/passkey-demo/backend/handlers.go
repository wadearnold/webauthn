package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Context keys
type contextKey string

const sessionIDKey contextKey = "sessionID"

func setSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

func getSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	return sessionID, ok
}

// Request/Response types
type RegisterBeginRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

type LoginBeginRequest struct {
	Username string `json:"username,omitempty"` // Optional for discoverable login
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// Username validation regex: alphanumeric, hyphens, underscores, dots (3-30 chars)
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,30}$`)

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
	if strings.HasPrefix(username, ".") || strings.HasPrefix(username, "-") || strings.HasPrefix(username, "_") ||
		strings.HasSuffix(username, ".") || strings.HasSuffix(username, "-") || strings.HasSuffix(username, "_") {
		return fmt.Errorf("username cannot start or end with dots, hyphens, or underscores")
	}
	
	return nil
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// App holds application dependencies
type App struct {
	webAuthn *webauthn.WebAuthn
	store    *InMemoryStore
}

// Registration handlers
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

	// Begin registration with discoverable credentials
	options, sessionData, err := app.webAuthn.BeginRegistration(
		user,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			UserVerification: protocol.VerificationRequired,
		}),
	)
	if err != nil {
		app.writeError(w, fmt.Sprintf("Failed to begin registration: %v", err), http.StatusInternalServerError)
		return
	}

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

	// Update user with new credential
	user.Credentials = append(user.Credentials, *credential)
	app.store.UpdateUser(user)

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

		options, sessionData, err := app.webAuthn.BeginLogin(user)
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
		// Discoverable login (passwordless)
		options, sessionData, err := app.webAuthn.BeginDiscoverableLogin(
			webauthn.WithUserVerification(protocol.VerificationRequired),
		)
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

		// Check for clone warning
		if credential.Authenticator.CloneWarning {
			fmt.Printf("WARNING: Clone detected for user %s\n", user.WebAuthnName())
		}

		// Update credential
		appUser := user.(*User)
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

	// Extract credential ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/user/passkeys/")
	if path == "" {
		app.writeError(w, "Credential ID required", http.StatusBadRequest)
		return
	}

	credentialID := []byte(path)
	err := app.store.DeleteUserPasskey(username, credentialID)
	if err != nil {
		app.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

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