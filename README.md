# Error ID - Go Error Tracking Package

A Go package for tracking errors with unique IDs, making debugging and customer support significantly easier.

## Features

**Core Features:**
- **Unique Error IDs** - Every error gets a unique tracking ID (format: `ERR-20251023-A3F9B2`)
- **Error Wrapping** - Add context to errors while preserving the original error
- **Metadata Support** - Attach structured data to errors for better debugging
- **Stack Traces** - Optional stack trace capture for deep debugging

**Integration Features:**
- **Custom Callbacks** - Hook into error events to send to external services (Sentry, Slack, etc.)
- **Flexible Logging** - Pluggable logger interface for custom logging solutions
- **HTTP Middleware** - Optional panic recovery middleware (catches uncaught panics only)

**Design:**
- **Singleton & Instance APIs** - Use simple global API or create custom instances
- **Thread-Safe** - Safe for concurrent use across goroutines
- **Zero Dependencies** - Only uses Go standard library  

## Installation

```bash
go get github.com/isaui/go-support-id-error
```

## Quick Start

### Simple Usage (Singleton API)

```go
package main

import (
    "errors"
    errorid "github.com/isaui/go-support-id-error"
)

func main() {
    err := doSomething()
    if err != nil {
        // Error is automatically logged with unique ID
        // Output: [ERROR-ID] ID=ERR-20251023-A3F9B2 | Context=database query | Error=connection timeout
    }
}

func doSomething() error {
    dbErr := errors.New("connection timeout")
    return errorid.Wrap(dbErr, "database query")
}
```

### HTTP Middleware (Optional - For Panic Recovery)

The middleware is **optional** and only needed to catch **uncaught panics** in HTTP handlers. For normal error handling, use `Wrap()` directly.

**When to use middleware:**
- Your code might panic (index out of range, nil pointer, etc.)
- You want automatic panic recovery with error ID responses
- You want to prevent server crashes from panics

**When NOT needed:**
- You're already handling all errors properly with `Wrap()`
- Your code doesn't panic
- You have custom panic recovery logic

```go
package main

import (
    "encoding/json"
    "net/http"
    errorid "github.com/isaui/go-support-id-error"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", getUsers)
    mux.HandleFunc("/api/orders", getOrders)
    
    // Optional: Wrap with recovery middleware to catch panics
    // This prevents server crashes from uncaught panics
    handler := errorid.RecoveryMiddleware(mux)
    
    http.ListenAndServe(":8080", handler)
}

// Example 1: Normal error handling (no middleware needed)
func getUsers(w http.ResponseWriter, r *http.Request) {
    users, err := fetchUsers()
    if err != nil {
        // Manually wrap and return error
        wrapped := errorid.Wrap(err, "fetch users failed")
        errorid.WriteError(w, wrapped)
        return
    }
    json.NewEncoder(w).Encode(users)
}

// Example 2: Code that might panic (middleware catches it)
func getOrders(w http.ResponseWriter, r *http.Request) {
    orders := fetchOrders()
    // Bug: if orders is empty, this panics!
    firstOrder := orders[0]  // panic: index out of range
    
    // Without middleware: server crashes
    // With middleware: returns JSON with error ID
    json.NewEncoder(w).Encode(firstOrder)
}
```

### Advanced Configuration

```go
package main

import (
    "errors"
    "fmt"
    "os"
    errorid "github.com/isaui/go-support-id-error"
)

func main() {
    // Configure global handler (must be called before any Wrap() calls)
    errorid.Configure(errorid.Config{
        OnError: func(err *errorid.ErrorWithID) {
            // Custom callback - send to external services
            sendToSentry(err)
            sendToSlack(fmt.Sprintf("Error %s: %v", err.ID, err.Original))
        },
        AsyncCallback:     true,  // Non-blocking (runs in goroutine)
        Logger:            errorid.NewDefaultLogger(os.Stdout),
        IncludeStackTrace: true,  // Capture stack traces
        Environment:       "production",
    })
    
    // Start your application
    startServer()
}

func processPayment(userID int, amount float64) error {
    // Simulate payment error
    if amount < 0 {
        // Wrap with additional metadata
        return errorid.WrapWithDetails(
            errors.New("invalid amount"),
            "process payment",
            map[string]interface{}{
                "user_id": userID,
                "amount":  amount,
                "gateway": "stripe",
            },
        )
        // When this error is returned:
        // 1. Error ID generated: ERR-20251023-A3F9B2
        // 2. Logged with metadata
        // 3. OnError callback triggered (sent to Sentry & Slack)
        // 4. Stack trace captured (if enabled)
    }
    return nil
}
```

### Custom Handler Instance

```go
// Create multiple handlers with different configs
productionHandler := errorid.New(errorid.Config{
    Environment: "production",
    Logger:      myProductionLogger,
})

developmentHandler := errorid.New(errorid.Config{
    Environment:       "development",
    IncludeStackTrace: true,
})

// Use them separately
err1 := productionHandler.Wrap(err, "prod error")
err2 := developmentHandler.Wrap(err, "dev error")
```

## Configuration Options

```go
type Config struct {
    // Callback executed when error is wrapped
    OnError func(*ErrorWithID)
    
    // Run OnError callback in goroutine (non-blocking)
    AsyncCallback bool
    
    // Custom logger implementation
    Logger Logger
    
    // Include stack trace in error details
    IncludeStackTrace bool
    
    // Environment: "production" or "development"
    // Affects error detail level in HTTP responses
    Environment string
    
    // Custom ID generator function
    IDGenerator func() string
}
```

## API Reference

### Singleton API

