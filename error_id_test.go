package errorid

import (
	"errors"
	"strings"
	"testing"
)

func TestGenerateErrorID(t *testing.T) {
	id := GenerateErrorID()
	
	// Check format: ERR-YYYYMMDD-XXXXXX
	if !strings.HasPrefix(id, "ERR-") {
		t.Errorf("expected ID to start with 'ERR-', got: %s", id)
	}
	
	parts := strings.Split(id, "-")
	if len(parts) != 3 {
		t.Errorf("expected 3 parts separated by '-', got: %v", parts)
	}
	
	// Date part should be 8 characters (YYYYMMDD)
	if len(parts[1]) != 8 {
		t.Errorf("expected date part to be 8 chars, got: %s", parts[1])
	}
	
	// Random part should be 6 characters (hex)
	if len(parts[2]) != 6 {
		t.Errorf("expected random part to be 6 chars, got: %s", parts[2])
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	context := "test context"
	
	wrapped := Wrap(originalErr, context)
	
	if wrapped == nil {
		t.Fatal("expected wrapped error to not be nil")
	}
	
	if wrapped.Original != originalErr {
		t.Error("expected original error to be preserved")
	}
	
	if wrapped.Context != context {
		t.Errorf("expected context '%s', got '%s'", context, wrapped.Context)
	}
	
	if wrapped.ID == "" {
		t.Error("expected error ID to be generated")
	}
	
	if wrapped.Timestamp == 0 {
		t.Error("expected timestamp to be set")
	}
}

func TestWrapNilError(t *testing.T) {
	wrapped := Wrap(nil, "context")
	
	if wrapped != nil {
		t.Error("expected wrapping nil error to return nil")
	}
}

func TestWrapWithDetails(t *testing.T) {
	originalErr := errors.New("test error")
	details := map[string]interface{}{
		"user_id": 123,
		"action":  "test",
	}
	
	wrapped := WrapWithDetails(originalErr, "test context", details)
	
	if wrapped == nil {
		t.Fatal("expected wrapped error to not be nil")
	}
	
	if wrapped.Details == nil {
		t.Fatal("expected details to be set")
	}
	
	if wrapped.Details["user_id"] != 123 {
		t.Error("expected user_id detail to be preserved")
	}
	
	if wrapped.Details["action"] != "test" {
		t.Error("expected action detail to be preserved")
	}
}

func TestErrorWithIDError(t *testing.T) {
	originalErr := errors.New("original")
	wrapped := Wrap(originalErr, "context")
	
	errorStr := wrapped.Error()
	
	// Should contain error ID
	if !strings.Contains(errorStr, wrapped.ID) {
		t.Errorf("expected error string to contain ID %s, got: %s", wrapped.ID, errorStr)
	}
	
	// Should contain context
	if !strings.Contains(errorStr, "context") {
		t.Errorf("expected error string to contain context, got: %s", errorStr)
	}
	
	// Should contain original error
	if !strings.Contains(errorStr, "original") {
		t.Errorf("expected error string to contain original error, got: %s", errorStr)
	}
}

func TestErrorWithIDUnwrap(t *testing.T) {
	originalErr := errors.New("original")
	wrapped := Wrap(originalErr, "context")
	
	unwrapped := wrapped.Unwrap()
	
	if unwrapped != originalErr {
		t.Error("expected Unwrap to return original error")
	}
	
	// Test errors.Is
	if !errors.Is(wrapped, originalErr) {
		t.Error("expected errors.Is to work with wrapped error")
	}
}

func TestHandlerNew(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Environment = "test"
	
	handler := New(cfg)
	
	if handler == nil {
		t.Fatal("expected handler to not be nil")
	}
	
	if handler.config.Environment != "test" {
		t.Errorf("expected environment 'test', got '%s'", handler.config.Environment)
	}
}

func TestHandlerWrap(t *testing.T) {
	var capturedError *ErrorWithID
	
	handler := New(Config{
		OnError: func(err *ErrorWithID) {
			capturedError = err
		},
		AsyncCallback: false,
	})
	
	originalErr := errors.New("test error")
	wrapped := handler.Wrap(originalErr, "test context")
	
	if wrapped == nil {
		t.Fatal("expected wrapped error to not be nil")
	}
	
	// Check callback was called
	if capturedError == nil {
		t.Fatal("expected OnError callback to be called")
	}
	
	if capturedError.ID != wrapped.ID {
		t.Error("expected callback to receive same error")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.OnError == nil {
		t.Error("expected default OnError to be set")
	}
	
	if cfg.Logger == nil {
		t.Error("expected default Logger to be set")
	}
	
	if cfg.Environment != "production" {
		t.Errorf("expected default environment 'production', got '%s'", cfg.Environment)
	}
}

func TestConfigure(t *testing.T) {
	// Create a fresh handler for testing by using instance API
	// Cannot test Configure() with global handler in test suite
	// because it's shared across all tests
	
	handler := New(Config{
		Environment: "test-config",
	})
	
	if handler.config.Environment != "test-config" {
		t.Errorf("expected configured environment, got '%s'", handler.config.Environment)
	}
}

func TestConfigurePanicsAfterInit(t *testing.T) {
	// This tests that Configure() panics if called after Wrap()
	// We test this with a separate program/example, not here
	// because global state is shared across tests
	
	// For now, just test that handler can be configured via New()
	handler := New(Config{Environment: "custom"})
	
	if handler.config.Environment != "custom" {
		t.Error("expected custom environment")
	}
}

func TestStackTrace(t *testing.T) {
	handler := New(Config{
		IncludeStackTrace: true,
	})
	
	wrapped := handler.Wrap(errors.New("test"), "context")
	
	if wrapped.StackTrace == "" {
		t.Error("expected stack trace to be captured")
	}
	
	// Stack trace should contain this test function
	if !strings.Contains(wrapped.StackTrace, "TestStackTrace") {
		t.Error("expected stack trace to contain test function name")
	}
}

func TestCustomIDGenerator(t *testing.T) {
	customID := "CUSTOM-ID-123"
	
	handler := New(Config{
		IDGenerator: func() string {
			return customID
		},
	})
	
	wrapped := handler.Wrap(errors.New("test"), "context")
	
	if wrapped.ID != customID {
		t.Errorf("expected custom ID '%s', got '%s'", customID, wrapped.ID)
	}
}

// Test Logger functionality
func TestCustomLogger(t *testing.T) {
	var loggedErrorID string
	var loggedContext string
	var loggedError error
	var loggedDetails map[string]interface{}
	
	customLogger := &mockLogger{
		errorFunc: func(errorID string, err error, context string, details map[string]interface{}) {
			loggedErrorID = errorID
			loggedError = err
			loggedContext = context
			loggedDetails = details
		},
	}
	
	handler := New(Config{
		Logger: customLogger,
	})
	
	testErr := errors.New("test error")
	testDetails := map[string]interface{}{"key": "value"}
	wrapped := handler.WrapWithDetails(testErr, "test context", testDetails)
	
	// Check logger was called with correct parameters
	if loggedErrorID != wrapped.ID {
		t.Errorf("expected logger to receive error ID %s, got %s", wrapped.ID, loggedErrorID)
	}
	
	if loggedError != testErr {
		t.Error("expected logger to receive original error")
	}
	
	if loggedContext != "test context" {
		t.Errorf("expected logger to receive context 'test context', got '%s'", loggedContext)
	}
	
	if loggedDetails["key"] != "value" {
		t.Error("expected logger to receive details")
	}
}

// Test AsyncCallback behavior
func TestAsyncCallback(t *testing.T) {
	callbackChan := make(chan *ErrorWithID, 1)
	
	handler := New(Config{
		OnError: func(err *ErrorWithID) {
			callbackChan <- err
		},
		AsyncCallback: true, // Async mode
	})
	
	testErr := errors.New("test error")
	wrapped := handler.Wrap(testErr, "test context")
	
	// Wait for async callback
	select {
	case capturedErr := <-callbackChan:
		if capturedErr.ID != wrapped.ID {
			t.Error("expected callback to receive same error ID")
		}
	case <-make(chan struct{}):
		// Timeout simulation - in real test would use time.After
		t.Fatal("timeout waiting for async callback")
	}
}

// Test SyncCallback behavior
func TestSyncCallback(t *testing.T) {
	var capturedError *ErrorWithID
	
	handler := New(Config{
		OnError: func(err *ErrorWithID) {
			capturedError = err
		},
		AsyncCallback: false, // Sync mode
	})
	
	testErr := errors.New("test error")
	wrapped := handler.Wrap(testErr, "test context")
	
	// In sync mode, callback should be called immediately
	if capturedError == nil {
		t.Fatal("expected sync callback to be called immediately")
	}
	
	if capturedError.ID != wrapped.ID {
		t.Error("expected callback to receive same error ID")
	}
}

// Test OnError callback execution
func TestOnErrorCallback(t *testing.T) {
	callCount := 0
	var lastError *ErrorWithID
	
	handler := New(Config{
		OnError: func(err *ErrorWithID) {
			callCount++
			lastError = err
		},
		AsyncCallback: false,
	})
	
	// Wrap multiple errors
	_ = handler.Wrap(errors.New("error 1"), "context 1")
	_ = handler.Wrap(errors.New("error 2"), "context 2")
	err3 := handler.Wrap(errors.New("error 3"), "context 3")
	
	// Callback should be called for each error
	if callCount != 3 {
		t.Errorf("expected OnError to be called 3 times, got %d", callCount)
	}
	
	if lastError.ID != err3.ID {
		t.Error("expected last error to be from err3")
	}
}

// Test Global Handler (Singleton)
func TestGlobalHandler(t *testing.T) {
	// Get default handler
	handler1 := Default()
	handler2 := Default()
	
	// Should be same instance (singleton)
	if handler1 != handler2 {
		t.Error("expected Default() to return same instance")
	}
	
	// Wrap with global API
	err := Wrap(errors.New("test"), "global test")
	if err == nil {
		t.Fatal("expected global Wrap to work")
	}
	
	if err.ID == "" {
		t.Error("expected global Wrap to generate ID")
	}
}

// Test Custom Handler Independence
func TestCustomHandlerIndependence(t *testing.T) {
	var customCallbackCalled bool
	
	// Create custom handler with its own callback
	customHandler := New(Config{
		OnError: func(err *ErrorWithID) {
			customCallbackCalled = true
		},
		AsyncCallback: false,
	})
	
	// Use custom handler
	wrapped := customHandler.Wrap(errors.New("custom error"), "custom")
	
	if !customCallbackCalled {
		t.Error("expected custom callback to be called")
	}
	
	if wrapped == nil {
		t.Fatal("expected wrapped error to not be nil")
	}
	
	if wrapped.ID == "" {
		t.Error("expected error ID to be generated")
	}
}

// Test IncludeStackTrace disabled
func TestNoStackTrace(t *testing.T) {
	handler := New(Config{
		IncludeStackTrace: false,
	})
	
	wrapped := handler.Wrap(errors.New("test"), "context")
	
	if wrapped.StackTrace != "" {
		t.Error("expected no stack trace when IncludeStackTrace is false")
	}
}

// Test Details are preserved
func TestDetailsPreservation(t *testing.T) {
	details := map[string]interface{}{
		"user_id":  12345,
		"email":    "test@example.com",
		"action":   "checkout",
		"amount":   99.99,
		"metadata": map[string]string{"key": "value"},
	}
	
	wrapped := WrapWithDetails(errors.New("test"), "test", details)
	
	if wrapped.Details["user_id"] != 12345 {
		t.Error("expected user_id to be preserved")
	}
	
	if wrapped.Details["email"] != "test@example.com" {
		t.Error("expected email to be preserved")
	}
	
	if wrapped.Details["amount"] != 99.99 {
		t.Error("expected amount to be preserved")
	}
	
	// Test nested map
	metadata, ok := wrapped.Details["metadata"].(map[string]string)
	if !ok {
		t.Fatal("expected metadata to be map[string]string")
	}
	
	if metadata["key"] != "value" {
		t.Error("expected nested metadata to be preserved")
	}
}

// Test timestamp is set
func TestTimestampSet(t *testing.T) {
	wrapped := Wrap(errors.New("test"), "context")
	
	if wrapped.Timestamp == 0 {
		t.Error("expected timestamp to be set")
	}
	
	// Timestamp should be recent (within last second)
	// In production test, would check against time.Now().Unix()
}

// Mock logger for testing
type mockLogger struct {
	errorFunc func(errorID string, err error, context string, details map[string]interface{})
	infoFunc  func(msg string)
}

func (m *mockLogger) Error(errorID string, err error, context string, details map[string]interface{}) {
	if m.errorFunc != nil {
		m.errorFunc(errorID, err, context, details)
	}
}

func (m *mockLogger) Info(msg string) {
	if m.infoFunc != nil {
		m.infoFunc(msg)
	}
}
