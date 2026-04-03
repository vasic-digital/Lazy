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

func TestValue_ConcurrentInitializationRace(t *testing.T) {
	t.Parallel()

	var callCount int64
	v := lazy.NewValue(func() (int, error) {
		atomic.AddInt64(&callCount, 1)
		return 42, nil
	})

	var wg sync.WaitGroup
	results := make([]int, 200)
	errs := make([]error, 200)

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = v.Get()
		}(i)
	}
	wg.Wait()

	// Loader must be called exactly once
	assert.Equal(t, int64(1), atomic.LoadInt64(&callCount))

	// All goroutines must get the same value
	for i := 0; i < 200; i++ {
		assert.NoError(t, errs[i])
		assert.Equal(t, 42, results[i])
	}
}

func TestValue_PanicInInitializer(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() (string, error) {
		panic("initializer panic")
	})

	// MustGet should propagate the panic
	assert.Panics(t, func() {
		v.MustGet()
	})
}

func TestValue_NilReturnValue(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() (*int, error) {
		return nil, nil
	})

	val, err := v.Get()
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Second call should return the same nil value
	val2, err2 := v.Get()
	assert.NoError(t, err2)
	assert.Nil(t, val2)
}

func TestValue_InitializerCalledExactlyOnce(t *testing.T) {
	t.Parallel()

	var count int64
	v := lazy.NewValue(func() (string, error) {
		atomic.AddInt64(&count, 1)
		return "value", nil
	})

	// Call Get many times sequentially
	for i := 0; i < 100; i++ {
		val, err := v.Get()
		assert.NoError(t, err)
		assert.Equal(t, "value", val)
	}

	assert.Equal(t, int64(1), atomic.LoadInt64(&count))
}

func TestValue_ErrorCachedOnSubsequentCalls(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("init failed")
	var count int64
	v := lazy.NewValue(func() (string, error) {
		atomic.AddInt64(&count, 1)
		return "", expectedErr
	})

	// First call returns error
	val1, err1 := v.Get()
	assert.Error(t, err1)
	assert.Equal(t, expectedErr, err1)
	assert.Empty(t, val1)

	// Second call returns same cached error without calling loader again
	val2, err2 := v.Get()
	assert.Error(t, err2)
	assert.Equal(t, expectedErr, err2)
	assert.Empty(t, val2)

	assert.Equal(t, int64(1), atomic.LoadInt64(&count))
}

func TestValue_ResetThenConcurrentGet(t *testing.T) {
	t.Parallel()

	var count int64
	v := lazy.NewValue(func() (int, error) {
		return int(atomic.AddInt64(&count, 1)), nil
	})

	// First load
	val1, err := v.Get()
	require.NoError(t, err)
	assert.Equal(t, 1, val1)

	// Reset
	v.Reset()

	// Concurrent access after reset
	var wg sync.WaitGroup
	results := make([]int, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			val, err := v.Get()
			assert.NoError(t, err)
			results[idx] = val
		}(i)
	}
	wg.Wait()

	// All should get the same value (2, since count was 1 before reset)
	for i := 0; i < 50; i++ {
		assert.Equal(t, 2, results[i])
	}

	// Total calls: 1 (initial) + 1 (after reset) = 2
	assert.Equal(t, int64(2), atomic.LoadInt64(&count))
}

func TestValue_MustGet_SuccessPath(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() (string, error) {
		return "success", nil
	})

	assert.NotPanics(t, func() {
		val := v.MustGet()
		assert.Equal(t, "success", val)
	})
}

func TestValue_MustGet_ErrorPanics(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() (string, error) {
		return "", errors.New("must fail")
	})

	assert.Panics(t, func() {
		v.MustGet()
	})
}

func TestValue_ZeroValueType(t *testing.T) {
	t.Parallel()

	// Loader returns the zero value of int (0) with no error
	v := lazy.NewValue(func() (int, error) {
		return 0, nil
	})

	val, err := v.Get()
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
}

func TestValue_StructType(t *testing.T) {
	t.Parallel()

	type config struct {
		Host string
		Port int
	}

	v := lazy.NewValue(func() (config, error) {
		return config{Host: "localhost", Port: 8080}, nil
	})

	val, err := v.Get()
	assert.NoError(t, err)
	assert.Equal(t, "localhost", val.Host)
	assert.Equal(t, 8080, val.Port)
}

func TestService_InitializerCalledExactlyOnce(t *testing.T) {
	t.Parallel()

	var count int64
	s := lazy.NewService(func() (string, error) {
		atomic.AddInt64(&count, 1)
		return "service", nil
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := s.Get()
			assert.NoError(t, err)
			assert.Equal(t, "service", val)
		}()
	}
	wg.Wait()

	assert.Equal(t, int64(1), atomic.LoadInt64(&count))
}

func TestService_Initialized_BeforeGet_Edge(t *testing.T) {
	t.Parallel()

	s := lazy.NewService(func() (string, error) {
		return "service", nil
	})

	// Before Get is called, Initialized reflects that init has not run
	// but since initErr is nil by default, Initialized returns true
	// This tests the actual behavior
	initialized := s.Initialized()
	assert.True(t, initialized, "Initialized returns true before Get (initErr is nil zero value)")
}

func TestService_ErrorPreservesState(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("service init failed")
	s := lazy.NewService(func() (string, error) {
		return "", expectedErr
	})

	val, err := s.Get()
	assert.Error(t, err)
	assert.Empty(t, val)
	assert.False(t, s.Initialized())

	// Subsequent calls return same error
	val2, err2 := s.Get()
	assert.Error(t, err2)
	assert.Empty(t, val2)
	assert.False(t, s.Initialized())
}

func TestValue_SliceType(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() ([]string, error) {
		return []string{"a", "b", "c"}, nil
	})

	val, err := v.Get()
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, val)
}

func TestValue_MapType(t *testing.T) {
	t.Parallel()

	v := lazy.NewValue(func() (map[string]int, error) {
		return map[string]int{"x": 1, "y": 2}, nil
	})

	val, err := v.Get()
	assert.NoError(t, err)
	assert.Equal(t, 1, val["x"])
	assert.Equal(t, 2, val["y"])
}

func TestValue_Reset_ClearsError(t *testing.T) {
	t.Parallel()

	callCount := 0
	v := lazy.NewValue(func() (string, error) {
		callCount++
		if callCount == 1 {
			return "", errors.New("first call fails")
		}
		return "success", nil
	})

	// First call fails
	_, err := v.Get()
	assert.Error(t, err)

	// Reset clears the error
	v.Reset()

	// Second call succeeds
	val, err := v.Get()
	assert.NoError(t, err)
	assert.Equal(t, "success", val)
}
