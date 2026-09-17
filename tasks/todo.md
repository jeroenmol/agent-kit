# Agent Kit implementation tasks

Status: implementation authorized on 2026-09-17; Tasks 02–03 specifications complete; Task 01 runtime probes remain blocked; runtime-independent Tasks 04–06 are authorized to proceed against the reviewed contracts. Sub-agents use `gpt-5.6-terra` at medium reasoning effort. Source: [PRD v0.4](../docs/PRD.md). Decisions and milestones: [plan](plan.md).

This file is the authoritative implementation checklist. Paths below are proposals, not existing files. `<ext>` and the language-specific skeleton paths are resolved in Task 1. Each task includes its tests within its file budget. Split any task that exceeds five files or one focused session before implementing it.

Verification convention: Task 4 establishes exact build, formatting, focused-test and full-test commands. Later implementation tasks must run the relevant focused tests and build; checkpoints run the accumulated suite and required integration scenarios. Until Task 4, these commands do not exist. Documentation/spec tasks use the explicit walkthrough below instead of fictitious test commands. Check off criteria only with evidence.

## Task 01: Prove Codex integration feasibility

**Execution status:** Partial progress: user-run normal-Terminal startup succeeded (`started`, exit 0, `ready_response`) on 2026-09-17. Coordinator verified the sanitized status artifact. Three user-run read-only reviewer probes completed with unchanged source. Schema diagnostics confirmed assistant-only output; revised explicit tool-call wording produced a refusal classification and still no command events. Tool execution remains unresolved; no permission boundary was exercised. Further prompt-only reruns are paused. Nested execution in this managed host remains blocked; activation and role enforcement still unproven. Independently reviewed interface evidence and blocker ADR recorded in [ADR 001](../docs/decisions/001-reference-runtime.md) and [probe results](../experiments/runtime-probe-results.md). Enforcement/runtime acceptance remains incomplete.

- [ ] Complete

**Description:** Exercise activation, isolated role execution and permission mediation in a disposable repository; target Codex first, select its supported integration surface/platform and choose a CLI language based on evidence.

**Acceptance criteria:**
- [ ] A versioned capability matrix links actual probe results for activation, child recursion, delegation, model/effort controls and telemetry.
- [ ] Alternate shell/tool attempts at reviewer source writes, builder delivery and protected effects are rejected; advisory or unsupported boundaries are explicit.
- [x] A decision record records the user-selected Codex runtime, settles the remaining H-01/H-02 surface/platform/language choices or records a blocker; failed enforcement does not become a supported integration.

**Verification:**
- [ ] Reproduce each documented probe on the selected runtime; record versions, exit results and redacted evidence.

**Implementation notes (contract audit):** For each protected effect, identify the host/tool enforcement boundary, bound actor, canonical resource, alternate-route probe and result. A capability matrix alone does not establish an enforceable adapter.

**Dependencies:** None

**Files likely touched:** `docs/decisions/001-reference-runtime.md`, `experiments/runtime-probe.md`, `experiments/runtime-probe-results.md`

**Estimated scope:** Small (3 proposed files).

**Coverage:** IN-04, PE-07; AC-02/12; H-01/02/14

## Task 02: Specify durable state and recovery

**Execution status:** Specification complete after independent contract review and coordinator verification of fixes. Evidence: [state](../docs/specs/state.md), [artifacts](../docs/specs/artifact-contracts.md), [recovery](../docs/specs/recovery.md). These are reviewed contracts and walkthroughs, not executed runtime tests.

- [x] Complete

**Description:** Define a minimal versioned artifact/index contract and guarded transitions before implementing persistence.

**Acceptance criteria:**
- [x] Schemas cover entity IDs, lineage, contract versions, change/run/delivery states and authoritative fields; Markdown summaries have a reconciliation rule.
- [x] Specify atomic replacement, event append recovery, ownership locks, interrupted operation reconciliation and unsupported-version behavior.
- [x] Set configurable source/config/state roots, evidence minimization and retention/reachability rules, resolving H-03/H-04/H-11.

**Verification:**
- [x] Walk through crashes before and after checkpoint, review and delivery side effects; verify every PRD entity has identity and lineage rules.

**Implementation notes (contract audit):** Define cross-artifact commit/recovery semantics for state, Markdown and JSONL: operation identity, state generation, owner fencing, durable commit point and reconciliation of partial writes. Define immutable artifact references and path/override rules. Break this task into separately reviewed document steps if its contract cannot fit one focused session.

**Dependencies:** Task 1 candidate interface/decision evidence for drafting; actual-runtime readiness remains blocked and is not implied by completing these platform-neutral specs.

**Files likely touched:** `docs/specs/state.md`, `docs/specs/artifact-contracts.md`, `docs/specs/recovery.md`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** DM-01/02, ST-01–06, TE-01/02

## Task 03: Specify the single-change execution contract

**Execution status:** Specification complete after independent review and reconciliation with Task 02. Evidence: [workflow](../docs/specs/local-workflow.md), [adapter](../docs/specs/runtime-adapter.md), [permissions](../docs/specs/permissions.md), [fixtures](../docs/specs/acceptance-fixtures.md). Actual-runtime fixtures remain blocked, not passed.

- [x] Complete

**Description:** Define the smallest complete request-to-reviewed-local-commit path, including the adapter and authorization interfaces.

