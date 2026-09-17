# State contract

**Status:** Task 02 reviewed specification for a Go implementation on the candidate macOS target. It is platform-neutral where possible and does not establish runtime readiness, activation, or permission enforcement.

**Covers:** DM-01/02/03, ST-01–06, TE-01/02, RE-06, GT-07.  This is a deliberately small run store, not a workflow engine.

## Authority and layout

There is one run directory at `<state-root>/runs/<run-id>/`. Its atomically replaced `state.json` is the authoritative current-state index. It has a supported `schema_version`, `run_id`, monotonically increasing `generation`, and the current fenced owner. An index is valid only if it parses, validates its references, and has a supported schema version.

Markdown documents are the human-readable, immutable record for a particular entity revision. JSONL is the append-only operational record. Neither overrides `state.json`; an artifact records `published_generation`, the generation that first referenced it, which is validated against the immutable index history/operation record rather than the reader's current generation. A current-status outcome may separately name the generation it observed. A reader uses the index for current state and the immutable artifact for its historical meaning. On disagreement it reports `reconciliation_required`, never chooses a narrative status.

The small index contains only IDs, state, lineage, immutable artifact references, current Git/delivery facts, active ownership, open finding IDs, grants applicable to an unfinished operation, operation intent, and retention timestamps. It does not duplicate objectives, reports, command output, prompts, secrets, or event payloads.

```text
<state-root>/
  runs/<run-id>/state.json
  runs/<run-id>/events.jsonl
  runs/<run-id>/artifacts/<kind>/<id>/<revision>.md
  runs/<run-id>/artifacts/operations/<operation-id>.json
  projects/<repository-id>/project.md
  worktrees/<run-id>/<change-id>/                 # source, not durable evidence
  quarantine/<run-id>/                            # recovery inspection only
```

`repository-id` is a stable opaque identifier derived from the canonical repository identity, not a filesystem path. `run-id` and all entity IDs are opaque randomly generated identifiers with a type prefix (`run_`, `chg_`, `rev_`, `review_`, `finding_`, `task_`, `obs_`, `eval_`, `proposal_`, `op_`, `evt_`). IDs are never reused. All timestamps are RFC 3339 UTC.

## Identity, lineage, and minimum index records

Every durable record has `schema_version`, `id`, `created_at`, `run_id` where applicable, and immutable `artifact_ref` (`relative_path`, SHA-256 digest, artifact schema version). A same-run reference must resolve inside its run directory and digest-match before it is used. A cross-run source reference (for example, an observation/eval source) additionally names the source `run_id`; its path is resolved and contained within that source run, never relative to the referring run or an arbitrary filesystem location.

| Entity | Immutable identity and required lineage |
|---|---|
| Request | `request_id`; preserved original request artifact; later interpretation/clarification artifacts name `source_request_id`. |
| Run | `run_id`, `request_id`, repository identity, provenance; a deliberate rerun adds `rerun_of_run_id`, while resume keeps the same ID. |
| Plan | `plan_id`, monotonically versioned `plan_version`; each version names `previous_plan_ref`, request, mapped criteria, and changes. |
| Change | `change_id`, owning run/plan version, objective, criteria/task IDs, base SHA, branch/worktree identity, predecessor IDs, risk/review/validation/delivery contracts. |
| Task | `task_id`, owning change and contract version; completion evidence refs. It has no implicit commit or review boundary. |
| Revision | `revision_id`, full SHA, tree SHA, base SHA, `previous_revision_id`, change and contract version; immutable once recorded. |
| Review | `review_id`, exact `revision_id` and full SHA, base SHA, contract version, review mode, reviewer provenance, verdict, finding IDs. |
| Finding | `finding_id`, originating review, stable category/severity/evidence, and append-only resolution history naming later review/revision. |
| Delivery attempt | `delivery_attempt_id`, change/revision, operation ID, intended remote/repository/base and requested effects; records push/PR facts and remote IDs separately. |
| Observation | `observation_id`, source run/artifact/event refs, confidence, category, fact and interpretation separately. |
| Eval | `eval_id`, version, fixture/input/grader/config refs and result refs; later versions name the predecessor. |
| Proposal | `proposal_id`, observation/eval refs, baseline/candidate refs, human decision, and implemented definition version when accepted. |

