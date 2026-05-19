#!/usr/bin/env bash
#
# challenges/lazy_describe_challenge.sh
#
# Round-199 deliverable — Lazy submodule deep-doc + test-matrix enrichment.
#
# Drives the full CONST-050(B) "Challenges" leg for the Lazy submodule:
#
#   Step 1: pre-build  -- go vet + go build
#   Step 2: post-build -- go test ./... -count=1 -race
#   Step 3: bundle load -- assert both fixture YAMLs exist + non-empty
#   Step 4: runtime end-to-end -- run challenges/runner against EN+SR
#                                  fixtures; assert every Describe state
#                                  flips to the locale-correct string
#   Step 5: paired anti-bluff mutation -- corrupt one SR entry, re-run,
#                                          expect non-zero exit; restore
#
# Anti-bluff invariants (CONST-035 / Article XI §11.9):
#   - every PASS is preceded by a real command + captured output
#   - the mutation leg PROVES the assertion would fail if Lazy regressed
#   - the script exits non-zero on the FIRST failure (no quiet skips)
#
# Exit 0 only if every step above succeeded.

set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
EVIDENCE_DIR="${SCRIPT_DIR}/.last-run"
mkdir -p "${EVIDENCE_DIR}"

cd "${REPO_ROOT}"

log() { printf '\n=== %s ===\n' "$*"; }
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }

# ---------------------------------------------------------------------------
# Step 1 -- pre-build floor
# ---------------------------------------------------------------------------
log "Step 1: go vet + go build (pre-build floor)"
go vet ./... 2>&1 | tee "${EVIDENCE_DIR}/01-vet.log" || fail "go vet"
go build ./... 2>&1 | tee "${EVIDENCE_DIR}/02-build.log" || fail "go build"

# ---------------------------------------------------------------------------
# Step 2 -- post-build floor: unit suite under race detector
# ---------------------------------------------------------------------------
log "Step 2: go test ./... -count=1 -race (post-build floor)"
go test ./... -count=1 -race 2>&1 | tee "${EVIDENCE_DIR}/03-test.log" || fail "unit suite"

# ---------------------------------------------------------------------------
# Step 3 -- bundle load sanity
# ---------------------------------------------------------------------------
log "Step 3: bilingual bundle load sanity"
EN_FIX="${SCRIPT_DIR}/fixtures/en.yaml"
SR_FIX="${SCRIPT_DIR}/fixtures/sr-Latn.yaml"
[[ -s "${EN_FIX}" ]] || fail "missing or empty fixture: ${EN_FIX}"
[[ -s "${SR_FIX}" ]] || fail "missing or empty fixture: ${SR_FIX}"
grep -q 'lazy.service.ready' "${EN_FIX}" || fail "en fixture missing lazy.service.ready"
grep -q 'lazy.service.ready' "${SR_FIX}" || fail "sr-Latn fixture missing lazy.service.ready"
printf 'fixtures OK: %s + %s\n' "${EN_FIX}" "${SR_FIX}" | tee "${EVIDENCE_DIR}/04-fixtures.log"

# ---------------------------------------------------------------------------
# Step 4 -- runtime end-to-end: real Translator + real Service.Describe
# ---------------------------------------------------------------------------
log "Step 4: runtime end-to-end (EN+SR Describe round-trip)"
go run ./challenges/runner 2>&1 | tee "${EVIDENCE_DIR}/05-runtime.log" || fail "runtime round-trip"

# ---------------------------------------------------------------------------
# Step 5 -- paired anti-bluff mutation
#
# We corrupt one entry in the SR bundle and assert the runner FAILS. If
# the runner still PASSES after corruption, the assertions are not
# actually checking the output and the suite is a bluff (CONST-035).
# ---------------------------------------------------------------------------
log "Step 5: paired anti-bluff mutation (corrupt SR bundle, expect runner FAIL)"
BACKUP="${SR_FIX}.bak.$$"
cp "${SR_FIX}" "${BACKUP}"
trap 'mv -f "${BACKUP}" "${SR_FIX}" 2>/dev/null || true' EXIT

# Replace the SR "ready" string with the wrong value -- runner MUST notice.
# We use a sed replacement that is purely a fixture mutation, never touches
# source code, and is reverted by the EXIT trap.
sed -i 's/Servis spreman/Servis NEISPRAVAN_MUTATION/' "${SR_FIX}"
grep -q 'NEISPRAVAN_MUTATION' "${SR_FIX}" || fail "mutation did not apply"

set +e
go run ./challenges/runner > "${EVIDENCE_DIR}/06-mutation.log" 2>&1
MUTATION_RC=$?
set -e

if [[ ${MUTATION_RC} -eq 0 ]]; then
    fail "paired-mutation leg: runner exited 0 with corrupted SR bundle -- assertions are not real (CONST-035 bluff)"
fi
printf 'mutation correctly rejected with exit code %d\n' "${MUTATION_RC}" \
    | tee -a "${EVIDENCE_DIR}/06-mutation.log"

# Restore explicitly (also restored by EXIT trap as belt-and-braces).
mv -f "${BACKUP}" "${SR_FIX}"
trap - EXIT

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
log "PASS: lazy_describe_challenge.sh -- all 5 steps green"
printf 'evidence directory: %s\n' "${EVIDENCE_DIR}"
ls -la "${EVIDENCE_DIR}"
exit 0