**Acceptance criteria:**
- [x] Specify role inputs/results, validation evidence, exact-revision approval, cancellation, scope escalation and source-writer ownership.
- [x] Define capability precedence, grants with resource/lifetime/source, cache/temp scopes and fail-closed handling; record which controls are host-enforced.
- [x] Define fixture scenarios and loop defaults; provisionally retain checkpoint commits and use three fix cycles with repeated-finding detection.

**Verification:**
- [x] Trace successful, failing, cancelled and denied requests through the transition table; map guards to AC-05/06/08/09/10/12/15/16.

**Implementation notes (contract audit):** Bind every enforcement claim to Task 01 evidence and runtime version. Specify transitive operation effects, resource matching, grant expiry and unsupported controls. Review this contract jointly with Task 02 recovery semantics.

**Dependencies:** 1, 2

**Files likely touched:** `docs/specs/local-workflow.md`, `docs/specs/runtime-adapter.md`, `docs/specs/permissions.md`, `docs/specs/acceptance-fixtures.md`

**Estimated scope:** Medium (4 proposed files).

**Coverage:** RO-01–04, VE-01–04, RE-01–07, PE-01–10; H-06/09/10

## Checkpoint after Tasks 01–03: Feasibility and contracts

**Status:** Not passed. Specs reviewed; nested runtime initialization failed before model invocation, so activation and enforcement remain unproven. Task 04 proceeds independently under the plan continuation decision; runtime readiness remains blocked.

- [ ] Runtime/enforcement probes support the proposed design; review the decision records before depending on unproven mechanisms.
- [ ] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [ ] Evidence and scope reviewed; record material decisions before the next phase.

## Task 04: Create an installable CLI skeleton

**Execution status:** Complete — implemented by `cli_foundation`, independently reviewed by `cli_review` with no findings. Go 1.26.1; `make focused-test`, `make check`, `make build`, fresh-directory build and binary exit/output probes passed. Lifecycle operations remain explicitly unavailable (69). Reviewed source hashes: root.go `56f59cb2`, main.go `ca0cdd52`, root_test.go `414c9c40`.

- [x] Complete

**Description:** Introduce the selected language/toolchain and a minimal executable with deterministic development commands.

**Acceptance criteria:**
- [x] A fresh checkout builds an executable with version identity and useful exit codes.
- [x] The four required lifecycle command names exist; unfinished operations explicitly report unavailable rather than success.
- [x] Document exact focused-test, full-test, formatting and build commands for subsequent tasks.

**Verification:**
- [x] Build from a clean checkout and exercise version, help, an invalid command and an unavailable operation.
- [x] Run the established focused tests and build; record commands/results.

**Implementation notes (contract audit):** The lifecycle commands are setup, doctor, init and update. Define version/help, exit-code taxonomy and unavailable-operation behavior; reserve status/resume grammar for Task 19.

**Dependencies:** Task 1 candidate language/interface decision and Task 3 reviewed contract; actual-runtime readiness is not required for this unavailable-only skeleton.

**Files likely touched:** `go.mod`, `go.sum` if needed, `cmd/agent-kit/`, `Makefile`, `README.md`, `.gitignore` (small scaffold exception to five-file estimate).

**Estimated scope:** Medium (5 proposed files).

**Coverage:** UX-01/03

## Task 05: Persist one run safely

**Execution status:** Complete — persistence kernel, lease transactions, durable intents and summary reconciliation independently reviewed. Focused race/shuffle tests passed five repetitions; combined build/tests passed. Summary review accepted exact-byte digest validation and stale/tampered-summary detection. Evidence integration remains under Task 06 review; see [review remediation](../CODE_REVIEW_TASKS.md).

- [x] Complete

**Description:** Implement versioned external run records, ownership and recoverable state replacement.

**Acceptance criteria:**
- [x] Create/read/update a run with stable IDs and valid lineage outside the target checkout.
- [x] Concurrent owner acquisition fails; crashes cannot replace a valid record with partial content.
- [x] Unknown schema versions fail visibly; reconciled summaries reflect authoritative records.

**Verification:**
- [x] Focused persistence tests inject interrupted writes, corrupt records and simultaneous ownership attempts in temporary state roots.
- [x] Run the established focused tests and build; record commands/results.

**Implementation notes (contract audit):** This is a generic versioned store plus run/ownership slice, not all lifecycle persistence. Cover stale-owner fencing, canonical path containment/symlink handling and the Task 02 cross-artifact protocol.

**Dependencies:** 2, 4

**Files likely touched:** `internal/state/store.<ext>`, `internal/state/store_test.<ext>`, `internal/state/schema.<ext>`, `templates/run.md`

**Estimated scope:** Medium (4 proposed files).

**Coverage:** DM-01/02, ST-01–05

## Task 06: Capture minimum safe run evidence

**Execution status:** Complete — synthetic outcomes, redaction, typed events and lease-owned interrupted-tail recovery independently reviewed. F5–F10 and final UTC validation correction resolved. `make check`, `make build`, and combined state/evidence race tests passed on the final code. Runtime settings are recorded as supplied facts with explicit unavailable values; these tests make no actual-runtime claim. Evidence scope split across record/redaction validation and lease-owned publication/recovery (eight files including templates).

- [x] Complete

**Description:** Record events, provenance and outcomes before any real builder execution.

