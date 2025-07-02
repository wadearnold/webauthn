package main

import (
	"net/http"
	"os"
	"strings"
)

// CORS middleware for multi-platform development with ngrok
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from multiple frontend platforms
		origin := r.Header.Get("Origin")
		
		// Get ngrok URL from environment
		ngrokURL := os.Getenv("NGROK_URL")
		
		allowedOrigins := []string{
			// Local development
			"http://localhost:3000",          // React dev server
			"http://localhost:5173",          // Vite dev server
			"https://localhost:3000",         // React dev server HTTPS
			"https://localhost:5173",         // Vite dev server HTTPS
			// ngrok tunnel (dynamic)
			ngrokURL,                         // Main ngrok URL
		}
		
		// Check if origin is allowed
		originAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				originAllowed = true
				break
			}
		}
		
		// Also allow any ngrok.io domain for flexibility
		if !originAllowed && origin != "" {
			if strings.Contains(origin, ".ngrok.io") || 
			   strings.Contains(origin, "localhost") {
				originAllowed = true
			}
		}
		
		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Default to ngrok URL
			if ngrokURL != "" {
				w.Header().Set("Access-Control-Allow-Origin", ngrokURL)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			}
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// JSON middleware sets content type for API routes only
func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only set JSON content type for API routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}

// Session middleware to extract session info
func (app *App) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract session ID from cookie
		cookie, err := r.Cookie("webauthn-session")
		if err == nil {
			// Add session ID to request context
			ctx := r.Context()
			ctx = setSessionID(ctx, cookie.Value)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
