# Examples

## Lazy Configuration Loading

Defer configuration loading until first access, useful for libraries that may not need all config:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "digital.vasic.lazy/pkg/lazy"
)

type DatabaseConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Database string `json:"database"`
}

func main() {
    dbConfig := lazy.NewValue(func() (*DatabaseConfig, error) {
        data, err := os.ReadFile("database.json")
        if err != nil {
            return nil, err
        }
        var cfg DatabaseConfig
        err = json.Unmarshal(data, &cfg)
        return &cfg, err
    })

    // Config is only loaded if this code path is reached
    if needsDatabase() {
        cfg, err := dbConfig.Get()
        if err != nil {
            panic(err)
        }
        fmt.Printf("Connecting to %s:%d\n", cfg.Host, cfg.Port)
    }
}
```

## Lazy Singleton Services

Initialize expensive services (HTTP clients, connection pools) only when needed:

```go
package main

import (
    "fmt"
    "net/http"
    "time"

    "digital.vasic.lazy/pkg/lazy"
)

var httpClient = lazy.NewService(func() (*http.Client, error) {
    return &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:    100,
            IdleConnTimeout: 90 * time.Second,
        },
    }, nil
})

func fetchData(url string) ([]byte, error) {
    client, err := httpClient.Get()
    if err != nil {
        return nil, err
    }
    resp, err := client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Read and return body...
    return nil, nil
}

func main() {
    fmt.Println("Client initialized:", httpClient.Initialized()) // false
    fetchData("https://example.com")
    fmt.Println("Client initialized:", httpClient.Initialized()) // true
}
```

## Concurrent Access Safety

Multiple goroutines can safely call Get() -- the loader executes exactly once:

```go
package main

import (
    "fmt"
    "sync"
    "sync/atomic"

    "digital.vasic.lazy/pkg/lazy"
)

func main() {
    var callCount atomic.Int64

    value := lazy.NewValue(func() (string, error) {
        callCount.Add(1)
        return "initialized", nil
    })

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            v, _ := value.Get()
            _ = v
        }()
    }
    wg.Wait()

    fmt.Printf("Loader called %d time(s)\n", callCount.Load()) // 1
}
```