**Acceptance criteria:**
- [x] Success, failure, cancellation and blocked runs produce parseable events and linked Markdown outcomes.
- [x] Record toolkit dirty-content identity, loaded-definition hashes and requested/actual runtime settings; missing usage remains unavailable.
- [x] Redact a seeded secret before persistence; interrupted event appends recover detectably and essential evidence failures block completion.

**Verification:**
- [x] Inject secret-bearing outputs, truncated records, unavailable telemetry and write failures; scan every persisted artifact for seeded secrets.
- [x] Run the established focused tests and build; record commands/results.

**Implementation notes (contract audit):** Use synthetic actors/attempts before builder support. Define mandatory evidence versus optional telemetry, redaction across all persisted channels, sanitized artifact references and fail-closed outcome handling. Consume Task 02 commit/recovery rules.

**Dependencies:** 5

**Files likely touched:** `internal/evidence/writer.<ext>`, `internal/evidence/writer_test.<ext>`, `internal/evidence/redaction.<ext>`, `templates/outcome.md`, `templates/event-schema.json`

**Estimated scope:** Medium (5 proposed files).

**Coverage:** PE-09, IM-01, TE-01–03; AC-16

## Checkpoint after Tasks 04–06: Evidence foundation

**Status:** Passed for runtime-independent foundations. CLI lifecycle commands remain unavailable; Task 01 and actual-runtime acceptance remain open.

- [x] CLI builds; state/evidence fault tests pass; no real execution is enabled without minimum evidence capture.
- [x] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [x] Evidence and scope reviewed; record material decisions before the next phase.

## Task 07: Make setup and doctor useful

- [ ] Complete

**Description:** Install the verified minimal bootstrap/adapter registration and diagnose actual readiness.

**Acceptance criteria:**
- [ ] Setup twice preserves unrelated settings and creates one registration with material changes reported.
- [ ] Doctor checks source/config/state, roles, tools, model mapping and enforcement/activation support with actionable remediation.
- [ ] Failed installation remains recoverable; doctor performs only documented temporary probes and reveals no credentials.

**Verification:**
- [ ] Run twice in an isolated user home; seed unrelated configuration and interrupt installation; compare before/after settings.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 4, 6

**Files likely touched:** `internal/install/setup.<ext>`, `internal/install/setup_test.<ext>`, `internal/install/doctor.<ext>`, `integrations/reference/bootstrap.md`, `integrations/reference/manifest.json`

**Estimated scope:** Medium (5 proposed files).

**Coverage:** IN-01, UX-01/02/04; AC-01

## Task 08: Register an existing repository externally

- [ ] Complete

**Description:** Implement init as non-invasive discovery with explicit provenance for validation candidates.

**Acceptance criteria:**
- [ ] Capture repository/base identity, dirty state, applicable instructions, language/build tools and candidate validation commands externally.
- [ ] Tracked and untracked project files remain byte-for-byte unchanged by default.
- [ ] Untrusted repository text cannot grant policy authority; candidate checks require suitability and capability validation before execution.

**Verification:**
- [ ] Initialize a dirty fixture with nested instructions and misleading command text; compare file manifests and inspect candidate provenance.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 5, 7

**Files likely touched:** `internal/project/init.<ext>`, `internal/project/init_test.<ext>`, `internal/project/discovery.<ext>`, `templates/project.md`

**Estimated scope:** Medium (4 proposed files).

**Coverage:** IN-02, GT-02, ST-01; AC-03

## Task 09: Enforce scoped role operations

- [ ] Complete

**Description:** Implement the approved permission contract using real adapter/tool restrictions, not prompt-only controls.

**Acceptance criteria:**
- [ ] Resolve allow/ask/deny against host ceilings, explicit user scope, role and resource; skill requirements never become grants.
- [ ] Required operation effects include execution, filesystem, network, Git and credentials; opaque credential use does not grant secret reads.
- [ ] Denied or unsupported protected actions fail closed without alternate-route retries; allowed scoped checks avoid redundant prompts.

**Verification:**
- [ ] Run the actual-runtime AC-12 adversarial probes plus deterministic grant-precedence tests; demonstrate allowed tests/cache writes.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 3, 6, 7

**Files likely touched:** `internal/permissions/policy.<ext>`, `internal/permissions/policy_test.<ext>`, `integrations/reference/permissions.<ext>`, `policies/permissions.yaml`

**Estimated scope:** Medium (4 proposed files).

**Coverage:** PE-01–10; AC-12

## Checkpoint after Tasks 07–09: Installed and bounded

- [ ] Setup/init preserve user data; real runtime enforces AC-12 including alternate shell/tool routes.
- [ ] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [ ] Evidence and scope reviewed; record material decisions before the next phase.

## Task 10: Route session intent into the right workflow

- [ ] Complete

**Description:** Connect ordinary supported-runtime prompts to proportionate workflow entry points.

**Acceptance criteria:**
- [ ] Questions/exploration create no implementation run/worktree; planning creates only planning artifacts; review-only starts no builder.
- [ ] First implementation prompts activate automatically, later implementation revalidates saved plans, and child roles cannot recurse into orchestration.
- [ ] Bootstrap records session/repository context and loads only stage-relevant definitions including applicable project instructions.

**Verification:**
- [ ] Exercise first and subsequent prompts and child-role invocations in the selected runtime; inspect created artifacts and loaded context.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 7, 8, 9

