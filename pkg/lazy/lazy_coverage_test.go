package lazy

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Value[T] — table-driven error conditions
// ---------------------------------------------------------------------------

func TestValue_Get_ErrorConditions(t *testing.T) {
	tests := []struct {
		name        string
		loader      func() (string, error)
		wantVal     string
		wantErr     string
		wantIsError bool
	}{
		{
			name: "nil error returns value",
			loader: func() (string, error) {
				return "hello", nil
			},
			wantVal:     "hello",
			wantErr:     "",
			wantIsError: false,
		},
		{
			name: "simple error",
			loader: func() (string, error) {
				return "", errors.New("simple failure")
			},
			wantVal:     "",
			wantErr:     "simple failure",
			wantIsError: true,
		},
		{
			name: "wrapped error",
			loader: func() (string, error) {
				return "", fmt.Errorf("outer: %w", errors.New("inner"))
			},
			wantVal:     "",
			wantErr:     "outer: inner",
			wantIsError: true,
		},
		{
			name: "error with non-empty value",
			loader: func() (string, error) {
				return "partial", errors.New("partial failure")
			},
			wantVal:     "partial",
			wantErr:     "partial failure",
			wantIsError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := NewValue(tc.loader)
			val, err := v.Get()
			assert.Equal(t, tc.wantVal, val)
			if tc.wantIsError {
				require.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Value[T] — zero value from loader
// ---------------------------------------------------------------------------

func TestValue_Get_ZeroValue(t *testing.T) {
	t.Run("int zero value", func(t *testing.T) {
		v := NewValue(func() (int, error) {
			return 0, nil
		})
		val, err := v.Get()
		require.NoError(t, err)
		assert.Equal(t, 0, val)
	})

	t.Run("empty string", func(t *testing.T) {
		v := NewValue(func() (string, error) {
			return "", nil
		})
		val, err := v.Get()
		require.NoError(t, err)
		assert.Equal(t, "", val)
	})

	t.Run("nil pointer", func(t *testing.T) {
		v := NewValue(func() (*int, error) {
			return nil, nil
		})
		val, err := v.Get()
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("nil slice", func(t *testing.T) {
		v := NewValue(func() ([]string, error) {
			return nil, nil
		})
		val, err := v.Get()
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("nil map", func(t *testing.T) {
		v := NewValue(func() (map[string]int, error) {
			return nil, nil
		})
		val, err := v.Get()
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("false bool", func(t *testing.T) {
		v := NewValue(func() (bool, error) {
			return false, nil
		})
		val, err := v.Get()
		require.NoError(t, err)
		assert.False(t, val)
	})
}

// ---------------------------------------------------------------------------
// Value[T] — generic type variants
// ---------------------------------------------------------------------------

type testStruct struct {
	Name  string
	Count int
}

func TestValue_Get_StructType(t *testing.T) {
	expected := testStruct{Name: "test", Count: 42}
	v := NewValue(func() (testStruct, error) {
		return expected, nil
	})
	val, err := v.Get()
	require.NoError(t, err)
	assert.Equal(t, expected, val)
}

func TestValue_Get_PointerType(t *testing.T) {
	expected := &testStruct{Name: "ptr", Count: 7}
	v := NewValue(func() (*testStruct, error) {
		return expected, nil
	})
	val, err := v.Get()
	require.NoError(t, err)
	assert.Same(t, expected, val)
}

func TestValue_Get_SliceType(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	v := NewValue(func() ([]int, error) {
		return expected, nil
	})
	val, err := v.Get()
	require.NoError(t, err)
	assert.Equal(t, expected, val)
}

func TestValue_Get_InterfaceType(t *testing.T) {
	v := NewValue(func() (error, error) {
		return errors.New("inner value"), nil
	})
	val, err := v.Get()
	require.NoError(t, err)
	assert.Equal(t, "inner value", val.Error())
}

// ---------------------------------------------------------------------------
// Value[T] — error caching (Get after error still returns same error)
// ---------------------------------------------------------------------------

func TestValue_Get_ErrorIsCached(t *testing.T) {
	callCount := 0
	expectedErr := errors.New("persistent failure")
	v := NewValue(func() (string, error) {
		callCount++
		return "", expectedErr
	})

	// First call: gets error
	val1, err1 := v.Get()
	assert.Error(t, err1)
	assert.Equal(t, expectedErr, err1)
	assert.Empty(t, val1)
	assert.Equal(t, 1, callCount)

	// Second call: same cached error, loader NOT called again
	val2, err2 := v.Get()
	assert.Error(t, err2)
	assert.Equal(t, expectedErr, err2)
	assert.Empty(t, val2)
	assert.Equal(t, 1, callCount, "loader must not be called again after cached error")
}

// ---------------------------------------------------------------------------
// Value[T] — MustGet panic verification
// ---------------------------------------------------------------------------

func TestValue_MustGet_PanicContainsError(t *testing.T) {
	expectedErr := errors.New("must-get-failure")
	v := NewValue(func() (string, error) {
		return "", expectedErr
	})

	defer func() {
		r := recover()
		require.NotNil(t, r, "MustGet should have panicked")
		// The panic value should be the error itself
		recoveredErr, ok := r.(error)
		require.True(t, ok, "panic value should be an error")
		assert.Equal(t, expectedErr, recoveredErr)
		assert.Equal(t, "must-get-failure", recoveredErr.Error())
	}()

	v.MustGet()
	t.Fatal("should not reach here")
}

func TestValue_MustGet_SuccessDoesNotPanic(t *testing.T) {
	v := NewValue(func() (int, error) {
		return 99, nil
	})

	// Should not panic
	val := v.MustGet()
	assert.Equal(t, 99, val)
}

func TestValue_MustGet_SuccessAfterReset(t *testing.T) {
	callCount := 0
	v := NewValue(func() (int, error) {
		callCount++
		return callCount * 10, nil
	})

	val1 := v.MustGet()
	assert.Equal(t, 10, val1)

	v.Reset()

	val2 := v.MustGet()
	assert.Equal(t, 20, val2)
}

// ---------------------------------------------------------------------------
// Value[T] — Reset behavior
// ---------------------------------------------------------------------------

func TestValue_Reset_ThenGetReturnsNewValue(t *testing.T) {
	sequence := 0
	v := NewValue(func() (string, error) {
		sequence++
		return fmt.Sprintf("value-%d", sequence), nil
	})

	val1, err1 := v.Get()
	require.NoError(t, err1)
	assert.Equal(t, "value-1", val1)

	v.Reset()

	val2, err2 := v.Get()
	require.NoError(t, err2)
	assert.Equal(t, "value-2", val2, "after reset, loader must run again producing a new value")
}

func TestValue_Reset_ClearsError(t *testing.T) {
	callCount := 0
	v := NewValue(func() (string, error) {
		callCount++
		if callCount == 1 {
			return "", errors.New("first call fails")
		}
		return "recovered", nil
	})

	// First call fails
	_, err1 := v.Get()
	require.Error(t, err1)

	// Reset and retry
	v.Reset()

	val2, err2 := v.Get()
	require.NoError(t, err2)
	assert.Equal(t, "recovered", val2, "after reset, a previously failed loader can succeed")
}

func TestValue_Reset_MultipleResets(t *testing.T) {
	callCount := 0
	v := NewValue(func() (int, error) {
		callCount++
		return callCount, nil
	})

	for i := 1; i <= 5; i++ {
		val, err := v.Get()
		require.NoError(t, err)
		assert.Equal(t, i, val)
		v.Reset()
	}
	assert.Equal(t, 5, callCount)
}

func TestValue_Reset_WithoutPriorGet(t *testing.T) {
	callCount := 0
	v := NewValue(func() (string, error) {
		callCount++
		return "loaded", nil
	})

	// Reset before any Get — should be a no-op
	v.Reset()
	assert.Equal(t, 0, callCount, "reset without prior Get should not call loader")

	val, err := v.Get()
	require.NoError(t, err)
	assert.Equal(t, "loaded", val)
	assert.Equal(t, 1, callCount)
}

// ---------------------------------------------------------------------------
// Value[T] — concurrent Reset + Get (sequential phases, safe pattern)
//
// Note: Value[T].Reset() replaces sync.Once without external synchronization,
// so truly simultaneous Reset+Get is inherently racy. The safe usage pattern
// is to Reset when no concurrent Get calls are in flight. This test exercises
// that pattern: concurrent reads in one phase, then a sequential reset.
// ---------------------------------------------------------------------------

func TestValue_SequentialResetWithConcurrentReads(t *testing.T) {
	var callCount atomic.Int64
	v := NewValue(func() (int64, error) {
		return callCount.Add(1), nil
	})

	for cycle := 0; cycle < 5; cycle++ {
		// Phase 1: concurrent reads — all goroutines see the same value
		var wg sync.WaitGroup
		expected, err := v.Get()
		require.NoError(t, err)
		assert.Greater(t, expected, int64(0))

		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				val, err := v.Get()
				require.NoError(t, err)
				assert.Equal(t, expected, val)
			}()
		}
		wg.Wait()

		// Phase 2: sequential reset (no concurrent Get in flight)
		v.Reset()
	}

	// After 5 cycles, loader should have been called 5 times
	assert.Equal(t, int64(5), callCount.Load())

	// Final Get should produce cycle 6
	finalVal, err := v.Get()
	require.NoError(t, err)
	assert.Equal(t, int64(6), finalVal)
}

// ---------------------------------------------------------------------------
// Service[T] — Initialized before Get
// ---------------------------------------------------------------------------

func TestService_Initialized_BeforeGet(t *testing.T) {
	s := NewService(func() (string, error) {
		return "svc", nil
	})

	// Before Get(), initErr is zero value (nil), so Initialized returns true.
	// This documents the actual behavior of the implementation.
	assert.True(t, s.Initialized(), "before Get(), initErr is nil so Initialized() returns true")
}

func TestService_Initialized_FalseAfterErrorGet(t *testing.T) {
	s := NewService(func() (string, error) {
		return "", errors.New("init failed")
	})

	// Before Get
	assert.True(t, s.Initialized(), "before Get(), initErr is nil")

	// After failed Get
	_, err := s.Get()
	require.Error(t, err)
	assert.False(t, s.Initialized(), "after failed Get(), Initialized should be false")
}

func TestService_Initialized_TrueAfterSuccessfulGet(t *testing.T) {
	s := NewService(func() (string, error) {
		return "running", nil
	})

	_, err := s.Get()
	require.NoError(t, err)
	assert.True(t, s.Initialized(), "after successful Get(), Initialized should be true")
}

// ---------------------------------------------------------------------------
// Service[T] — table-driven error conditions
// ---------------------------------------------------------------------------

func TestService_Get_ErrorConditions(t *testing.T) {
	tests := []struct {
		name            string
		initFn          func() (string, error)
		wantVal         string
		wantErr         bool
		wantInitialized bool
	}{
		{
			name: "successful initialization",
			initFn: func() (string, error) {
				return "ok", nil
			},
			wantVal:         "ok",
			wantErr:         false,
			wantInitialized: true,
		},
		{
			name: "initialization error",
			initFn: func() (string, error) {
				return "", errors.New("failed")
			},
			wantVal:         "",
			wantErr:         true,
			wantInitialized: false,
		},
		{
			name: "error with partial value",
			initFn: func() (string, error) {
				return "partial", errors.New("degraded")
			},
			wantVal:         "partial",
			wantErr:         true,
			wantInitialized: false,
		},
		{
			name: "returns zero value on success",
			initFn: func() (string, error) {
				return "", nil
			},
			wantVal:         "",
			wantErr:         false,
			wantInitialized: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(tc.initFn)
			val, err := s.Get()
			assert.Equal(t, tc.wantVal, val)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.wantInitialized, s.Initialized())
		})
	}
}

// ---------------------------------------------------------------------------
// Service[T] — cached result on multiple Get calls
// ---------------------------------------------------------------------------

func TestService_Get_CachesResult(t *testing.T) {
	callCount := 0
	s := NewService(func() (string, error) {
		callCount++
		return "cached-service", nil
	})

	val1, err1 := s.Get()
	require.NoError(t, err1)
	assert.Equal(t, "cached-service", val1)
	assert.Equal(t, 1, callCount)

	val2, err2 := s.Get()
	require.NoError(t, err2)
	assert.Equal(t, "cached-service", val2)
	assert.Equal(t, 1, callCount, "init function should only be called once")
}

func TestService_Get_CachesError(t *testing.T) {
	callCount := 0
	expectedErr := errors.New("permanent failure")
	s := NewService(func() (string, error) {
		callCount++
		return "", expectedErr
	})

	_, err1 := s.Get()
	assert.Equal(t, expectedErr, err1)

	_, err2 := s.Get()
	assert.Equal(t, expectedErr, err2)
	assert.Equal(t, 1, callCount, "init function should only be called once even on error")
}

// ---------------------------------------------------------------------------
// Service[T] — generic type variants
// ---------------------------------------------------------------------------

func TestService_Get_StructType(t *testing.T) {
	expected := testStruct{Name: "svc", Count: 100}
	s := NewService(func() (testStruct, error) {
		return expected, nil
	})

	val, err := s.Get()
	require.NoError(t, err)
	assert.Equal(t, expected, val)
	assert.True(t, s.Initialized())
}

func TestService_Get_PointerType(t *testing.T) {
	expected := &testStruct{Name: "ptr-svc", Count: 200}
	s := NewService(func() (*testStruct, error) {
		return expected, nil
	})

	val, err := s.Get()
	require.NoError(t, err)
	assert.Same(t, expected, val)
}

func TestService_Get_NilInterfaceType(t *testing.T) {
	s := NewService(func() (fmt.Stringer, error) {
		return nil, nil
	})

	val, err := s.Get()
	require.NoError(t, err)
	assert.Nil(t, val)
}

// ---------------------------------------------------------------------------
// Service[T] — concurrent access
// ---------------------------------------------------------------------------

func TestService_ConcurrentGetAndInitialized(t *testing.T) {
	var initCount atomic.Int64
	s := NewService(func() (int, error) {
		initCount.Add(1)
		return 999, nil
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			val, err := s.Get()
			require.NoError(t, err)
			assert.Equal(t, 999, val)
		}()
		go func() {
			defer wg.Done()
			// Initialized may return true before or after init runs
			_ = s.Initialized()
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(1), initCount.Load(), "init should run exactly once")
	assert.True(t, s.Initialized())
}

// ---------------------------------------------------------------------------
// Value[T] — idempotent Get
// ---------------------------------------------------------------------------

func TestValue_Get_Idempotent(t *testing.T) {
	callCount := 0
	v := NewValue(func() (string, error) {
		callCount++
		return "stable", nil
	})

	for i := 0; i < 10; i++ {
		val, err := v.Get()
		require.NoError(t, err)
		assert.Equal(t, "stable", val)
	}
	assert.Equal(t, 1, callCount, "loader should be called exactly once across multiple Gets")
}

// ---------------------------------------------------------------------------
// Value[T] — MustGet after Reset with error on second call
// ---------------------------------------------------------------------------

func TestValue_MustGet_PanicAfterResetWithError(t *testing.T) {
	callCount := 0
	v := NewValue(func() (int, error) {
		callCount++
		if callCount == 2 {
			return 0, errors.New("second call fails")
		}
		return callCount, nil
	})

	// First call succeeds
	val := v.MustGet()
	assert.Equal(t, 1, val)

	// Reset
	v.Reset()

	// Second call should panic
	defer func() {
		r := recover()
		require.NotNil(t, r, "MustGet should panic on error after reset")
		recoveredErr, ok := r.(error)
		require.True(t, ok)
		assert.Equal(t, "second call fails", recoveredErr.Error())
	}()

	v.MustGet()
	t.Fatal("should not reach here")
}

// ---------------------------------------------------------------------------
// Value[T] — loader returning both value and error
// ---------------------------------------------------------------------------

func TestValue_Get_ValueAndErrorTogether(t *testing.T) {
	v := NewValue(func() (string, error) {
		return "partial-data", errors.New("partial error")
	})

	val, err := v.Get()
	assert.Equal(t, "partial-data", val, "value should be returned even with error")
	assert.Error(t, err)
	assert.Equal(t, "partial error", err.Error())
}

// ---------------------------------------------------------------------------
// NewValue / NewService — constructor verification
// ---------------------------------------------------------------------------

func TestNewValue_ReturnsNonNil(t *testing.T) {
	v := NewValue(func() (int, error) {
		return 0, nil
	})
	require.NotNil(t, v)
}

func TestNewService_ReturnsNonNil(t *testing.T) {
	s := NewService(func() (int, error) {
		return 0, nil
	})
	require.NotNil(t, s)
}

// ---------------------------------------------------------------------------
// Value[T] — large sequential reset cycles with concurrent reads
// ---------------------------------------------------------------------------

func TestValue_ResetCyclesWithConcurrentReads(t *testing.T) {
	var callCount atomic.Int64
	v := NewValue(func() (int64, error) {
		return callCount.Add(1), nil
	})

	var wg sync.WaitGroup
	for cycle := 0; cycle < 10; cycle++ {
		// Get the value for this cycle
		val, err := v.Get()
		require.NoError(t, err)
		assert.Greater(t, val, int64(0))

		// Concurrent reads during stable state
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(expected int64) {
				defer wg.Done()
				got, err := v.Get()
				assert.NoError(t, err)
				assert.Equal(t, expected, got)
			}(val)
		}
		wg.Wait()

		// Reset for next cycle (no concurrent Get in flight)
		v.Reset()
	}

	finalVal, err := v.Get()
	require.NoError(t, err)
	assert.Equal(t, int64(11), finalVal)
}
