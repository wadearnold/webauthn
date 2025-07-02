package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Global logger instance
var logger = NewLogger("passkey-backend")

func main() {
	// Parse command line flags
	localhost := flag.Bool("localhost", false, "Force localhost mode (ignore NGROK_URL)")
	flag.Parse()

	// Get ngrok URL from environment variable or force localhost
	var ngrokURL string
	if *localhost {
		ngrokURL = "https://your-tunnel.ngrok.io" // Force localhost mode
		logger.Printf("🏠 Localhost mode forced via -localhost flag")
	} else {
		ngrokURL = os.Getenv("NGROK_URL")
		if ngrokURL == "" {
			ngrokURL = "https://your-tunnel.ngrok.io" // Placeholder
		}
	}
	
	// Extract domain from ngrok URL for RPID
	rpid := "localhost" // Default fallback
	if ngrokURL != "https://your-tunnel.ngrok.io" {
		// Extract domain from ngrok URL (e.g., "https://abc123.ngrok.io" -> "abc123.ngrok.io")
		if len(ngrokURL) > 8 { // Remove "https://"
			rpid = ngrokURL[8:]
		}
	}

	// Initialize WebAuthn with ngrok-based configuration
	// This enables passkey sharing across web and iOS platforms using ngrok tunneling
	config := &webauthn.Config{
		RPDisplayName: "WebAuthn Passkey Demo",
		RPID:          rpid, // Use ngrok domain for cross-platform compatibility
		RPOrigins: []string{
			// ngrok tunnel (primary for production-like testing)
			ngrokURL,
			// Localhost fallback for development
			"http://localhost:5173", // React dev server fallback
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

	// Setup routes - organized by middleware requirements

	// Main mux for all routes
	mainMux := http.NewServeMux()
	
	// Create separate API mux with proper routing
	apiMux := http.NewServeMux()
	
	// Registration endpoints
	apiMux.HandleFunc("/api/register/begin", app.handleRegisterBegin)
	apiMux.HandleFunc("/api/register/finish", app.handleRegisterFinish)

	// Authentication endpoints  
	apiMux.HandleFunc("/api/login/begin", app.handleLoginBegin)
	apiMux.HandleFunc("/api/login/finish", app.handleLoginFinish)

	// Other endpoints
	apiMux.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		app.handleLogout(w, r)
	})

	// Health check
	apiMux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// User routes handler - handles all /api/user/* routes
	apiMux.HandleFunc("/api/user/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		
		if strings.HasPrefix(path, "/api/user/passkeys/") && len(path) > len("/api/user/passkeys/") {
			// Handle passkey deletion: /api/user/passkeys/{id}
			switch r.Method {
			case "DELETE":
				app.handleDeletePasskey(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if path == "/api/user/passkeys" {
			// Handle passkey listing: /api/user/passkeys
			switch r.Method {
			case "GET":
				app.handleGetPasskeys(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.Contains(path, "/profile") || path == "/api/user/" {
			// Handle profile: /api/user/ or /api/user/{username}/profile
			if r.Method != "GET" {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			app.handleGetProfile(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	
	// Apply middleware to API routes
	apiHandler := corsMiddleware(
		logger.LogHTTP(
			app.sessionMiddleware(
				jsonMiddleware(apiMux),
			),
		),
	)
	
	// Mount API handler  
	mainMux.Handle("/api/", apiHandler)
	
	// Static files without middleware
	mainMux.HandleFunc("/.well-known/apple-app-site-association", func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("🍎 AASA file requested from: %s (User-Agent: %s)", r.RemoteAddr, r.UserAgent())
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "static/.well-known/apple-app-site-association")
	})
	mainMux.Handle("/.well-known/", http.StripPrefix("/.well-known/", http.FileServer(http.Dir("static/.well-known/"))))
	mainMux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	
	// React app serving
	reactDistPath := "../frontend-react/dist"
	if _, err := os.Stat(reactDistPath); err == nil {
		fmt.Println("📦 Serving React build from /frontend-react/dist")
		
		// Serve React static assets
		mainMux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(reactDistPath+"/assets/"))))
		mainMux.HandleFunc("/vite.svg", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, reactDistPath+"/vite.svg")
		})
		
		// Catch-all: serve index.html for SPA routing
		mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("Serving HTML for: %s", r.URL.Path)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFile(w, r, reactDistPath+"/index.html")
		})
	}

	// Start server with mode-aware output
	fmt.Println("🚀 WebAuthn Passkey Demo Backend")
	fmt.Println("=================================")
	fmt.Printf("🔐 RPID: %s\n", config.RPID)
	
	if rpid == "localhost" {
		fmt.Println("📍 Mode: Local Development")
		fmt.Println("🏠 API: http://localhost:8080")
		fmt.Println("")
		fmt.Println("🌐 Access frontend at:")
		fmt.Println("   http://localhost:5173 (with hot reload)")
		fmt.Println("")
		fmt.Println("⚠️  Note: Cross-platform passkeys won't work in localhost mode")
		fmt.Println("   Use ngrok mode for iOS/cross-platform testing")
	} else {
		fmt.Println("🌍 Mode: ngrok (Cross-Platform)")
		fmt.Printf("📡 Public API: %s/api\n", ngrokURL)
		fmt.Println("")
		if _, err := os.Stat("../frontend-react/dist"); err == nil {
			fmt.Println("🌐 Access app at:")
			fmt.Printf("   %s\n", ngrokURL)
			fmt.Println("   (serving React build)")
		} else {
			fmt.Println("⚠️  React build not found!")
			fmt.Println("   Run: cd frontend-react && npm run build")
		}
		fmt.Println("")
		fmt.Println("✅ Cross-platform passkeys enabled")
	}

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mainMux,
	}

	fmt.Println("🌟 Starting server on port 8080...")
	log.Fatal(server.ListenAndServe())
}