**Files likely touched:** `integrations/reference/activation.<ext>`, `internal/routing/intent.<ext>`, `internal/routing/intent_test.<ext>`, `agents/orchestrator.md`

**Estimated scope:** Medium (4 proposed files).

**Coverage:** IN-03–05, CX-01/02; AC-02

## Task 11: Create one executable change contract

- [ ] Complete

**Description:** Turn an implementation request into an immutable-versioned acceptance contract with preflight.

**Acceptance criteria:**
- [ ] Preserve original request and interpretation; assign stable criteria/tasks and included/excluded scope, validation, risk and delivery scope.
- [ ] Check required tools/capabilities before builder assignment; missing optional remote rights do not block local work.
- [ ] Changed objectives or needed uncommitted user input trigger explicit resolution; ordinary routine choices remain autonomous.

**Verification:**
- [ ] Use small, ambiguous, unavailable-check and local-only requests; verify criterion mapping and no execution before readiness.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 8, 9, 10

**Files likely touched:** `internal/planning/contract.<ext>`, `internal/planning/contract_test.<ext>`, `templates/change.md`, `templates/plan.md`

**Estimated scope:** Medium (4 proposed files).

**Coverage:** PL-01/05, PE-04/05, GT-02

## Task 12: Isolate the change in its own worktree

- [ ] Complete

**Description:** Create the dedicated branch/worktree from the recorded base with one source owner.

**Acceptance criteria:**
- [ ] Worktree is external to the user checkout; only the orchestrator mutates branch/worktree lifecycle.
- [ ] Dirty user files remain untouched; no implicit stash, stage or adoption of uncommitted content.
- [ ] Duplicate writers are rejected and partial worktree creation reconciles safely on retry.

**Verification:**
- [ ] Exercise dirty tracked/untracked files, concurrent acquisition and interrupted Git operations in disposable repositories.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 5, 9, 11

**Files likely touched:** `internal/git/worktree.<ext>`, `internal/git/worktree_test.<ext>`, `internal/state/ownership.<ext>`, `skills/git-workflow/SKILL.md`

**Estimated scope:** Medium (4 proposed files).

**Coverage:** GT-01/02, RO-01; AC-10

## Checkpoint after Tasks 10–12: Ready isolated request

- [ ] Intent/contract/worktree path works; dirty user checkout and second-writer tests pass.
- [ ] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [ ] Evidence and scope reviewed; record material decisions before the next phase.

## Task 13: Run the assigned builder

- [ ] Complete

**Description:** Execute one builder against the change contract and retain its handoff evidence.

**Acceptance criteria:**
- [ ] Builder receives scoped contract, relevant rules and one suitable coding/debugging skill; it cannot manage delivery or delegate additional builders.
- [ ] Build reports record task progress, deviations and factual evidence without broadening authorized scope.
- [ ] Cancellation pauses source writing and persists an outcome; replacement requires a recorded handoff preserving findings and provenance.

**Verification:**
- [ ] Run a bounded fixture and interrupt it; attempt forbidden builder operations and verify explicit replacement ownership.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 6, 9, 12

**Files likely touched:** `agents/builder.md`, `integrations/reference/builder.<ext>`, `internal/workflow/build.<ext>`, `internal/workflow/build_test.<ext>`, `skills/coding/SKILL.md`

**Estimated scope:** Medium (5 proposed files).

**Coverage:** RO-01/04, CX-02, SK-01/03

## Task 14: Gate progress on recorded verification

- [ ] Complete

**Description:** Run cheap meaningful checks progressively and the required contract checks before review.

**Acceptance criteria:**
- [ ] Evidence records command, context, content identity, tool versions when available, timing, output reference and exit/result.
- [ ] Failure, timeout, unavailable, skipped and flaky results remain distinct; required non-passes block approval.
- [ ] Tests that mutate tracked source invalidate earlier content identity; pre-existing failures need visible authorized contract decisions.

**Verification:**
- [ ] Use passing, failing, missing, timed-out and source-generating checks; ensure only unchanged-source results attach to checkpoints.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 6, 11, 13

**Files likely touched:** `internal/verification/runner.<ext>`, `internal/verification/runner_test.<ext>`, `templates/validation.md`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** VE-01–04, RO-02; AC-08

## Task 15: Checkpoint attributable source for review

- [ ] Complete

**Description:** Pause the builder and create evidence-bearing immutable revisions.

**Acceptance criteria:**
- [ ] Stage only attributable paths after required verification and record SHA, tree, base, contract and previous revision.
- [ ] Checkpoint creation verifies source stability and rejects races or unexpected source changes.
- [ ] Retain evidence revisions with explicit Git refs and reuse the approved checkpoint as final commit; rewriting remains unsupported initially.

**Verification:**
- [ ] Inject source changes between validation/staging/commit and failures around commit creation; verify refs and no unrelated staged content.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 12, 14

**Files likely touched:** `internal/git/checkpoint.<ext>`, `internal/git/checkpoint_test.<ext>`, `internal/state/revision.<ext>`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** GT-03/04, RO-02, ST-06; AC-09

## Checkpoint after Tasks 13–15: Verified immutable revision

- [ ] A real builder produces meaningful validation and an attributable stable checkpoint; failures block review.
- [ ] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [ ] Evidence and scope reviewed; record material decisions before the next phase.

