package errorid

import (
	"io"
	"log"
	"os"
)

// Config holds configuration for error handler
type Config struct {
	// OnError callback executed when error is wrapped
	// Use this to send errors to external services (Sentry, etc)
	OnError func(*ErrorWithID)

	// AsyncCallback determines if OnError runs in goroutine
	// true = non-blocking, false = blocking
	AsyncCallback bool

	// Logger for error logging. If nil, uses default logger
	Logger Logger

	// IncludeStackTrace adds stack trace to error details
	IncludeStackTrace bool

	// Environment affects detail level in responses
	// "production" = minimal details, "development" = full details
	Environment string

	// IDGenerator custom function to generate error IDs
	// If nil, uses default generator
	IDGenerator func() string
}

// Logger interface for custom logging implementations
type Logger interface {
	Error(errorID string, err error, context string, details map[string]interface{})
	Info(msg string)
}

// DefaultConfig returns sensible default configuration
func DefaultConfig() Config {
	return Config{
		OnError:           defaultOnError,
		AsyncCallback:     false,
		Logger:            NewDefaultLogger(os.Stdout),
		IncludeStackTrace: false,
		Environment:       "production",
		IDGenerator:       nil, // uses default generator
	}
}

// defaultOnError is the default callback that just logs
func defaultOnError(err *ErrorWithID) {
	// Logging is handled separately, so this is just a no-op placeholder
	// Users can override with custom behavior
}

// DefaultLogger implements Logger interface using standard log package
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger(out io.Writer) *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(out, "[ERROR-ID] ", log.LstdFlags),
	}
}

// Error logs error with ID and context
func (l *DefaultLogger) Error(errorID string, err error, context string, details map[string]interface{}) {
	l.logger.Printf("ID=%s | Context=%s | Error=%v | Details=%+v", errorID, context, err, details)
}

// Info logs informational message
func (l *DefaultLogger) Info(msg string) {
	l.logger.Printf("INFO: %s", msg)
}
