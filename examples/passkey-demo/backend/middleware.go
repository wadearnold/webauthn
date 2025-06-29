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
			// HTTPS origins (preferred for production-like testing)
			"https://passkey-demo.local:5173", // React frontend (HTTPS)
			"https://passkey-demo.local:3000", // Alternative React port (HTTPS)
			"https://passkey-demo.local:8080", // API server access (HTTPS)
			"https://localhost:5173",          // Localhost (HTTPS)
			"https://localhost:3000",          // Alternative localhost port (HTTPS)
			"https://localhost:8080",          // Backend API localhost access (HTTPS)
			// iOS Simulator/Device access (add your Mac's IP here)
			"https://192.168.1.100:8080",     // Example IP - replace with your Mac's IP
			"http://192.168.1.100:8080",      // HTTP fallback for IP access
			// HTTP origins (fallback for development)
			"http://passkey-demo.local:5173",  // React frontend (HTTP)
			"http://passkey-demo.local:3000",  // Alternative React port (HTTP)
			"http://passkey-demo.local:8080",  // API server access (HTTP)
			"http://localhost:5173",           // Localhost (HTTP)
			"http://localhost:3000",           // Alternative localhost port (HTTP)
			"http://localhost:8080",           // Backend API localhost access (HTTP)
			// Mobile app origins
			"capacitor://passkey-demo.local",  // Capacitor hybrid apps
			"ionic://passkey-demo.local",     // Ionic hybrid apps
			"capacitor://localhost",
			"ionic://localhost",
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
			// Default to HTTPS local domain for cross-platform compatibility
			// Fall back to HTTP if HTTPS not available
			w.Header().Set("Access-Control-Allow-Origin", "https://passkey-demo.local:5173")
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