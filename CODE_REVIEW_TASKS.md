# Remediation Plan: Persistence foundation

Review ID: foundation-state-20260917. Full review of untracked Task 05 code on Go 1.26.1, darwin/arm64. Reviewed hashes: schema.go `57cef989…f452`, store.go `dad9d61…a12cb4`, store_test.go `521ce80…570b17`. Reviewer `state_review` candidates C1/C2 map to F1/F2 below. Fixes are part of the user's authorized implementation. The state implementer owns `internal/state/**`; the evidence implementer owns `internal/evidence/**` and must coordinate the transaction API.

## Constraints

Preserve the reviewed storage contract and user changes. No runtime invocation, permission changes, commits or remote effects. Stale-owner takeover and full workflow semantics remain outside this kernel. Regression tests must demonstrate the corrected boundaries; test success alone does not prove macOS power-loss guarantees.

## Tasks

### [x] F1: Persist new directory entries — blocking correctness

`Store.CreateRun` creates the runs/run directories but did not sync their containing directories. Sync the state root after creating `runs` and `runs` after creating a run directory. Failed post-creation sync must report uncertainty rather than successful creation or invented rollback. Owner: `state_foundation`; files: store/persistence code and focused state tests. Verify injected failures at both boundaries preserve inspectable data and do not return success. Dependencies: none. Scope: small.

### [x] F2: Reject malformed empty operations — blocking input validation

`validateRun` skipped operation validation when the ID was empty, even when other fields contained invalid references. Accept only the exact empty operation representation, otherwise require valid identity, preconditions and reference shapes. Owner: `state_foundation`; files: state validation and tests. Verify persisted empty-ID operations carrying artifact refs, event ranges, intent or nonzero preconditions fail visibly. Dependencies: none. Scope: small.

## Execution and verification

Fix both sequentially in the shared state package, add focused regressions, then re-review. Task 06 separately requires a lease-owned publication transaction; its new code receives review alongside the corrected kernel, without treating it as a remedy for F1/F2. Run scoped state race tests, then combined state/evidence tests, vet and build once the integration is stable. Record actual results before checking items off.

### [x] F3: Verify failed preparation is non-authoritative — should-fix tests

Re-review resolves F1/F2. Its new test-coverage finding (reported as C1 on the second pass; assigned distinct F3 here) requires a callback that writes inspectable bytes then fails: generation and operation must remain unchanged; bytes remain inspectable; a subsequent valid commit can succeed. Invalid callback refs must also fail without index replacement. Owner: state_foundation, state tests only unless a real defect is exposed. Verify focused race tests and inspect final cases.

## Evidence integration review

Review ID: foundation-evidence-20260917. Reviewer `evidence_review` C1–C7 map to F4–F10. Implementation remains authorized; these findings must be resolved before Task 06 completes. Ownership: state intent work in `internal/state/**` belongs to `state_foundation`; remaining evidence files belong to `evidence_foundation`. Shared transaction changes are sequential and coordinated.

### [x] F4: Persist intent before preparation — blocking recovery

`Lease.Commit` must publish a durable immutable operation intent before invoking preparation. Record expected generation/fence/state and identity so interrupted outputs can be classified. Verify intent exists at the callback boundary, preserves failed attempts, rejects conflicting identity reuse and is synced before side effects.

### [x] F5: Guard and durably record tail recovery — blocking path security/recovery

Replace arbitrary-path recovery with a lease-owned, run-relative operation. Persist recovery intent, validate trusted committed boundary, sync sanitized quarantine, then truncate/sync only the incomplete uncommitted tail. Reject symlinks/escapes, complete malformed records and overlap with committed evidence. Test each forbidden effect and failure boundary.

### [x] F6: Validate complete event envelopes — blocking corruption validation

Require the supported schema, IDs, actor, phase, timestamp and generation/fence semantics. Distinguish complete corruption from an eligible partial tail. Verify `{}`, newer schema, missing fields and corruption before the committed boundary fail rather than allowing further appends.

### [x] F7: Persist evidence directory entries — should-fix durability

Create each missing evidence directory with a parent sync before reporting publication success. Verify failures at ancestor creation/sync boundaries prevent authoritative commit.

### [x] F8: Use one canonical artifact digest — should-fix integrity

The reference digest must equal frontmatter's canonical SHA-256 excluding its digest field. Test canonical/ref equality independently of the full-file hash.

### [x] F9: Bound output at UTF-8 boundaries — should-fix artifact format

Truncation must preserve valid UTF-8, including multibyte characters crossing the byte limit. Keep the visible truncation marker and test the boundary.

### [x] F10: Represent required outcome facts — should-fix evidence coverage

Add structured criteria, validation/review refs, revision/branch, local/delivery state, blockers/cancellation and adapter/probe facts. Represent absent or not-reached stages explicitly; synthetic evidence must not become an actual-runtime claim. Redact all persisted fields and test present/absent cases.

### [x] F11: Hash canonical operation-intent bytes — blocking integrity

The new operation record hashed JSON containing an empty `content_sha256` field, conflicting with its declared omitted-field convention. Owner: state_foundation. Define deterministic JSON canonicalization with the digest field omitted and test by independently decoding/removing/reserializing the persisted record. The recorded hash must match. This is the final state-review digest finding (C1 in that pass), not a reopening of F1.

State review closure: F1–F4 and F11 are resolved by inspected fixes and focused race/shuffle tests (five repetitions). Final reviewed hashes: store `f5b00363…edd67`, transaction `f60a0ec0…79d6`, operation `5cb989e4…d3275`. Evidence F5–F10 remain pending final verification.

Final evidence re-review: F5 remains blocking because recovery does not validate committed range integrity or complete predecessor records. F6 remains blocking because key presence does not validate envelope types/semantics and partial-tail classification masks prior corruption. F10 still needs revision identity and explicit unavailable/not-reached facts. F7/F8 implementations look correct but required failure/canonical digest regressions remain outstanding. F9 accepted by inspection and UTF-8 regression. Corrective work remains assigned to evidence_foundation.

### [x] F12: Verify summary integrity before reconciliation

The new current-summary projection preserves historical files, but reconciliation initially accepted matching IDs/state/generation without verifying required frontmatter or digest. Require complete valid envelope, reject duplicate fields, and verify exact canonical bytes excluding the digest field. Test tampered body, missing/bad digest and duplicate keys. Owner: state_foundation; files summary.go and state tests.

F12 final review accepted required envelope, duplicate rejection and exact-byte digest checks with tampered/missing/bad digest and stale-generation regressions. Reviewed summary hash `2ae4bedd…875f5`; state tests `480b53ca…a13aa`.

Evidence re-review accepted F5–F10 with typed invalid-envelope, corruption/digest no-mutation, injected ancestor-sync failure, independent canonical digest and explicit/redacted facts regressions. One final UTC timestamp-contract correction remains before Task 06 closure.

### [x] F13: Reject non-UTC event timestamps

Event envelope validation accepts RFC3339 offsets despite the UTC persisted-event contract. Require UTC and add an offset-timestamp rejection regression. Owner: evidence_foundation.

Final closure: UTC validation independently accepted (writer `9963865e…ecbac`, validation tests `8ba4f197…a0c37`). All F1–F13 resolved. Final `make check`, `make build`, and `go test -race ./internal/state ./internal/evidence` passed using repository cache settings. No runtime acceptance is inferred.
