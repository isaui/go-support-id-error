# Quick Start - 5 Menit Langsung Pakai

## 1. Install

```bash
go get github.com/isaui/go-support-id-error
```

## 2. Import

```go
import errorid "github.com/isaui/go-support-id-error"
```

## 3. Pakai!

### Cara Paling Sederhana

```go
func doSomething() error {
    err := database.Query()
    if err != nil {
        return errorid.Wrap(err, "database query failed")
        // Auto-logged: [ERROR-ID] ID=ERR-20251023-A3F9B2 | Context=database query failed | Error=...
    }
    return nil
}
```

### HTTP API

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api", handleAPI)
    
    // Catch panic, return error ID
    handler := errorid.RecoveryMiddleware(mux)
    http.ListenAndServe(":8080", handler)
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
    err := processRequest()
    if err != nil {
        wrapped := errorid.Wrap(err, "process request")
        errorid.WriteError(w, wrapped)
        // Client gets: {"error_id": "ERR-20251023-A3F9B2", "message": "...", "timestamp": ...}
        return
    }
}
```

## 4. Setup Production (Optional)

```go
func main() {
    // Configure di awal sebelum pakai Wrap()
    errorid.Configure(errorid.Config{
        OnError: func(err *errorid.ErrorWithID) {
            sendToSentry(err)
        },
        AsyncCallback: true,
        Environment:   "production",
    })
    
    // Run app
    startServer()
}
```

## Done!

Sekarang setiap error punya ID unik untuk tracking!

**Customer:** "Error ID: ERR-20251023-A3F9B2"  
**Developer:** `grep ERR-20251023-A3F9B2 app.log` → Langsung ketemu!

---

## Cheat Sheet

```go
// Basic wrap
errorid.Wrap(err, "context")

// With metadata
errorid.WrapWithDetails(err, "context", map[string]interface{}{
    "user_id": 123,
})

// HTTP middleware (catch panic)
errorid.RecoveryMiddleware(handler)

// Return error response
errorid.WriteError(w, wrappedError)

// Configure global
errorid.Configure(config)

// Custom handler
handler := errorid.New(config)
handler.Wrap(err, "context")
```

## Examples

Lihat folder `examples/` untuk kode lengkap:
- `examples/simple/` - Basic usage
- `examples/advanced/` - Configuration
- `examples/http/` - HTTP server

## Docs

- **README.md** - Full documentation (English)
- **USAGE.md** - Panduan lengkap (Indonesian)
- **STRUCTURE.md** - Package architecture

## Support

- GitHub: https://github.com/isaui/go-support-id-error
- Issues: https://github.com/isaui/go-support-id-error/issues
