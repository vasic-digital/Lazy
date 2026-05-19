// Lazy describe-Challenge runner.
//
// This is a CHALLENGE program (not production code) — it exercises
// the real Lazy Service[T].Describe API against a real Translator
// reading real YAML bundles off disk in two locales (en + sr-Latn)
// and asserts the three Describe states each resolve to the
// locale-correct localized string.
//
// Anti-bluff posture (CONST-035 / CONST-050):
//   - no mocks beyond the Translator implementation itself, which IS the
//     "real consumer integration" the Challenge is verifying;
//   - every assertion captures the actual returned string verbatim so
//     a regression cannot disguise itself as "test still passing";
//   - the paired-mutation leg (driven from challenges/lazy_describe_challenge.sh)
//     corrupts a YAML entry and re-runs this program — failure expected.
//
// Exit codes:
//   0 — all assertions held; bilingual round-trip green.
//   1 — bundle load failure, Describe regression, or locale-mismatch.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	_ "digital.vasic.lazy/pkg/i18n" // structural Translator contract (used via Service.SetTranslator)
	"digital.vasic.lazy/pkg/lazy"
)

// bundleTranslator is the Challenge's real Translator implementation.
// It reads a single-locale YAML bundle into memory once and resolves
// message IDs from the flat key->string map. Lookup-miss returns the
// message ID verbatim so a missing-key regression is loud, not silent.
type bundleTranslator struct {
	locale  string
	entries map[string]string
}

func loadBundle(path, locale string) (*bundleTranslator, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read bundle %s: %w", path, err)
	}
	entries := map[string]string{}
	if err := yaml.Unmarshal(raw, &entries); err != nil {
		return nil, fmt.Errorf("parse bundle %s: %w", path, err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("bundle %s parsed but contains no entries", path)
	}
	return &bundleTranslator{locale: locale, entries: entries}, nil
}

func (b *bundleTranslator) T(_ context.Context, id string, _ map[string]any) (string, error) {
	if v, ok := b.entries[id]; ok {
		return v, nil
	}
	return id, fmt.Errorf("translator(%s): missing key %s", b.locale, id)
}

func (b *bundleTranslator) TPlural(ctx context.Context, id string, _ int, data map[string]any) (string, error) {
	return b.T(ctx, id, data)
}

// expectedFor is the source-of-truth assertion table the Challenge
// compares Describe output against. Hardcoded here ON PURPOSE — this is
// challenge/test code, not production output (CONST-046 carve-out:
// fixtures + test assertions are not user-facing text).
func expectedFor(locale, msgID string) string {
	table := map[string]map[string]string{
		"en": {
			"lazy.service.uninitialized": "Service uninitialized",
			"lazy.service.ready":         "Service ready",
			"lazy.service.failed":        "Service failed",
		},
		"sr-Latn": {
			"lazy.service.uninitialized": "Servis nije inicijalizovan",
			"lazy.service.ready":         "Servis spreman",
			"lazy.service.failed":        "Servis neuspeo",
		},
	}
	return table[locale][msgID]
}

func mustEqual(label, got, want string) error {
	if got != want {
		return fmt.Errorf("%s: got %q want %q", label, got, want)
	}
	fmt.Printf("  OK  %s -> %q\n", label, got)
	return nil
}

// runLocale drives a single locale end-to-end through the three
// Describe states and asserts each lands on the expected localized
// string. Returns the first assertion failure, or nil on success.
func runLocale(ctx context.Context, fixturePath, locale string) error {
	tr, err := loadBundle(fixturePath, locale)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(fixturePath, locale+".yaml") {
		return fmt.Errorf("fixture path %q does not match locale %q", fixturePath, locale)
	}

	fmt.Printf("--- locale: %s (%s)\n", locale, fixturePath)

	// State 1: uninitialized
	svcUninit := lazy.NewService(func() (int, error) { return 42, nil })
	svcUninit.SetTranslator(tr)
	got, err := svcUninit.Describe(ctx)
	if err != nil {
		return fmt.Errorf("describe(uninitialized): %w", err)
	}
	if err := mustEqual(
		fmt.Sprintf("[%s] uninitialized", locale),
		got, expectedFor(locale, "lazy.service.uninitialized")); err != nil {
		return err
	}

	// State 2: ready
	svcReady := lazy.NewService(func() (int, error) { return 42, nil })
	svcReady.SetTranslator(tr)
	if _, err := svcReady.Get(); err != nil {
		return fmt.Errorf("svcReady.Get unexpectedly failed: %w", err)
	}
	got, err = svcReady.Describe(ctx)
	if err != nil {
		return fmt.Errorf("describe(ready): %w", err)
	}
	if err := mustEqual(
		fmt.Sprintf("[%s] ready", locale),
		got, expectedFor(locale, "lazy.service.ready")); err != nil {
		return err
	}

	// State 3: failed
	svcFailed := lazy.NewService(func() (int, error) { return 0, errors.New("boom") })
	svcFailed.SetTranslator(tr)
	if _, err := svcFailed.Get(); err == nil {
		return fmt.Errorf("svcFailed.Get should have returned an error")
	}
	got, err = svcFailed.Describe(ctx)
	if err != nil {
		return fmt.Errorf("describe(failed): %w", err)
	}
	if err := mustEqual(
		fmt.Sprintf("[%s] failed", locale),
		got, expectedFor(locale, "lazy.service.failed")); err != nil {
		return err
	}

	return nil
}

// crossLocaleSanity asserts that the EN and SR strings differ for every
// state — a sanity check that we are not accidentally serving the same
// language for both locales (CONST-046 regression guard).
func crossLocaleSanity(ctx context.Context, enFixture, srFixture string) error {
	en, err := loadBundle(enFixture, "en")
	if err != nil {
		return err
	}
	sr, err := loadBundle(srFixture, "sr-Latn")
	if err != nil {
		return err
	}
	for _, id := range []string{"lazy.service.uninitialized", "lazy.service.ready", "lazy.service.failed"} {
		eVal, _ := en.T(ctx, id, nil)
		sVal, _ := sr.T(ctx, id, nil)
		if eVal == sVal {
			return fmt.Errorf("cross-locale sanity: en and sr-Latn returned identical %q for %s", eVal, id)
		}
		if eVal == id || sVal == id {
			return fmt.Errorf("cross-locale sanity: bundle returned verbatim message id for %s (NoopTranslator regression?)", id)
		}
		fmt.Printf("  OK  cross-locale differs for %s: en=%q sr-Latn=%q\n", id, eVal, sVal)
	}
	return nil
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: cwd: %v\n", err)
		os.Exit(1)
	}
	fixturesDir := filepath.Join(root, "challenges", "fixtures")
	enFixture := filepath.Join(fixturesDir, "en.yaml")
	srFixture := filepath.Join(fixturesDir, "sr-Latn.yaml")

	ctx := context.Background()

	if err := runLocale(ctx, enFixture, "en"); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL en: %v\n", err)
		os.Exit(1)
	}
	if err := runLocale(ctx, srFixture, "sr-Latn"); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL sr-Latn: %v\n", err)
		os.Exit(1)
	}
	if err := crossLocaleSanity(ctx, enFixture, srFixture); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL cross-locale: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("PASS: lazy describe-Challenge — EN+SR round-trip + cross-locale sanity green")
}
