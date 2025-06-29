package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
			// HTTPS origins (preferred for production-like testing)
			"https://passkey-demo.local:5173", // React frontend (HTTPS)
			"https://passkey-demo.local:3000", // Alternative React port (HTTPS)
			"https://passkey-demo.local:8080", // API server (HTTPS)
			// Cross-platform domain (HTTP fallback)
			"http://passkey-demo.local:5173",  // React frontend (HTTP)
			"http://passkey-demo.local:3000",  // Alternative React port (HTTP)
			"http://passkey-demo.local:8080",  // API server (HTTP)
			"capacitor://passkey-demo.local",  // Capacitor hybrid apps
			"ionic://passkey-demo.local",     // Ionic hybrid apps
			// Development fallback (WebAuthn works without HTTPS on localhost)
			"https://localhost:5173",         // React dev server (HTTPS)
			"https://localhost:3000",         // Alternative localhost port (HTTPS)
			"https://localhost:8080",         // Backend API localhost access (HTTPS)
			"http://localhost:5173",          // React dev server fallback (HTTP)
			"http://localhost:3000",          // Alternative localhost port (HTTP)
			"http://localhost:8080",          // Backend API localhost access (HTTP)
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

	// Check for HTTPS certificates
	certFile := filepath.Join("certs", "passkey-demo.local+4.pem")
	keyFile := filepath.Join("certs", "passkey-demo.local+4-key.pem")
	
	// Check if we're in the correct directory or need to look in parent
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		// Try parent directory (when running from backend/)
		certFile = filepath.Join("..", "certs", "passkey-demo.local+4.pem")
		keyFile = filepath.Join("..", "certs", "passkey-demo.local+4-key.pem")
	}
	
	useHTTPS := false
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			useHTTPS = true
		}
	}

	// Start server with cross-platform configuration info
	fmt.Println("🚀 WebAuthn Passkey Demo Server")
	fmt.Println("===============================")
	fmt.Println("🌐 Cross-Platform Configuration:")
	fmt.Printf("🔐 WebAuthn RPID: %s\n", config.RPID)
	
	if useHTTPS {
		fmt.Println("🔒 HTTPS Mode: ENABLED")
		fmt.Println("📱 React Frontend: https://passkey-demo.local:5173")
		fmt.Println("📡 Backend API: https://passkey-demo.local:8080")
		fmt.Printf("📜 Certificate: %s\n", certFile)
	} else {
		fmt.Println("🔓 HTTP Mode: Fallback (HTTPS certificates not found)")
		fmt.Println("📱 React Frontend: http://passkey-demo.local:5173 (⚠️  requires localhost for WebAuthn)")
		fmt.Println("📡 Backend API: http://passkey-demo.local:8080")
		fmt.Println("💡 Run './setup-https.sh' to enable HTTPS for full WebAuthn support")
	}
	
	fmt.Println("🔄 Allowed Origins:")
	for _, origin := range config.RPOrigins {
		fmt.Printf("   • %s\n", origin)
	}
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: Add '127.0.0.1 passkey-demo.local' to your /etc/hosts file")
	fmt.Println("🔗 See backend/README.md for setup instructions")
	fmt.Println()

	// Start server
	server := &http.Server{
		Addr:    "0.0.0.0:8080", // Listen on all interfaces for iOS device access
		Handler: handler,
	}

	if useHTTPS {
		// Configure TLS
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		
		fmt.Println("🌟 Starting HTTPS server on :8080...")
		log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
	} else {
		fmt.Println("🌟 Starting HTTP server on :8080...")
		log.Fatal(server.ListenAndServe())
	}
}