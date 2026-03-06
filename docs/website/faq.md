# FAQ

## Is the lazy value thread-safe?

Yes. Both `Value[T]` and `Service[T]` use `sync.Once` internally, which guarantees that the loader/init function runs exactly once even when called concurrently from multiple goroutines. All subsequent calls return the cached result.

## What happens if the loader function returns an error?

The error is cached along with the zero value. All subsequent calls to `Get()` return the same error without re-running the loader. Use `Reset()` to clear the cached error and allow the loader to run again.

## When should I use Value vs Service?

Use `Value[T]` for lazily computing data values (configuration, computed results, parsed files). Use `Service[T]` for lazily initializing service objects (database connections, HTTP clients). `Service[T]` adds the `Initialized()` method for checking initialization status.

## Can I force the loader to run again?

Yes, but only for `Value[T]`. Call `val.Reset()` to clear the `sync.Once`, which allows the loader to run on the next `Get()` call. `Service[T]` does not expose a Reset method because service reinitialization typically requires cleanup of the previous instance.

## Does MustGet recover from panics?

No. `MustGet()` intentionally panics if the loader returns an error. Use it only for values that are truly required and whose absence means the program cannot function. For optional values, use `Get()` and handle the error explicitly.
