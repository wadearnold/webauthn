package main

import (
	"net/http"
)

// CORS middleware for multi-platform development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from multiple frontend platforms
		// React dev server, iOS simulator, Android emulator, etc.
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:5173",     // React frontend
			"http://localhost:3000",     // Alternative React port
			"http://10.0.2.2:8080",      // Android emulator
			"http://127.0.0.1:8080",     // iOS simulator
			"capacitor://localhost",     // Capacitor apps
			"ionic://localhost",        // Ionic apps
		}
		
		// Check if origin is allowed
		originAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				originAllowed = true
				break
			}
		}
		
		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Default to React dev server for backward compatibility
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
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

// JSON middleware sets content type
func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple request logging
		println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}