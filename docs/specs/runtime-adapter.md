# Runtime adapter contract

**Status:** Task 03 reviewed specification. This is a platform-neutral boundary for a later implementation. It does not make the candidate Codex CLI a supported execution runtime. ADR 001 records that its required host enforcement and activation evidence is currently blocked.

## Purpose and authority

The adapter translates a role assignment into a particular runtime invocation and translates only observed results back to the orchestrator. It owns neither workflow state nor policy: the orchestrator resolves permissions and transition guards, publishes the operation through the `state.json` protocol, and sends a fully resolved assignment to the adapter. The adapter may refuse a request; it may never replace a refusal with a less constrained invocation.

`AdapterVersion` identifies the adapter build, runtime product/version, host OS/architecture, and content hash of the probe suite. A capability answer is valid only for that exact tuple. The authoritative current capability matrix is probe evidence referenced from the run, not a runtime self-description.

## Capability evidence

Each capability has an operation and canonical resource scope, status, tested runtime tuple, probe ref, and host-control description.

| Status | Meaning and assignment rule |
|---|---|
| `enforced` | A versioned adversarial probe proved the named host/tool boundary for the actor, resource, and alternate route. The adapter may execute an allowed assignment within the proven scope. |
| `advisory` | The runtime can express an instruction or preferred route, but no interception proof exists. It cannot satisfy a protected role boundary. |
| `unsupported` | The adapter/runtime cannot express or mediate the boundary. It cannot execute an assignment requiring it. |
| `unknown` | No applicable current probe exists, it is stale, or evidence is incomplete. Treat exactly as unsupported for a required effect. |

The candidate runtime is presently `unknown`/unproven for automatic activation, reviewer source-write prevention, builder delivery prevention, protected external effects, requested/actual model effort, and runtime usage. No protected assignment requiring those unproven boundaries is supported. Missing optional model/usage telemetry does not itself block a launch when every required permission boundary is independently enforced. Deterministic fake adapters may model `enforced` effects for unit tests, but are never evidence for an actual-runtime acceptance criterion.

## Requests and results

All calls carry `run_id`, `change_id`, `operation_id`, state generation, owner fence, actor ID, role, expected adapter version, and correlation ID. The adapter rejects an expired/cancelled assignment, stale generation/fence, unknown role, mismatched runtime tuple, or resource outside the assignment. It also receives an immutable assignment digest.

| Request | Required input | Result |
|---|---|---|
| `Probe` | tuple, probe case, disposable resources, adversarial routes | redacted observation; no user configuration mutation. |
| `Preflight` | assignment, transitive effects, resolved grants, required capability statuses | a disposition for every effect before role start. |
| `Launch` | assignment, allowed resource handles, cancellation handle, requested model/effort, output limit | one actor only after permitted preflight. |
| `Quiesce` | actor/operation/fence and reason | observed completed/stopped process proof before source snapshot; otherwise unavailable/uncertain. |
| `Cancel` | actor/operation/fence and reason | observed acknowledgement or exit only. |
| `Inspect` | run/actor correlation | liveness, versions and safe usage facts; unavailable stays unavailable. |

An assignment contains no raw credentials, unredacted environment, or authority the role does not need. The builder receives contract, task scope, worktree, rules/skills and reporting sink. The reviewer receives only the factual review package. Child roles carry `child_role` and must not activate an orchestrator; if this cannot be enforced or verified, automatic activation is unavailable.

`LaunchResult` is factual and immutable: `started`, `completed`, `cancelled`, `denied`, `unavailable`, `failed`, or `uncertain`; timestamps; versions; exit/signal if available; redacted output ref/digest; declared temporary output identities; and actual model/effort/usage only when runtime-reported. It lists attempted effects and their adapter disposition. Actor prose never proves source, check, review, or delivery completion.

`denied` is a policy/host refusal; `unavailable` is missing runtime support, proof, or tool; `uncertain` is interruption with unknown process/external outcome. `completed` only says the actor exited normally: workflow guards still validate artifacts, source, checks and review. Output is bounded/redacted before persistence. Failed redaction, digesting, or required result/provenance publication makes the operation fail and prevents a completion claim. Optional unavailable telemetry is `unavailable`, never zero.

A completed builder or reviewer stage also requires a digest-linked build/report or review artifact naming the assignment, fence and content identity, plus its `LaunchResult` reference, in the same Task 02 operation/index transition. A stored output or actor result lacking that binding cannot advance a stage.

## Fail-closed behavior and probes

No alternate command/tool route follows a denied, unsupported, or unknown protected effect. The adapter returns every missing effect and its probe/capability reference once; the orchestrator can seek a narrow grant or block. It must not recursively delegate, bypass a runtime boundary, or turn a host prompt into an Agent Kit grant.

Cancellation is durably recorded before invocation and invalidates the assignment. Quiescing requires observed completed/stopped process proof. Until either is observed and source is reconciled, source writing, checkpointing, review, and delivery are blocked. A late old-fence result is observation only. A fence cannot prevent stale filesystem/Git writes; replacement is allowed only after the old process is proven dead. Otherwise block and reconcile late effects as uncertain.

To promote a capability to `enforced`, a probe artifact names the actor, operation/effects, canonical resource, expected decision, host/tool interceptor, normal route and alternate shell/tool route. It records exact runtime/host/adapter versions, configuration hash, exit/result, redacted evidence and limitations. A prompt, worktree, role label, or unit test alone cannot prove enforcement. Actual AC-02 and AC-12 remain blocked until the corresponding probes pass on the selected environment.
