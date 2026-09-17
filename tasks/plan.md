# Implementation Plan: Agent Kit for actual use

**Status:** Approved to begin implementation on 2026-09-17; Tasks 02–03 specifications complete; Tasks 04–06 foundations complete and independently reviewed. External Codex startup passed; Task 01 enforcement remains inconclusive and managed-host execution remains blocked.

**Date:** 2026-09-17

**Baseline:** [PRD v0.4, target v0.1](../docs/PRD.md)

**Task list target:** [tasks/todo.md](todo.md), authoritative for task completion.

**Confirmed user decisions:** Codex is the first supported coding runtime. The coordinator manages implementation; all sub-agents use `gpt-5.6-terra` with medium reasoning effort.

## Overview

Build a personal local toolkit that takes an ordinary engineering request through an isolated build, meaningful verification, independent review, fixes and a truthful outcome. Reach a usable local workflow first, then add multi-change policy, authorized PR delivery and evidence-backed improvement. Full v0.1 still requires every PRD acceptance criterion; the earlier local milestone is not a reduction of that release contract.

Repository inspection found only `docs/PRD.md`: no executable, runtime adapter, build/test commands, project agent rules or existing task files. These documents establish the initial backlog without replacing an incomplete plan. Repository planning documents live here because this is planning the toolkit itself; Agent Kit's future execution artifacts must remain outside target projects by default.

## Architecture decisions and provisional choices

| Area | Decision or proposal | Decision point |
|---|---|---|
| First runtime | **Confirmed: Codex.** Select a specific supported surface/version; do not assume all Codex surfaces have identical integration controls. | Task 1 |
| Reference platform | Propose macOS first, matching this workspace; prove enforcement on the recorded OS/runtime versions. Additional platforms remain deferred. | Task 1 |
| Runtime integration | Thin CLI and one adapter. Probe native activation/integration and a controlled session interface; choose the smallest path that satisfies activation and role enforcement. | Tasks 1–3 |
| Implementation language | Go, recorded in ADR 001 and implemented in Task 04. | Selected |
| State authority | Propose a small validated state index for current state, immutable/versioned contracts and evidence, Markdown summaries and append-only JSONL telemetry. Define reconciliation before implementation. | Task 2 |
| Storage | Separate source, machine config and external state/worktrees with configurable OS-appropriate roots. No default target-repository installation files. | Task 2 |
| Execution | Sequential changes, one logical builder/source writer per change; orchestrator owns workflow Git and delivery; reviewer cannot modify application source. | PRD invariant |
| Approval | Bind to exact SHA, base and contract version. Retain checkpoint commits as final commits initially; no squash/rebase feature in this backlog. | Tasks 3, 15, 28 |
| Local dependencies | Start a dependent change from an immutable approved predecessor checkpoint; preserve individual change/review identity. | Task 22 |
| Remote dependencies | Propose wait for predecessor merge, then rebuild/revalidate the dependent base. Report waiting explicitly. No automatic stack maintenance. | Task 28 |
| Permissions | Capability grants plus actual host/tool controls. A worktree, prompt or command spelling restriction alone does not prove enforcement. | Tasks 1, 3, 9 |
| Review/compute defaults | Propose standard review, capable/medium compute and three fix cycles with repeated-finding detection; validate and configure separately. | Tasks 3, 23, 25 |
| Skills | Reuse or adapt one suitable coding/debugging skill and add Git/delivery lifecycle procedures; no large library migration. | Tasks 12, 13, 24, 28 |
| Improvement | Capture evidence from the first execution. Candidate evaluation never changes canonical definitions automatically. | Tasks 6, 31–34 |

### Codex documentation baseline

