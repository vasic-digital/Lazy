# Lazy — API Reference

**Package path**: `digital.vasic.lazy/pkg/lazy`

**Import**:

```go
import "digital.vasic.lazy/pkg/lazy"
```

**Module**: `digital.vasic.lazy`  
**Minimum Go version**: 1.25  
**External dependencies**: none (standard library only)

---

## Generic Constraints

Both types accept a single type parameter `T` constrained to `any`. This means `T` can be any
Go type: a pointer, an interface, a struct, a scalar, a slice, or a map.

```go
// T can be any of these:
*sql.DB          // pointer
http.RoundTripper // interface
AppConfig        // struct value
int              // scalar
[]byte           // slice
map[string]int   // map
```

---

## Type: Value[T]

`Value[T]` lazily loads a value of type `T` on first access. The loader function is called at
most once, even under concurrent access. The result — value and error — is cached and returned
on all subsequent calls.

Unlike `Service[T]`, `Value[T]` supports `Reset()` for invalidating the cache and triggering a
fresh load on the next `Get()` call.

### Declaration

```go
type Value[T any] struct {
    // unexported fields: mu sync.Mutex, once sync.Once,
    // value T, err error, loader func() (T, error)
}
```

### Fields

All fields are unexported. Callers interact exclusively through the methods below.

---

### func NewValue

```go
func NewValue[T any](loader func() (T, error)) *Value[T]
```

Creates and returns a new `*Value[T]` backed by the provided loader function. The loader is not
called at construction time.

**Parameters**

| Name | Type | Description |
|------|------|-------------|
| `loader` | `func() (T, error)` | Function that produces the value. Called at most once per `Value` lifetime (or per reset cycle). Must not be `nil`. |

**Returns**

| Type | Description |
|------|-------------|
| `*Value[T]` | Pointer to the new lazy value. Never nil. |

**Example**

```go
cfg := lazy.NewValue(func() (AppConfig, error) {
    return loadConfigFromDisk("config.json")
})
```

---

### func (*Value[T]) Get

```go
func (v *Value[T]) Get() (T, error)
```

Returns the lazily-loaded value. On the first call, `Get()` invokes the loader, caches both the
returned value and error, and returns them. All subsequent calls return the cached result without
invoking the loader again.

`Get()` is safe for concurrent use. If multiple goroutines call `Get()` simultaneously before
the value is loaded, all goroutines block until the loader finishes; the loader runs in exactly
one goroutine.

**Returns**

| Type | Description |
|------|-------------|
| `T` | The loaded value. Zero value of `T` if the loader returned an error. |
| `error` | The error returned by the loader, or `nil` on success. Cached and replayed on every call until `Reset()` is invoked. |

**Example**

```go
conn, err := db.Get()
if err != nil {
    return fmt.Errorf("database unavailable: %w", err)
}
rows, err := conn.Query("SELECT id FROM users")
```

---

### func (*Value[T]) MustGet

```go
func (v *Value[T]) MustGet() T
```

Returns the lazily-loaded value, panicking if the loader returns a non-nil error. This is a
convenience wrapper around `Get()` for contexts where an error is unrecoverable.

**Returns**

| Type | Description |
|------|-------------|
| `T` | The loaded value. |

**Panics** if `Get()` returns a non-nil error. The panic value is the error itself.

**Example**

```go
// At application startup — missing DATABASE_URL should stop the process.
dsn := lazy.NewValue(func() (string, error) {
    v, ok := os.LookupEnv("DATABASE_URL")
    if !ok {
        return "", fmt.Errorf("DATABASE_URL not set")
    }
    return v, nil
})

connectionString := dsn.MustGet() // panics if env var is absent
```

**Note**: Do not use `MustGet()` inside request handlers or long-running goroutines. Prefer
`Get()` with explicit error handling in those contexts.

---

### func (*Value[T]) Reset

```go
func (v *Value[T]) Reset()
```

Clears the cached value and error so that the next `Get()` call will invoke the loader again.
`Reset()` is safe for concurrent use with `Get()`.

After `Reset()` returns, the internal `sync.Once` is replaced with a fresh zero value. The very
next `Get()` call (whichever goroutine wins) re-runs the loader and populates the new cache.

**Example**

