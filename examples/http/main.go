package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	
	errorid "github.com/isaui/go-support-id-error"
)

func main() {
	fmt.Println("=== HTTP Middleware Example ===")
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("\nEndpoints:")
	fmt.Println("  - GET  /ok       -> Returns success")
	fmt.Println("  - GET  /error    -> Returns handled error")
	fmt.Println("  - GET  /panic    -> Triggers panic (recovered by middleware)")
	fmt.Println()
	
	// Configure error handling
	errorid.Configure(errorid.Config{
		OnError: func(err *errorid.ErrorWithID) {
			// Could send to Sentry, Slack, etc.
			fmt.Printf("🚨 Error logged: %s\n", err.ID)
		},
		AsyncCallback:     false,
		Logger:            errorid.NewDefaultLogger(os.Stdout),
		IncludeStackTrace: true,
		Environment:       "development", // Will show detailed errors
	})
	
	// Create router
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", handleOK)
	mux.HandleFunc("/error", handleError)
	mux.HandleFunc("/panic", handlePanic)
	
	// Wrap with recovery middleware
	handler := errorid.RecoveryMiddleware(mux)
	
	// Start server
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

func handleOK(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok", "message": "Everything is working!"}`))
}

func handleError(w http.ResponseWriter, r *http.Request) {
	// Simulate a business logic error
	err := processUserData()
	if err != nil {
		// Wrap the error with ID and write response
		wrapped := errorid.WrapWithDetails(err, "user data processing failed", map[string]interface{}{
			"user_id": r.URL.Query().Get("user_id"),
			"path":    r.URL.Path,
		})
		
		// Manually write error response
		errorid.WriteError(w, wrapped)
		return
	}
	
	w.Write([]byte(`{"status": "success"}`))
}

func handlePanic(w http.ResponseWriter, r *http.Request) {
	// This will panic, but middleware will recover and return error ID
	var data []string
	_ = data[100] // index out of range panic
}

func processUserData() error {
	// Simulate various error scenarios
	return errors.New("database connection failed")
}