The official App Server documentation describes a bidirectional JSON-RPC interface, version-specific schema generation and approval requests. It also marks experimental surfaces explicitly. This makes it a candidate to investigate, not an already selected or proven production integration. Task 1 must compare the installed version and reproduce the required controls before adoption. [Official Codex App Server documentation](https://learn.chatgpt.com/docs/app-server).

No activation, sandbox, model-selection or telemetry capability is asserted as verified by this planning pass. In particular, document and test shared Git metadata protection, alternate tool/shell routes, reviewer check outputs and network/credential effects. If a necessary boundary cannot be enforced, report the blocker and revisit the integration design; do not silently weaken AC-12 or change the user's runtime choice.

## Intended user flow

```text
setup → doctor → init an existing repository
                     ↓
ordinary prompt → classify intent
  question/explore → read-only answer
  plan             → saved external plan; no implementation
  review           → independent review; no builder
  implement        → contract/preflight → external worktree
                     → builder → checks → checkpoint → reviewer
                         ↑                         │
                         └── bounded fixes ───────┘
                     → exact-revision gate → local outcome
                     → authorized push/PR when requested

interruption → status → reconcile ownership, Git, evidence and remote facts → resume
```

## Task list and milestones

All acceptance checkboxes, dependencies, proposed files and verification procedures are in [todo.md](todo.md). This index groups them without duplicating task status.

| Tasks | Increment | Evidence of readiness |
|---|---|---|
| 01–03 | Codex feasibility, durable state/recovery spec, local execution/permission spec | Supported surface and enforcement approach demonstrated; contracts ready for implementation |
| 04–06 | CLI skeleton, safe state, redacted evidence/provenance | Buildable executable and fault-tested state/evidence before real execution |
| 07–09 | Setup/doctor, external repository init, enforced role operations | Repeatable installation preserves settings; actual permission bypass tests pass |
| 10–12 | Intent activation, change contract, dedicated worktree | Ordinary implementation prompts reach an isolated ready change |
| 13–15 | Builder, meaningful checks, immutable checkpoint | Real implementation produces attributable verified source |
| 16–18 | Independent review, bounded fixes, exact local completion | One real Go request completes without manual role coordination |
| 19–21 | Status/resume, safe update, local acceptance walkthrough | **M1: first usable local workflow** with interruption recovery and recorded limitations |
| 22–24 | Sequential slices, proportional review, selective skills | Multi-concern requests complete as coherent reviewed changes |
| 25–27 | Model/effort policy, useful reports, retention/cleanup | **M2: routine local-use readiness** with operational controls |
| 28–30 | Delivery contract, approved push, idempotent PR lifecycle | **M3: authorized PR workflow**, including uncertain-response recovery and dependent waiting |
| 31–33 | Observations, versioned evals, controlled comparison | Improvement candidates measured against baseline and regressions |
| 34–36 | Human proposal decisions, full acceptance, operating guide | **M4: complete v0.1**; all AC-01–18 evidenced on the supported environment |

Task 18 is an end-to-end development demonstration. Task 21 is the first use milestone because a successful happy path alone is insufficient for recoverable daily work. Tasks 22–36 remain required for the full PRD release even if local use begins earlier.

## Dependency graph and execution order

```text
01 Codex feasibility
 └─02 state/recovery contract
    └─03 execution/permission contract
       └─04–06 executable + state + evidence
          └─07–09 setup + init + enforced operations
             └─10–12 routing + contract + isolation
                └─13–15 builder + checks + checkpoint
                   └─16–18 review + fixes + local completion
                      └─19–21 recovery + update + M1 acceptance
                         └─22–24 slices + review policy + skills
                            └─25–27 allocation + reports + retention
                               └─28–30 delivery → M3
                                  └─31–34 improvement
                                     └─35–36 release gate → M4
```

The backlog intentionally follows a conservative single-implementer order; explicit per-task dependencies identify actual prerequisites. Each implementation task includes its relevant tests, and every three tasks have an evidence checkpoint. These checkpoints are validation boundaries, not a requirement to produce one giant PR per three tasks. Group implementation changes by coherent behavior and independently review each change.

Once interfaces are stable, fixture authoring, documentation and labeled policy cases can be worked on independently in separate sessions. Shared persistence, permissions, orchestration and Git lifecycle changes remain sequential. No parallel agent execution is required by this plan or by v0.1.

## Verification strategy and definition of done

Task 04 established `make check`, `make build`, `make test` and `make focused-test`; the focused target covers the CLI. Later packages use scoped `go test` commands with the repository cache settings in Makefile. Record actual executed checks per task. Specification tasks are verified through contract walkthroughs and mapped failure cases.

- Per task: meet its three acceptance criteria, run focused behavior/error-path checks and applicable build/format checks, and attach actual evidence. Documentation-only tasks use their stated walkthrough.
- Per implemented change: verify the full required contract, obtain independent review of the immutable revision, resolve blocking findings, update public behavior documentation and preserve scope/provenance.
- Per checkpoint: run the accumulated relevant suite and the stated integration scenario. Failures remain visible; unavailable required checks are not passes.
- Before first local use: use the actual Codex runtime against disposable Go repositories; prove permission enforcement, dirty-checkout preservation, seeded-defect correction and interruption recovery.
- Before delivery support: use a controlled test remote and scoped credentials; verify wrong-remote/stale-approval rejection, local-only non-mutation and recovery after uncertain remote effects.
- Before v0.1: run all AC-01–18 with fixture/configuration/version links, retained artifacts and a limitations matrix. Tests with runtime doubles complement actual-runtime evidence.

The user authorized coordinator-managed implementation on 2026-09-17. Future execution should proceed within the user's authorized scope without asking again at every routine checkpoint; material objective or authority changes require resolution. This plan grants no merge, deployment or publication authority.

## PRD acceptance coverage

| Acceptance | Primary tasks and evidence |
|---|---|
| AC-01 setup/update | 07, 20, 21, 35 |
| AC-02 activation/intents | 01, 10, 21, 35 |
| AC-03 external init | 08, 21, 35 |
| AC-04 decomposition/replan | 11, 22, 35 |
| AC-05 real Go local change | 13–18, 21, 35 |
| AC-06 blocking finding/fix | 16–17, 21, 35 |
| AC-07 review rigor | 23, 35 |
| AC-08 verification/loop gates | 14, 17, 21, 35 |
| AC-09 exact approval | 15, 18, 29, 35; rewrite variant intentionally excluded |
| AC-10 dirty checkout/one writer | 12, 19, 21, 35 |
| AC-11 authorized delivery | 28–30, 35 |
| AC-12 enforced role boundaries | 01, 09, 16, 21, 35 |
| AC-13 relevant skills | 24, 33, 35 |
| AC-14 allocation honesty | 25, 35 |
| AC-15 interruption recovery | 05, 19, 29–30, 35 |
| AC-16 evidence/secrets/provenance | 06, 18, 21, 26, 35 |
| AC-17 improvement lifecycle | 31–34, 35 |
| AC-18 aggregate/local/remote completion | 22–23, 28–30, 35 |

## Hypothesis resolution

| PRD hypothesis | Resolution task |
|---|---|
| H-01 runtime/platform | Codex selected by user; surface/platform/enforcement proved in 01 |
| H-02 language/thin adapter | 01, interface contract in 03 |
| H-03 storage roots | 02, installed in 07 |
| H-04 persistence/index | 02, failure-tested in 05–06 |
| H-05 dependent delivery | Local rule in 22; remote decision in 28 and test in 30 |
| H-06 final commits | 03, 15, 28; retain checkpoints initially |
| H-07 compute defaults | 25 |
| H-08 risk checklist | 23 |
| H-09 loop budget | 03, 17 |
| H-10 permission profiles | 03, 09 |
| H-11 retention | Decide defaults in 02 before use; guarded cleanup in 27 |
| H-12 context budgets | Measure in 24; regression comparison in 33 |
| H-13 eval runner | 32–33 |
| H-14 reporting/usage | Probe in 01; minimum evidence in 06; report in 26 |

## Risks and mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Codex surface cannot mediate required role effects | High; blocks supported path | Probe real bypass attempts first; separate enforced/advisory/unsupported capabilities; do not claim AC-12 from role prompts |
| Activation works only through an explicit command | High; AC-02/IN-04 gap | Test ordinary first/subsequent prompts and child sessions in Task 1; document the precise supported entry point |
| Toolkit becomes a general agent platform | High scope growth | One adapter/platform, thin deterministic lifecycle operations, no generic workflow language or concurrent builders |
| Interrupted state disagrees with Git/remote reality | High data/approval risk | Persist operation intent, atomic state, exclusive ownership, retained refs and fault-injection recovery tests |
| Shell checks escape allowed effects | High permission risk | Validate real host controls, shared Git metadata and network/credential scope; deny unmediated protected actions |
| Evidence leaks secrets or grows indefinitely | High | Minimize/redact before persistence, seeded-secret tests, defined retention before use and guarded cleanup |
| Native runtime behavior changes | Medium/high | Pin tested versions, capture schema/capabilities and active-run definitions; compatibility-aware update and doctor |
| Agent evals fluctuate or overfit | Medium | Separate deterministic checks from rubrics; declare sampling, controls and uncertainty; include routine successes |
| Tasks exceed the estimated file/session budget | Medium | Proposed paths are estimates; split in place before implementation while preserving criteria/dependencies and history |

## Open decisions and next action

The user has settled the runtime. Task 1 still needs to decide the supported Codex surface/version, reference platform and implementation language from concrete probes. Task 2 decides storage authority/roots and retention; Task 28 selects the remote provider and confirms the dependent delivery policy. These are scheduled decisions, not reasons to request every preference upfront.

The next unblocking action is to rerun Task 01 in an environment that permits the documented isolated Codex probes under its own authorized controls. The current host denies nested client initialization and offers no escalation. Do not bypass its restrictions or enable runtime-dependent features before the feasibility evidence supports the candidate design. Runtime-independent foundations may proceed under the continuation decision below.

## Execution update: 2026-09-17

Task 01 found that nested Codex execution fails during client initialization under the current host sandbox, before any model call. Runtime activation and AC-12 remain unproven. The candidate runtime/language decision is recorded in [ADR 001](../docs/decisions/001-reference-runtime.md); it is not a support claim.

The coordinator permits Tasks 02–03 platform-neutral contract drafting against this explicitly provisional decision while runtime evidence is blocked. This refines the original strict dependency: Task 01's interface/decision evidence is sufficient to draft contracts, but its actual-runtime readiness gate remains required for execution integration. No task or acceptance criterion is waived. Tasks 02 and 03 are drafted in parallel with explicit reconciliation of shared state and ownership semantics. Independent review must check the combined contracts before source implementation.

### Reviewed outcome

Tasks 02 and 03 are complete as specifications after independent Terra/medium review and corrections. They define state/artifact/recovery contracts and the local workflow/adapter/permission/fixture contracts. They are not implemented runtime behavior. Task 01 remains incomplete: its reviewed probe report records a candidate Codex CLI/macOS/Go target but no successful activation or enforcement fixture. The first checkpoint remains open; Tasks 04–36 have not started.

Validation performed: local Markdown links, balanced code fences, and whitespace checks on all added documents. No application build or runtime acceptance suite exists yet. All sub-agents in this implementation session used `gpt-5.6-terra` with medium reasoning effort.

## Continuation decision: 2026-09-17

The user requested continued implementation while resolving Task 01. The coordinator narrows the former blanket dependency: Tasks 04–06 (CLI skeleton, local persistence, synthetic evidence) may proceed against reviewed contracts and the candidate Go decision, with meaningful deterministic tests. Runtime-dependent setup, activation, assignment and enforcement remain unavailable until Task 01 actual-runtime probes pass. The first feasibility checkpoint remains open; completing local foundations does not satisfy it.

Task 04 is assigned to a Terra/medium Go implementer. A separate Terra/medium worker prepares a bounded host-initialization check for the user to run in normal Terminal. This is an environment diagnostic only; a successful startup does not prove automatic activation or AC-12.

### Foundation progress

Task 04 is implemented and independently reviewed: Cobra CLI grammar, version/build identity, explicit unavailable lifecycle commands and reproducible build/check commands. Task 05 persistence is in progress. The external diagnostic is available as [check-runtime.sh](../experiments/check-runtime.sh); only sanitized startup status is requested from the user.

### External startup evidence

On 2026-09-17 the user ran the reviewed startup diagnostic from normal Terminal. The coordinator read its sanitized artifact and confirmed `host_startup_check=started`, `process_exit=0`, `diagnostic=ready_response`, requested `gpt-5.6-terra` / `medium`. This establishes actual session startup outside the managed host, not automatic Agent Kit activation, role enforcement, actual model/effort reporting or a bypass of the current host restrictions. Task 01 remains partially complete pending the separate probes.

The subsequent user-run reviewer probe completed with the sentinel unchanged, but no shell/Python/cache command events were recognized. This is inconclusive: neither tool execution nor sandbox denial was established. The next probe change must diagnose event shape/tool activity before requesting another external run.

### Feasibility and activation evidence ordering

Task 01 establishes native runtime feasibility using disposable probes. Agent Kit automatic activation cannot be accepted until Tasks 07 and 10 implement registration and routing. Keep AC-02 and AC-12 unpassed until the real integration fixtures demonstrate them. Runtime-independent installation bookkeeping, diagnostic reporting and policy evaluation may be developed with explicit unavailable results; no such result establishes a supported registration or enforced boundary. This removes the circular implementation prerequisite without lowering the acceptance requirements.
