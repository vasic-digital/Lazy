# Lesson 1: Lazy Values with sync.Once

## Learning Objectives

- Understand the Proxy pattern for deferred initialization
- Use `Value[T]` to lazily load values with error handling
- Verify thread safety under concurrent access

## Key Concepts

- **Proxy Pattern**: `Value[T]` acts as a proxy for the actual value. The expensive loader function is not called until `Get()` is invoked for the first time. Subsequent calls return the cached result immediately.
- **sync.Once Guarantee**: Go's `sync.Once.Do()` ensures the loader runs exactly once, even if hundreds of goroutines call `Get()` simultaneously. The first goroutine to arrive runs the loader; all others block until it completes.
- **Error Caching**: If the loader returns an error, both the error and the zero value are cached. This prevents repeated calls to a failing loader. Use `Reset()` to clear the cache and retry.
- **MustGet**: A convenience method that panics on error. Use it for values that are absolutely required at startup or in init functions.

## Code Walkthrough

### Source: `pkg/lazy/lazy.go`

The `Value[T]` struct:

```go
type Value[T any] struct {
    once   sync.Once
    value  T
    err    error
    loader func() (T, error)
}
```

`Get()` delegates to `sync.Once.Do`, which calls the loader exactly once:

```go
func (v *Value[T]) Get() (T, error) {
    v.once.Do(func() {
        v.value, v.err = v.loader()
    })
    return v.value, v.err
}
```

`Reset()` replaces the `sync.Once` with a fresh instance, allowing the loader to run again:

```go
func (v *Value[T]) Reset() {
    v.once = sync.Once{}
}
```

## Practice Exercise

1. Create a `Value[[]byte]` that reads a file. Call `Get()` twice and verify the file is only read once (use a counter in the loader).
2. Create a `Value[int]` with a loader that returns an error. Call `Get()` and verify the error is returned. Call `Reset()` and provide a working loader to verify the retry succeeds.
3. Launch 50 goroutines that all call `Get()` on the same `Value`. Verify the loader runs exactly once and all goroutines receive the same result.
