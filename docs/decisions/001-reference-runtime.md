# ADR 001: Codex reference-runtime candidate remains gated on actual enforcement evidence

**Status:** proposed; runtime readiness blocked

**Date:** 2026-09-17

**Deciders:** user selected Codex; Task 01 investigation

## Context

Agent Kit needs one declared runtime/platform path that can automatically activate an ordinary implementation request, separate builder and reviewer work, and enforce the protected effects in AC-12. The PRD explicitly rejects treating role prompts and Git worktrees as security sandboxes.

The user selected Codex. Investigation on the current host found standalone `codex-cli 0.154.0` and Codex Desktop 26.908.70816 on macOS 27 arm64. The CLI offers noninteractive `exec`, `exec fork`, sandbox modes, JSON event output, and an experimental app-server that can generate TypeScript protocol bindings and JSON Schema. The evidence is recorded in [the runtime probe](../../experiments/runtime-probe.md).

OpenAI's [model guidance](https://developers.openai.com/api/docs/guides/latest-model) documents reasoning controls at the API/model level. It does not substitute for evidence that a specific installed Codex surface exposes or applies a requested effort setting.

## Decision

Choose the following **candidate reference environment**, pending the gate below:

| Decision | Choice | Basis |
|---|---|---|
| Runtime | Native standalone Codex CLI, pinned at 0.154.0 for probe reproduction | installed interface and user decision |
| Platform | macOS arm64, currently macOS 27.0 | native CLI and macOS seatbelt sandbox surface are present |
| Integration surface | Thin local CLI adapter that launches/observes Codex sessions; app-server protocol is experimental and is not a v0.1 dependency until versioned and exercised | `codex --help` exposes the CLI; `codex app-server --help` explicitly labels app-server experimental; neither has demonstrated compatibility or enforcement here |
| Implementation language | Go | Go 1.26.1 and Node 24 are installed. TypeScript has an app-server binding generator, but Agent Kit does not need app-server in v0.1; Go keeps the distributed adapter to one binary and avoids a Node runtime dependency. A later app-server use can consume its generated JSON Schema through a small Go protocol client. |

This decides H-01/H-02 as a constrained implementation target, **not** as an accepted supported runtime. Version matching must be exact in `agent-kit doctor`; an upgrade requires renewed probes.

## Enforcement contract for the candidate

The adapter may provide durable contracts, role prompts, worktree selection, evidence collection, and fail-closed preflight. It must only label an effect enforced when a tested runtime interceptor binds all of:

| Effect | Needed tested control | Current state |
|---|---|---|
| Reviewer source write | actor-specific reviewer process plus source resource restriction, tested through shell and alternative executable/tool | unproven |
| Builder delivery | actor-specific delivery/Git/network control, tested through alternate route | unproven |
| Protected external effect | operation/resource-specific host control that cannot be escaped through arbitrary execution | unproven |
| Routine scoped build/test | allowed working directory plus declared temp/cache writes | unproven |

Until a row is tested, its enforcement is unproven. Role prompts and worktrees are advisory; real builder/reviewer assignment remains unavailable. Adapter preflight must reject an assignment requiring an unmediated protected effect and must not retry via a different shell or tool. It cannot claim that a prompt, a worktree, or a broad workspace-write process sandbox enforces actor-specific policy.

## Probe outcome and gate

In the present managed host, a disposable `codex -a never exec --ephemeral ...` exited 1 with `failed to initialize in-process app-server client: Operation not permitted`. No model call, write, JSONL event stream, fork, or sandboxed command ran. This differs from doctor’s successful handshake with the already-running Desktop app-server. Direct `codex sandbox` requires a named permission profile. Its inspected `-p` option layers a file under `$CODEX_HOME`; no documented isolated profile-root route was identified, so no profile was created or changed. The host's `Never` approval policy also prevents an `ask` mediation demonstration.

Therefore:

- Automatic activation and child recursion prevention are **blocked**, not accepted.
- Reviewer-write, builder-delivery, protected-effect, and alternate-route denials are **unproven**; no builder-delivery or Git-specific per-role control was found in inspected CLI help.
- The local model catalogue lists `gpt-5.6-terra` and medium effort, but no requested/actual `gpt-5.6-terra` medium result exists. `exec --help` lacks a direct effort flag; possible configuration-based selection remains untested.
- Per-run tokens, cost, tool counts, and context usage are **unavailable** in this evidence.

Do not enable real execution, claim AC-02/AC-12, or mark Task 01 complete based on this ADR. Reopen the gate only in an isolated macOS environment that permits nested Codex execution and disposable policy-profile setup. The required rerun is specified in the probe document.

## Consequences

Tasks that only specify contracts may proceed while carrying this limitation. Any task that registers automatic activation or executes a builder/reviewer must remain unavailable and report why until the actual-runtime fixtures prove the required boundary. Alternative OSes, Codex surfaces, or model mappings remain unvalidated until separately evidenced.
