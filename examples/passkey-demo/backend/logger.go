package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// CustomLogger implements logging with Console.app compatible timestamps
type CustomLogger struct {
	serviceName string
}

// NewLogger creates a new logger with the given service name
func NewLogger(serviceName string) *CustomLogger {
	return &CustomLogger{
		serviceName: serviceName,
	}
}

// getTimestamp returns a timestamp in Console.app format: HH:MM:SS.microseconds-ZZZZ
func (l *CustomLogger) getTimestamp() string {
	now := time.Now()
	_, offset := now.Zone()
	offsetHours := offset / 3600
	offsetMinutes := (offset % 3600) / 60
	
	// Format: HH:MM:SS.microseconds-HHMM
	timestamp := fmt.Sprintf("%02d:%02d:%02d.%06d%+03d%02d",
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond()/1000, // Convert to microseconds
		offsetHours,
		offsetMinutes,
	)
	
	return timestamp
}

// Printf logs a formatted message with timestamp
func (l *CustomLogger) Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("%s    %s\n", l.getTimestamp(), message)
}

// Errorf logs an error message with timestamp
func (l *CustomLogger) Errorf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s    ERROR: %s\n", l.getTimestamp(), message)
}

// Info logs an info message with timestamp
func (l *CustomLogger) Info(message string) {
	fmt.Printf("%s    %s\n", l.getTimestamp(), message)
}

// Error logs an error message with timestamp
func (l *CustomLogger) Error(message string) {
	fmt.Fprintf(os.Stderr, "%s    ERROR: %s\n", l.getTimestamp(), message)
}

// HTTP middleware for logging requests
func (l *CustomLogger) LogHTTP(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Log incoming request
		l.Printf("→ %s %s (from %s, User-Agent: %s)",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.UserAgent(),
		)
		
		// Wrap response writer to capture status
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Handle request
		handler.ServeHTTP(wrapped, r)
		
		// Log response
		duration := time.Since(start)
		l.Printf("← %s %s [%d] (%v)",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}