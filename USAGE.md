# Usage Guide

## Panduan Penggunaan Error ID Package

### 1. Cara Termudah (Singleton/Default)

```go
package main

import (
    "errors"
    errorid "github.com/isaui/go-support-id-error"
)

func main() {
    // Langsung pakai tanpa setup
    err := doDatabase()
    if err != nil {
        // Error sudah otomatis di-log dengan ID
        // Output: [ERROR-ID] ID=ERR-20251023-A3F9B2 | Context=database query | Error=...
    }
}

func doDatabase() error {
    dbErr := errors.New("connection timeout")
    
    // Wrap error dengan context
    return errorid.Wrap(dbErr, "database query failed")
}
```

### 2. Dengan Metadata Tambahan

```go
func processPayment(userID int, amount float64) error {
    err := chargeCard()
    if err != nil {
        // Wrap dengan details untuk debugging
        return errorid.WrapWithDetails(err, "payment processing", map[string]interface{}{
            "user_id": userID,
            "amount":  amount,
            "gateway": "stripe",
        })
    }
    return nil
}
```

### 3. HTTP Middleware (Catch Panic)

```go
package main

import (
    "net/http"
    errorid "github.com/isaui/go-support-id-error"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api", handleAPI)
    
    // Middleware akan catch panic dan return error ID ke client
    handler := errorid.RecoveryMiddleware(mux)
    
    http.ListenAndServe(":8080", handler)
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
    // Kalau panic, middleware akan catch dan return JSON:
    // {"error_id": "ERR-20251023-A3F9B2", "message": "...", "timestamp": 1729652400}
    
    panic("something went wrong!")
}
```

### 4. Manual Error Response di Handler

```go
func handleUser(w http.ResponseWriter, r *http.Request) {
    user, err := getUser(userID)
    if err != nil {
        // Wrap error dengan ID
        wrapped := errorid.Wrap(err, "get user failed")
        
        // Return error response ke client
        errorid.WriteError(w, wrapped)
        // Client dapat JSON dengan error ID
        return
    }
    
    // Success response
    json.NewEncoder(w).Encode(user)
}
```

### 5. Custom Configuration (Advanced)

```go
package main

import (
    "os"
    errorid "github.com/isaui/go-support-id-error"
)

func main() {
    // Setup global config sekali di awal (sebelum Wrap() apapun!)
    errorid.Configure(errorid.Config{
        OnError: func(err *errorid.ErrorWithID) {
            // Kirim ke Sentry
            sendToSentry(err)
            
            // Kirim notif ke Slack
            sendToSlack(fmt.Sprintf("🚨 Error %s: %v", err.ID, err.Original))
        },
        AsyncCallback:     true,  // Non-blocking
        Logger:            errorid.NewDefaultLogger(os.Stdout),
        IncludeStackTrace: true,  // Untuk debugging
        Environment:       "production",
    })
    
    // Setelah configure, langsung pakai
    err := errorid.Wrap(someError, "context")
}

func sendToSentry(err *errorid.ErrorWithID) {
    // Implementasi kirim ke Sentry
}

func sendToSlack(msg string) {
    // Implementasi kirim ke Slack
}
```

### 6. Multiple Handlers (Advanced)

```go
package main

import errorid "github.com/isaui/go-support-id-error"

var (
    criticalHandler *errorid.Handler
    normalHandler   *errorid.Handler
)

func init() {
    // Handler untuk critical errors
    criticalHandler = errorid.New(errorid.Config{
        OnError: func(err *errorid.ErrorWithID) {
            sendUrgentAlert(err) // Page on-call engineer
        },
        Environment: "production",
    })
    
    // Handler untuk normal errors
    normalHandler = errorid.New(errorid.Config{
        OnError: func(err *errorid.ErrorWithID) {
            logToFile(err) // Just log
        },
        Environment: "production",
    })
}

func processPayment() error {
    err := chargeCard()
    if err != nil {
        // Critical error - use critical handler
        return criticalHandler.Wrap(err, "payment failed")
    }
    return nil
}

func processNotification() error {
    err := sendEmail()
    if err != nil {
        // Not critical - use normal handler
        return normalHandler.Wrap(err, "email failed")
    }
    return nil
}
```

## Scenario Use Cases

### Scenario 1: User Support