## Task 16: Review an exact revision independently

- [ ] Complete

**Description:** Invoke a source-stable reviewer with a factual package and structured verdict.

**Acceptance criteria:**
- [ ] Package contains exact revision/base, contract, risk, diff, relevant rules and checks; builder private reasoning/self-assessment is excluded.
- [ ] Reviewer can inspect surrounding source and run scoped checks but cannot modify source or workflow Git state.
- [ ] Stable findings include severity, evidence, impact and required correction; verdict is approved, changes requested or blocked/inconclusive.

**Verification:**
- [ ] Actual-runtime seeded-defect review plus package isolation and write-denial tests; approval is impossible with unresolved blockers.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 9, 15

**Files likely touched:** `agents/reviewer.md`, `internal/review/package.<ext>`, `internal/review/package_test.<ext>`, `integrations/reference/reviewer.<ext>`, `templates/review.md`

**Estimated scope:** Medium (5 proposed files).

**Coverage:** CX-04, RE-01/03/04/06; AC-05/06/12

## Task 17: Close findings through bounded re-review

- [ ] Complete

**Description:** Return actionable findings to the same logical builder and assess each resulting revision.

**Acceptance criteria:**
- [ ] Fixes map stable finding IDs to new checks/revisions; only review resolves blocking findings.
- [ ] Focused review checks fixes, introduced regressions and resulting contract compliance; changed scope/base/risk broadens review.
- [ ] Configured cycle limit and repeated unresolved findings trigger reassessment or a blocker without lowering acceptance.

**Verification:**
- [ ] Seed a fixable defect and an unfixable/repeating defect; verify new-SHA approval, status progression and finite termination.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 13, 14, 16

**Files likely touched:** `internal/workflow/fix_loop.<ext>`, `internal/workflow/fix_loop_test.<ext>`, `templates/fix-report.md`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** RE-05/07, MO-04; AC-06/08

## Task 18: Report a truthful local outcome

- [ ] Complete

**Description:** Complete a single change only when all acceptance, validation and exact-revision gates hold.

**Acceptance criteria:**
- [ ] Completion checks current content/base/contract, criteria coverage and unresolved blockers before recording local completion.
- [ ] Progress and outcome link validation/review evidence, branch/commit, limitations and separate delivery status.
- [ ] Local-only requests cause no remote mutation; missing evidence or stale approval yields an incomplete/blocked result.

**Verification:**
- [ ] Run one real small Go change end to end without manual role coordination; mutate base/content/contract after review and verify rejection.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 6, 16, 17

**Files likely touched:** `internal/workflow/complete.<ext>`, `internal/workflow/complete_test.<ext>`, `internal/reporting/outcome.<ext>`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** DM-03, IN-05, RE-06/08; AC-05/09/16

## Checkpoint after Tasks 16–18: Complete local execution

- [ ] Actual-runtime AC-05/06/08/09 succeeds; local completion includes independent review and no remote effects.
- [ ] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [ ] Evidence and scope reviewed; record material decisions before the next phase.

## Task 19: Inspect and resume interrupted local runs

- [ ] Complete

**Description:** Expose status and resume with reconciliation before any restarted work.

**Acceptance criteria:**
- [ ] Status exposes stage, next action, blocker, owners, evidence and delivery facts.
- [ ] Resume checks filesystem/Git reality after build, verification, checkpoint and review interruption without duplicate writers or lost evidence.
- [ ] Preserve original definitions/grants and their lifetimes; changed configuration or contract invalidates affected evidence explicitly.

**Verification:**
- [ ] Fault-inject at each local stage boundary, restart the process and complete or report an actionable blocker; test stale owner recovery.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 5, 18

**Files likely touched:** `internal/workflow/resume.<ext>`, `internal/workflow/resume_test.<ext>`, `internal/cli/status.<ext>`, `internal/cli/resume.<ext>`

**Estimated scope:** Medium (4 proposed files).

**Coverage:** ST-03–05, TE-02, UX-03; AC-15

## Task 20: Update an installation safely

- [ ] Complete

**Description:** Implement intentional definition/integration updates with compatibility checks and recovery.

**Acceptance criteria:**
- [ ] Update identifies source/target versions and preserves local edits/unrelated settings, stopping safely on conflicts.
- [ ] Active runs retain pinned definitions; unsupported schema/integration changes report incompatibility before replacement.
- [ ] Interrupted update leaves a recoverable installation and doctor reports accurate readiness.

**Verification:**
- [ ] Exercise clean update, dirty toolkit, active run, incompatible version and mid-update failure under temporary config/state roots.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 7, 19

**Files likely touched:** `internal/install/update.<ext>`, `internal/install/update_test.<ext>`, `docs/operations/update.md`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** UX-01/02; AC-01

## Task 21: Validate the first local-use milestone

- [ ] Complete

**Description:** Package reproducible local acceptance scenarios and a practical first-run guide.

**Acceptance criteria:**
- [ ] Actual-runtime evidence covers activation, Go implementation, seeded fix loop, dirty checkout, permission bypass attempts and interrupted resume.
- [ ] Successful and unsuccessful runs preserve valid redacted evidence; unsupported runtime capabilities are documented.
- [ ] A user can follow setup → doctor → init → ordinary request → status/resume with no manual role coordination.

