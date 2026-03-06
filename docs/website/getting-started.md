# Getting Started

## Installation

```bash
go get digital.vasic.lazy
```

## Lazy Value

Defer expensive computation until the value is first needed:

```go
package main

import (
    "database/sql"
    "fmt"

    "digital.vasic.lazy/pkg/lazy"
)

func main() {
    // The loader function runs once, on the first Get() call
    db := lazy.NewValue(func() (*sql.DB, error) {
        fmt.Println("Opening database connection...")
        return sql.Open("sqlite3", "app.db")
    })

    // Nothing happens yet
    fmt.Println("Application started")

    // Now the connection opens
    conn, err := db.Get()
    if err != nil {
        panic(err)
    }

    // Subsequent calls return the cached result
    conn2, _ := db.Get()
    fmt.Println(conn == conn2) // true
}
```

## MustGet for Infallible Values

Use `MustGet()` when the loader should never fail (panics on error):

```go
config := lazy.NewValue(func() (*AppConfig, error) {
    return loadConfigFromFile("config.yaml")
})

// Panics if loading fails -- appropriate for required startup resources
cfg := config.MustGet()
fmt.Println(cfg.Port)
```

## Reset for Cache Invalidation

Clear the cached value so the next `Get()` re-runs the loader:

```go
config := lazy.NewValue(func() (*AppConfig, error) {
    return loadConfigFromFile("config.yaml")
})

cfg1, _ := config.Get() // Loads from file
config.Reset()          // Clear cache
cfg2, _ := config.Get() // Reloads from file
```

## Lazy Service

Use `Service[T]` for one-time initialization of services that should never be reset:

```go
package main

import (
    "fmt"

    "digital.vasic.lazy/pkg/lazy"
)

func main() {
    svc := lazy.NewService(func() (*MyService, error) {
        // Expensive startup: open connections, load caches, etc.
        return &MyService{ready: true}, nil
    })

    fmt.Println(svc.Initialized()) // false

    instance, err := svc.Get()
    if err != nil {
        panic(err)
    }

    fmt.Println(svc.Initialized()) // true
    fmt.Println(instance.ready)     // true
}
```
