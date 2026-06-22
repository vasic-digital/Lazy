# Lazy — Test-Type Coverage Matrix

**Authority**: CONST-050(B) "100%-Test-Type-Coverage" mandate (cascaded from HelixConstitution submodule §11.4.27).
**Scope**: this document is the Lazy submodule's coverage ledger. It enumerates every test type CONST-050(B) recognises and records the current status against Lazy's surface (`pkg/lazy` + `pkg/i18n`).

A row may be `covered`, `planned`, or `n/a (out of scope for a library of this shape)`. `n/a` rows MUST justify themselves — silent omission is a CONST-048 violation per §11.4.25.

---

## Coverage Ledger

| Test type        | Status   | Artefact / location                                                                                              | Notes |
|------------------|----------|------------------------------------------------------------------------------------------------------------------|-------|
| Unit             | covered  | `pkg/lazy/lazy_test.go`, `lazy_coverage_test.go`, `lazy_edge_test.go`, `edge_test.go`, `describe_test.go`        | Mocks permitted per CONST-050(A); race-detector enforced; `fakeTranslator` in `describe_test.go` proves SetTranslator wiring. |
| Integration      | planned  | recommend: real-Translator wire-up against a consumer's `internal/i18nadapter`                                   | A real Translator reading a YAML bundle off-disk; assert all 3 message IDs resolve to expected EN+SR strings. Owner: the consuming project's integration layer, not Lazy itself (CONST-051(B) — Lazy stays project-not-aware). |
| E2E              | covered  | `challenges/lazy_describe_challenge.sh`                                                                          | Bash-orchestrated full round-trip from EN+SR fixture bundles through a real Translator into `Describe()` output, plus paired anti-bluff mutation. |
| Full automation  | planned  | recommend: re-run the Challenge under every supported Go minor (1.25, 1.26) on every host platform (linux/darwin/windows) | CONST-048 coverage matrix dimension is feature × platform × invariant; Lazy is pure Go so platform coverage = Go-supported set. |
| Security         | planned  | recommend: nil-handling, panic-safety on `MustGet` with broken loader, concurrent access fuzz, race-condition assertion | Existing `lazy_edge_test.go` covers some nil/panic paths — formalise as a tagged `_security_test.go` suite with explicit threat-model annotation. |
| DDoS             | n/a      | —                                                                                                                | Lazy is an in-process primitive — no network surface, no request fan-in. The consuming service exposes the DDoS surface, not Lazy. |
| Scaling          | planned  | recommend: benchmark `Get()` under N goroutines (N ∈ {1, 10, 100, 1000, 10000}) to verify sync.Once cost stays flat | Pure-CPU scaling test; not a network-tier scaling test. |
| Chaos            | planned  | recommend: chaos-style assertion that `Translator.T` returning an error propagates correctly through `Describe`; loader returning panic vs error | Failure-injection scope is narrow because the surface is narrow — but a complete CONST-050(B) ledger still names it. |
| Stress           | planned  | recommend: `go test -count=10000` of the concurrent-access tests to surface flakiness                            | Stress = sustained load above advertised tier; for an in-process primitive that means iteration count. |
| Performance      | planned  | recommend: `BenchmarkValue_Get`, `BenchmarkService_Get`, `BenchmarkService_Describe` with `b.ReportAllocs()` + historical p95 drift | Lazy's value prop is "init runs once" — benchmark MUST prove the cached-read path stays at ~few-ns allocation-free. |
| Benchmarking     | planned  | recommend: micro-benchmarks listed above + macro-benchmark embedded inside the consuming project's profile-run               | Macro tier lives outside Lazy (CONST-051(B)). |
| UI               | n/a      | —                                                                                                                | Lazy ships no UI. |
| UX               | covered  | bilingual locale verification inside `challenges/lazy_describe_challenge.sh`                                     | UX dimension Lazy actually owns: does `Describe` output flip language when the Translator's locale flips. Asserted EN→SR transition. |
| Challenges       | covered  | `challenges/lazy_describe_challenge.sh` (added round 199)                                                        | Incorporates the `vasic-digital/Challenges` pattern; captures stdout/stderr as wire evidence per §11.4.2; paired mutation per §1.1 / CONST-055 meta-test. |
| HelixQA          | planned  | recommend: register Lazy as a target in HelixQA's autonomous QA bank                                             | HelixQA submodule (`HelixDevelopment/HelixQA`) is incorporated at HelixCode root per CONST-050; Lazy enrolment is a HelixCode-meta-repo task, not a Lazy-internal task. |

---

## Anti-Bluff Posture

Every `covered` row above carries captured runtime evidence:

- **Unit**: `go test ./... -count=1 -race` exits 0; coverage `>= 70%` measured by `go test -cover`.
- **E2E (Challenge)**: `challenges/lazy_describe_challenge.sh` writes `challenges/.last-run/` artefacts containing stdout + stderr + assertion log + mutation-rejection proof.
- **UX**: the Challenge's bilingual leg captures the actual EN vs SR strings returned by `Describe` and diff-asserts both differ from each other AND from the verbatim message-id, ruling out NoopTranslator regression.

Rows marked `planned` are **deliverables for future rounds**, NOT bluffs — CONST-048 (Six Invariants) tolerates documented gaps in the ledger only when the gap is explicit, dated, and owner-assigned. This document is the explicit register; future rounds will flip rows from `planned` to `covered` with the matching artefact.

---

## Four-Layer Floor (CONST-048 invariant 6)

Per §1 of the constitution, every test artefact MUST sit on the four-layer floor:

| Layer       | Lazy artefact today                                                                            |
|-------------|------------------------------------------------------------------------------------------------|
| Pre-build   | `go vet ./...`, `go build ./...` — invoked by `challenges/lazy_describe_challenge.sh` step 1   |
| Post-build  | `go test ./... -count=1 -race` — invoked by Challenge step 2                                   |
| Runtime     | bilingual round-trip + Describe state-transition probe — Challenge step 4                      |
| Paired mut. | corrupt one YAML entry, assert Challenge FAILs — Challenge step 5                              |

A future round that adds a new test type to a `covered` row MUST extend the Challenge to keep the four-layer floor intact.

---

## Owner / Cadence

- **Owner**: Lazy submodule maintainer (vasic-digital). Consumers MAY contribute upstream but MUST NOT inject project-specific context (CONST-051(B)).
- **Cadence**: ledger reviewed at every governance-cascade round; planned → covered transitions land as their own commits with verbatim mandate quotes per CONST-049 §11.4.17.
