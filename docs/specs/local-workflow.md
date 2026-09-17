# Single-change local workflow contract

**Status:** Task 03 reviewed specification. It defines the smallest request-to-reviewed-local-commit path. Real builder/reviewer execution is unavailable until the reference runtime proves required enforcement.

## Preconditions, ownership, and role boundary

The orchestrator creates a versioned request, plan and immutable change contract before source writing. The contract names objective, inclusions/exclusions, criteria/tasks, base SHA, worktree/branch, checks, risk/review mode, delivery scope, required capabilities, selected rules/skills and completion guards. Preflight resolves effects and adapter evidence. A missing required check, tool, worktree, or enforcement capability gives `blocked`/`unavailable`, never a pass. Missing optional delivery capability permits local completion only when the contract permits it.

Exactly one logical builder writes source. The orchestrator owns worktree/branch lifecycle, staging, checkpoints, grants and delivery. Before checkpoint/review it invokes adapter `Quiesce`: observed completed/stopped process proof and `Inspect`/termination evidence must confirm it before a worktree snapshot. If the runtime cannot provide that proof, checkpoint/review is unavailable or blocked. Reviewer checks run read-only against reviewed source and may write only declared cache/temp paths. State mutation follows Task 02 generation/fence and operation protocol. A fence cannot prevent a stale process writing Git/filesystem: replacement requires the old process proven dead and a committed handoff; otherwise block and reconcile.

| Role | Receives | May return | Must not do |
|---|---|---|---|
| Orchestrator | request/contract, policy, capability matrix, state facts | assignments, transitions, checkpoints/delivery decisions, outcome | implement app source or claim unknown/failed gates pass. |
| Builder | immutable assignment, contract/task scope, worktree, rules/skills | factual build report, deviation, check evidence, cancellation/handoff | silently broaden scope; manage delivery/Git lifecycle; delegate builders. |
| Reviewer | factual review package | independent verdict/findings/re-review | modify source, push/create PR, alter branches, receive builder private reasoning. |

Reports name assignment/operation, actor, current content identity, output/artifact refs and observed result. A completed builder/reviewer stage requires its digest-linked build/report or review artifact and adapter `LaunchResult` reference in the same state operation/index transition. They become authoritative only when validated there. Required report, output redaction, provenance or evidence publication failures block completion.

## Transition guards

| Transition | Required guard |
|---|---|
| `planned → ready` | Valid contract, base/repository, criteria mapping. |
| `ready → building` | Exclusive live writer, ready worktree, successful capability preflight and builder assignment. |
| `building → verifying` | Builder quiesced with observed completed/stopped proof; bounded report and content identity committed. |
| `verifying → reviewing` | Required checks attach to same content identity and pass, or have explicit authorized visible exception. Immediately after observed quiescence, re-snapshot tracked and attributable untracked source, reconfirm identity, compare the attributable staged Git index/tree to it, create checkpoint, and assert checkpoint tree/SHA maps to it. Otherwise invalidate checks and return to building/verifying. Failed/skipped/timed-out/flaky/unavailable required checks block. |
| `reviewing → approved` | Checkpoint exists; independent review examined exact base/revision/contract; approved verdict and no blocking finding. |
| `reviewing → fixing` | Actionable `changes_requested` or an open blocking finding. |
| `reviewing → blocked` | Blocked/inconclusive review without actionable findings; persist reason/reassessment artifact. |
| `fixing → verifying` | Finding mapping/rationale and new content identity; prior checks/approval do not transfer. |
| `approved → locally_complete` | Exact current approval and checks still apply, all local criteria pass, no blocker, truthful outcome committed. |

Any nonterminal stage may become blocked, failed or cancelled with reason evidence. Resume first reconciles state. Cancellation durably invalidates in-flight assignment, pauses source writing and preserves outcome; checkpoint/review/delivery wait for observed termination/reconciliation. A cancelled change needs explicit resumed/new-attempt decision.

## Verification, review, approval, and bounded fixes

Builders run cheap meaningful checks during work, then contract checks before review. Evidence records check/command, content identity, context, versions where available, timing, exit/result and bounded redacted output ref. Checks changing tracked source invalidate earlier evidence. Pre-existing failures/exceptions remain visible.

After quiescence, the orchestrator re-snapshots tracked source and contract-attributable untracked files, confirms the validated identity, and compares the attributable staged Git index/tree before creating an immutable checkpoint. The checkpoint tree/SHA must map to that identity; a mismatch or unexpected staged/untracked path invalidates checks and returns to building/verifying. Initial policy retains the approved checkpoint as final local commit; squash/rebase/reword is unsupported. A revision has full SHA, tree SHA, base SHA, predecessor, contract and check refs.

The reviewer gets contract/criteria, risk/mode, base and revision SHA/tree, diff, relevant project instructions, validation evidence, constraints, prior findings/fix mapping and surrounding code. It excludes builder reasoning/self-assessment. Findings have stable ID, severity/category, evidence/location, impact, required correction and status (`open`, `addressed_pending_review`, `resolved`, `rejected_with_rationale`). Only reviewer resolution closes a blocking finding. Verdicts are approved, changes requested, blocked or inconclusive.

Approval applies only when change, revision/full SHA/tree, base, contract digest/version, supported review schema, required check identities and zero blocking finding snapshot exactly match. Changes invalidate it. Three automatic fix/re-review cycles maximum. Recurrence means the same finding ID or materially equivalent category/location/impact, or reviewer-declared recurrence. On the third unsuccessful cycle or recurrence without new diagnosis, stop and record reassessment/replan or blocker; never lower criteria.

## Scope and outcome

Material out-of-contract work pauses the task and records cause, scope, criteria impact, dependencies and options. The orchestrator versions contract/plan, splits, defers, or seeks resolution for changed objective/authority. Changed contract invalidates approval and rechecks effects/validation.

Outcome separately states criteria, checks, review, checkpoint/branch, local completion, delivery state, cancellation, limitations and blockers. It never calls local work delivered or unavailable usage zero. Provenance records toolkit and adapter/runtime versions, probes/capabilities, requested/actual model/effort when available, effective non-secret config, selected definitions, base/revision. Required output/provenance failure prevents completion.
