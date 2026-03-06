# Getting Started

## Installation

```bash
go get digital.vasic.lazy
```

## Lazy Value Loading

Create a lazy value that loads expensive data on first access:

```go
package main

import (
    "fmt"
    "digital.vasic.lazy/pkg/lazy"
)

func main() {
    config := lazy.NewValue(func() (map[string]string, error) {
        fmt.Println("Loading config...") // runs only once
        return map[string]string{
            "db_host": "localhost",
            "db_port": "5432",
        }, nil
    })

    // First call triggers the loader
    cfg, err := config.Get()
    if err != nil {
        panic(err)
    }
    fmt.Println(cfg["db_host"]) // "localhost"

    // Second call returns cached value immediately
    cfg2, _ := config.Get()
    fmt.Println(cfg2["db_port"]) // "5432" (no "Loading config..." printed)
}
```

## Lazy Service Initialization

Initialize a service exactly once:

```go
import "digital.vasic.lazy/pkg/lazy"

type Database struct {
    Host string
    Port int
}

dbService := lazy.NewService(func() (*Database, error) {
    // Expensive initialization runs once
    return &Database{Host: "localhost", Port: 5432}, nil
})

db, err := dbService.Get()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Connected to %s:%d\n", db.Host, db.Port)

// Check initialization status
fmt.Println(dbService.Initialized()) // true
```

## Using MustGet and Reset

```go
// MustGet panics on error -- use for required resources
val := lazy.NewValue(func() (string, error) {
    return "hello", nil
})
fmt.Println(val.MustGet()) // "hello"

// Reset clears the cache so loader runs again
val.Reset()
fmt.Println(val.MustGet()) // "hello" (loader runs again)
```