```go
// A database connection became stale after a network partition.
// Close the old connection, reset, and let the next caller reconnect.
if conn, err := db.Get(); err == nil {
    conn.Close()
}
db.Reset()

// Next Get() opens a fresh connection.
freshConn, err := db.Get()
```

**Note**: If the cached value holds open resources, close them explicitly before calling
`Reset()` to avoid leaks.

---

## Type: Service[T]

`Service[T]` lazily initializes a service of type `T` exactly once. It is semantically identical
to `Value[T]` with two differences: it exposes `Initialized()` instead of `Reset()`, and it is
designed for long-lived objects that are expected to be initialized once for the lifetime of the
process.

### Declaration

```go
type Service[T any] struct {
    // unexported fields: mu sync.Mutex, once sync.Once,
    // service T, initErr error, init func() (T, error)
}
```

### Fields

All fields are unexported.

---

### func NewService

```go
func NewService[T any](initializer func() (T, error)) *Service[T]
```

Creates and returns a new `*Service[T]` backed by the provided initializer function. The
initializer is not called at construction time.

**Parameters**

| Name | Type | Description |
|------|------|-------------|
| `initializer` | `func() (T, error)` | Function that initializes the service. Called at most once. Must not be `nil`. |

**Returns**

| Type | Description |
|------|-------------|
| `*Service[T]` | Pointer to the new lazy service. Never nil. |

**Example**

```go
emailClient := lazy.NewService(func() (*EmailClient, error) {
    return dialSMTP("smtp.example.com:465")
})
```

---

### func (*Service[T]) Get

```go
func (s *Service[T]) Get() (T, error)
```

Returns the lazily-initialized service. On the first call, `Get()` invokes the initializer,
caches the result, and returns it. All subsequent calls return the cached result.

`Get()` is safe for concurrent use. If multiple goroutines call `Get()` simultaneously before
the service is initialized, all goroutines block until the initializer finishes.

**Returns**

| Type | Description |
|------|-------------|
| `T` | The initialized service. Zero value of `T` if initialization failed. |
| `error` | The error returned by the initializer, or `nil` on success. Cached permanently (no reset). |

**Example**

```go
client, err := httpSvc.Get()
if err != nil {
    return fmt.Errorf("http service unavailable: %w", err)
}
resp, err := client.Get("https://api.example.com/health")
```

---

### func (*Service[T]) Initialized

```go
func (s *Service[T]) Initialized() bool
```

Reports whether the service was successfully initialized (i.e., `Get()` has been called and the
initializer returned a `nil` error).

`Initialized()` is safe for concurrent use.

**Returns**

| Type | Description |
|------|-------------|
| `bool` | `true` if `initErr == nil`, `false` otherwise. |

**Behavior before first Get()**

`Initialized()` returns `true` before `Get()` has ever been called because the internal
`initErr` field is `nil` at zero value. This is an observable implementation detail. To get a
meaningful result, always call `Get()` first.

**Example**

```go
// Readiness probe: report ready only after successful initialization.
http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
    if !dbSvc.Initialized() {
        http.Error(w, "not ready", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
})
```

---

## Concurrency Guarantees

| Guarantee | Detail |
|-----------|--------|
| Loader called at most once | `sync.Once` ensures a single execution per `Value` or `Service` instance (per reset cycle for `Value[T]`). |
| All goroutines receive the same result | Every concurrent caller of `Get()` receives the identical cached value and error. |
| `Reset()` is safe under concurrency | After `Reset()` returns, the next `Get()` re-runs the loader. Goroutines that entered `Get()` before `Reset()` may still receive the old cached result. |
| No data races | All internal state is protected by `sync.Mutex`. Verified with `go test -race`. |

---

## Error Caching Behavior

Both `Value[T]` and `Service[T]` cache errors returned by the loader. A failed initialization is
not retried unless `Reset()` is called (only available on `Value[T]`).

```go
var attempts int
v := lazy.NewValue(func() (string, error) {
    attempts++
    return "", errors.New("always fails")
})

_, _ = v.Get() // attempts == 1, error cached
_, _ = v.Get() // attempts == 1, same error returned from cache
_, _ = v.Get() // attempts == 1, same error returned from cache

v.Reset()

_, _ = v.Get() // attempts == 2, loader called again
```
