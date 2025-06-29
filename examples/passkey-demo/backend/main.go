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
	config := &webauthn.Config{
		RPDisplayName: "WebAuthn Passkey Demo",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:5173"}, // React dev server
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
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