package lazy

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	lazyi18n "digital.vasic.lazy/pkg/i18n"
)

// fakeTranslator is a unit-test-only stand-in that satisfies
// lazyi18n.Translator by upper-casing the message ID. It demonstrates
// that SetTranslator wiring actually routes through the injected
// implementation rather than the default Noop.
type fakeTranslator struct{ prefix string }

func (f fakeTranslator) T(_ context.Context, id string, _ map[string]any) (string, error) {
	return f.prefix + strings.ToUpper(id), nil
}

func (f fakeTranslator) TPlural(_ context.Context, id string, _ int, _ map[string]any) (string, error) {
	return f.prefix + strings.ToUpper(id), nil
}

func TestService_Describe_NoopDefault_Uninitialized(t *testing.T) {
	s := NewService(func() (int, error) { return 42, nil })
	got, err := s.Describe(context.Background())
	require.NoError(t, err)
	// NoopTranslator returns the message ID verbatim.
	assert.Equal(t, "lazy.service.uninitialized", got)
}

func TestService_Describe_NoopDefault_Ready(t *testing.T) {
	s := NewService(func() (int, error) { return 42, nil })
	_, _ = s.Get()
	got, err := s.Describe(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "lazy.service.ready", got)
}

func TestService_Describe_NoopDefault_Failed(t *testing.T) {
	s := NewService(func() (int, error) { return 0, errors.New("init failed") })
	_, _ = s.Get()
	got, err := s.Describe(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "lazy.service.failed", got)
}

func TestService_Describe_InjectedTranslator_ChangesOutput(t *testing.T) {
	s := NewService(func() (int, error) { return 42, nil })

	// With default Noop translator: passthrough.
	got, err := s.Describe(context.Background())
	require.NoError(t, err)
	require.Equal(t, "lazy.service.uninitialized", got, "default Noop should passthrough")

	// Inject a real translator (fake here for unit-test isolation):
	s.SetTranslator(fakeTranslator{prefix: "FAKE>"})

	got2, err2 := s.Describe(context.Background())
	require.NoError(t, err2)
	assert.Equal(t, "FAKE>LAZY.SERVICE.UNINITIALIZED", got2,
		"injected Translator MUST change output; if equal to message ID, "+
			"SetTranslator wiring is broken (CONST-051(B) decoupling proof-of-life)")
}

func TestService_SetTranslator_NilIsNoOp(t *testing.T) {
	s := NewService(func() (int, error) { return 42, nil })
	s.SetTranslator(fakeTranslator{prefix: "X>"})

	got1, _ := s.Describe(context.Background())
	require.Equal(t, "X>LAZY.SERVICE.UNINITIALIZED", got1)

	// Nil should NOT clobber the previously-set translator.
	s.SetTranslator(nil)
	got2, _ := s.Describe(context.Background())
	assert.Equal(t, "X>LAZY.SERVICE.UNINITIALIZED", got2,
		"SetTranslator(nil) MUST be a no-op; otherwise consumers could "+
			"accidentally reset i18n by passing a nil from upstream config")
}

func TestNoopTranslator_PassthroughIDs(t *testing.T) {
	// Direct unit test of the Noop fallback — also documents the
	// contract that NoopTranslator never returns an error.
	n := lazyi18n.NoopTranslator{}
	for _, id := range []string{"foo", "bar.baz", ""} {
		got, err := n.T(context.Background(), id, nil)
		require.NoError(t, err)
		assert.Equal(t, id, got)

		got2, err2 := n.TPlural(context.Background(), id, 5, nil)
		require.NoError(t, err2)
		assert.Equal(t, id, got2)
	}
}