**Verification:**
- [ ] Execute the guide from isolated config/state against disposable Go fixtures using the selected runtime; link results by version/configuration.

**Dependencies:** 18, 19, 20

**Files likely touched:** `evals/local/manifest.yaml`, `evals/local/scenarios.md`, `docs/quickstart.md`, `docs/limitations.md`, `docs/acceptance/local.md`

**Estimated scope:** Medium (5 proposed files).

**Coverage:** AC-01/02/03/05/06/08/09/10/12/15/16

## Checkpoint after Tasks 19–21: First usable local milestone

- [ ] Clean install/update/resume walkthrough passes and core local acceptance evidence is linked; suitable for controlled personal use, not full v0.1.
- [ ] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [ ] Evidence and scope reviewed; record material decisions before the next phase.

## Task 22: Execute sequential multi-change plans

- [ ] Complete

**Description:** Extend the working single-change flow to dependency-ordered coherent slices.

**Acceptance criteria:**
- [ ] Plans reject dependency cycles and map every request criterion to tasks/changes; changes execute sequentially.
- [ ] A local dependent change starts only from an immutable approved predecessor checkpoint under a recorded readiness rule.
- [ ] Material scope growth records measured size and a preserved plan revision rather than silently redefining success.

**Verification:**
- [ ] Run a three-slice fixture with independent/dependent concerns, a cycle, an unfinished predecessor and unexpected scope growth.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 21

**Files likely touched:** `internal/planning/dependencies.<ext>`, `internal/planning/dependencies_test.<ext>`, `internal/workflow/sequence.<ext>`, `internal/workflow/sequence_test.<ext>`

**Estimated scope:** Medium (4 proposed files).

**Coverage:** PL-02–05, GT-08; AC-04/18; H-05

## Task 23: Select proportional review rigor

- [ ] Complete

**Description:** Add inspectable risk signals and a documented light/standard/thorough policy.

**Acceptance criteria:**
- [ ] Low-risk, ordinary and tiny security-sensitive fixtures receive justified modes; default is standard and no implemented change skips review.
- [ ] Mode changes respond to new risk independently of model selection and diff size.
- [ ] Review remains at change boundaries; aggregate acceptance/integration checks request only justified additional review.

**Verification:**
- [ ] Evaluate labeled risk cases and inspect review counts on multi-task/multi-slice fixtures.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 16, 22

**Files likely touched:** `policies/review.yaml`, `internal/review/risk.<ext>`, `internal/review/risk_test.<ext>`, `evals/policy/review.md`

**Estimated scope:** Medium (4 proposed files).

**Coverage:** RE-01/02/08; AC-07/18; H-08

## Task 24: Route skills with observable context use

- [ ] Complete

**Description:** Add a small versioned catalog with stage-aware selection and loading evidence.

**Acceptance criteria:**
- [ ] Included skills declare use/non-use, procedure, constraints, required capabilities, version and completion conditions.
- [ ] Track considered/selected/loaded/observably-used states and reasons; unobservable adherence remains unknown.
- [ ] Positive/negative fixtures cover relevant coding/debugging, irrelevant installed skills and lifecycle timing; measure bootstrap/context baseline.

**Verification:**
- [ ] Replay labeled feature/debugging/local-only fixtures; check selected hashes, load events, adherence evidence and baseline overhead.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 6, 13, 23

**Files likely touched:** `internal/skills/catalog.<ext>`, `internal/skills/routing.<ext>`, `internal/skills/routing_test.<ext>`, `evals/policy/skills.md`, `docs/specs/skill-contract.md`

**Estimated scope:** Medium (5 proposed files).

**Coverage:** CX-01–04, SK-01–05; AC-13; H-12

## Checkpoint after Tasks 22–24: Proportional multi-change work

- [ ] Three-slice request completes with risk-appropriate review and relevant context only.
- [ ] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [ ] Evidence and scope reviewed; record material decisions before the next phase.

## Task 25: Allocate models and effort explicitly

- [ ] Complete

**Description:** Resolve portable compute profiles and bounded reassessment based on stage, uncertainty and risk.

**Acceptance criteria:**
- [ ] Record requested and actual profile/model/effort separately; unsupported controls and unavailable usage are honest.
- [ ] Initial capable/medium mapping is evaluated against labeled fixtures; lower/higher allocation has recorded reasons.
- [ ] Difficult/repeated-failure cases trigger bounded reassessment, preserving review rigor and configured budgets.

**Verification:**
- [ ] Use adapter doubles for unsupported controls and an actual-runtime difficult fixture; compare full attempt costs only where observable.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 17, 24

**Files likely touched:** `policies/models.yaml`, `internal/allocation/policy.<ext>`, `internal/allocation/policy_test.<ext>`, `evals/policy/allocation.md`

**Estimated scope:** Medium (4 proposed files).

**Coverage:** MO-01–05; AC-14; H-07

## Task 26: Report run quality and effort

- [ ] Complete

**Description:** Produce a basic evidence-derived report without requiring a dashboard.

**Acceptance criteria:**
- [ ] Report completion, blockers/cancellations, review cycles/findings, validation failures, scope growth and permission friction.
- [ ] Include skill routing/adherence and allocation comparisons when labeled/evidenced; missing token/cost data includes availability coverage.
- [ ] Metrics retain provenance and separate facts, estimates and suspected causal explanations; ordinary successes remain represented.

