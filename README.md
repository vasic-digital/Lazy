# Lazy

Generic, reusable Go module for lazy-loading primitives. Defers expensive initialization until first access, guaranteeing the loader runs exactly once even under concurrent access. Uses Go generics for type-safe operation.

**Module**: `digital.vasic.lazy`

## Packages

- **pkg/lazy** -- Generic `Value[T]` and `Service[T]` types that wrap `sync.Once` to provide lazy initialization with error handling.

## Quick Start

```go
import "digital.vasic.lazy/pkg/lazy"

// Lazy database connection -- opened only when first queried
db := lazy.NewValue(func() (*sql.DB, error) {
    return sql.Open("sqlite3", "app.db")
})

// Nothing happens until Get() is called
conn, err := db.Get()

// Subsequent calls return the same *sql.DB without re-opening
conn2, _ := db.Get()
fmt.Println(conn == conn2) // true
```

## Testing

```bash
go test ./... -count=1 -race
```
