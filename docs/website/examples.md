# Examples

## Lazy Configuration from File

Load configuration from disk only when first needed:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "digital.vasic.lazy/pkg/lazy"
)

type AppConfig struct {
    Port     int    `json:"port"`
    LogLevel string `json:"log_level"`
}

var config = lazy.NewValue(func() (*AppConfig, error) {
    data, err := os.ReadFile("config.json")
    if err != nil {
        return nil, fmt.Errorf("reading config: %w", err)
    }
    var cfg AppConfig
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parsing config: %w", err)
    }
    return &cfg, nil
})

func main() {
    cfg, err := config.Get()
    if err != nil {
        fmt.Printf("Config error: %v\n", err)
        return
    }
    fmt.Printf("Server port: %d, log level: %s\n", cfg.Port, cfg.LogLevel)
}
```

## Concurrent Access Safety

Multiple goroutines can safely access a lazy value simultaneously:

```go
import (
    "sync"
    "digital.vasic.lazy/pkg/lazy"
)

callCount := 0
val := lazy.NewValue(func() (int, error) {
    callCount++
    return 42, nil
})

var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        result, _ := val.Get()
        fmt.Println(result) // always 42
    }()
}
wg.Wait()

fmt.Println(callCount) // 1 (loader ran exactly once)
```

## Lazy Service Registry

Build a service registry where each service initializes on first access:

```go
import "digital.vasic.lazy/pkg/lazy"

type ServiceRegistry struct {
    db    *lazy.Service[*Database]
    cache *lazy.Service[*Cache]
}

func NewRegistry() *ServiceRegistry {
    return &ServiceRegistry{
        db: lazy.NewService(func() (*Database, error) {
            return ConnectDB("localhost:5432")
        }),
        cache: lazy.NewService(func() (*Cache, error) {
            return ConnectCache("localhost:6379")
        }),
    }
}

func (r *ServiceRegistry) DB() (*Database, error) {
    return r.db.Get()
}

func (r *ServiceRegistry) Cache() (*Cache, error) {
    return r.cache.Get()
}
```
