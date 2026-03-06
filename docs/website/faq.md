# FAQ

## What is the difference between Value[T] and Service[T]?

`Value[T]` supports `Reset()` for scenarios where a cached value must be invalidated (e.g., configuration reload). `Service[T]` omits `Reset()` because services are typically initialized once for the lifetime of the process and should not be re-created. Both use `sync.Once` for thread-safe initialization.

## Is the lazy initialization thread-safe?

Yes. Both `Value[T]` and `Service[T]` use `sync.Once` internally, which guarantees the loader function executes at most once regardless of how many goroutines call `Get()` concurrently. All goroutines that call `Get()` while initialization is in progress will block until it completes and then receive the same cached result.

## What happens if the loader function returns an error?

If the loader returns an error, `Get()` returns the zero value of `T` along with the error. The error is cached -- subsequent `Get()` calls return the same error without re-running the loader. For `Value[T]`, you can call `Reset()` and then `Get()` to retry. `MustGet()` panics if the loader returns an error.

## Does this module have any external dependencies?

No. The module depends solely on the Go standard library (`sync` package). There are no external runtime dependencies.