**Verification:**
- [ ] Compare a mixed success/failure/cancelled fixture dataset to hand-calculated totals; verify unavailable metrics are not zero.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 6, 25

**Files likely touched:** `internal/reporting/metrics.<ext>`, `internal/reporting/metrics_test.<ext>`, `templates/report.md`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** TE-03/04, IM-07; H-14

## Task 27: Protect retained evidence during cleanup

- [ ] Complete

**Description:** Implement explicit retention policy with safe preview and guarded cleanup.

**Acceptance criteria:**
- [ ] Cleanup refuses active/dirty/unrecorded worktrees and evidence whose configured retention has not expired.
- [ ] Retained review/checkpoint refs and artifacts remain recoverable after eligible worktree removal.
- [ ] Preview reports affected resources; interrupted cleanup is recoverable and sensitive raw-output retention is bounded by policy.

**Verification:**
- [ ] Use expired, active, dirty and only-copy evidence fixtures; interrupt removal and verify retained revisions/artifacts are still accessible.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 19, 26

**Files likely touched:** `internal/state/retention.<ext>`, `internal/state/retention_test.<ext>`, `docs/operations/retention.md`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** GT-09, ST-06, PE-06; H-11

## Checkpoint after Tasks 25–27: Routine local-use readiness

- [ ] Recovery, policy allocation, reports and safe retention work together; local-use limitations are explicit.
- [ ] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [ ] Evidence and scope reviewed; record material decisions before the next phase.

## Task 28: Specify authorized remote delivery

- [ ] Complete

**Description:** Define one delivery provider and dependent-PR policy on top of reviewed local completion.

**Acceptance criteria:**
- [ ] Specify remote identity, grant scope, exact-revision gate and durable push/PR attempt/reconciliation records.
- [ ] Adopt or replace the provisional wait-for-predecessor-merge policy; new bases require revalidation/review and dependency waiting stays distinct.
- [ ] Retain checkpoint commits initially; merge/deploy/publish/force-push remain outside v0.1 authority.

**Verification:**
- [ ] Walk through local-only, pre-authorized PR, denied rights, wrong remote, stale approval and uncertain network responses.

**Dependencies:** 22, 27

**Files likely touched:** `docs/specs/delivery.md`, `skills/delivery/SKILL.md`, `evals/delivery/scenarios.md`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** GT-04–08, PE-05; AC-09/11/18; H-05/06

## Task 29: Push the exact approved revision

- [ ] Complete

**Description:** Deliver a reviewed branch only to the authorized repository/remote.

**Acceptance criteria:**
- [ ] Gate on current SHA/base/contract, required checks, blockers, attributable contents and all scoped capabilities.
- [ ] Persist attempt before push and reconcile uncertain outcomes through remote refs before retry.
- [ ] Local-only performs no remote mutation; existing scoped authorization avoids redundant Agent Kit prompts and host prompts remain distinguishable.

**Verification:**
- [ ] Use a controlled remote fixture for accepted/denied push, stale SHA/base and interruption before/after remote acceptance.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 9, 18, 28

**Files likely touched:** `internal/delivery/push.<ext>`, `internal/delivery/push_test.<ext>`, `integrations/reference/remote.<ext>`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** GT-05/07, PE-05; AC-09/11/15

## Task 30: Create or update one PR idempotently

- [ ] Complete

**Description:** Finish requested delivery with an evidence-based PR and recoverable remote identity.

**Acceptance criteria:**
- [ ] Create/update targets recorded repository, branch/base and PR identity; content states final behavior, validation and limitations.
- [ ] Uncertain creation reconciles before retry; updates use the existing PR and do not duplicate it.
- [ ] Dependent changes wait visibly, then rebuild/revalidate against the merged base; request completion distinguishes local/delivered/waiting outcomes.

**Verification:**
- [ ] Exercise a pre-authorized controlled-remote PR flow, response loss, existing PR update and a predecessor merged with a changed SHA/base.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 22, 29

**Files likely touched:** `internal/delivery/pr.<ext>`, `internal/delivery/pr_test.<ext>`, `internal/delivery/dependency.<ext>`, `internal/delivery/dependency_test.<ext>`, `templates/pr.md`

**Estimated scope:** Medium (5 proposed files).

**Coverage:** GT-06–08, DM-03; AC-11/15/18

## Checkpoint after Tasks 28–30: Authorized PR delivery

- [ ] Controlled-remote delivery/recovery and dependency waiting pass AC-09/11/15/18; no merge authority is introduced.
- [ ] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [ ] Evidence and scope reviewed; record material decisions before the next phase.

## Task 31: Promote sanitized observations

- [ ] Complete

**Description:** Make useful run evidence into explicit observations without modifying canonical behavior.

**Acceptance criteria:**
- [ ] Observations link source runs/revisions/events and separate facts, hypotheses, confidence and follow-up.
- [ ] Support candidate/confirmed/resolved/dismissed states with historical decisions preserved.
- [ ] Promotion is deliberate and sanitized; normal successes and unsuccessful runs both remain eligible evidence.

**Verification:**
- [ ] Promote one successful and one failed run, dismiss an unsupported explanation and verify no secret/raw transcript enters source control.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 26, 30

