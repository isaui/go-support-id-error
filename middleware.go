package errorid

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorResponse is the JSON structure returned to clients
type ErrorResponse struct {
	ErrorID   string `json:"error_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// RecoveryMiddleware recovers from panics and returns error ID to client
// Uses the default singleton handler
func RecoveryMiddleware(next http.Handler) http.Handler {
	return Default().RecoveryMiddleware(next)
}

// RecoveryMiddleware creates middleware using this handler instance
func (h *Handler) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Wrap panic as error
				var err error
				switch v := rec.(type) {
				case error:
					err = v
				default:
					err = &panicError{value: rec}
				}
				
				// Wrap with error ID
				wrapped := h.WrapWithDetails(err, "panic recovered in HTTP handler", map[string]interface{}{
					"method": r.Method,
					"path":   r.URL.Path,
					"remote": r.RemoteAddr,
				})
				
				// Return error response to client
				h.writeErrorResponse(w, wrapped)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// writeErrorResponse writes JSON error response to client
func (h *Handler) writeErrorResponse(w http.ResponseWriter, err *ErrorWithID) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	
	message := "An internal error occurred. Please contact support with this error ID."
	
	// In development, show more details
	if h.config.Environment == "development" {
		message = err.Error()
	}
	
	response := ErrorResponse{
		ErrorID:   err.ID,
		Message:   message,
		Timestamp: err.Timestamp,
	}
	
	json.NewEncoder(w).Encode(response)
}

// WriteError is a helper to manually write error responses in handlers
func WriteError(w http.ResponseWriter, err *ErrorWithID) {
	Default().writeErrorResponse(w, err)
}

// WriteErrorWithHandler writes error using specific handler instance
func (h *Handler) WriteError(w http.ResponseWriter, err *ErrorWithID) {
	h.writeErrorResponse(w, err)
}

// panicError wraps a panic value as an error
type panicError struct {
	value interface{}
}

func (e *panicError) Error() string {
	return fmt.Sprint("panic: ", e.value)
}
