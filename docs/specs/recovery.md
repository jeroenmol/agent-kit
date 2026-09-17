# Recovery and reconciliation contract

**Status:** Task 02 reviewed specification for later implementation and fault injection. It relies on the state and artifact contracts and is independent of unproven Codex runtime behavior.

## Entry conditions

Recovery runs before resuming a nonterminal run, reclaiming an expired owner, retrying a Git/delivery operation, cleanup, or migrating state. It obtains a new fenced owner before mutation. If `state.json` has an unsupported schema or integrity failure, recovery is inspection-only: preserve the directory, mark the run blocked in any readable project index, and report the exact artifact/version; it must not guess or rewrite it.

The reconciler compares authoritative index references with immutable artifacts, event stream committed boundary, local filesystem/worktree, Git repository, and, only for a recorded delivery attempt, the intended remote. It records every observation as a recovery operation/event. It never claims an unobserved command outcome.

## Ordered reconciliation

1. Canonicalize roots and validate containment/no symlink escape before opening any path. Acquire lock and new fence; if the former owner may still be running, do not reclaim on lease timeout—prove it exited or complete an explicit operator stop procedure, otherwise reject resume as already owned.
2. Parse and validate index, operation records and referenced artifacts/digests. Preserve and block on corruption or unsupported required schemas.
3. Repair an interrupted event append only if it is an invalid trailing line beyond the index committed event boundary. First durably record a recovery operation, quarantine/copy the suffix, truncate only that uncommitted non-event suffix, then append its stable-ID recovery event and update the index through the normal commit protocol. Never silently truncate an event needed by the committed index.
4. Examine each prepared/uncommitted operation. A recovering owner never commits it using the former owner's fence. If all expected artifact digests, event range and prior generation validate and no external effect is involved, create a new recovery operation at the current fence that references the prepared operation/events and commits the equivalent transition idempotently. Otherwise leave current index unchanged, mark artifacts orphaned/quarantined, and require inspection or a new operation.
5. Reconcile Git/worktree facts and then delivery facts. Update state only with observed facts and a fresh operation. Re-evaluate review approval and grants after every base/SHA/contract change.

Recovery has no “best effort success” path for required evidence, a missing worktree, uncertain external mutation, a stale owner, or an unsupported schema. Those produce `waiting`, `blocked`, or `uncertain` with a reason artifact.

## Git and worktree reconciliation

For each change, inspect the recorded repository identity, worktree path, branch, recorded base SHA, current HEAD SHA/tree, and cleanliness. Do not run destructive Git commands, stash, reset, checkout, clean, rebase, or delete branches/worktrees during recovery.

| Observed condition | Required result |
|---|---|
| Recorded worktree/branch/HEAD match state | Resume only after owner and stage checks. |
| Worktree exists but HEAD/tree differs from recorded revision | Record observed SHA/tree; invalidate validation/review/approval applicability; move to `building` or `blocked` according to whether an owner can explain the change. |
| Worktree missing, but recorded checkpoint exists in Git | Preserve evidence; mark source workspace unavailable and require explicit recreation from recorded SHA before source writing. |
| Worktree missing and checkpoint/ref absent | Block; preserve artifacts and do not recreate from a guessed base. |
| Git operation in progress (`MERGE_HEAD`, rebase/cherry-pick state, lock) | Block and report exact Git state. Do not continue/abort automatically. |
| Dirty worktree with unrecorded content | Block cleanup/delivery; create no commit. A later explicit handoff can adopt it only with a new contract/evidence operation. |
| User checkout differs/is dirty | Informational unless the change was explicitly recorded as depending on it; never stash, stage, or modify it. |

An interrupted checkpoint command is reconciled by inspecting the exact commit and tree. A commit that exists but lacks a committed revision artifact/state reference is not reviewable until recovery records it through a new operation; a revision record that names a missing commit blocks the run. This preserves the distinction between Git's side effect and Agent Kit's durable evidence.

## Delivery reconciliation

Delivery attempts are explicit external operations. Before push or PR mutation, state persists the intended remote URL identity, repository identity, source branch, target base, exact revision/tree/contract, authorization ref, operation ID, provider idempotency key where supported, and state `in_progress`. A timeout, lost process, or ambiguous response changes it to `uncertain`; retry is forbidden until reconciliation.

Reconciliation queries only the recorded remote/provider and searches for the recorded source branch, exact SHA/tree where observable, and recorded provider idempotency key or PR identifier. It records raw response only after redaction/minimization.

| Observation | Result |
|---|---|
| Exact intended push/PR is confirmed | Record remote IDs/URLs and `succeeded`; no retry. |
| Provider positively establishes no effect and no request remains in flight, while authorization/approval is still applicable | Create a new attempt linked to the uncertain one; retry once through the authorized path. |
| Effect exists but target/base/revision differs | `blocked`; do not update, force-push, delete, or create another PR. |
| Remote cannot be queried | Remain `uncertain`; do not infer failure or repeat mutation. |
| Current SHA/base/contract or authorization differs | `blocked` pending fresh validation/review/authorization, even if a prior attempt was prepared. |

PR creation recovery first checks the recorded PR ID; if absent, it searches only the intended repository/source branch and requested target base. More than one candidate is a blocker, never a reason to select one or create another. “Local only” has no delivery attempt and recovery must issue no remote request.

## Ownership handoff and stale operations

A recovered run records the previous owner as stopped/reclaimed and retains its attempt/provenance. A replacement builder can be assigned only after a handoff artifact names the old owner, current contract/revision, open findings, validation evidence, deviations, and remaining tasks. It becomes the one source writer at a new fence only after the old writer is proved stopped. A stale process that later wakes cannot commit state or append events based on its old fence, but fencing alone cannot stop it writing source or Git; any possibility it remains running blocks handoff. Any observed external effect from it is reconciled as uncertain.

## Scenario walkthrough evidence

| Crash point | Evidence preserved | Recovery result |
|---|---|---|
| Before checkpoint Git commit | prepared operation and prior valid index | No revision is invented; return to building after owner recovery. |
| After Git commit, before state commit | commit plus uncommitted operation artifacts | Inspect commit/tree; record a revision only through a fresh validated operation, otherwise block/quarantine. |
| After review side effect/artifact, before index | immutable review/report and events may exist | State remains previous stage until atomic commit validation; no approval/finding status is inferred from Markdown. |
| After index commits approved review | exact review/ref, generation and approval facts | Approval is current only if SHA/tree/base/contract/findings/checks still match. |
| During JSONL append | valid prefix plus partial suffix | Keep valid prefix; quarantine incomplete suffix; reconstruct no event. |
| After remote PR creation, before response/state | in-progress attempt and idempotency/search facts | Set uncertain; query recorded remote/branch before any retry; confirm one PR or block. |

These walkthroughs are specification evidence only. Task 05/06 must implement fault injection at each named boundary and Task 19/29–30 must exercise Git/remote reconciliation in disposable fixtures.

## Open runtime-dependent assumptions

The selected Codex CLI cannot be launched nested on this host and has not proved role enforcement, runtime JSONL, model/effort actuals, or a supported authorization mediator. Runtime telemetry is optional and absent fields remain `unavailable`; builder/reviewer results and required verification evidence remain mandatory for a completed change. Essential state/evidence failures block completion. The file protocol's macOS fsync/rename guarantees, Git provider query APIs, and the exact lock implementation remain implementation tests, not claims of current readiness.