An artifact's frontmatter repeats only its own ID, schema version, lineage fields and `published_generation`; those values are checked against the index generation/operation that first referenced it. Historical artifacts are never edited to represent a new state: create a new artifact revision and move the index reference.

## States and guarded transitions

Run states are `active`, `waiting`, `completed`, `failed`, and `cancelled`. Change states are `planned`, `ready`, `building`, `verifying`, `reviewing`, `fixing`, `approved`, `locally_complete`, `blocked`, `failed`, and `cancelled`. Delivery states are `not_requested`, `pending`, `in_progress`, `succeeded`, `uncertain`, `blocked`, and `failed`.

Normal change transitions are:

```text
planned → ready → building → verifying → reviewing → approved → locally_complete
                         ^                   |
                         └──── fixing ←──────┘
```

The graph's arrows are allowed state transitions: `reviewing → approved` requires an approved review; `reviewing → fixing` requires a `changes_requested` verdict or an open blocking finding; `fixing → verifying` requires a new source revision or a documented no-source correction; and `verifying → reviewing` requires contract-required check evidence. `blocked`, `failed`, and `cancelled` may be entered from any nonterminal state with a reason artifact. Resuming moves a nonterminal condition only after reconciliation; `failed` and `cancelled` need an explicit new attempt or plan decision. `approved` is only valid with an applicable approved review and no open blocking findings. `locally_complete` additionally requires required validation, the current approved revision, and all non-delivery completion conditions. Delivery may move independently from `not_requested` to `pending`, `in_progress`, then `succeeded`; an interrupted external effect becomes `uncertain`, not failed or succeeded, until reconciled.

| Entity | Legal transition | Guard |
|---|---|---|
| Run | `active ↔ waiting`; `active/waiting → completed/failed/cancelled` | Waiting/resume needs a reconciled blocker; completed needs every change locally complete and delivery resolved as requested. Failed/cancelled terminal states require a new attempt or plan decision, never an in-place silent resume. |
| Change | `planned → ready → building → verifying → reviewing → approved → locally_complete`; `reviewing → fixing → verifying` | Guards are those above plus contract/base/dependency readiness. Any nonterminal state may become blocked/failed/cancelled with a reason; `blocked →` its pre-block state needs reconciled facts and still-valid guards. |
| Delivery | `not_requested → pending → in_progress → succeeded`; `in_progress → uncertain`; `pending/in_progress/uncertain → blocked/failed` | `in_progress` needs a fresh effective authorization snapshot and exact current approval. `uncertain → succeeded` needs positive remote confirmation; `uncertain → pending` needs positive authoritative absence and no in-flight request; blocked/failed need a new attempt after explicit resolution. |

Each index mutation declares expected prior generation, entity state, owner fence, and (when relevant) revision/base/contract version. Invalid transitions, missing evidence, an unsupported schema, a stale fence, or unmet guard reject the mutation without changing the last valid index.

If rename succeeds but a required directory sync fails, the operation is `durability_uncertain`: it performs no further mutation and does not claim either index generation authoritative. Recovery re-reads and validates the physically present `state.json` and referenced bytes to determine whether the old or new index is authoritative; it never overwrites the file to force the expected old result.

## Exact review approval applicability

An approval record applies iff all of these match the delivery/current revision gate: `revision_id` and full SHA, tree SHA, base SHA, change ID, contract artifact digest/version, review verdict `approved`, review schema supported, and zero unresolved blocking findings. It also requires checks marked required by that contract to be attached to the same source-content identity and to have passed (or carry a visible, separately authorized exception).

Changing content, base, contract, required validation, or a checkpoint SHA invalidates approval applicability. A rewritten but tree-equivalent commit remains inapplicable until a focused equivalence review records old/new SHAs, tree SHA, base, and the new approval. A grant is separate: a still-valid user grant may authorize a later retry, but each operation needs a fresh effective-authorization snapshot bound to the exact current SHA/base/contract and that grant. An old approval or old effective snapshot is never carried to another operation or delivery attempt.

