# Lazy Module

`digital.vasic.lazy` provides generic lazy-loading primitives for Go. It defers expensive initialization until first access, guaranteeing that the loader runs exactly once even under concurrent access. The module uses Go generics so it works with any type without casting.

## Key Features

- **Value[T]** -- Generic lazy value with `Get()`, `MustGet()`, and `Reset()` for resettable caching
- **Service[T]** -- Generic lazy service initialization with `Get()` and `Initialized()` status check
- **Thread-safe** -- Uses `sync.Once` internally for goroutine-safe initialization
- **Zero dependencies** -- No external dependencies beyond the Go standard library
- **Minimal API** -- Two types, five methods total

## Package Overview

| Package | Purpose |
|---------|---------|
| `pkg/lazy` | Generic `Value[T]` and `Service[T]` types for lazy initialization |

## Installation

```bash
go get digital.vasic.lazy
```

Requires Go 1.24 or later.

## Quick Example

```go
import "digital.vasic.lazy/pkg/lazy"

// Lazy database connection -- opened only when first queried
db := lazy.NewValue(func() (*sql.DB, error) {
    return sql.Open("sqlite3", "app.db")
})

conn, err := db.Get()  // Connection opens here
conn2, _ := db.Get()   // Returns the same connection
fmt.Println(conn == conn2) // true
```
