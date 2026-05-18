// Package i18n declares the minimal Translator contract that lazy
// consumers may inject for locale-aware status / diagnostic messages.
//
// This package depends on nothing outside Go's standard library: the
// Lazy module remains fully decoupled, project-not-aware, and
// reusable by any consumer. A consumer wires its real i18n stack to
// satisfy Translator (signature parity is sufficient — Go's
// structural typing handles the rest). Consumers that do not need
// localized diagnostics may pass NoopTranslator{}, which returns the
// message ID verbatim.
package i18n

import "context"

// Translator resolves message IDs to localized strings. The contract
// intentionally mirrors the minimal subset of go-i18n's Localizer API
// that Lazy needs, without importing go-i18n.
type Translator interface {
	// T resolves messageID with the given optional templateData. A nil
	// data map MUST be treated as "no template variables".
	T(ctx context.Context, messageID string, templateData map[string]any) (string, error)

	// TPlural resolves messageID with CLDR-aware plural selection for
	// the given count. A nil data map MUST be treated as "no template
	// variables beyond the implicit PluralCount".
	TPlural(ctx context.Context, messageID string, count int, templateData map[string]any) (string, error)
}

// NoopTranslator returns the messageID verbatim as a SAFETY DEFAULT
// for tests + consumers without a real i18n stack. Production
// consumers MUST inject a real Translator at construction time —
// shipping a NoopTranslator into production silently breaks
// non-English users (the messageID is returned as-is, which is
// untranslated technical text). The error return is always nil.
type NoopTranslator struct{}

// T implements Translator by returning messageID unchanged.
func (NoopTranslator) T(_ context.Context, messageID string, _ map[string]any) (string, error) {
	return messageID, nil
}

// TPlural implements Translator by returning messageID unchanged.
func (NoopTranslator) TPlural(_ context.Context, messageID string, _ int, _ map[string]any) (string, error) {
	return messageID, nil
}
