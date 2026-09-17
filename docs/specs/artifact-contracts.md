# Artifact contracts

**Status:** Task 02 reviewed specification. This specifies persisted artifacts only; it does not claim that the selected runtime can generate them.

## Common envelope

Each Markdown artifact begins with small YAML frontmatter. Required fields are:

```yaml
schema_version: 1
kind: change                 # closed supported kind for this schema
id: chg_...
run_id: run_...
created_at: "2026-09-17T12:00:00Z"
published_generation: 42
content_sha256: <sha256 of canonical bytes excluding this field>
```

Line endings are LF and UTF-8. Canonical bytes are the exact persisted artifact bytes with `content_sha256` omitted from frontmatter; writers compute the digest before atomic publication. A same-run reference contains `kind`, ID, relative path, schema version and digest. A cross-run source reference also contains its source `run_id`; validation contains the path within that source run. Parsers reject duplicate frontmatter keys, missing required fields, invalid IDs, unsupported major schema versions, and references escaping their owning run directory. Unknown optional fields are preserved where possible but do not become authority.

Bodies are Markdown intended for people. Frontmatter gives only metadata that needs validation or correlation. Prose does not alter the state index, authorization, validation result, approval, or delivery fact. Each immutable artifact revision has a distinct filename/revision identity; corrected narrative produces a successor linked by `supersedes_ref`.

## Required artifact kinds

| Kind | Required metadata beyond common envelope | Body must record |
|---|---|---|
| request | `request_id`, `request_revision` | original user request verbatim enough to preserve intent; clarifications link back. |
| plan | `plan_id`, `plan_version`, `request_ref`, `previous_plan_ref` | ordered changes, criterion mapping, assumptions, risks, replan rationale. |
| change contract | `change_id`, `plan_ref`, `contract_version`, `base_sha`, `delivery_scope`, `risk`, `review_mode` | objective, inclusions/exclusions, criteria, tasks, dependencies, validation and completion guards. |
| task/build report | owning `change_id`, `task_id` where applicable, `attempt_id`, evidence refs | bounded work, completion/deviation, checks and factual evidence. |
| revision | `revision_id`, `change_id`, `full_sha`, `tree_sha`, `base_sha`, `previous_revision_id`, `contract_ref` | checkpoint provenance and linked verification. |
| review | `review_id`, `change_id`, `revision_id`, `full_sha`, `base_sha`, `contract_ref`, `mode`, `verdict` | scope, evidence reviewed, findings, limitations. |
| finding | `finding_id`, `review_id`, severity/category, status, resolution history refs | evidence/location, impact, correction required when blocking. |
| outcome | `outcome_id`, run/change IDs, final state, current revision/delivery refs | achieved criteria, validation/review/delivery facts and blockers. |
| observation, eval, proposal | their stable ID/version and required source refs from `state.md` | the corresponding DM-01 information, with fact/evidence distinct from interpretation/decision. |

Event records are one JSON object per UTF-8 LF-terminated line in `events.jsonl`. Required fields: `schema_version`, `event_id`, `timestamp`, `event`, `phase`, `run_id`, `operation_id`, `state_generation`, `owner_fence`, actor/role, and relevant entity IDs. `event_id` is deterministically derived from operation ID plus its ordinal/name, so a retry repairs or recognizes the same event rather than inventing another. Before state commit, `phase` is `prepared` and the event must not claim a completed transition; only an index that references its range establishes that transition. After index commit, a best-effort `phase: committed` confirmation may be appended. Events name artifact refs and full SHAs when relevant. Event schemas are forward-tolerant for unknown optional fields; an unknown required semantic or unsupported major version blocks state transition/recovery until upgraded.

## Publication and cross-artifact protocol

An operation is an idempotent durable intent, not merely a command invocation. Before a transition, the owner allocates `op_<id>` and writes an immutable operation record containing intent, affected entity IDs, expected generation/fence/state, expected current references, and any external idempotency key. It emits `operation.prepared` only after that record is durable.

For a state transition that names new Markdown and events:

