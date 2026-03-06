package lazy

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValue_Get(t *testing.T) {
	callCount := 0
	loader := func() (string, error) {
		callCount++
		return "loaded", nil
	}

	v := NewValue(loader)

	val1, err1 := v.Get()
	require.NoError(t, err1)
	assert.Equal(t, "loaded", val1)
	assert.Equal(t, 1, callCount)

	val2, err2 := v.Get()
	require.NoError(t, err2)
	assert.Equal(t, "loaded", val2)
	assert.Equal(t, 1, callCount, "Loader should only be called once")
}

func TestValue_Get_Error(t *testing.T) {
	expectedErr := errors.New("load failed")
	loader := func() (string, error) {
		return "", expectedErr
	}

	v := NewValue(loader)

	val, err := v.Get()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, val)
}

func TestValue_MustGet(t *testing.T) {
	v := NewValue(func() (string, error) {
		return "success", nil
	})

	val := v.MustGet()
	assert.Equal(t, "success", val)
}

func TestValue_MustGet_Panic(t *testing.T) {
	v := NewValue(func() (string, error) {
		return "", errors.New("failed")
	})

	defer func() {
		r := recover()
		require.NotNil(t, r, "Should have panicked")
	}()

	v.MustGet()
	t.Error("Should have panicked")
}

func TestValue_Reset(t *testing.T) {
	callCount := 0
	loader := func() (int, error) {
		callCount++
		return callCount, nil
	}

	v := NewValue(loader)

	val1, _ := v.Get()
	assert.Equal(t, 1, val1)

	v.Reset()

	val2, _ := v.Get()
	assert.Equal(t, 2, val2, "After reset, loader should be called again")
}

func TestValue_Concurrent(t *testing.T) {
	var callCount int64
	var mu sync.Mutex

	loader := func() (int, error) {
		mu.Lock()
		callCount++
		mu.Unlock()
		return 42, nil
	}

	v := NewValue(loader)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := v.Get()
			require.NoError(t, err)
			assert.Equal(t, 42, val)
		}()
	}

	wg.Wait()

	mu.Lock()
	count := callCount
	mu.Unlock()

	assert.Equal(t, int64(1), count, "Loader should only be called once even with concurrent access")
}

func TestService_Get(t *testing.T) {
	initCalled := false
	initFn := func() (string, error) {
		initCalled = true
		return "service-initialized", nil
	}

	s := NewService(initFn)

	_, _ = s.Get()

	assert.True(t, initCalled)
	assert.True(t, s.Initialized())
}

func TestService_Get_Error(t *testing.T) {
	initFn := func() (string, error) {
		return "", errors.New("init failed")
	}

	s := NewService(initFn)

	val, err := s.Get()
	assert.Error(t, err)
	assert.Empty(t, val)
	assert.False(t, s.Initialized())
}

func TestService_Concurrent(t *testing.T) {
	var initCount int64
	var mu sync.Mutex

	initFn := func() (int, error) {
		mu.Lock()
		initCount++
		mu.Unlock()
		return 100, nil
	}

	s := NewService(initFn)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := s.Get()
			require.NoError(t, err)
			assert.Equal(t, 100, val)
		}()
	}

	wg.Wait()

	mu.Lock()
	count := initCount
	mu.Unlock()

	assert.Equal(t, int64(1), count, "Init should only be called once")
}