**Customer complain:**
> "Saya tidak bisa checkout, muncul error"

**Without Error ID:**
- Support harus tanya: jam berapa? browser apa? username apa?
- Developer cari di log dengan filter user + timestamp (susah!)

**With Error ID:**
- Customer: "Error ID: ERR-20251023-A3F9B2"
- Support: `grep ERR-20251023-A3F9B2 /var/log/app.log`
- Langsung dapat: error detail, user ID, request context, stack trace

### Scenario 2: Production Debugging

```go
// Error terjadi di production
func processOrder(orderID int) error {
    err := validateOrder(orderID)
    if err != nil {
        return errorid.WrapWithDetails(err, "order validation", map[string]interface{}{
            "order_id": orderID,
            "user_id":  getCurrentUser(),
            "ip":       getClientIP(),
        })
    }
    return nil
}

// Di log file:
// [ERROR-ID] ID=ERR-20251023-A3F9B2 | Context=order validation | Error=invalid product | 
// Details=map[order_id:12345 user_id:67890 ip:192.168.1.1]

// Developer search:
// grep ERR-20251023-A3F9B2 app.log
// -> Langsung dapat semua context tanpa tebak-tebakan
```

### Scenario 3: Integration dengan External Services

```go
errorid.Configure(errorid.Config{
    OnError: func(err *errorid.ErrorWithID) {
        // Auto report ke Sentry
        sentry.CaptureException(err.Original, map[string]interface{}{
            "error_id": err.ID,
            "context":  err.Context,
            "details":  err.Details,
        })
        
        // Auto alert di Slack channel #errors
        slack.Send("#errors", fmt.Sprintf(
            "🚨 Error detected\nID: %s\nContext: %s\nDetails: %+v",
            err.ID, err.Context, err.Details,
        ))
    },
    AsyncCallback: true, // Jangan block request
})
```

## Tips & Best Practices

### ✅ DO

```go
// 1. Context yang jelas
errorid.Wrap(err, "database query for user orders")

// 2. Tambahkan metadata yang berguna
errorid.WrapWithDetails(err, "API call", map[string]interface{}{
    "endpoint": "/users",
    "method":   "POST",
    "status":   500,
})

// 3. Configure di main() sekali
func main() {
    errorid.Configure(loadConfig())
    // ...
}

// 4. Async callback di production
Config{AsyncCallback: true}
```

### ❌ DON'T

```go
// 1. Context generic/tidak jelas
errorid.Wrap(err, "error")  // ❌ Tidak membantu

// 2. Hardcode sensitive data
errorid.WrapWithDetails(err, "auth", map[string]interface{}{
    "password": userPassword,  // ❌ NEVER!
})

// 3. Blocking callback dengan external call
Config{
    OnError: func(err) {
        http.Post("https://slow-api.com", ...)  // ❌ Block request!
    },
    AsyncCallback: false,  // ❌ Blocking
}
```

## Testing

Untuk test, buat mock handler:

```go
func TestMyFunction(t *testing.T) {
    var capturedError *errorid.ErrorWithID
    
    handler := errorid.New(errorid.Config{
        OnError: func(err *errorid.ErrorWithID) {
            capturedError = err
        },
    })
    
    // Test function
    err := handler.Wrap(errors.New("test"), "test context")
    
    // Assert
    if capturedError == nil {
        t.Fatal("expected error to be captured")
    }
    if capturedError.Context != "test context" {
        t.Errorf("expected context 'test context', got '%s'", capturedError.Context)
    }
}
```

## FAQ

**Q: Apakah error ID collision-resistant?**  
A: Ya, menggunakan random hex (3 bytes = 16 juta kombinasi per hari). Collision sangat kecil.

**Q: Apakah thread-safe?**  
A: Ya, semua API thread-safe dan bisa dipakai concurrent.

**Q: Apakah bisa customize format ID?**  
A: Ya, lewat `Config.IDGenerator`:
```go
Config{
    IDGenerator: func() string {
        return fmt.Sprintf("MY-CUSTOM-%d", time.Now().Unix())
    },
}
```

**Q: Performance impact?**  
A: Minimal, hanya ~1-2μs per wrap. Zero allocation kalau tanpa stack trace.

**Q: Apakah harus setup Configure()?**  
A: Tidak, ada default config yang sensible. Tapi lebih baik configure untuk production.
