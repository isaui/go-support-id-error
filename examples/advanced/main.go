package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	
	errorid "github.com/isaui/go-support-id-error"
)

func main() {
	fmt.Println("=== Advanced Error ID Example ===\n")
	
	// Example 1: Configure global singleton
	configureGlobalHandler()
	testGlobalHandler()
	
	// Example 2: Create custom handler instances
	testCustomHandler()
}

func configureGlobalHandler() {
	errorid.Configure(errorid.Config{
		OnError: func(err *errorid.ErrorWithID) {
			// Custom callback: could send to Sentry, save to DB, etc.
			fmt.Printf("📢 Custom callback triggered for error: %s\n", err.ID)
		},
		AsyncCallback:     false, // Synchronous for this example
		Logger:            errorid.NewDefaultLogger(os.Stdout),
		IncludeStackTrace: true,
		Environment:       "development",
		IDGenerator:       nil, // use default
	})
}

func testGlobalHandler() {
	fmt.Println("--- Testing Global Handler ---")
	
	err := errors.New("something went wrong")
	wrapped := errorid.Wrap(err, "processing user request")
	
	if wrapped != nil {
		fmt.Printf("Error ID: %s\n", wrapped.ID)
		fmt.Printf("Context: %s\n", wrapped.Context)
		fmt.Printf("Original: %v\n\n", wrapped.Original)
	}
}

func testCustomHandler() {
	fmt.Println("--- Testing Custom Handler Instance ---")
	
	// Create a custom handler with different config
	customHandler := errorid.New(errorid.Config{
		OnError: func(err *errorid.ErrorWithID) {
			// Different behavior for this handler
			log.Printf("🔴 CRITICAL ERROR [%s]: %v", err.ID, err.Original)
		},
		AsyncCallback:     true, // Async for performance
		Logger:            errorid.NewDefaultLogger(os.Stderr),
		IncludeStackTrace: false,
		Environment:       "production",
		IDGenerator: func() string {
			// Custom ID format
			return fmt.Sprintf("CUSTOM-%s", errorid.GenerateErrorID())
		},
	})
	
	// Use the custom handler
	err := errors.New("critical system failure")
	wrapped := customHandler.WrapWithDetails(err, "system initialization", map[string]interface{}{
		"component": "auth-service",
		"severity":  "critical",
	})
	
	if wrapped != nil {
		fmt.Printf("Error ID: %s\n", wrapped.ID)
		fmt.Printf("Details: %+v\n\n", wrapped.Details)
	}
}