## Ownership and fencing

Only one process may mutate a run index or append its event stream. It acquires `<run>/lock` by exclusive creation/open and writes `owner_id`, host/process identity, heartbeat timestamps, and a new monotonic `fence`. The index stores that fence. Every state/event write includes the fence; a writer that loses the lock or sees a greater fence stops before any further Agent Kit mutation. Lock files are contained in the run directory and are not symlinks.

Fencing protects the store, not arbitrary filesystem or Git writes by an old process. A missed heartbeat or elapsed 10-minute diagnostic lease is **not** authority to take over. Reclaim is permitted only after the old local process is proved exited and its lock can be acquired, or after an explicit operator stop procedure has terminated and verified the old writer. Otherwise the run stays `waiting`/`blocked`; it must not start another source writer. Reclaim creates a new owner/fence and an `ownership.reclaimed` transition, then reconciles source/Git before any source mutation. A stale process that could still run is therefore treated as a safety blocker, not made harmless by the fence.

## Roots, precedence, and containment

Roots are resolved once at setup/run creation and recorded as sanitized effective configuration. Precedence is explicit CLI option, then environment variable, then user config, then OS default. Empty values are ignored; conflicting explicit values fail. Defaults are macOS `~/Library/Application Support/agent-kit` for config and `~/Library/Application Support/agent-kit/state` for state; Linux follows XDG config/state directories; Windows uses known-folder equivalents. Toolkit source is an explicit installation path, never inferred from the target repository.

Source root (`AGENT_KIT_SOURCE_ROOT`) is read-only toolkit definitions; config root (`AGENT_KIT_CONFIG_ROOT`) holds non-secret settings; state root (`AGENT_KIT_STATE_ROOT`) holds runs, project metadata and worktrees. Worktree root defaults below state root but may be explicitly configured. A project checkout can never be a source, state, or worktree root by default. A configured root is canonicalized before use, must exist or be created as a directory, and must not resolve through a symlink outside the configured canonical root. Every generated child path is made from validated opaque IDs, then checked with `filepath.Rel`; absolute paths, `..`, separators in IDs, and symlink traversal fail closed. Artifact references are relative to a run and cannot name an external path.

## Data minimization and retention defaults

Persist only evidence needed to prove an outcome: the required original user request artifact, command identity, exit/result, timing, content/revision identity, selected provenance, and redacted output excerpt or a protected local output ref. Do not persist environment dumps, credentials, runtime prompts/transcripts, unbounded tool output, or runtime usage values that are unavailable. The original request is retained as the request entity, distinct from generated runtime prompts, but is sanitized before persistence: redacted spans preserve their position/type while no secret value is written. Secret redaction therefore applies before every artifact, event, diagnostic, quarantine, or filename is written.

Defaults before routine use: retain completed-run metadata, Markdown, indexes, events, referenced revision identifiers, and every referenced verification/review/approval/grant summary for 180 days; retain raw command/runtime output for 30 days; retain failed/blocked/cancelled runs for 180 days; retain active/waiting runs indefinitely; retain an approved/delivered revision and all evidence it reaches for 180 days after its last delivery record. Cleanup is opt-in/guarded and may delete only expired raw outputs first. It retains the required safe verification summary plus a tombstone with IDs, hashes, output-expiry/deletion time, and reason. It never deletes active worktrees, dirty worktrees, an artifact reachable from retained evidence, or the only copy of a revision/evidence link. Promotion to toolkit source requires a new sanitized artifact, never copying raw evidence.

## Runtime-dependent assumptions

The file/lock protocol is intended for a Go implementation using same-volume temporary-file replacement, file sync, and directory sync where the candidate OS supports them. A transition is called durable only after every required sync succeeds; status/reporting writes after that point are best-effort and never alter authority. Exact macOS durability primitives and cross-volume rejection must be verified in Task 05. This contract does not assume that Codex can start, lock a process, mediate a role, or report telemetry; ADR 001 leaves those capabilities blocked or unsupported.
