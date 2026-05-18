// Package lazy provides generic lazy-loading primitives using sync.Once.
//
// Value[T] lazily loads a value on first access. Service[T] wraps
// a service initialization that runs exactly once. Both are safe for
// concurrent use.
//
// Design patterns: Proxy (deferred initialization), Singleton (sync.Once).
//
// Optional i18n: Service[T] accepts an injected
// digital.vasic.lazy/pkg/i18n.Translator via SetTranslator for
// locale-aware Describe output. Consumers that don't need
// localization can ignore SetTranslator; Describe falls back to a
// NoopTranslator that returns the message ID verbatim.
package lazy

import (
	"context"
	"sync"

	lazyi18n "digital.vasic.lazy/pkg/i18n"
)

// Value lazily loads a value of type T on first access via Get().
// The loader function is called at most once, even under concurrent access.
type Value[T any] struct {
	mu     sync.Mutex
	once   sync.Once
	value  T
	err    error
	loader func() (T, error)
}

// NewValue creates a new lazy value with the given loader function.
func NewValue[T any](loader func() (T, error)) *Value[T] {
	return &Value[T]{
		loader: loader,
	}
}

// Get returns the lazily-loaded value. The loader is called at most once.
func (v *Value[T]) Get() (T, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.once.Do(func() {
		v.value, v.err = v.loader()
	})
	return v.value, v.err
}

// MustGet returns the lazily-loaded value, panicking on error.
// This follows the Go "Must" convention (see template.Must, regexp.MustCompile).
// Use Get() instead in request handlers or goroutines where panics would
// crash the server — MustGet is intended for package-level initialisation only.
func (v *Value[T]) MustGet() T {
	val, err := v.Get()
	if err != nil {
		panic(err)
	}
	return val
}

// Reset clears the cached value so the loader will run again on next Get().
// Safe for concurrent use with Get().
func (v *Value[T]) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.once = sync.Once{}
	var zero T
	v.value = zero
	v.err = nil
}

// Service lazily initializes a service of type T exactly once.
//
// Service supports OPTIONAL i18n injection for human-readable status
// reports via SetTranslator + Describe. The default translator is a
// NoopTranslator that returns the message ID verbatim — production
// consumers MUST inject a real Translator to surface localized
// diagnostics to end users.
type Service[T any] struct {
	mu      sync.Mutex
	once    sync.Once
	service T
	initErr error
	init    func() (T, error)
	tr      lazyi18n.Translator
	// initCalled is set to true under mu whenever Get's init runs,
	// regardless of init success. Describe() uses this to distinguish
	// "uninitialized" from "ready / failed" without forcing init.
	initCalled bool
}

// NewService creates a new lazy service with the given init function.
// The translator defaults to NoopTranslator (message-id passthrough);
// call SetTranslator to inject a real localization stack.
func NewService[T any](init func() (T, error)) *Service[T] {
	return &Service[T]{
		init: init,
		tr:   lazyi18n.NoopTranslator{},
	}
}

// SetTranslator injects a Translator used by Describe. Passing nil is
// a no-op (the existing translator is retained); to reset, pass
// lazyi18n.NoopTranslator{}. Safe for concurrent use with Describe.
func (s *Service[T]) SetTranslator(tr lazyi18n.Translator) {
	if tr == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tr = tr
}

// Describe returns a localized human-readable status string for this
// service: one of the message IDs
//   - "lazy.service.uninitialized"  (init has not been called yet)
//   - "lazy.service.ready"          (init succeeded)
//   - "lazy.service.failed"         (init returned an error)
//
// Resolution goes through the injected Translator so the consumer's
// locale determines the output. With the default NoopTranslator the
// message ID is returned verbatim, which is safe for logs but
// untranslated for end users.
func (s *Service[T]) Describe(ctx context.Context) (string, error) {
	s.mu.Lock()
	tr := s.tr
	initialized := false
	failed := false
	called := false
	// Probe state without triggering once.Do: we replay the cached
	// outcome only if the user has already called Get(). We detect
	// this by attempting a non-blocking sync.Once check via a sentinel
	// — but sync.Once has no public probe, so we use the cached
	// initErr / service fields directly under the mutex.
	// Convention: if init has never run, initErr is the zero value
	// AND the typed-T service is the zero value. We surface this as
	// "uninitialized" without forcing init to run (Describe is a
	// diagnostic, not a trigger).
	//
	// To reliably detect "has Get been called", we inspect a small
	// flag set by Get under the same mutex.
	called = s.initCalled
	if called {
		if s.initErr != nil {
			failed = true
		} else {
			initialized = true
		}
	}
	s.mu.Unlock()

	switch {
	case failed:
		return tr.T(ctx, "lazy.service.failed", nil)
	case initialized:
		return tr.T(ctx, "lazy.service.ready", nil)
	default:
		return tr.T(ctx, "lazy.service.uninitialized", nil)
	}
}

// Get returns the lazily-initialized service.
func (s *Service[T]) Get() (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.once.Do(func() {
		s.service, s.initErr = s.init()
		s.initCalled = true
	})
	return s.service, s.initErr
}

// Initialized returns true if the service was initialized without error.
func (s *Service[T]) Initialized() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.initErr == nil
}
