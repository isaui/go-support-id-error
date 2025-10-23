package main

import (
	"errors"
	"fmt"
	
	errorid "github.com/isaui/go-support-id-error"
)

func main() {
	fmt.Println("=== Simple Error ID Example ===\n")
	
	// Example 1: Basic usage with default singleton
	err := doSomething()
	if err != nil {
		fmt.Printf("Error occurred: %v\n\n", err)
	}
	
	// Example 2: Wrap with additional details
	err2 := doSomethingElse()
	if err2 != nil {
		fmt.Printf("Error occurred: %v\n\n", err2)
	}
	
	// Example 3: Unwrap to check original error
	err3 := doAnotherThing()
	if err3 != nil {
		fmt.Printf("Error occurred: %v\n", err3)
		
		// Check if it's a specific error type
		var myErr *MyCustomError
		if errors.As(err3, &myErr) {
			fmt.Printf("It's a custom error with code: %s\n\n", myErr.Code)
		}
	}
}

func doSomething() error {
	// Simulate a database error
	dbErr := errors.New("connection timeout")
	
	// Wrap it with error ID (using default singleton)
	return errorid.Wrap(dbErr, "database connection failed")
}

func doSomethingElse() error {
	// Simulate an API error
	apiErr := errors.New("HTTP 500: Internal Server Error")
	
	// Wrap with additional metadata
	return errorid.WrapWithDetails(apiErr, "external API call failed", map[string]interface{}{
		"endpoint": "https://api.example.com/users",
		"method":   "GET",
		"retry":    3,
	})
}

func doAnotherThing() error {
	// Custom error type
	customErr := &MyCustomError{
		Code:    "USR_001",
		Message: "user not found",
	}
	
	return errorid.Wrap(customErr, "user lookup failed")
}

// MyCustomError is a custom error type
type MyCustomError struct {
	Code    string
	Message string
}

func (e *MyCustomError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
