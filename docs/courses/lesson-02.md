# Lesson 2: Lazy Services and Practical Patterns

## Learning Objectives

- Use `Service[T]` for one-time service initialization
- Build lazy service registries for dependency injection
- Understand when to use lazy loading vs. eager initialization

## Key Concepts

- **Service vs Value**: `Service[T]` is semantically identical to `Value[T]` but designed for service objects (database connections, HTTP clients, caches). It adds `Initialized() bool` to check whether initialization succeeded.
- **Singleton Pattern**: Combined with package-level variables, `Service[T]` implements the Singleton pattern: `var db = lazy.NewService(connectDB)`. The connection is established on first `db.Get()` call.
- **Lazy Registry**: A struct holding multiple `Service[T]` fields creates a registry where each service initializes independently on first access. Services that are never accessed are never initialized, saving resources.
- **Trade-offs**: Lazy loading defers errors to runtime rather than startup. For critical services, consider calling `Get()` during initialization to fail fast. For optional services, lazy loading avoids unnecessary resource allocation.

## Code Walkthrough

### Source: `pkg/lazy/lazy.go`

The `Service[T]` struct mirrors `Value[T]` but uses `init` (renamed to avoid shadowing the built-in) as the initialization function:

```go
type Service[T any] struct {
    once    sync.Once
    service T
    initErr error
    init    func() (T, error)
}
```

`Initialized()` returns true only if the init function ran and returned no error:

```go
func (s *Service[T]) Initialized() bool {
    return s.initErr == nil
}
```

Note: `Initialized()` returns `true` before `Get()` is called (since `initErr` is the zero value `nil`). It is most meaningful after `Get()` has been called at least once.

## Practice Exercise

1. Create a `Service[*http.Client]` that configures a custom HTTP client with timeout. Call `Get()` and verify the client is configured correctly. Call `Get()` again and verify the same pointer is returned.
2. Build a `ServiceRegistry` struct with `db`, `cache`, and `search` services. Access only the `db` service and verify that `cache` and `search` are never initialized (use counters in init functions).
3. Create a `Service` with a failing init function. Call `Get()` and verify the error. Call `Initialized()` and verify it returns false.
