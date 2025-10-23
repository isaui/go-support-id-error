package errorid

import (
	"fmt"
	"runtime"
	"sync"
)

// ErrorWithID wraps an error with a unique tracking ID
type ErrorWithID struct {
	ID           string                 // Unique error identifier
	Original     error                  // Original error
	Context      string                 // Context where error occurred
	StackTrace   string                 // Stack trace (if enabled)
	Details      map[string]interface{} // Additional metadata
	Timestamp    int64                  // Unix timestamp when error was wrapped
}

// Error implements error interface
func (e *ErrorWithID) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("[%s] %s: %v", e.ID, e.Context, e.Original)
	}
	return fmt.Sprintf("[%s] %v", e.ID, e.Original)
}

// Unwrap returns the original error for errors.Is and errors.As
func (e *ErrorWithID) Unwrap() error {
	return e.Original
}

var (
	defaultHandler = New(DefaultConfig())  // Direct initialization like stdlib
	configureMu    sync.Mutex
	configured     bool
)

// Configure sets up the global default handler with custom config
// Must be called at program startup, before any Wrap() calls
// Can only be called once - subsequent calls will panic
func Configure(cfg Config) {
	configureMu.Lock()
	defer configureMu.Unlock()
	
	if configured {
		panic("errorid: Configure() called multiple times. Call Configure() only once at program startup.")
	}
	
	defaultHandler = New(cfg)
	configured = true
}

// Wrap wraps an error with a unique ID using the default handler
// This is the main singleton API for simple use cases
func Wrap(err error, context string) *ErrorWithID {
	return defaultHandler.Wrap(err, context)
}

// WrapWithDetails wraps error with additional metadata
func WrapWithDetails(err error, context string, details map[string]interface{}) *ErrorWithID {
	return defaultHandler.WrapWithDetails(err, context, details)
}

// Default returns the singleton handler instance
// Useful for accessing handler methods directly
func Default() *Handler {
	return defaultHandler
}

// captureStackTrace captures current stack trace
func captureStackTrace(skip int) string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
