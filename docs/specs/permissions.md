# Permission and operation-effects contract

**Status:** Task 03 reviewed specification. This defines decisions, not host enforcement; only versioned adapter probes can make a boundary enforced.

## Model and scope

A capability decision is `(action, resource, disposition, source, lifetime, enforcement)`. Disposition is `allow`, `ask`, or `deny`; enforcement is adapter evidence: `enforced`, `advisory`, `unsupported`, or `unknown`. A skill declares requirements but never grants them.

Resources are canonical identities, not command strings: repository/ref, worktree canonical path, state run, external filesystem path, network destination, Git remote/branch, provider repository/PR/base, and credential handle. Filesystem matching resolves containment and rejects symlink/relative escape; network matching uses connection destination. A resource scope cannot silently cover another repository, renamed remote, branch, base, or external path.

The baseline actions are repository/filesystem read/write, process execute/test/build, network read/write, Git read/branch/commit/history-write/push, PR create/update/merge, delegation, state read/write, and credential use/read. `credentials.use` permits an opaque helper for its listed destination; it never permits secret-value access or persistence.

## Transitive operation effects

Preflight expands every direct and foreseeable effect. An allow for `process.test` does not allow every side effect of a script.

| Operation | Minimum effects |
|---|---|
| Builder edit | repository read/write and filesystem worktree limited to assigned worktree. |
| Test/build/lint/format | process test/build/execute, repository/worktree read, explicit cache/temp writes, plus discovered network, credentials, generated-source, or external effects. |
| Checkpoint | Git read/branch/commit, state write, and orchestrator-controlled worktree/shared Git metadata access. |
| Review | repository read, declared check/cache/temp scopes, state write; deny source write and Git delivery/history effects. |
| Push/PR | Git push, network write, credential use, PR create/update, state write, and recorded remote/repository/source branch/base/revision. |
| Delete/rewrite/merge/deploy | exact protected effect and resource; no v0.1 default grant. |

If expansion cannot determine an effect, the operation is unavailable unless an enforced boundary captures it. Repository scripts, hooks and generators are potential effects, not harmless command spellings.

## Precedence and results

Resolve every effect in order: immutable host/runtime ceiling; explicit user grant/run instruction; reusable user profile; role restriction; resource restriction; operation restriction. An applicable deny wins. A more-specific selector can narrow allow but cannot override deny. Missing match is deny. A user grant outside the host ceiling is unavailable.

| Result | Meaning |
|---|---|
| `allowed` | All effects allow and required boundaries are enforced. |
| `needs_grant` | No denial, but an effect is ask; request one narrow deduplicated grant. |
| `denied` | Policy/role/resource denies an effect; do not retry through an equivalent route. |
| `unavailable` | Tool/capability/proof is advisory, unsupported, unknown, or absent. |

The development profile allows routine reads, builder writes inside the assigned worktree (including deletion of contract-attributable source), and declared check/cache/temp scopes only where enforced. Its role policy is: builder and reviewer deny `git.push`, `delivery.pr_create`, and `delivery.pr_update`; the orchestrator defaults those three actions to `ask` for its recorded repository/change, so an applicable user grant can allow them. Reviewer source writes and builder Git lifecycle/delivery remain denied. Force-push, history rewrite, merge, deploy, release, publish, destructive external deletion, destructive database effects, and secret-value reads are denied for every role. The orchestrator owns worktree/branch lifecycle, checkpoints and delivery subject to grants; separately scoped authorized cleanup remains subject to retention and recovery guards. If the runtime cannot enforce a required role boundary, preflight returns unavailable.

## Grants and authorization snapshots

A user grant is reusable authority scoped to action/resource, repository/change where applicable, source and lifetime. “Implement locally; do not push” creates denials for remote mutation. “Implement and create a PR” can grant push/PR creation for the named repository/change during its lifetime, subject to host ceilings. This user grant may predate a later checkpoint and remains usable through bounded fixes/retries while it stays valid; it does not itself approve a revision.

Before a protected operation, the orchestrator creates an immutable authorization snapshot under the durable operation intent. It binds the applicable grant IDs and effective policy to the exact operation, repository/remote/source branch/target base, exact current revision/tree, contract digest/version, approval reference, and attempt ID. A new revision/base/contract makes the old snapshot inapplicable and requires a fresh snapshot, but not a repeat user grant if the original grant still covers it. An uncertain remote attempt cannot reuse its snapshot: recovery needs positive absence, then creates a linked new attempt/snapshot under the still-valid grant.

No grant permits force-push, merge, deployment, publishing, destructive external deletion, shared-history rewrite, another repository, secret-value read, or a capability outside its selector. Routine in-worktree source deletion follows its builder write scope; expired worktree/evidence cleanup requires its own narrow authorization and the retention/recovery guards. Protected actions default deny. If their host mediation is not enforced, preflight is unavailable even with user authority. An expired grant ceases to match and resolution returns to the baseline `ask` or `deny`; it is not a perpetual denial. A declined same-scope grant suppresses equivalent prompts for its recorded lifetime and is recorded.

## Evidence and failures

Each decision records operation/effects, canonical resources, effective sources, grants/expiry, adapter capability/probe refs, and redacted result. It never records credential contents or full environments. Required actor result, permission decision, redaction, output, or provenance publication failure blocks the transition. Routine successes may aggregate, but denials, asks, unavailable effects and host prompts remain individually attributable.
