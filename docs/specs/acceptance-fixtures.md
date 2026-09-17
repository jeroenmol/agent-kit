# Execution-contract acceptance fixtures

**Status:** Task 03 reviewed specification. These are walkthrough specifications, not executed evidence. The current host has not proved a supported execution runtime.

## Evidence classes

`deterministic` fixtures use a fake adapter and disposable Git repository to test orchestration, state generation/fence, grants, guards and artifacts. They establish implementation behavior only. `actual-runtime` fixtures run the selected adapter/runtime on a disposable repository and prove activation and host enforcement. They are mandatory for AC-02 and AC-12; a fake adapter cannot satisfy them. `controlled-remote` additionally uses a disposable remote.

Every fixture records fixture/version, policy/configuration hashes, adapter/runtime/host tuple, probe refs, operation/artifact/event refs and PRD IDs. It uses synthetic data and asserts a seeded secret is absent from Markdown, JSONL, diagnostics, filenames and output. Results are `passed`, `failed`, `blocked`, or `unavailable`; the last two never pass. Host prompts are recorded separately from Agent Kit policy.

| Fixture | Class | Setup/action | Required result | Coverage |
|---|---|---|---|---|
| `local-happy-go` | deterministic, then actual-runtime | Clean disposable Go repo; small scoped request. | One builder, checks, checkpoint, independent exact review, local outcome. | AC-05, VE, RE |
| `blocking-finding` | deterministic, then actual-runtime | Seed review-detectable defect. | Stable blocker → fix → fresh checks/revision → re-review only approves new revision. | AC-06 |
| `required-check-nonpass` | deterministic | Required check fails, unavailable, skipped, times out, flakes. | States stay distinct; no completion; three unsuccessful cycles reassess/block. | AC-08 |
| `approval-stale` | deterministic | Approve then change content/base/contract/check identity or SHA. | Approval gate fails pending new evidence. | AC-09 |
| `dirty-second-writer` | deterministic, actual-runtime where enforceable | Dirty user checkout; concurrent builder/reclaim attempt. | User checkout untouched; no second writer/takeover until old process proven dead. | AC-10, AC-15 |
| `checkpoint-stability` | deterministic | Mutate source after checks before checkpoint and after quiesce while review input is assembled. | No checkpoint/review approval from stale evidence; checks invalidate and state returns to building/verifying. | VE-02, RO-02, AC-08 |
| `cancel-build` | deterministic | Cancel during source work; inject late old-fence result and source/Git write after cancellation/reclaim. | Durable cancellation; no stale checkpoint/review; late write is reconciled as uncertain; handoff required. | RO-01, AC-15 |
| `role-adversarial` | actual-runtime | Reviewer source write and builder delivery via normal and alternate shell/tool; ungranted protected effect; allowed scoped test/build. | Denied effects receive host-enforced rejection; allowed check works. Missing enforcement returns unavailable/blocked and AC-12 remains unmet. | AC-12 |
| `activation-child` | actual-runtime | Ordinary implementation/question/plan/later implement/review-only/child prompts. | Intended activation only; no child root workflow. Unsupported is unavailable. | AC-02 |
| `output-provenance` | deterministic, actual-runtime for runtime fields | Complete/fail/cancel; seed secret; omit usage. | Valid artifacts/events/outcomes, secret absent, missing usage unavailable; required evidence-write failure blocks. | AC-16 |
| `scope-escalation` | deterministic | Discover material unrelated work. | Recorded deviation; version/split/defer/block; old approval invalidated. | PL-05, RO-04 |

## Transition walkthroughs

The success trace is `planned → ready → building → verifying → reviewing → approved → locally_complete`. Every arrow asserts expected state generation/fence, prepared operation then immutable refs/event correlation and atomic index transition. It retains the checkpoint commit and makes no delivery attempt for local-only work.

The failing trace stops at verifying or blocked and cannot create approved/local-complete state. Cancellation preserves work/evidence, enters cancelled and needs recovery. A denial stops at preflight and names canonical effect/resource/policy/grant source; it makes no alternate-route attempt. An unavailable result similarly makes no invocation but names missing runtime support/evidence.

For fixes assert a fresh content identity, checks, checkpoint and review per attempt. Mutate every exact-approval input (revision/full SHA/tree, base, contract digest/version, checks, findings) and assert the gate fails. Remote retry fixtures must show positive absence after reconciliation before a linked new authorization snapshot; the reusable user grant need not be re-requested if valid.

Actual-runtime evidence names exact runtime/adapter/host versions, probe configuration, enforcing interceptor, normal and alternate-route result and redacted evidence. Role prompts, worktrees, documentation and fake tests are insufficient. Current nested-runtime initialization failure leaves actual-runtime fixtures blocked; no AC claim may be satisfied until they pass.

Review fixtures assert that reviewer input contains contract/criteria, risk/mode, base/revision/tree, diff, project instructions, check evidence, constraints, prior finding/fix mapping and surrounding code, but no builder private reasoning. Findings require factual evidence/impact; preference-only feedback cannot extend the loop.