1. Validate the current index and owner fence; derive all new immutable artifacts and event records with `operation_id` and next generation.
2. Atomically publish each artifact at its final unique path, syncing the file and parent directory. Existing immutable paths may only be accepted when their digest matches exactly.
3. Append complete `phase: prepared` JSONL records as one serialized single-writer append, sync the file, then record the event-stream offset/end digest in the pending index update. A partial final line is never interpreted as an event. These records are intent/audit evidence, not a claim that the transition committed.
4. Atomically replace `state.json` with generation `n+1`, references to the artifacts/event range, and operation status `committed`; sync its parent directory. **This replacement is the durable commit point only when all required syncs succeed.** A rename followed by failed directory sync is `durability_uncertain`, not a rollback or a prior-index result.
5. Best-effort append `operation.committed`/committed event confirmations after the commit point. Their absence does not roll back the index.

No Markdown artifact is published by overwriting a prior revision. Event publication before index replacement is permitted because events alone do not change authority. If a crash occurs, recovery uses the index/operation record to decide whether staged artifacts/events are orphaned, committed, or need inspection. A writer must never state a completed external operation before state reconciliation records its observed fact.

Atomic replacement requires a temp file made in the destination directory, exclusive permissions, complete write, flush, close, rename over the old index, and directory sync. Replacement across filesystems, a nonregular target, or a symlink target is rejected. A transition is durable only if all required file and directory sync calls succeed. A failure before rename leaves the prior valid index authoritative; a failure after rename (including directory sync) is `durability_uncertain` and blocks further mutation until recovery reads the physically present index. Post-commit confirmations and presentation-only diagnostics are explicitly best-effort. The implementation must serialize all JSONL appends under the run owner lock; separate processes cannot append concurrently.

## Schema evolution and corruption

Schema versions use `major.minor` semantics even if serialized initially as integer `1` (meaning 1.0). Readers accept supported major versions and unknown optional minor fields. Writers emit only the current supported version. A newer/unsupported major index, operation record, or event required for recovery puts the run into an inspectable blocked state without modifying it; the user must upgrade or use an explicit export/migration tool. Never discard an unknown record to make a run appear complete.

Malformed index, digest mismatch, dangling reference, duplicate immutable identity with different bytes, or corrupt JSONL before the recorded committed offset is corruption. Preserve bytes, quarantine copies without parsing secrets into diagnostics, block mutation, and require recovery. A malformed trailing JSONL line after the index's recorded offset is an interrupted append: it is never an event. Controlled recovery durably records its recovery intent, copies the suffix to quarantine, truncates only that uncommitted suffix (the narrow exception to append-only), then appends the stable-ID recovery event and commits the updated index. It does not append after an unterminated line or reconstruct missing evidence.

## Approval and grants in artifacts

Review frontmatter must name the exact revision/full SHA, base SHA, contract reference/digest/version, verdict and finding status snapshot. An `approved` review does not grant delivery by itself. A durable user grant is separate and names its source, capability/action, scoped repository/remote/base, change ID where known, lifetime, and revocation state. Thus “implement and create a PR” can authorize the later normal push/PR operation without knowing its future SHA. It cannot cover force push, merge, deployment, or a different remote. Each operation creates an immutable effective-authorization snapshot that names this grant plus the exact current revision/full SHA, base, contract digest, owner/policy resolution, and operation ID. The snapshot is revalidated for each retry; a valid unexpired grant may support a new snapshot, but no old snapshot or review approval silently applies to later content.

## Scenario evidence

**Checkpoint crash:** the process publishes a revision artifact and `revision.created` **prepared** event, then crashes before replacing `state.json`. The index still names the old revision; recovery finds `op_` prepared with matching orphan files, retains them for inspection, and does not treat the checkpoint as reviewable. If index replacement had completed, its digest/reference makes that exact revision current even if best-effort `operation.committed` is missing.

**Review crash:** a review artifact with `changes_requested` is published but the transition is uncommitted. The change stays `reviewing`, and no finding is open in authoritative state. Recovery either commits the whole prepared transition after validating all refs and fence/generation or quarantines it; it never lets the prose alone send the builder to `fixing`.
