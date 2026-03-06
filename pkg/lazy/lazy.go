// Package lazy provides generic lazy-loading primitives using sync.Once.
//
// Value[T] lazily loads a value on first access. Service[T] wraps
// a service initialization that runs exactly once. Both are safe for
// concurrent use.
//
// Design patterns: Proxy (deferred initialization), Singleton (sync.Once).
package lazy

import (
	"sync"
)

// Value lazily loads a value of type T on first access via Get().
// The loader function is called at most once, even under concurrent access.
type Value[T any] struct {
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
	v.once.Do(func() {
		v.value, v.err = v.loader()
	})
	return v.value, v.err
}

// MustGet returns the lazily-loaded value, panicking on error.
func (v *Value[T]) MustGet() T {
	val, err := v.Get()
	if err != nil {
		panic(err)
	}
	return val
}

// Reset clears the cached value so the loader will run again on next Get().
func (v *Value[T]) Reset() {
	v.once = sync.Once{}
}

// Service lazily initializes a service of type T exactly once.
type Service[T any] struct {
	once    sync.Once
	service T
	initErr error
	init    func() (T, error)
}

// NewService creates a new lazy service with the given init function.
func NewService[T any](init func() (T, error)) *Service[T] {
	return &Service[T]{
		init: init,
	}
}

// Get returns the lazily-initialized service.
func (s *Service[T]) Get() (T, error) {
	s.once.Do(func() {
		s.service, s.initErr = s.init()
	})
	return s.service, s.initErr
}

// Initialized returns true if the service was initialized without error.
func (s *Service[T]) Initialized() bool {
	return s.initErr == nil
}
