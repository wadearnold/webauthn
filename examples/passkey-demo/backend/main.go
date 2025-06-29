package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func main() {
	// Initialize WebAuthn
	// Cross-platform WebAuthn configuration using local domain
	// This enables passkey sharing across web, iOS, and Android platforms
	config := &webauthn.Config{
		RPDisplayName: "WebAuthn Passkey Demo",
		RPID:          "passkey-demo.local", // Local domain for cross-platform compatibility
		RPOrigins: []string{
			"http://passkey-demo.local:5173",  // React frontend
			"http://passkey-demo.local:3000",  // Alternative React port
			"http://passkey-demo.local:8080",  // API server (for mobile apps)
			"capacitor://passkey-demo.local",  // Capacitor hybrid apps
			"ionic://passkey-demo.local",     // Ionic hybrid apps
			// Native mobile apps will use app-specific origins but same RPID
		},
		AttestationPreference: protocol.PreferNoAttestation,
		// Default authenticator selection - will be overridden per-request
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			// Platform authenticators (built-in biometrics) preferred but not required
			AuthenticatorAttachment: protocol.Platform,
			// Require resident keys for discoverable credentials (passkeys)
			ResidentKey: protocol.ResidentKeyRequirementPreferred,
			RequireResidentKey: protocol.ResidentKeyNotRequired(),
			// User verification preferred to allow fallback if biometrics unavailable
			UserVerification: protocol.VerificationPreferred,
		},
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

	webAuthn, err := webauthn.New(config)
	if err != nil {
		log.Fatalf("Failed to create WebAuthn instance: %v", err)
	}

	// Initialize in-memory store
	store := NewInMemoryStore()

	// Create app with dependencies
	app := &App{
		webAuthn: webAuthn,
		store:    store,
	}

	// Start cleanup routine for expired sessions
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			store.CleanupExpiredSessions()
		}
	}()

	// Setup routes
	mux := http.NewServeMux()

	// Registration endpoints
	mux.HandleFunc("/api/register/begin", app.handleRegisterBegin)
	mux.HandleFunc("/api/register/finish", app.handleRegisterFinish)

	// Authentication endpoints  
	mux.HandleFunc("/api/login/begin", app.handleLoginBegin)
	mux.HandleFunc("/api/login/finish", app.handleLoginFinish)

	// User management endpoints
	mux.HandleFunc("/api/user/passkeys", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			app.handleGetPasskeys(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/user/passkeys/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "DELETE":
			app.handleDeletePasskey(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		app.handleLogout(w, r)
	})

	// Protected profile endpoint
	mux.HandleFunc("/api/user/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		app.handleGetProfile(w, r)
	})

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Apply middleware
	handler := corsMiddleware(
		loggingMiddleware(
			app.sessionMiddleware(
				jsonMiddleware(mux),
			),
		),
	)

	// Start server
	fmt.Println("🚀 WebAuthn Passkey Demo Server starting on :8080")
	fmt.Println("📱 Frontend should be running on http://localhost:5173")
	fmt.Println("🔐 WebAuthn RPID: localhost")
	fmt.Println("🌐 Allowed origins: http://localhost:5173")

	log.Fatal(http.ListenAndServe(":8080", handler))
}