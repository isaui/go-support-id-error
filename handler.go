package errorid

import (
	"fmt"
	"time"
)

// Handler manages error wrapping and tracking
type Handler struct {
	config Config
}

// New creates a new Handler instance with custom configuration
// This is the instance-based API for advanced use cases
func New(cfg Config) *Handler {
	// Use default ID generator if not provided
	if cfg.IDGenerator == nil {
		cfg.IDGenerator = GenerateErrorID
	}
	
	// Use default logger if not provided
	if cfg.Logger == nil {
		cfg.Logger = DefaultConfig().Logger
	}
	
	return &Handler{
		config: cfg,
	}
}

// Wrap wraps an error with a unique ID and logs it
func (h *Handler) Wrap(err error, context string) *ErrorWithID {
	return h.WrapWithDetails(err, context, nil)
}

// WrapWithDetails wraps error with additional metadata
func (h *Handler) WrapWithDetails(err error, context string, details map[string]interface{}) *ErrorWithID {
	if err == nil {
		return nil
	}
	
	errorID := h.config.IDGenerator()
	
	wrapped := &ErrorWithID{
		ID:        errorID,
		Original:  err,
		Context:   context,
		Details:   details,
		Timestamp: time.Now().Unix(),
	}
	
	// Capture stack trace if enabled
	if h.config.IncludeStackTrace {
		wrapped.StackTrace = captureStackTrace(2) // skip this function and Wrap
	}
	
	// Log the error
	h.logError(wrapped)
	
	// Execute OnError callback
	if h.config.OnError != nil {
		if h.config.AsyncCallback {
			// Async: run in goroutine
			go h.safeCallback(wrapped)
		} else {
			// Sync: blocking call
			h.safeCallback(wrapped)
		}
	}
	
	return wrapped
}

// logError logs the error using configured logger
func (h *Handler) logError(err *ErrorWithID) {
	if h.config.Logger == nil {
		return
	}
	
	// Copy user details (don't mutate original)
	details := make(map[string]interface{})
	if err.Details != nil {
		for k, v := range err.Details {
			details[k] = v
		}
	}
	
	// Add timestamp
	details["timestamp"] = err.Timestamp
	
	// Log with stack trace as separate parameter (not in details)
	h.config.Logger.Error(err.ID, err.Original, err.Context, details, err.StackTrace)
}

// safeCallback executes OnError callback with panic recovery
func (h *Handler) safeCallback(err *ErrorWithID) {
	defer func() {
		if r := recover(); r != nil {
			// Callback panicked, log it but don't crash
			if h.config.Logger != nil {
				h.config.Logger.Info("OnError callback panicked: " + fmt.Sprint(r))
			}
		}
	}()
	
	h.config.OnError(err)
}

// Config returns current handler configuration (read-only)
func (h *Handler) Config() Config {
	return h.config
}
