# Agent Kit — Product Requirements Document

**Status:** Consolidated draft, ready for technical specification  
**Document version:** 0.4  
**Target release:** v0.1  
**Date:** 2026-09-16  
**Owner / primary user:** Personal software engineering practitioner  
**Source:** [Create Agent Repository Structure](chatgpt-conversation://6aaa7c6f-ff44-83ed-88e7-38cb9a5531a7)

This document replaces the fragmented PRD drafts and additions in the source discussion. It is the product baseline for writing technical specifications and implementation tasks. It does not claim that any runtime integration or enforcement mechanism has already been implemented or verified.

## 1. Reading this document

**Requirements** describe required v0.1 behavior and use stable identifiers. **MUST** denotes a release requirement; **SHOULD** denotes the expected default, with any exception recorded and explained. **MAY** denotes an optional capability that cannot be relied on for v0.1 acceptance.

**Hypotheses** in section 19 are proposed implementation choices to validate, not settled requirements. **Future ideas** in section 20 are outside v0.1. Examples of paths, field names, events, and commands beyond the four named lifecycle commands are illustrative unless explicitly required.

The consolidation resolves earlier alternatives as follows: a worktree belongs to a **change**, not a run or agent; v0.1 executes changes sequentially with one builder per change; Markdown replaces prose-heavy YAML artifacts; permissions are capabilities rather than a shell-command list; every implemented change receives risk-based review, while individual tasks receive progressive verification.

## 2. Vision, user needs, and principles

Agent Kit is a personal, local, version-controlled toolkit that turns ordinary engineering requests into small, verified, independently reviewed changes. After setup, the user should be able to ask for work in a supported coding runtime and receive a clear outcome without manually coordinating agents.

The primary user needs to implement and maintain software, particularly Go software; keep changes digestible; preserve unfinished work; avoid repeated operational approvals; and learn which instructions, skills, models, and workflows actually help. Roles remain language-agnostic. Language and ecosystem knowledge belongs in composable skills.

Success means reliably completing an authorized request with reasonable total effort, a clean evidence trail, and few unnecessary interruptions. Saving tokens matters, but cannot justify bypassing correctness or review gates.

Product principles:

1. **Small coherent changes.** Decompose by purpose and behavior, not arbitrary line counts.
2. **Changes own execution.** A change owns its branch, worktree, revisions, evidence, and delivery state; agents are assigned workers.
3. **One builder per change.** The same logical builder handles its tasks and fixes. Replacement after interruption is a documented handoff, not concurrent implementation.
4. **Verify frequently; review at meaningful boundaries.** Use cheap relevant checks during implementation and independent reasoning at the change boundary.
5. **Risk determines rigor.** A one-line authorization change can require more review than a large mechanical edit.
6. **Load context when needed.** Availability does not imply loading or use.
7. **Maximum useful autonomy within explicit scope.** Routine engineering proceeds autonomously; protected actions require applicable authorization.
8. **Human-readable evidence.** Use Markdown for communication, small metadata for coordination, and structured events for measurement.
9. **Evidence precedes improvement.** Observe, evaluate, propose, validate, and obtain human acceptance before changing canonical behavior.
10. **Honest outcomes.** Distinguish implemented, verified, reviewed, committed, pushed, PR-created, blocked, and incomplete.

## 3. Goals and non-goals

### Goals for v0.1

- Automatically route implementation requests into the appropriate workflow and keep questions and planning lightweight.
- Turn a request into explicit acceptance criteria, ordered changes, and implementation tasks.
- Execute a complete build → verify → review → fix → re-review loop.
- Isolate work from the user's checkout and preserve durable execution history externally.
- Create local commits and optionally push and create/update PRs when authorized.
- Select relevant skills and allocate model capability and reasoning effort intentionally.
- Capture provenance, skill behavior, permission friction, and outcomes from the first real run.
- Support reproducible evals and human-approved improvement proposals.
- Provide usable setup, diagnostics, initialization, and update operations.

### Non-goals for v0.1

- Company-specific policy, compliance rules, CDK conventions, or organization-specific infrastructure knowledge.
- Multiple concurrent builders, task-level branch merging, distributed execution, or multi-repository orchestration.
- Fully automated stacked-PR maintenance, autonomous PR merging, deployment, releases, or package publishing.
- Automatic self-modification or automatic acceptance of proposals.
- A general workflow language, graphical dashboard, vector memory, or reinforcement-learning system.
- A complete catalog of language or personal-life skills.
- Universal runtime compatibility, model availability, or security enforcement from prompt instructions alone.

The architecture SHOULD permit additional skills and integrations without duplicating the core role definitions. v0.1 MUST prove one complete supported runtime path rather than offering several incomplete adapters.

## 4. Domain model and relationships

**DM-01 — Identity and lineage.** Every durable entity MUST have a stable identifier, schema version, and links to its owning or source entities. Historical records MUST retain the contract and revision they evaluated; later edits must not silently change their meaning.

| Entity | Meaning and minimum information |
|---|---|
| Request | Original user intent, clarifications, constraints, acceptance criteria, and authorized delivery scope. Preserve the original request alongside subsequent interpretation. |
| Plan | Versioned strategy for a request: changes, dependency order, acceptance-criterion mapping, assumptions, risks, and completion checks. |
| Change / PR slice | One coherent review and delivery unit: objective, scope, exclusions, criteria, tasks, dependencies, base revision, branch/worktree, risk, review mode, validation contract, and current state. A remote PR is optional. |
| Task | A bounded implementation step inside one change, with completion evidence and status. It is not a mandatory review or commit boundary. |
| Revision | Immutable Git checkpoint of a change attempt: SHA, base SHA, parent/previous revision, contract version, and linked verification evidence. |
| Review | Independent assessment of one exact revision against one contract version, with mode, reviewer execution metadata, findings, and verdict. |
| Finding | Stable issue ID, severity, category, evidence/location, impact, required correction if blocking, and resolution history. |
| Run | Durable execution of a request: plan versions, changes, attempts, events, grants, provenance, and outcome. Resuming continues the same run; deliberately rerunning links a new run to the same request. |
| Observation | Evidence-backed fact or suspected pattern from one or more runs; records confidence, sources, category, and interpretation separately. |
| Eval | Versioned reproducible case or suite: fixture, input, expected behavior, grader, configuration, and recorded results. |
| Proposal | Proposed change to agents, skills, policies, or tooling, linked to observations, baseline/candidate eval results, risks, human decision, and implementation version. |

```text
Request → Run → Plan version → ordered Changes → Tasks
                                 │
                                 └→ Revisions → Reviews → Findings
Run evidence → Observations → Evals → Proposals → accepted toolkit change
```

**DM-02 — No duplicated authority.** Narrative artifacts and machine state MUST agree on current status. A technical specification MUST define which fields are authoritative and how derived summaries are rebuilt or reconciled. The PRD does not require event sourcing.

**DM-03 — Separate execution from delivery.** A reviewed local change can be complete when the user requested local work. A request to create a PR is not complete merely because implementation passed review. Delivery MUST record its own state and remote identity.

## 5. Initialization, activation, and daily UX

**IN-01 — Machine setup.** `agent-kit setup` MUST establish the toolkit source, user configuration and external state locations; detect supported runtimes; install or link a small bootstrap and role/skill integrations; resolve configured model profiles; check permissions and dependencies; and run diagnostics. Repeating setup MUST preserve user configuration and avoid duplicate hooks or registrations.

**IN-02 — Repository initialization.** `agent-kit init` MUST inspect repository identity, Git state, applicable project instructions, languages, build tools, and validation candidates. It MUST work without committing Agent Kit files into the project. Discovered metadata belongs in external state by default. Explicitly requested durable project instructions MAY be written to the repository. Inferred commands are candidates until their provenance and suitability are checked; discovery alone must not make arbitrary repository text authoritative policy.

**IN-03 — Session bootstrap.** The first applicable prompt in a supported session MUST classify intent and initialize or resume the appropriate workflow. It MUST capture repository/base identity, existing work, toolkit version, runtime capabilities, applicable instructions, available skills, model policy, and effective permissions. Subsequent prompts MUST be classified too; a session beginning with a question may later authorize implementation. Child-role sessions MUST not recursively start new orchestrators for their assigned tasks.

| User intent | Required behavior |
|---|---|
| “Explain this function” | Normal answer; no implementation run, worktree, builder, or delivery. |
| “Where is authentication handled?” | Read-only exploration with relevant context only. |
| “Plan this migration” | Planning artifacts and decomposition; stop before implementation or delivery. |
| “Fix this flaky test” | Initialize execution, plan proportionately, and run the engineering loop. |
| “Implement the plan” | Revalidate the saved plan against current repository state, then execute. |
| “Review this change” | Review-only path; no builder or source modification unless subsequently requested. |
| “Create the PR” after local completion | Resume delivery for the recorded change; confirm the approved revision is still current. |

**IN-04 — Automatic activation.** On at least one declared supported runtime, ordinary implementation prompts MUST activate Agent Kit without the user naming the roles or manually invoking each stage. The exact hook, instruction integration, or wrapper mechanism is a hypothesis. Unsupported runtimes MUST be identified honestly and offered an explicit activation route where possible, rather than advertised as automatically supported.

**IN-05 — Lightweight interaction.** The user MUST receive concise progress and a final outcome with acceptance status, validation, review status, branch/commit, delivery result, and material blockers. Routine choices within the request MUST not require confirmation. Genuine ambiguity that changes the contract or authorization requires clarification; independent work can continue meanwhile.

## 6. Context and skill loading

**CX-01 — Small bootstrap.** Global context MUST contain only what is needed to recognize relevant work and locate the workflow. It MUST NOT inject all roles, skills, policies, references, or historical runs into every session.

**CX-02 — Stage-specific context.** The orchestrator loads planning and workflow policy; the builder receives the change contract, relevant project rules, and selected skills; the reviewer receives a curated review package. Each can retrieve additional context when needed. Lazy loading must not omit applicable mandatory project instructions.

**CX-03 — Context states.** Distinguish available, considered, selected, loaded, and observably used components. Record selected/loaded definitions and reasons for relevant routing decisions. Unobservable use MUST be labeled unknown, not inferred merely from loading. Catalog-level availability can be recorded once per version rather than emitting an event for every installed skill.

**CX-04 — Independence.** Reviewer context MUST exclude the builder's private reasoning and persuasive self-assessment. Factual summaries, contracts, diffs, test evidence, and prior findings are appropriate. Reviewers can inspect surrounding source rather than relying solely on the builder's summary.

## 7. Roles and boundaries

| Role | Required responsibility | Boundary |
|---|---|---|
| Orchestrator | Interpret intent; establish contracts; classify risk/complexity; decompose; select skills and compute; manage worktrees, state, permissions, verification gates, review loops, Git, delivery, and completion. | SHOULD NOT implement application code. Cannot bypass host restrictions or claim a failed gate passed. |
| Builder | Implement the assigned change's tasks, verify progressively, report evidence and deviations, and fix findings in the same change worktree. | MUST NOT silently broaden scope, own branch/commit/delivery decisions, or delegate concurrent builders. |
| Reviewer | Independently assess an exact revision, identify actionable findings, and perform focused re-review. | MUST NOT fix application source, push, create PRs, or change workflow branches. |

**RO-01 — Single builder ownership.** v0.1 MUST execute changes sequentially, with one logical builder per change and one source writer at a time. A stopped builder can be replaced with an explicit handoff preserving contract, state, findings, and provenance.

**RO-02 — Source stability.** Builders MUST be paused during checkpoint creation and review. Reviewer-run checks may write declared temporary outputs or caches, but MUST NOT modify the reviewed source. Checks that regenerate tracked files require controlled handling and a new revision if content changes.

**RO-03 — Skills rather than extra roles.** Git workflow and delivery MUST initially be orchestrator-owned procedures/tools or skills. Deterministic verification does not require a separate reasoning agent. A specialized reviewer fleet is not required.

**RO-04 — Scope escalation.** On discovering material extra work, the builder MUST report the new scope, cause, acceptance impact, and options. The orchestrator decides whether to adjust the plan, split the change, or ask about a changed objective. Unrelated improvements become follow-ups rather than automatic additions.

## 8. Planning and small-change decomposition

**PL-01 — Contract before implementation.** Every change MUST define objective, included/excluded scope, constraints, acceptance criteria with stable IDs, dependencies, likely affected areas, required validation, risk/review mode, and completion conditions. Small requests may have one short plan and change contract; document volume must be proportional.

**PL-02 — Reviewable slices.** Each change SHOULD have one primary purpose, include necessary tests/documentation, leave a valid repository state, and be independently mergeable where practical. A dependent change MUST declare its dependency; “small” must not mean an unusable fragment missing its tests.

**PL-03 — Dependency readiness.** A plan MUST have an acyclic dependency order and a recorded base/readiness rule. A dependent change MUST NOT start against an unfinished or changing predecessor. The initial delivery policy is resolved in section 19; automated stack maintenance is not required.

**PL-04 — Size feedback.** Record estimated and actual files, lines, packages/modules, task count, and concerns where practical. Material growth MUST trigger reassessment. There is no universal line limit; risk and conceptual coherence remain decisive. Large generated or mechanical changes need an explained review approach.

**PL-05 — Coverage and replanning.** Every request acceptance criterion MUST map to changes and completion evidence. Plan revisions MUST preserve prior scope and record why work was added, split, deferred, or removed. Replanning cannot silently redefine success to exclude unfinished requested work.

Example decomposition for retry support: introduce the reusable behavior and tests; migrate one client with tests; migrate another client with tests; remove obsolete behavior after its consumers are gone. Each slice is reviewed once as a whole, not after each task.

## 9. Verification, review, and loop control

### Verification

**VE-01 — Progressive verification.** Builders MUST run the cheapest meaningful checks after logical steps, then all contract-required checks before review. Depending on the project these include tests, compilation, formatting, linting, static analysis, type checking, and generated-output checks. “Full validation” means the complete required set for the change, not every conceivable test.

**VE-02 — Evidence.** Each check MUST record the command/check identity, working context, relevant revision or pre-checkpoint content identity, environment/tool versions where available, result, exit status, timing, and output reference. Skipped, unavailable, timed-out, flaky, and failed checks MUST remain distinguishable from passed checks. A pre-checkpoint result can attach to a checkpoint only after unchanged source content is confirmed.

**VE-03 — Gate.** Known failing or unavailable required checks MUST prevent ordinary approved completion and delivery. Pre-existing failures must be documented and distinguished from regressions. Any exception requires an explicit authorized contract/policy decision and remains visible as a limitation; it cannot be silently converted to a pass.

**VE-04 — Determinism.** Repeatable automated evidence is preferred. Nondeterministic tests require a declared repeat/seed strategy and honest results. An LLM's statement that tests “should pass” is not verification. Tests must check relevant behavior rather than mirror implementation mechanically.

### Risk-based review

**RE-01 — Review boundary.** Every implemented change MUST receive independent review after required verification. Individual tasks do not require full review. The default workflow has no `review: none` fast path.

| Mode | Intended use | Minimum review scope |
|---|---|---|
| Light | Demonstrably low-risk changes | Acceptance, diff sanity, obvious correctness, scope, and appropriateness of verification. |
| Standard | Ordinary implementation and refactoring | Light scope plus logic, regressions, tests, maintainability, and affected interfaces. |
| Thorough | High-consequence or hard-to-reason-about behavior | Standard scope plus relevant failure modes, security boundaries, edge cases, concurrency, persistence, compatibility, and wider regression surface. |

**RE-02 — Classification.** The orchestrator MUST record risk signals and selected mode. Authentication, authorization, secrets/cryptography, concurrency, persistence/schema changes, financial behavior, and public compatibility boundaries require explicit consideration for thorough review. Size alone cannot justify reduced review. Classification combines inspectable signals with explained judgment and can escalate when new facts emerge.

**RE-03 — Review package.** Provide the versioned contract and criteria; risk and mode; exact base and revision SHA; final diff; relevant project instructions; validation evidence; and known constraints. Reviewers MUST be able to retrieve necessary surrounding code.

**RE-04 — Findings and verdict.** Findings MUST be specific, evidence-backed, and tied to the contract or material regression risk. Record stable ID, category, severity, location where applicable, impact, and required correction. Default severities are critical/major (blocking) and minor/suggestion (non-blocking). Verdicts MUST distinguish approved, changes requested, and blocked/inconclusive. Approval requires satisfied criteria and no unresolved blocking findings. Stylistic preferences alone must not cause endless fix loops.

**RE-05 — Focused re-review.** A fix produces fresh verification and a new immutable revision. Re-review receives previous findings, fix mapping, the delta from the previous revision, the resulting revision, and updated evidence. It MUST confirm resolution, inspect introduced regressions, and reassess contract compliance. Changed scope, base, contract, or newly discovered risk can require broader review. Finding statuses distinguish open, addressed-pending-review, resolved, and rejected-with-rationale; a builder cannot mark its own blocking finding resolved.

**RE-06 — Exact approval.** Approval applies only to a recorded revision and contract. Any later content, base, or contract change invalidates its applicability. Even a commit rewrite with identical files requires explicit approval reconciliation under the Git lifecycle below; approval must never silently move to a new SHA.

**RE-07 — Bounded loops.** Configure a finite fix/re-review attempt budget and detect repeated unresolved findings. On exhaustion, stop the automatic loop and reassess diagnosis, scope, tools, or model allocation. Record a replan or blocker; do not retry indefinitely or lower the acceptance bar to finish.

**RE-08 — Request completion.** After slices pass, the orchestrator MUST check aggregate acceptance coverage, dependency integration, required final validation, review applicability, and remaining blockers. A second exhaustive review of all prior diffs is not the default. Cross-slice interactions can justify targeted integration review.

## 10. Git isolation and delivery lifecycle

**GT-01 — Worktree per change.** Each executing change MUST have one dedicated branch and worktree outside the user's project checkout. The orchestrator owns creation and mutation of Git workflow state. Record the starting repository, branch/base SHA, selected remote, and pre-existing work. Never silently stash, discard, stage, or commit the user's unrelated work.

**GT-02 — Explicit base.** Do not guess whether the request includes uncommitted user changes. When required input exists only in the user's checkout, record and resolve how it enters the isolated worktree before depending on it. Otherwise build from the recorded committed base. Shared Git metadata is not isolated by a worktree, so destructive repository-wide operations require separate protection.

**GT-03 — Checkpoints and revisions.** After implementation and validation, the orchestrator stages only attributable files and creates an immutable local checkpoint for review. Fixes create subsequent checkpoints. These are evidence-bearing revisions, not necessarily the final public commit structure.

**GT-04 — Final commits.** Deliver coherent commits rather than requiring one public commit per agent attempt. The simplest v0.1 path retains an approved checkpoint as the final commit. Optional squash/reword/rebase requires separately permitted local history editing. A resulting new SHA MUST receive an explicit final approval record: content/base changes require fresh relevant verification and review; a demonstrably content-equivalent rewrite may use a focused equivalence review with recorded old/new SHAs, trees, and base. Verification affected by commit identity must be rerun. Shared or published history must not be rewritten under this local exception.

```text
contract → worktree → builder + progressive checks → required validation
   → checkpoint → review → fixes/checks/new checkpoint → focused re-review
   → final commit preparation → exact final-revision gate → optional push/PR
```

**GT-05 — Delivery gate.** Before a push or PR mutation, confirm authorization, intended repository/remote/base, attributable contents, required checks, current revision approval, and absence of unresolved blockers. Local completion requires no push. Creating a PR does not authorize merging it.

**GT-06 — PR content.** PR title and description MUST reflect the final scope, actual behavior, relevant validation, and limitations using recorded evidence. They must not claim human review or CI success merely because an agent approved the change.

**GT-07 — Recovery and idempotency.** Persist delivery attempts and confirmed remote identifiers. After interruption or uncertain network response, reconcile remote state before retrying; do not duplicate PRs or repeat an already completed mutation blindly. An existing PR update MUST target the recorded PR and intended branch.

**GT-08 — Dependency delivery.** Record each change's base, predecessor, local readiness, and remote delivery state separately. Dependent delivery must use the declared policy; do not silently collapse all slices into one large PR. If delivery must wait for a predecessor to merge, report that dependency as waiting rather than completed or failed. No automated stack rebasing is required.

**GT-09 — Cleanup.** Do not delete an active worktree, dirty/unrecorded work, or the only reachable copy of review evidence. Completed worktree cleanup MAY be explicit or policy-controlled once required revisions and artifacts remain recoverable. Retention defaults remain a hypothesis.

## 11. Permissions and useful autonomy

**PE-01 — Capability model.** Agent Kit MUST express capabilities with `allow`, `ask`, or `deny` policies and operation/resource scope. At minimum it must distinguish:

| Area | Capability vocabulary |
|---|---|
| Repository/filesystem | `repository.read`, `repository.write`, `filesystem.project`, `filesystem.worktree`, `filesystem.external` |
| Execution | `process.execute`, `process.test`, `process.build` |
| Network | `network.read`, `network.write` |
| Git | `git.read`, `git.branch`, `git.commit`, `git.history_write`, `git.push` |
| Delivery | `delivery.pr_create`, `delivery.pr_update`, `delivery.pr_merge` |
| Coordination | `agent.delegate`, `state.read`, `state.write` |
| Credentials | `credentials.use`; distinguish secret-value access (`credentials.read`) and do not grant it by default |

The vocabulary is a product contract; serialization and granular sub-capabilities are implementation choices. A broad execution grant MUST NOT implicitly authorize every side effect a shell command can perform.

**PE-02 — Effective authorization.** Runtime/system limits are the ceiling. Within that ceiling, resolve user policy and explicit run instructions, then role and resource restrictions. A skill's declared requirements are checked against those grants; skill requirements are not grants themselves. Check all capabilities required by an operation, including its filesystem and external effects. A denied capability is not recoverable by changing the command spelling or tool.

**PE-03 — Development defaults.** The default profile MUST allow routine reading, editing within the assigned worktree, formatting, tests, builds, linting, and relevant local analysis without repeated approval where the runtime supports it. Scope cache/temp writes explicitly. Network reads needed for documentation and dependency resolution SHOULD be configurable once rather than approved repeatedly.

| Action | Orchestrator | Builder | Reviewer |
|---|---|---|---|
| Read relevant repository and Git data | Allow | Allow | Allow |
| Modify application source | Delegate to builder by default | Allow within change | Deny |
| Execute scoped checks | Allow | Allow | Allow with source stability |
| Manage branches/worktrees and local commits | Allow within run | Deny | Deny |
| Rewrite history | Deny by default; explicit scoped exception for unpublished checkpoints | Deny | Deny |
| Push / create or update PR | Ask unless already authorized | Deny | Deny |
| Merge / deploy / release / publish | Deny in v0.1 workflow | Deny | Deny |
| Delegate workflow roles | Allow | Deny by default | Deny by default |
| Write run state | Coordinate all state | Assigned build/task reports | Assigned review/finding reports |

**PE-04 — Preflight.** Before assignment, compare the task and selected skills' required capabilities with effective permissions and actual runtime support. Resolve foreseeable validation blockers early. Missing optional delivery rights need not block authorized local work. Distinguish a policy denial from an unavailable runtime feature or missing tool.

**PE-05 — Existing authorization.** “Implement and create a PR” authorizes the necessary scoped push and PR creation, subject to host limits and explicit policy restrictions; do not ask again merely because defaults say `ask`. “Implement locally; do not push” forbids remote mutation. A grant has an action, resource, lifetime, and source. It cannot silently expand to force-push, merge, deploy, or unrelated repositories.

**PE-06 — Protected operations.** Deleting external data, modifying external infrastructure, rewriting shared history, force-pushing, deleting remote branches, destructive database actions, and accessing secret values MUST remain separately controlled. A permitted test command is not inherently safe: repository scripts and dependency hooks can have external effects. Supported integrations must enforce restrictions through appropriate tool/runtime controls or clearly expose their inability to do so.

**PE-07 — Enforcement honesty.** Worktrees and role prompts are not security sandboxes. The adapter MUST declare which boundaries are enforced, advisory, or unsupported. It MUST fail closed for protected operations it cannot mediate and MUST NOT advertise read-only or network isolation based solely on instructions. The v0.1 selected runtime must demonstrate the permission acceptance tests in section 18.

**PE-08 — Escalation.** On a missing permission, stop repeated attempts, record the required capability and reason, and route the issue to the orchestrator. Adapt within existing authority, ask for a narrow grant when appropriate, or report a blocker. A declined grant must not produce repeated equivalent prompts.

**PE-09 — Credentials and logs.** Prefer credential helpers and opaque credential use. Secrets MUST NOT be persisted in artifacts or telemetry. Capture safe metadata and redacted output; avoid collecting sensitive environment contents or full transcripts by default. Configuration and diagnostics must not reveal secret values.

**PE-10 — Profiles and friction.** Support reusable user-level permission profiles independent of agent definitions. Record grants, denials, blocked routine operations, and repeated/unnecessary prompts. Exact profile names are hypotheses. Routine successful checks may be aggregated.

## 12. Model capability and reasoning effort

**MO-01 — Separate axes.** Role identifies responsibility; model profile identifies capability; reasoning effort identifies requested reasoning allocation. Agent definitions MUST avoid hardcoded vendor/model names where practical.

**MO-02 — Portable policy.** Support abstract capability profiles such as `fast`, `capable`, and `deep`, and conceptual effort levels such as low/medium/high. Adapters MUST resolve supported concrete settings and record both requested and actual settings. Profiles can map to the same model when only one is available. Unsupported effort must be reported as unsupported, not claimed as applied.

**MO-03 — Allocation.** The orchestrator MUST choose and explain allocation by stage, complexity, uncertainty, and risk. Complexity affects implementation reasoning; consequence and risk affect review rigor. Review mode and model choice remain independent: stronger models do not eliminate required review coverage.

**MO-04 — Escalation.** Unexpected architecture, repeated ineffective attempts, unresolved findings, or newly discovered security/concurrency concerns MUST trigger reassessment. Options include more effort, a stronger available model, better context, replanning, or clarification. Do not simply retry indefinitely at the same inadequate allocation. Escalation remains bounded by configured availability, policy, and budget.

**MO-05 — De-escalation and accounting.** Lower allocation MAY be used for demonstrably straightforward tasks or focused fixes, with a reason. Record attempts, escalations, elapsed time, and available usage/cost metrics. Optimize total cost of a successful reviewed change, including failed attempts and rework, rather than cost per invocation.

Initial defaults and thresholds are hypotheses; v0.1 does not require a learned model router or provider price engine.

## 13. Storage, artifacts, and recovery

**ST-01 — Three storage classes.** Keep version-controlled toolkit definitions separate from machine configuration and runtime evidence. Keep project-owned instructions and architecture documentation in the project. Execution plans, attempts, reviews, temporary state, and telemetry MUST live outside target project working trees by default.

Illustrative layout, with configurable OS-appropriate roots:

```text
toolkit-source/                    # version controlled
  agents/                          # orchestrator, builder, reviewer
  skills/                          # definitions, references, optional skill evals
  policies/
  templates/
  integrations/
  evals/
  observations/                    # deliberately promoted, sanitized evidence
  proposals/
  scripts/

user-config/                       # machine-specific settings, no secret values
state-root/
  projects/<repository-id>/
  runs/<run-id>/
    run.md
    plan.md
    changes/<change-id>/
      change.md
      tasks/
      builds/
      reviews/
    events.jsonl
    outcome.md
  worktrees/<run-id>/<change-id>/
  eval-results/
```

**ST-02 — Artifact format.** Plans, contracts, tasks, build reports, reviews, outcomes, observations, eval descriptions, and proposals MUST use Markdown with small YAML frontmatter where machine metadata is needed. Use JSONL for telemetry. Small YAML/JSON indexes or configuration are acceptable; do not put prose-heavy documents into nested schemas.

Illustrative change metadata:

```yaml
---
schema_version: 1
id: change-002
run_id: run-123
status: ready
risk: medium
review_mode: standard
depends_on: [change-001]
---
```

The Markdown body contains objective, scope, criteria, constraints, and validation. A task can be a short section rather than a separate file if stable identity and tracking are preserved. Schemas MUST validate required metadata and references without prescribing document verbosity.

**ST-03 — Recoverability.** At stage boundaries, persist enough state to identify the active owner, current branch/revision, validation and review applicability, pending findings, grants, and delivery attempts. On resume, compare persisted records with filesystem/Git/remote reality. Do not assume an interrupted command succeeded, reuse a stale approval, start a second writer, or silently lose completed evidence.

**ST-04 — Integrity.** State updates MUST avoid partially replacing valid authoritative records. Event writers MUST prevent interleaved/corrupt records; interrupted partial writes must be detectable and recoverable. Exact locking, atomic-write, and indexing mechanisms belong in the technical spec.

**ST-05 — Lifecycle states.** The implementation MUST distinguish at least planned, ready, building, verifying, reviewing, fixing, approved, and locally complete change stages, plus blocked, failed, and cancelled outcomes. Delivery separately distinguishes not requested, pending, in progress, succeeded, uncertain, and blocked/failed, with push and PR facts retained. Runs distinguish active, waiting, completed, failed, and cancelled. Names may change; the semantic distinctions and transition guards may not.

**ST-06 — Retention.** Historical revisions cited by evidence MUST remain identifiable and recoverable for the configured retention period, including after optional squashing. Raw runtime evidence remains local by default. Promotion into the toolkit repository is deliberate and sanitized. Retention/cleanup must not erase active work or silently break evidence links.

## 14. Skills and invocation evaluation

**SK-01 — Skill contract.** Each included skill MUST define purpose, use conditions, non-use conditions, procedure, constraints, required capabilities, version identity, and completion criteria where applicable. References load only when needed. Skills do not grant themselves permissions.

**SK-02 — Routing.** Contextual skills such as Go and debugging are selected for the current work, not merely because matching files exist somewhere in the repository. Lifecycle skills such as Git workflow and delivery are invoked from workflow state where practical. Avoid asking an LLM to rediscover a mandatory lifecycle step.

**SK-03 — Initial scope.** v0.1 MUST demonstrate Git workflow, delivery, and at least one coding/debugging skill path sufficient for an end-to-end engineering fixture. Existing suitable skills MAY be reused; a large new skill library is not required.

**SK-04 — Three evaluation dimensions.** Evaluate routing (should/should not select), observable adherence (was relevant procedure followed), and effectiveness (did it improve outcomes relative to its cost). A loaded skill is not proof of adherence. A successful run is not proof that a skill caused success.

**SK-05 — Positive and negative cases.** Routing evals MUST include required skills, forbidden/unnecessary skills, and lifecycle timing. For example, debugging is appropriate for an unknown regression but not automatically for a straightforward feature; delivery is not invoked during a local-only build.

**SK-06 — Controlled comparisons.** The eval mechanism MUST allow the same fixture/configuration to run against baseline and candidate skill or routing definitions. Comparative outcome claims require recorded controls and limitations; advanced automatic scorecards are future work.

## 15. Observations, evals, and proposals

**IM-01 — Capture from day one.** Every run MUST leave an outcome, including ordinary successes, failures, cancellations, and blocked attempts. Allow observations on agent behavior, verification quality, review value, routing, context cost, model allocation, permission friction, and workflow defects. Not every run needs an observation or proposal.

**IM-02 — Observation evidence.** Store source run/revision/event references, factual behavior, suspected explanation, confidence, and follow-up. Separate what happened from why it may have happened. Support candidate → confirmed → resolved and a way to dismiss an unsupported observation.

**IM-03 — Reproducible evals.** An eval MUST define a sanitized fixture, request/input, expected behavior, capabilities/environment, model and toolkit configuration, and scoring method. Prefer deterministic checks for checkable facts; use explicit rubrics for judgment. Record grader identity/version and failures. Nondeterministic agent outputs require declared sampling and uncertainty, not claims of identical reproduction.

**IM-04 — Evaluation coverage.** Maintain cases for orchestrator behavior, builder outcomes, reviewer findings, activation, skill routing/adherence, permission boundaries, model/effort policy, and end-to-end workflow. Eval lifecycle distinguishes candidate, accepted, active, and retired. A small meaningful suite is sufficient initially.

**IM-05 — Proposal contract.** Every behavior-change proposal MUST describe the problem, linked observations/evals, proposed change, expected benefit, risks, baseline results, candidate results, regression results, and rollback approach. Lifecycle distinguishes draft, proposed, accepted/rejected, and implemented. An accepted proposal is not implemented until linked to a toolkit revision.

**IM-06 — Human acceptance.** Canonical agents, skills, policies, and routing MUST NOT change automatically after a run. The improvement loop is observation → eval → proposal → before/after validation → human acceptance → versioned implementation. Autonomous software delivery permission is not permission for unattended self-modification.

**IM-07 — Avoid overfitting.** Keep ordinary successful work alongside failures and difficult cases. Improvement evals MUST check for regressions such as redundant tests, unnecessary skill loading, excessive review, or reduced autonomy. Outcome comparisons must distinguish task mix from actual policy effect.

## 16. Telemetry, metrics, and provenance

**TE-01 — Event contract.** Maintain an append-only JSONL event stream with schema version, unique event ID, timestamp, run ID, event type, actor/role, and relevant entity/attempt IDs. Include revision and artifact references where relevant. Ordering and correlation must be sufficient to reconstruct stage transitions; consumers must tolerate unknown optional fields and detect unsupported schema versions.

Capture run/plan/change/task transitions; validation results; revision creation; review/finding activity; skill selection/loading; model allocation/escalation; scope deviations/replans; permission decisions; delivery attempts/results; and final outcome. Essential evidence capture failure MUST be visible and must not permit unsupported completion claims.

Illustrative event:

```json
{"schema_version":1,"event_id":"evt-017","timestamp":"2026-09-16T10:00:00Z","event":"review.completed","run_id":"run-123","change_id":"change-001","review_id":"review-002","revision":"<full-sha>","mode":"standard","verdict":"approved"}
```

**TE-02 — Version provenance.** Each run MUST record toolkit commit plus dirty-state/content identity when definitions differ from that commit; loaded agent/skill/policy/template versions or hashes; runtime/adapter versions; requested and actual model/effort; effective non-secret configuration; project base and reviewed/delivered SHAs; and available relevant tool versions. A toolkit commit alone is insufficient when local definitions are modified. Resumed runs must retain original provenance and record any explicitly adopted configuration change.

**TE-03 — Minimum measurements.** Record elapsed times, changes/tasks, build attempts, review cycles and modes, findings by severity/category, validation failures, replans, scope growth, selected/loaded skills, permissions and interruptions, and delivery outcome. Collect input/output/cached tokens, tool calls, context size, and monetary cost only where available. Unknown metrics MUST remain unavailable, never zero. Separate runtime-reported usage from estimates and retain pricing assumptions for estimated cost.

**TE-04 — Useful derived metrics.** The implementation MUST support a basic report from recorded data; a dashboard is unnecessary. Evaluate:

| Dimension | Measures and interpretation |
|---|---|
| Completion | Accepted outcomes / attempted runs, with blocked and cancelled outcomes separately visible. |
| Quality | Acceptance coverage, unresolved blockers, validation regressions, and later reported defects linked when available. |
| Review | Cycles per change, findings by mode/severity, and cost per blocking finding as a diagnostic—not a quality target. |
| Skills | Routing precision/recall against labeled evals, adherence, unused loaded context, and controlled outcome comparisons. |
| Efficiency | Total time/tokens/cost per successfully reviewed change, including rework; report unavailable coverage. |
| Autonomy | Prompts per change, blocked routine operations, duplicate prompts, denied escalations, and fulfilled pre-authorized delivery. |
| Compute policy | Outcomes and total cost by complexity/risk/profile, escalation frequency, and repeated ineffective attempts. |

Do not reward finding volume, low token counts, or fewer permission prompts at the expense of correctness and authorization. Numerical optimization targets beyond acceptance fixtures require a baseline from real usage.

## 17. CLI and operational experience

**UX-01 — Required commands.** Expose these product operations through a CLI:

| Command | Required behavior |
|---|---|
| `agent-kit setup` | Install/configure the integration idempotently, preserving unrelated user settings; report changes and readiness. |
| `agent-kit doctor` | Diagnose source/config/state, runtime/adapter, role/skill availability, model/effort mapping, tools, capability enforcement, and activation readiness. Provide actionable remediation without exposing secrets. |
| `agent-kit init` | Discover and register a repository externally; optionally write explicitly requested durable project context. |
| `agent-kit update` | Update toolkit/integration definitions intentionally, preserve local modifications, validate compatibility and health, and identify the resulting version. Do not silently replace active-run definitions. |

**UX-02 — Safe maintenance.** Setup/update MUST show material installation/configuration changes, avoid overwriting unrelated configuration, and leave a diagnosable recoverable installation on failure. An update with conflicting local toolkit modifications must stop safely or preserve them through an explicit supported strategy. Active runs retain their recorded definitions or require explicit migration.

**UX-03 — Status and resume.** Users MUST be able to inspect a run's stage, next action, blocker, evidence, and delivery state, and resume an interrupted run. Dedicated `status`, `runs`, and `resume` subcommands are suggested, but equivalent supported runtime interactions satisfy v0.1. Commands must return meaningful success/failure status; machine-readable output MAY be added where needed for integration.

**UX-04 — Diagnostics are not repair authority.** Doctor SHOULD be read-only except documented temporary probes. Installation, updates, network access, and repair obey the same permissions model as other operations. Unsupported capabilities must be visible rather than hidden behind a generic “ready.”

## 18. v0.1 acceptance criteria and release gate

All required criteria below MUST pass on the declared reference environment. Each result must link to fixture/version, configuration, execution evidence, and relevant requirement IDs. Use disposable repositories and a controlled remote test repository for delivery tests. Core orchestration transition tests can use deterministic runtime doubles; the end-to-end gate must also use the actual selected runtime.

| ID | Scenario and measurable pass condition | Requirements |
|---|---|---|
| AC-01 | Run setup twice: one functioning integration remains, existing settings are preserved, and doctor reports accurate readiness. Update preserves local edits and reports incompatibility or a validated new version. | IN-01, UX-01–04 |
| AC-02 | On first prompts, implementation activates without naming roles; question/exploration creates no implementation worktree; planning stops before source edits; later “implement” continues correctly. A child-role invocation does not activate another root run. | IN-03–04 |
| AC-03 | Initialize an existing repository with no Agent Kit files: external metadata is created, applicable project instructions are respected, and tracked/untracked user files remain unchanged. | IN-02, ST-01 |
| AC-04 | A fixture with multiple independent concerns produces coherent ordered slices and tasks with acceptance coverage; material unexpected growth triggers a recorded replan rather than an unbounded diff. | PL-01–05 |
| AC-05 | Complete a real small Go change with one builder, meaningful verification, an immutable checkpoint, independent review, local commit, and an evidence-linked final outcome. No manual role coordination is required. | RO-01, VE-01–04, GT-01–04 |
| AC-06 | Seed a blocking defect: review creates a stable finding; builder fixes it; new checks and revision are recorded; focused re-review resolves it and approves only the new revision. | RE-03–06 |
| AC-07 | A low-risk fixture gets light review, ordinary logic gets standard, and a tiny security-sensitive fixture gets thorough consideration under recorded risk policy. No task-by-task full review occurs. | RE-01–02 |
| AC-08 | A failing or unavailable required check prevents ordinary approved completion/delivery. A skipped check is never reported passed. Repeated unsuccessful fixes terminate at the configured budget with an explicit reassessment/blocker. | VE-02–03, RE-07 |
| AC-09 | Mutate content or base after approval: delivery is blocked pending new applicable evidence. If final-commit rewriting is implemented, a new SHA cannot inherit approval silently. | RE-06, GT-04–05 |
| AC-10 | Begin with a dirty user checkout: isolated work completes without stashing, staging, modifying, or committing unrelated work. Attempts to start a second writer are rejected. | RO-01–02, GT-01–02 |
| AC-11 | “Local only” produces no remote mutation. “Implement and create a PR” completes one authorized push/PR without a redundant Agent Kit approval prompt. Simulated interruption after PR creation resumes without duplication. Host-required prompts are separately reported. | PE-05, GT-05–08 |
| AC-12 | Builder delivery, reviewer source writes, and an ungranted protected operation are rejected by the supported enforcement path, including an attempted alternate shell/tool route. Allowed scoped test/build commands work without escalation. | PE-02–08 |
| AC-13 | Lazy loading records only relevant selected definitions. Skill evals cover required and unnecessary invocation plus observable adherence; an installed irrelevant skill is not automatically loaded. | CX-01–04, SK-01–05 |
| AC-14 | Model policy records requested/actual settings; an unsupported effort level is reported truthfully; a difficult fixture triggers a bounded allocation reassessment. | MO-01–05 |
| AC-15 | Interrupt during build, review handoff, and uncertain delivery; resume reconciles state without duplicate writers, stale approval, lost completed evidence, or duplicate remote effects. | GT-07, ST-03–05 |
| AC-16 | A successful and an unsuccessful run both produce valid Markdown artifacts, parseable JSONL, outcome and provenance. Missing runtime token data remains unavailable. A seeded secret does not appear in persisted artifacts/logs. | ST-02, TE-01–03, PE-09 |
| AC-17 | Promote an observation into an eval and proposal; run baseline/candidate/regression cases; record a human decision and implementation version. No canonical definition changes before acceptance. | IM-01–07, SK-06 |
| AC-18 | Finish a multi-slice request with acceptance/integration checks and no mandatory duplicate exhaustive review. Distinguish local completion, requested delivery completion, and dependency waiting. | DM-03, RE-08, GT-08 |

Release evidence MUST also include a concise known-limitations list: supported platform/runtime/version, model mappings, enforcement coverage, missing telemetry, and deferred features. No broad claim of runtime portability or hard isolation may exceed the demonstrated evidence.

## 19. Open questions and implementation hypotheses

These choices require a technical decision or experiment. They do not weaken the product invariants above.

| ID | Question / initial hypothesis | How to decide; latest decision point |
|---|---|---|
| H-01 | Start with one local coding runtime and one platform; runtime choice remains open. | Spike first-prompt activation, independent roles, permission mediation, worktrees, model controls, and usage capture. Decide before adapter implementation. |
| H-02 | Use a thin CLI plus runtime adapter; implementation language remains open. | Compare integration/tooling needs and maintenance cost. Avoid building a general agent platform. Decide before CLI/state spec. |
| H-03 | Use a version-controlled source checkout, separate user config, and OS-appropriate state roots; prior examples used `~/.agent-kit`, `~/.config/agent-kit`, and `~/.local/state/agent-kit`. | Validate platform conventions, overrides, installation/linking, and update behavior. Decide in installation spec. |
| H-04 | Use Markdown/frontmatter plus a small validated state index, not a database initially. | Test atomic updates, crash recovery, querying, and stale ownership handling. Decide before persistence implementation. |
| H-05 | Sequential changes; for dependent remote PRs initially wait for predecessor merge, then recreate/revalidate the next base. For local-only runs, allow a declared immutable predecessor checkpoint as the next base. | Exercise a three-slice request and examine waiting time versus complexity. Choose/document one delivery policy before multi-slice delivery acceptance. Automated stack maintenance stays deferred. |
| H-06 | Keep checkpoint commits unchanged for initial delivery; add history cleanup only if reviewable commit quality needs it. | Test traceability and public history usability. If rewriting is included, implement GT-04 explicitly. Decide before delivery spec. |
| H-07 | Default all roles to capable/medium; lower for proven straightforward work and raise for risk/complexity. | Run a labeled representative fixture set across available configurations; compare quality and total effort. Decide initial mapping before execution-policy release. |
| H-08 | Start with a small explicit risk checklist and a default of standard review; no universal LOC cap. | Evaluate misclassification and review cost on light/ordinary/high-risk fixtures. Tune before review-policy acceptance. |
| H-09 | Start with a fix-loop limit of three cycles, plus repeated-finding detection. | Compare successful recovery versus wasted retries. Fix a documented configurable default before loop-control acceptance. |
| H-10 | A development permission profile plus optional authorized-delivery profile is enough initially. | Verify all acceptance paths and runtime enforcement limits; define exact scopes/cache/network behavior. Decide before permission implementation. |
| H-11 | Default raw evidence retention should preserve early learning without retaining unlimited sensitive output. | Decide artifact/output minimization, retention duration, cleanup protection, and revision reachability. Resolve before routine use. |
| H-12 | Bootstrap and loaded-context budgets should be measured rather than guessed as hard limits. | Measure baseline prompt overhead and irrelevant loads. Define initial regression tolerances once a baseline exists. |
| H-13 | Begin with a small explicit eval runner and deterministic graders plus rubric-based review. | Validate grading quality, runtime variability, fixture isolation, sampling cost, and baseline/candidate comparison. Decide before improvement-loop release. |
| H-14 | Basic summary reports are enough; costs and usage remain optional when the runtime does not expose them. | Probe telemetry support and document provenance/coverage. Resolve before claiming measured efficiency improvements. |

## 20. Future ideas, explicitly outside v0.1

- Parallel execution of independent changes and concurrent specialized reviewers on immutable revisions.
- Automated stacked-PR updates, CI monitoring, provider expansion, and richer delivery policies.
- Additional coding, domain, and personal skills; company-specific/compliance/CDK extensions only as separately scoped work.
- Richer constrained permissions such as conditional grants and deeper sandbox integrations.
- Controlled skill ablation studies at scale, richer scorecards, automatic routing suggestions, and model-policy optimization.
- Searchable historical evidence, richer memory, a dashboard, distributed execution, and broader runtime/platform support.

Future proposals still require evidence and human acceptance. None of these ideas implies autonomous canonical self-modification or unrestricted external actions.

## 21. Suggested phased delivery roadmap

This is a planning sequence, not an implementation backlog. Each phase should become one or more technical specifications with interfaces, schema/transition details, failure behavior, permission boundaries, and mapped acceptance tests. Break those specifications into small deliverable changes using Agent Kit's own principles.

| Phase | Product increment | Technical specs to derive | Exit evidence / dependencies |
|---|---|---|---|
| 0 — Resolve integration feasibility | Choose the reference platform/runtime and demonstrate the capabilities needed for honest automation. | Adapter contract; activation mechanism; enforcement boundaries; initial model mapping. | Resolve H-01/H-02 and major permission risks. Produce concrete supported/unsupported capability matrix. |
| 1 — Foundation and evidence | Installable toolkit, role/skill contracts, external state, IDs, templates, and telemetry available before real runs. | Source/config/state layout; artifact schemas; events/provenance; setup/doctor/init/update; recovery primitives. | AC-01/03 and foundational AC-16. No real execution without minimum evidence capture. |
| 2 — Local vertical slice | First-prompt activation through a single isolated change, progressive checks, checkpoint review, fix loop, and local outcome. | Intent routing; orchestration transitions; worktree/Git ownership; builder/reviewer handoff; verification and review contracts. | AC-02/05/06/08/09/10/12/15 on a real fixture. Start collecting observations. Depends on phase 1. |
| 3 — Decomposition and efficient policy | Sequential multi-change plans, risk modes, lazy skills, focused re-review, allocation/escalation, and integration completion. | Change/dependency planning; risk policy; skill routing; model/effort policy; context loading. | AC-04/07/13/14/18. Resolve initial policy hypotheses with evals. Depends on phase 2. |
| 4 — Authorized delivery | Scoped push and PR creation/update, accurate final-commit approval, and recovery from uncertain remote outcomes. | Delivery adapter; grants/pre-authorization; final revision gate; remote reconciliation; dependent delivery policy. | AC-11 plus delivery variants of AC-09/12/15/18. No merge/deploy feature. Depends on phase 3. |
| 5 — Improvement and release | Eval runner, observations/proposals workflow, baseline/candidate comparison, operational documentation, and complete release evidence. | Eval/grader/result schema; proposal lifecycle; metrics reports; retention/update compatibility; acceptance harness. | AC-16/17 and full AC-01–18 regression suite. Publish v0.1 limitations and validated configuration. |

Evaluation fixtures and observations should accumulate throughout the phases; phase 5 completes the improvement workflow rather than starting measurement late.

## 22. Handoff into specifications and tasks

For each technical specification, identify covered requirement and acceptance IDs, settle relevant hypotheses explicitly, define inputs/outputs and transition guards, describe permission enforcement and failure recovery, and name the evidence that proves completion. Do not convert future ideas or illustrative paths into mandatory tasks by accident.

The first implementation planning deliverable should be a decision record for the reference runtime and enforcement approach, followed by the foundation and single-change vertical-slice specifications. The PRD is complete enough to drive those decisions while keeping untested mechanisms visibly open.