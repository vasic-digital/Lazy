# Lazy — User Guide

`digital.vasic.lazy` is a small, zero-dependency Go module that provides generic lazy-loading
primitives. It defers expensive initialization until the value is first accessed, then caches the
result so the work is never repeated. All types are safe for concurrent use.

---

## Installation

```bash
go get digital.vasic.lazy
```

Minimum Go version: **1.25**

---

## Package Overview

| Package | Import Path | Description |
|---------|-------------|-------------|
| `lazy` | `digital.vasic.lazy/pkg/lazy` | Generic `Value[T]` and `Service[T]` types backed by `sync.Once` |

---

## Quick Start

### Lazy database connection with Value[T]

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/mattn/go-sqlite3"
    "digital.vasic.lazy/pkg/lazy"
)

func main() {
    // The connection is not opened until Get() is called for the first time.
    db := lazy.NewValue(func() (*sql.DB, error) {
        conn, err := sql.Open("sqlite3", "app.db")
        if err != nil {
            return nil, fmt.Errorf("open database: %w", err)
        }
        return conn, nil
    })

    // First call: opens the connection.
    conn, err := db.Get()
    if err != nil {
        log.Fatal(err)
    }

    // Subsequent calls: return the already-opened *sql.DB, no re-open.
    conn2, _ := db.Get()
    fmt.Println(conn == conn2) // true
}
```

### Lazy configuration loading with Value[T]

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "digital.vasic.lazy/pkg/lazy"
)

type AppConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

func main() {
    cfg := lazy.NewValue(func() (AppConfig, error) {
        f, err := os.Open("config.json")
        if err != nil {
            return AppConfig{}, fmt.Errorf("open config: %w", err)
        }
        defer f.Close()

        var c AppConfig
        if err := json.NewDecoder(f).Decode(&c); err != nil {
            return AppConfig{}, fmt.Errorf("decode config: %w", err)
        }
        return c, nil
    })

    // config.json is read only on the first call.
    c, err := cfg.Get()
    if err != nil {
        panic(err)
    }
    fmt.Printf("listening on %s:%d\n", c.Host, c.Port)
}
```

### Lazy service initialization with Service[T]

```go
package main

import (
    "fmt"
    "net/http"

    "digital.vasic.lazy/pkg/lazy"
)

func main() {
    client := lazy.NewService(func() (*http.Client, error) {
        // Expensive transport setup happens only once.
        return &http.Client{Timeout: 30 * 1e9}, nil
    })

    if svc, err := client.Get(); err == nil {
        resp, _ := svc.Get("https://example.com")
        fmt.Println(resp.Status)
    }

    fmt.Println("initialized:", client.Initialized()) // true
}
```

---

## Advanced Usage

### Reset for reconnection (Value[T])

Use `Reset()` when a cached value has become stale and must be reloaded — for example, after a
network error invalidates a database connection or a configuration file is updated on disk.

```go
db := lazy.NewValue(func() (*sql.DB, error) {
    return sql.Open("sqlite3", "app.db")
})

conn, err := db.Get()
if err != nil {
    // Connection attempt failed. Clear the cache so the next caller
    // triggers a fresh dial instead of receiving the same error.
    db.Reset()
}

// The next call re-runs the loader.
conn, err = db.Get()
```

`Reset()` is safe to call concurrently with `Get()`. After `Reset()` returns, the very next
`Get()` call (whichever goroutine wins) re-runs the loader and caches the new result.

### MustGet for application startup

Use `MustGet()` when failure is unrecoverable — typically during application startup where a
missing dependency should stop the process immediately.

```go
dsn := lazy.NewValue(func() (string, error) {
    dsn, ok := os.LookupEnv("DATABASE_URL")
    if !ok {
        return "", fmt.Errorf("DATABASE_URL is required")
    }
    return dsn, nil
})

// Panics immediately if DATABASE_URL is not set.
connectionString := dsn.MustGet()
```

Do not call `MustGet()` inside request handlers or background goroutines where panics would be
difficult to recover. Prefer `Get()` with explicit error handling in those contexts.

### Checking initialization state (Service[T])

`Service[T].Initialized()` reports whether the initialization function has already run and
completed without error. It is useful for health checks and readiness probes.

```go
emailSvc := lazy.NewService(func() (*EmailClient, error) {
    return dialSMTP("smtp.example.com:465")
})

// Health endpoint
http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    if !emailSvc.Initialized() {
        http.Error(w, "email service not ready", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
})
```

Note: `Initialized()` returns `true` before `Get()` is ever called because the internal error
field is `nil` at construction time. A `true` result is only meaningful after `Get()` has
returned without error.

---

## Best Practices

- **Prefer `Value[T]` for data** (configs, file contents, computed values) and **`Service[T]`
  for long-lived objects** (HTTP clients, database handles, caches). The distinction is
  semantic: `Value[T]` offers `Reset()`; `Service[T]` offers `Initialized()`.
- **Wrap errors with context.** Errors returned by the loader are cached and replayed on every
  subsequent `Get()` call until `Reset()` is called (for `Value[T]`). Include enough context in
  the error message to diagnose the failure without re-running the loader.
- **Do not share mutable state in the loader.** The loader is called at most once, but it must
  not capture shared mutable variables without its own synchronization, because the loader itself
  is not re-entrant.
- **Close resources before `Reset()`.** If the cached value holds open resources (file
  descriptors, connections), close them explicitly before calling `Reset()` to avoid leaks.
- **Use `MustGet()` only at startup.** At program startup, panicking on a missing dependency is
  acceptable. In request-handling code, always use `Get()` and return errors to the caller.

---

## FAQ

**Q: Is it safe to call `Get()` from multiple goroutines simultaneously?**

Yes. Both `Value[T]` and `Service[T]` use a `sync.Mutex` and `sync.Once` internally. The loader
runs in exactly one goroutine; all others block until it completes and then receive the cached
result.

**Q: What happens if the loader panics?**

`sync.Once` does not recover panics. If the loader panics, the panic propagates to the calling
goroutine and `sync.Once` marks the execution as complete — the loader will not be called again.
For `Value[T]`, call `Reset()` after recovering from a panic to allow re-loading.

**Q: Does `Service[T]` support reset?**

No. Services are intended to be initialized once for the lifetime of the process. If you need
reset semantics, use `Value[T]` instead.

**Q: What does `Initialized()` return before the first `Get()` call?**

It returns `true` because the internal `initErr` field is `nil` at zero value. Call `Get()`
first, then check `Initialized()` to get a meaningful result.

**Q: Can the loader return a nil pointer?**

Yes. A nil pointer is a valid value of any pointer type. `Get()` returns it without error, and
subsequent calls return the same nil pointer from cache.

**Q: Is there a way to pre-warm the cache?**

Call `Get()` (or `MustGet()`) before the first concurrent access. Because the result is cached,
later concurrent callers will never need to run the loader.