```go
// Configure global handler (optional, has sensible defaults)
// Must be called before any Wrap() calls to take effect
// Can only be called once (subsequent calls ignored)
errorid.Configure(cfg Config)

// Wrap error with ID
errorid.Wrap(err error, context string) *ErrorWithID

// Wrap with additional metadata
errorid.WrapWithDetails(err error, context string, details map[string]interface{}) *ErrorWithID

// Get default handler instance
errorid.Default() *Handler

// HTTP middleware for panic recovery
errorid.RecoveryMiddleware(next http.Handler) http.Handler

// Write error response to HTTP client
errorid.WriteError(w http.ResponseWriter, err *ErrorWithID)
```

### Instance API

```go
// Create new handler instance
handler := errorid.New(cfg Config)

// Instance methods
handler.Wrap(err error, context string) *ErrorWithID
handler.WrapWithDetails(err error, context string, details map[string]interface{}) *ErrorWithID
handler.RecoveryMiddleware(next http.Handler) http.Handler
handler.WriteError(w http.ResponseWriter, err *ErrorWithID)
```

## Error ID Format

Default format: `ERR-YYYYMMDD-XXXXXX`

- `ERR` - Prefix for easy identification
- `YYYYMMDD` - Date (e.g., 20251023)
- `XXXXXX` - 6-character random hex (e.g., A3F9B2)

Example: `ERR-20251023-A3F9B2`

You can customize this by providing a custom `IDGenerator` function in the config.

## Use Cases

### 1. Customer Support

**Without Error ID:**
```
Customer: "I got an error when trying to checkout"
Support: "Can you describe the error? When did it happen? What browser?"
Customer: "I don't remember, it was yesterday"
Support: [searches through thousands of log entries...]
```

**With Error ID:**
```
Customer: "Error ID: ERR-20251023-A3F9B2"
Support: grep ERR-20251023-A3F9B2 app.log
# Instantly found: Database timeout on payment gateway, user ID 12345
```

### 2. Production Debugging

```go
// Error occurs in production
func processCheckout(userID int, cartID string) error {
    err := chargeCard()
    if err != nil {
        return errorid.WrapWithDetails(err, "checkout failed", map[string]interface{}{
            "user_id": userID,
            "cart_id": cartID,
        })
        // Error logged with ID: ERR-20251023-A3F9B2
    }
    return nil
}

// Developer debugging:
// $ grep ERR-20251023-A3F9B2 /var/log/app.log
//
// [ERROR-ID] ID=ERR-20251023-A3F9B2 | Context=checkout failed |
// Error=payment gateway timeout | Details=map[cart_id:abc123 user_id:12345]
//
// Instantly know: which user, which cart, what failed
```

### 3. Integration with External Services

```go
func main() {
    errorid.Configure(errorid.Config{
        OnError: func(err *errorid.ErrorWithID) {
            // Auto-send to Sentry
            sentry.CaptureException(err.Original, map[string]interface{}{
                "error_id": err.ID,
                "context":  err.Context,
                "details":  err.Details,
            })
            
            // Auto-alert on Slack
            slack.Send("#alerts", fmt.Sprintf(
                "Error %s: %v\nContext: %s\nDetails: %+v",
                err.ID, err.Original, err.Context, err.Details,
            ))
        },
        AsyncCallback: true, // Non-blocking (important for performance)
    })
    
    // Now every Wrap() automatically sends to Sentry + Slack
    startServer()
}
```

## Examples

See the `examples/` directory:

- `simple/` - Basic singleton usage
- `advanced/` - Custom handlers and configuration
- `http/` - HTTP middleware and API errors

Run examples:

```bash
# Simple example
go run examples/simple/main.go

# Advanced example
go run examples/advanced/main.go

# HTTP server example
go run examples/http/main.go
# Then visit: http://localhost:8080/error
```

## Best Practices

1. **Use context wisely** - Provide meaningful context that helps debugging
   ```go
   // Bad: Generic context
   errorid.Wrap(err, "error")
   
   // Good: Specific context
   errorid.Wrap(err, "database query for user orders")
   ```

2. **Add relevant details** - Include data that helps troubleshooting
   ```go
   errorid.WrapWithDetails(err, "payment processing", map[string]interface{}{
       "user_id":   userID,
       "amount":    amount,
       "gateway":   "stripe",
   })
   ```

3. **Configure at startup** - Set up global config BEFORE any Wrap() calls
   ```go
   func main() {
       errorid.Configure(loadConfig())  // Must be first!
       // ... rest of app (Wrap() calls come after)
   }
   ```

4. **Use async callbacks in production** - Don't block requests
   ```go
   Config{
       AsyncCallback: true,  // Non-blocking external calls
   }
   ```

5. **Different configs per environment** - Hide details in production
   ```go
   if os.Getenv("ENV") == "production" {
       cfg.Environment = "production"        // Minimal details to clients
       cfg.IncludeStackTrace = true          // But log full details
   }
   ```

## Thread Safety

- Singleton API is thread-safe
- Handler instances are thread-safe
- Safe to use across goroutines
- Configuration is immutable after initialization

## Performance

- **Minimal overhead** - Only adds ~1-2μs per error wrap
- **Zero allocations** - When no stack trace is captured
- **Async callbacks** - Optional non-blocking mode
- **No external dependencies** - No network calls unless you add them

## License

MIT License - feel free to use in commercial projects!

## Contributing

Contributions welcome! Please open issues or pull requests.

## Support

- **Issues**: [GitHub Issues](https://github.com/isaui/go-support-id-error/issues)
- **Documentation**: See this README and code examples
- **Questions**: Open a discussion on GitHub
