# Lazy

**Module**: `digital.vasic.lazy` · **Status**: i18n-migrated (round 93) · **Production LOC**: ~200 across 2 packages

Generic, reusable Go module for lazy-loading primitives. Defers expensive initialization until first access, guaranteeing the loader function runs exactly once even under concurrent access. Uses Go generics for type-safe operation. Optionally accepts an injected `Translator` for locale-aware diagnostic output.

---

## Status Banner

- **2026-05-19 (round 199)** — deep-doc + test-matrix enrichment per operator's broader directive; this README expanded, `docs/test-coverage.md` added, `challenges/lazy_describe_challenge.sh` added.
- **2026-05-16 (round 93)** — i18n migration landed: `pkg/i18n.Translator` contract + `Service[T].SetTranslator` + `Service[T].Describe(ctx)` for locale-aware status strings. EN + SR-Latn bilingual bundle seeded.
- **2026-05-15** — CONST-047..061 governance cascade.
- **2026-05-02** — anti-bluff amendments adopted (see `CLAUDE.md`).

---

## Purpose

Two concrete primitives wrap `sync.Once` so callers can defer expensive work (database handles, network clients, parsed config trees, large embeddings) until the first consumer asks for it — and never pay the cost again. Both are safe under concurrent access.

| Type | Use case |
|------|----------|
| `Value[T]` | A value computed by a loader. Read with `Get()` / `MustGet()`, invalidate with `Reset()`. |
| `Service[T]` | A service initialized by an `init` function. Read with `Get()`, probe with `Initialized()`, surface human-readable status with `Describe(ctx)`. |

---

## Integration Seams

Lazy is **project-not-aware** (CONST-051(B)). Consumers integrate via three seams:

1. **Loader / init function** — closure that does the real work (open DB, dial gRPC, parse YAML, …). Returns `(T, error)`.
2. **Optional `Translator` injection** — `Service[T].SetTranslator(tr)` lets the consumer route `Describe` output through its own i18n stack (HelixCode wires this to `helix_code/internal/i18nadapter`). The default `NoopTranslator` returns the message ID verbatim — safe for logs, **untranslated for end users** (production consumers MUST inject a real Translator per CONST-046).
3. **Context propagation** — `Describe(ctx)` accepts a `context.Context` so the consumer's `Translator` can read locale + request-scoped data from `ctx`.

No reverse coupling: Lazy never reaches into a consumer's tree, never imports a project-specific package, never assumes a HelixCode-specific layout.

---

## API Surface

### `pkg/lazy`

```go
// Lazy value
func NewValue[T any](loader func() (T, error)) *Value[T]
func (v *Value[T]) Get() (T, error)
func (v *Value[T]) MustGet() T          // panics on error — package-init only
func (v *Value[T]) Reset()              // re-arms the loader for the next Get

// Lazy service
func NewService[T any](init func() (T, error)) *Service[T]
func (s *Service[T]) Get() (T, error)
func (s *Service[T]) Initialized() bool
func (s *Service[T]) SetTranslator(tr i18n.Translator)
func (s *Service[T]) Describe(ctx context.Context) (string, error)
```

`Describe(ctx)` resolves one of three message IDs through the injected translator:

| Message ID | Triggered when |
|------------|----------------|
| `lazy.service.uninitialized` | `Get()` has not been called yet (probe-only — `Describe` does NOT trigger init). |
| `lazy.service.ready`         | `Get()` succeeded. |
| `lazy.service.failed`        | `Get()` returned an error. |

### `pkg/i18n`

```go
type Translator interface {
    T(ctx context.Context, messageID string, templateData map[string]any) (string, error)
    TPlural(ctx context.Context, messageID string, count int, templateData map[string]any) (string, error)
}

type NoopTranslator struct{}      // message-id passthrough; SAFETY DEFAULT, not production-ready
```

The contract intentionally mirrors the minimal subset of go-i18n's Localizer API that Lazy needs, **without** importing go-i18n — consumers can wire any real i18n stack (go-i18n, gotext, fluent, in-house) via structural typing.

---

## Usage Examples

