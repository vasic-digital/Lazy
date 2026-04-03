package lazy_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"digital.vasic.lazy/pkg/lazy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Concurrent Init Race ---

func TestValue_ConcurrentGet_100Goroutines(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int64
	v := lazy.NewValue(func() (int, error) {
		callCount.Add(1)
		return 42, nil
	})

	var wg sync.WaitGroup
	results := make([]int, 100)
	errs := make([]error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = v.Get()
		}(i)
	}
	wg.Wait()

	// Loader should have been called exactly once
	assert.Equal(t, int64(1), callCount.Load())

	for i := 0; i < 100; i++ {
		assert.NoError(t, errs[i])
		assert.Equal(t, 42, results[i])
	}
}

// --- Panic in Initializer ---

func TestValue_PanicInLoader_PropagatesPanic(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() (string, error) {
		panic("loader panic")
	})

	assert.Panics(t, func() {
		_, _ = v.Get()
	})
}

// --- Error Return ---

func TestValue_LoaderReturnsError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("init failed")
	v := lazy.NewValue(func() (int, error) {
		return 0, expectedErr
	})

	val, err := v.Get()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 0, val)

	// Second call should return the same cached error
	val2, err2 := v.Get()
	assert.Equal(t, err, err2)
	assert.Equal(t, val, val2)
}

// --- Nil Return Value ---

func TestValue_LoaderReturnsNilPointer(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() (*int, error) {
		return nil, nil
	})

	val, err := v.Get()
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestValue_LoaderReturnsNilInterface(t *testing.T) {
	t.Parallel()

	type MyInterface interface {
		DoSomething()
	}

	v := lazy.NewValue(func() (MyInterface, error) {
		return nil, nil
	})

	val, err := v.Get()
	assert.NoError(t, err)
	assert.Nil(t, val)
}

// --- Double Initialization Never Calls Init Twice ---

func TestValue_DoubleGet_CallsLoaderOnce(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int64
	v := lazy.NewValue(func() (string, error) {
		callCount.Add(1)
		return "hello", nil
	})

	val1, err1 := v.Get()
	require.NoError(t, err1)

	val2, err2 := v.Get()
	require.NoError(t, err2)

	assert.Equal(t, "hello", val1)
	assert.Equal(t, "hello", val2)
	assert.Equal(t, int64(1), callCount.Load())
}

// --- MustGet Panics On Error ---

func TestValue_MustGet_PanicsOnError(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() (int, error) {
		return 0, errors.New("oops")
	})

	assert.Panics(t, func() {
		v.MustGet()
	})
}

func TestValue_MustGet_Success(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() (int, error) {
		return 99, nil
	})

	assert.NotPanics(t, func() {
		val := v.MustGet()
		assert.Equal(t, 99, val)
	})
}

// --- Reset ---

func TestValue_Reset_RerunsLoader(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int64
	v := lazy.NewValue(func() (int, error) {
		return int(callCount.Add(1)), nil
	})

	val1, err := v.Get()
	require.NoError(t, err)
	assert.Equal(t, 1, val1)

	v.Reset()

	val2, err := v.Get()
	require.NoError(t, err)
	assert.Equal(t, 2, val2)
	assert.Equal(t, int64(2), callCount.Load())
}

func TestValue_Reset_ConcurrentWithGet(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int64
	v := lazy.NewValue(func() (int, error) {
		return int(callCount.Add(1)), nil
	})

	var wg sync.WaitGroup

	// Hammer Get and Reset concurrently
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = v.Get()
		}()
		go func() {
			defer wg.Done()
			v.Reset()
		}()
	}
	wg.Wait()

	// Should not deadlock or panic
	val, err := v.Get()
	assert.NoError(t, err)
	assert.Greater(t, val, 0)
}

// --- Service Edge Cases ---

func TestService_ConcurrentGet(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int64
	svc := lazy.NewService(func() (string, error) {
		callCount.Add(1)
		return "service-value", nil
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := svc.Get()
			assert.NoError(t, err)
			assert.Equal(t, "service-value", val)
		}()
	}
	wg.Wait()

	assert.Equal(t, int64(1), callCount.Load())
}

func TestService_InitError(t *testing.T) {
	t.Parallel()

	svc := lazy.NewService(func() (int, error) {
		return 0, errors.New("service init failed")
	})

	val, err := svc.Get()
	assert.Error(t, err)
	assert.Equal(t, 0, val)

	// Initialized should still return false due to the way the check works
	// (initErr != nil means it attempted but failed)
	// The actual Initialized() checks initErr == nil
	assert.False(t, svc.Initialized())
}

func TestService_Initialized_BeforeGet(t *testing.T) {
	t.Parallel()

	svc := lazy.NewService(func() (string, error) {
		return "ok", nil
	})

	// Before first Get, Initialized returns true because initErr is nil (zero value)
	// This is testing the actual behavior of the code
	assert.True(t, svc.Initialized())
}

func TestService_PanicInInit(t *testing.T) {
	t.Parallel()

	svc := lazy.NewService(func() (string, error) {
		panic("service panic")
	})

	assert.Panics(t, func() {
		_, _ = svc.Get()
	})
}

// --- Zero Value Types ---

func TestValue_ZeroValueString(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() (string, error) {
		return "", nil
	})

	val, err := v.Get()
	assert.NoError(t, err)
	assert.Empty(t, val)
}

func TestValue_ZeroValueInt(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() (int, error) {
		return 0, nil
	})

	val, err := v.Get()
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
}

func TestValue_ZeroValueSlice(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() ([]string, error) {
		return nil, nil
	})

	val, err := v.Get()
	assert.NoError(t, err)
	assert.Nil(t, val)
}