**Files likely touched:** `internal/improvement/observation.<ext>`, `internal/improvement/observation_test.<ext>`, `templates/observation.md`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** IM-01/02/07

## Task 32: Run versioned evaluations

- [ ] Complete

**Description:** Introduce a small explicit eval runner around fixtures already accumulated during development.

**Acceptance criteria:**
- [ ] Eval cases identify sanitized input, expected behavior, environment/capabilities, definition versions, graders and lifecycle state.
- [ ] Deterministic graders record failures; rubric graders record identity/rubric and declared sampling/uncertainty.
- [ ] Coverage includes roles, activation, skills, permissions, allocation and end-to-end execution with isolated result directories.

**Verification:**
- [ ] Run a passing and deliberately failing fixture, unavailable runtime and failing grader; confirm results reproduce configuration and limitations.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 21, 25, 31

**Files likely touched:** `internal/eval/runner.<ext>`, `internal/eval/runner_test.<ext>`, `internal/eval/grader.<ext>`, `evals/suite.yaml`, `templates/eval.md`

**Estimated scope:** Medium (5 proposed files).

**Coverage:** IM-03/04; H-13

## Task 33: Compare baseline and candidate definitions

- [ ] Complete

**Description:** Run controlled comparisons without replacing canonical definitions.

**Acceptance criteria:**
- [ ] Same fixtures and declared controls run against pinned baseline/candidate definitions with isolated artifacts.
- [ ] Report routing, adherence, outcomes and available total effort with uncertainty and task-mix limits.
- [ ] Regression suite includes unnecessary loading, redundant checks/excess review, lost autonomy and ordinary successful cases.

**Verification:**
- [ ] Compare a known-regressing candidate to baseline; prove canonical hashes remain unchanged and regression evidence is visible.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 24, 26, 32

**Files likely touched:** `internal/eval/compare.<ext>`, `internal/eval/compare_test.<ext>`, `templates/comparison.md`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** SK-04/06, IM-03/07

## Checkpoint after Tasks 31–33: Measured improvement

- [ ] Versioned baseline/candidate evaluation detects regression without changing canonical definitions.
- [ ] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [ ] Evidence and scope reviewed; record material decisions before the next phase.

## Task 34: Record human decisions on improvement proposals

- [ ] Complete

**Description:** Complete the observation-to-proposal lifecycle with explicit acceptance and versioned implementation.

**Acceptance criteria:**
- [ ] Proposal links problem, evidence, expected benefit, risks, baseline/candidate/regressions and rollback.
- [ ] Draft/proposed/accepted/rejected/implemented states preserve human decision source; acceptance alone is not implementation.
- [ ] No unattended canonical mutation occurs; implemented status requires the resulting toolkit revision after acceptance.

**Verification:**
- [ ] Exercise accepted and rejected proposals, unauthorized transition attempts and a separately implemented accepted change with linked revision.
- [ ] Run the established focused tests and build; record commands/results.

**Dependencies:** 31, 33

**Files likely touched:** `internal/improvement/proposal.<ext>`, `internal/improvement/proposal_test.<ext>`, `templates/proposal.md`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** IM-05/06; AC-17

## Task 35: Assemble complete v0.1 acceptance evidence

- [ ] Complete

**Description:** Run the PRD release gate across the declared supported environment.

**Acceptance criteria:**
- [ ] Every AC-01–18 has versioned fixture/configuration, requirement links, result and evidence; no unavailable/failing criterion is marked passed.
- [ ] Actual runtime covers end-to-end and enforcement gates; controlled remote covers delivery and interruption recovery.
- [ ] Results include known limitations, supported runtime/platform versions, model mappings, enforcement coverage and missing telemetry.

**Verification:**
- [ ] Run the complete acceptance suite and build/format/test gates from a clean install; inspect failure artifacts as well as success.

**Dependencies:** 27, 30, 34

**Files likely touched:** `evals/release/manifest.yaml`, `docs/acceptance/v0.1.md`, `docs/limitations.md`

**Estimated scope:** Medium (3 proposed files).

**Coverage:** AC-01–18

## Task 36: Document and rehearse daily operation

- [ ] Complete

**Description:** Finish operational instructions and demonstrate readiness for personal routine use.

**Acceptance criteria:**
- [ ] Quickstart and operations cover installation, ordinary prompts, local/PR scope, progress, status/resume, updates and retention.
- [ ] Document recovery/removal of managed integration entries while preserving unrelated settings and retained evidence.
- [ ] A fresh-environment walkthrough and upgrade walkthrough match actual behavior; release readiness records unresolved limitations and human review.

**Verification:**
- [ ] Follow the documentation in an isolated home/state and disposable repository; verify cleanup removes only managed registrations.

**Dependencies:** 35

**Files likely touched:** `README.md`, `docs/quickstart.md`, `docs/operations/recovery.md`, `docs/operations/removal.md`, `docs/acceptance/v0.1.md`

**Estimated scope:** Medium (5 proposed files).

**Coverage:** IN-05, UX-01–04; v0.1 release gate

## Checkpoint after Tasks 34–36: v0.1 readiness

- [ ] All AC-01–18 pass on the declared environment; documentation/limitations and release evidence are ready for human review.
- [ ] Applicable accumulated tests/build pass; no unresolved acceptance failures are hidden.
- [ ] Evidence and scope reviewed; record material decisions before the next phase.