### 1. Lazy DB handle

```go
import "digital.vasic.lazy/pkg/lazy"

// Opened only when first queried
db := lazy.NewValue(func() (*sql.DB, error) {
    return sql.Open("sqlite3", "app.db")
})

conn, err := db.Get()    // sql.Open runs HERE
conn2, _ := db.Get()     // returns the same *sql.DB — no re-open
fmt.Println(conn == conn2) // true
```

### 2. Lazy service with localized status

```go
import (
    "digital.vasic.lazy/pkg/i18n"
    "digital.vasic.lazy/pkg/lazy"
)

svc := lazy.NewService(func() (*OllamaClient, error) {
    return NewOllamaClient(os.Getenv("OLLAMA_HOST"))
})

// Inject the consumer's real i18n stack (HelixCode example)
svc.SetTranslator(helixI18n.LazyAdapter(localizer))

// Probe status without triggering init
status, _ := svc.Describe(ctx)   // → "Service uninitialized" (EN)
                                  // → "Servis nije inicijalizovan" (sr-Latn)

client, err := svc.Get()         // init runs HERE
status, _ = svc.Describe(ctx)    // → "Service ready"
```

### 3. Default NoopTranslator (logs only)

```go
svc := lazy.NewService(func() (*Cache, error) { return NewCache(), nil })
// No SetTranslator call → NoopTranslator
status, _ := svc.Describe(ctx)   // → "lazy.service.uninitialized" (verbatim ID)
```

CONST-046 reminder: this output is **untranslated technical text** and is safe ONLY for logs / structured telemetry. Surfacing it to end users is a CONST-046 violation.

---

## Bilingual Locale Support

The round-93 cascade established Lazy as the canonical i18n contract for `Service[T].Describe`. Reference bundle entries (used by `challenges/lazy_describe_challenge.sh`) live under `challenges/fixtures/`:

| Message ID                     | en          | sr-Latn (Serbian Latin) |
|--------------------------------|-------------|-------------------------|
| `lazy.service.uninitialized`   | Service uninitialized | Servis nije inicijalizovan |
| `lazy.service.ready`           | Service ready         | Servis spreman             |
| `lazy.service.failed`          | Service failed        | Servis neuspeo             |

These are **fixtures the Challenge consumes** — Lazy itself ships no bundled translations (it is a pure library; bundles are the consumer's responsibility). The fixtures prove that a minimal Translator implementation reading a 2-locale YAML satisfies the contract end-to-end.

---

## Testing

```bash
# Unit suite (race detector on)
go test ./... -count=1 -race

# Coverage report
go test ./... -cover -count=1

# Full Challenge (build + unit + bilingual round-trip + anti-bluff mutation)
bash challenges/lazy_describe_challenge.sh
```

The Challenge exits non-zero if any of:
- `go build ./...` fails
- unit suite fails or skips
- EN+SR bundles fail to load / parse
- `Describe()` returns the wrong localized string for any of the three states
- the anti-bluff mutation (corrupt one YAML entry) is NOT caught by the round-trip assertion

See [`docs/test-coverage.md`](docs/test-coverage.md) for the full test-type matrix (CONST-050(B) compliance ledger).

---

## Governance

- Anti-bluff prime directive: [`CLAUDE.md`](CLAUDE.md) preamble + Article XI §11.9 anchor.
- CONST-035 (anti-bluff), CONST-046 (no hardcoded content), CONST-050 (test-type coverage), CONST-051 (decoupling) — see [`CONSTITUTION.md`](CONSTITUTION.md).
- Canonical-root inheritance (CONST-059): governance text in this submodule is consumer-side; universal rules live in the [HelixConstitution submodule](https://github.com/HelixDevelopment/HelixConstitution).

---

## Module Identity

| Field            | Value                              |
|------------------|------------------------------------|
| Go module        | `digital.vasic.lazy`               |
| Go version       | 1.25                               |
| Direct deps      | `github.com/stretchr/testify` (test-only) |
| Upstream remotes | `vasic-digital/Lazy` (GitHub + GitLab) |
| License          | See repository                     |
