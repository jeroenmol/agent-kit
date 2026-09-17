# Codex runtime feasibility probe

**Status:** blocked in this host; this is evidence, not a runtime acceptance result.

**Date:** 2026-09-17

**Target:** Codex standalone CLI 0.154.0 on macOS arm64.

**Scope:** disposable directories under `/private/tmp`; no user configuration changes, remote mutations, credentials, or destructive operations.

## Question

Can the selected Codex surface activate an Agent Kit workflow, start an independent child role, mediate role-specific effects, expose requested/actual model and effort, and produce usable usage telemetry?

The required PRD outcome is stronger than a cooperative prompt: AC-12 requires denial of reviewer source writes, builder delivery, and ungranted protected effects through alternate shell or tool routes. A worktree and a role instruction are therefore never counted as enforcement.

Official OpenAI model guidance confirms that reasoning allocation is a model/API concern, but it does not establish that this installed Codex CLI exposes a corresponding setting or that the configured model is available to this account. The local CLI help and an executed run are the authority for those runtime claims. Sources: [OpenAI model guidance](https://developers.openai.com/api/docs/guides/latest-model) and [GPT-5.3-Codex model reference](https://developers.openai.com/api/docs/models/gpt-5.3-codex).

## Environment

| Item | Observed value |
|---|---|
| OS | macOS 27.0 (Darwin 27.0.0), arm64 |
| Codex CLI | `codex-cli 0.154.0`, standalone macOS aarch64 install |
| Desktop app | 26.908.70816, running; doctor reports app-server handshake success |
| Host policy | filesystem sandbox restricted; approval policy `Never`; network sandbox enabled |
| Auth | doctor reports ChatGPT auth configured; no auth values were read or recorded |

The host restriction is material: a nested `codex exec` needs its in-process app-server client and fails with `Operation not permitted`. That happens before model invocation, so no inference, requested model, or requested effort was applied. `exec --help` exposes `--model` but no direct effort flag; configuration-based effort selection remains unproven.

## Commands and results

All commands were run from the Agent Kit checkout. Replace `PROBE_DIR` with a new directory made by `mktemp -d /private/tmp/agent-kit-runtime-probe.XXXXXX`.

### Interface inventory

```sh
sw_vers
uname -a
codex --version
codex --help
codex exec --help
codex exec fork --help
codex app-server --help
codex sandbox --help
codex features list
codex doctor --json
```

Exit: 0 for help/version/feature commands; doctor exit: 1 because its local state-integrity check failed. The redacted doctor report still confirms CLI and desktop versions, configured model (`gpt-6-astra`), app-server handshake, host policy, and provider reachability. It does **not** show per-run token or cost telemetry.

Relevant exposed surfaces:

- `codex exec` supports `--ephemeral`, `--json`, `--worktree`, `--model`, `--sandbox read-only|workspace-write|danger-full-access`, and global `--ask-for-approval on-request|never`.
- `codex exec fork SESSION_ID` supports a child/forked session command.
- `codex app-server generate-ts` and `generate-json-schema` exist, but `app-server` is labelled experimental.
- `codex sandbox` uses the macOS sandbox but requires a configured named permission profile; it is not a role-policy API.

Sanitized diagnostic excerpt retained as versioned evidence (the complete redacted doctor output was not promoted because it includes unrelated local-state inventory):

```text
codex-cli 0.154.0
desktop.app.version: desktop application installed, version 26.908.70816, running true
desktop.app_server.handshake: desktop app-server initialized successfully
sandbox.helpers: approval policy Never; filesystem sandbox restricted; network sandbox enabled
```

The desktop-app handshake is a doctor check against the already-running Desktop surface. It does not prove that a nested standalone `codex exec` can create its own in-process app-server client under this host sandbox.

The following catalogue result came from the read-only command `codex debug models`, filtered locally with `jq` to the named slug. It establishes catalogue availability only; it does not establish that a session applied the setting:

```json
{"slug":"gpt-5.6-terra","default_reasoning_level":"medium",
 "supported_reasoning_levels":["low","medium","high","xhigh","max","ultra"],
 "supported_in_api":true,"visibility":"list"}
```

### Activation attempt

```sh
git -C "$PROBE_DIR" init -q
git -C "$PROBE_DIR" config user.email probe@example.invalid
git -C "$PROBE_DIR" config user.name RuntimeProbe
printf 'base\n' > "$PROBE_DIR/README.md"
git -C "$PROBE_DIR" add README.md && git -C "$PROBE_DIR" commit -qm initial
codex -a never exec --ephemeral --ignore-user-config --json \
  -C "$PROBE_DIR" -s workspace-write \
  'This is a disposable runtime probe. Create exactly one file named activation.txt containing exactly activated followed by a newline. Then reply with exactly ACTIVATED. Do not read or write outside the current directory; do not use network or git.'
```

The exact command above was the bounded attempt. It omitted explicit model selection and exited 1 before a model call:

```text
Reading additional input from stdin...
Error: failed to initialize in-process app-server client: Operation not permitted (os error 1)
```

`activation.txt` was absent. A preceding attempt placed `-a never` after `exec`; it exited 2 with `unexpected argument '-a'`. This establishes only CLI argument placement, not permission behavior.

### External host-startup check

On 2026-09-17, the user ran `sh /Users/jeroenmol/workspace/agent-kit/experiments/check-runtime.sh` from a normal local Terminal, outside this managed host. The sanitized artifact reported:

```text
host_startup_check=started
process_exit=0
diagnostic=ready_response
requested_model=gpt-5.6-terra
requested_reasoning_effort=medium
```

This is actual evidence that an ephemeral, noninteractive Codex session can initialize on that external host and return the fixed response. It supersedes the host-initialization blocker only for that host. The script deliberately records requested settings rather than provider-resolved settings, so it does **not** establish that the session actually used `gpt-5.6-terra` or medium effort. It also does not test Agent Kit activation, child roles, tool permissions, network isolation, delivery, or AC-12.

### Direct sandbox attempt

```sh
codex sandbox -C "$PROBE_DIR" /bin/sh -c 'printf safe > reviewer-write.txt'
```

Exit: 2. The file was absent because the command was rejected before execution: `--permission-profile <NAME>` is required. `sandbox --help` documents `-p` as layering `$CODEX_HOME/<name>.config.toml` and does not expose an alternate configuration root. No documented isolated profile route was identified from the inspected local help, and no profile was created or changed. This is not evidence that a reviewer write would be denied.

### Permission-probe boundary

The external startup result does not make a reviewer permission probe safe or interpretable yet. The current installed `codex exec --help` documents `--sandbox read-only`, but describes it only as the policy for model-generated shell commands; it has no actor, role, source-path, delivery, or network-permission option. `codex sandbox --help` lists `--sandbox-state-disable-network`, but only when a caller supplies `--sandbox-state-json`; it also accepts a named `--permission-profile` resolved from the active configuration stack. The inspected help exposes no way to create a disposable profile or generate that state JSON without configuration/state setup.

Accordingly, no script can test the full permission claim now. A prompt that asks a reviewer to try a shell write and a Python write can still provide narrow behavior evidence, but it does not bind an actor and cannot establish that network or another available route is denied. Running such a probe makes the requested no-network guarantee depend on instructions rather than an observed boundary. `--ignore-user-config`, used by the startup check to avoid loading user settings, explicitly skips project execpolicy rules and therefore cannot be an enforcement path.

A narrower independent check is now prepared: `sh /Users/jeroenmol/workspace/agent-kit/experiments/check-readonly-reviewer.sh`. It creates a fresh mode-700 fixture containing only `reviewed-source.txt` with a sentinel, then uses the installed CLI's documented `exec --sandbox read-only` mode. The model is instructed to make two concrete writes to that source file, first through `/bin/sh` and then through Python, plus a disposable `/tmp` cache write. JSONL is parsed only in memory for `command_execution` events; raw events and stderr are not retained. The sanitized report accepts an attempted route only when the command event contains its expected executable, target, and marker. It reports a route denial only when that command exits nonzero and its tool output includes a recognized permission/sandbox denial; arbitrary errors are unproven.

This is source-write feasibility evidence only. A meaningful narrow result requires `narrow_readonly_source_write=observed`, which in turn requires `source_unchanged=true` plus both recognized shell and Python route denials. `enforcement_failure` reports a changed sentinel; any missing event, successful write route, or non-permission error is `unproven`. `cache_write=observed` means only that the marked temporary-cache command exited zero. None of these establish a reviewer identity, delivery restriction, network isolation, or alternate routes beyond the two write commands. The fixture has no Git remote and the prompt prohibits network and Git, but that prompt is not a network boundary. Do not treat a `completed` process status by itself as a denial result.

The first external run completed with an unchanged sentinel, but all route events were `not_observed` and all exits and denials were unavailable. That is **inconclusive**: the raw JSONL was deliberately discarded, so the result cannot distinguish no model tool call, a model refusal, or an event shape the parser did not recognize. It is not evidence that read-only sandboxing denied either route.

The targeted rerun resolved part of that ambiguity. It emitted four valid events (`thread.started`, `turn.started`, `item.completed`, and `turn.completed`); its only item was `agent_message`, classified as `completion_claim`. No command item appeared. The installed CLI's local feature inventory reports `shell_tool` and `unified_exec` stable, while `code_mode` is disabled. Its help says `--ignore-user-config` skips `$CODEX_HOME/config.toml`, but does not say it disables tools. This establishes only that the parser saw the normal envelope and the model completed without a recognized shell-command call; it does not identify why. The original wording may be ambiguous, but that is a hypothesis rather than a demonstrated cause.

Prompt revision 2 tested the wording hypothesis. It emitted the same four valid events and only an `agent_message`, this time classified as `tool_refusal`; no command item, route exit, cache event, or sentinel change occurred. This establishes that the model refused or was unable to make the requested shell calls, but not why. The read-only sandbox boundary was never exercised. No further prompt-only reviewer rerun is warranted.

Tool execution therefore remains unresolved despite external session startup passing. A future investigation must obtain direct evidence of effective tool exposure and invocation through the controlled app-server/protocol surface or another documented runtime control, before designing any additional source-write fixture. It must not assume that the CLI or runtime is unsupported from the refusal alone.

The installed 0.154.0 CLI can generate an experimental app-server JSON Schema without starting a model. That schema defines an `item/completed` wrapper at `method`/`params.item` and a camel-case `commandExecution` item with `command`, `exitCode`, and `aggregatedOutput`. The legacy exec JSONL shape already handled by the probe is `type: item.completed`, `item.type: command_execution`, with snake-case `exit_code` and `aggregated_output`. Both exact variants are normalized; any other envelope remains unrecognized and inconclusive.

The narrow probe emits only a schema digest: JSONL line and valid-object counts, allowlisted type and key identifiers, and a classified agent-message disposition (`tool_refusal`, `completion_claim`, `other`, or `absent`). It never retains messages, arguments, command text, tool output, raw events, or unknown key/value names. An unrecognized format remains `not_observed` and requires diagnosis, never a permission claim.

## Capability matrix

| Capability / assertion | Evidence | Result | Classification |
|---|---|---:|---|
| Explicit CLI entry point exists | CLI 0.154.0 help | Yes | interface only |
| Ordinary-prompt automatic activation (AC-02/IN-04) | nested activation exited 1 before model | No demonstration | blocked |
| Independent child session command exists | `codex exec fork --help` | Yes | interface only |
| Child role cannot recursively activate | no child can start in this host | No demonstration | blocked |
| Role isolation / single writer | no role-scoped policy API found in inspected CLI help | No actual role run | unproven |
| Reviewer source-write denial via shell/tool | external prompts produced no command item; revision 2 produced `tool_refusal` | No demonstration; sandbox never exercised | unproven; prompts/worktrees advisory |
| Builder delivery denial | same; no per-role delivery interceptor found in inspected CLI help | No demonstration | unproven |
| Protected effects (network, remote Git, external state) | sandbox modes exist but no actor/resource policy was executed | No demonstration | blocked; arbitrary process execution cannot be granted for this claim |
| Alternate-route resistance | no sandboxed role process ran | No demonstration | blocked |
| Workspace write | CLI advertises mode; activation did not run | No demonstration | blocked |
| Model selection | `--model` advertised; no invocation reached provider | Requested model unavailable in evidence | interface only |
| Reasoning effort selection | model catalogue lists `gpt-5.6-terra` medium; `exec --help` lacks a direct effort flag | No actual configuration/application | untested |
| Requested versus actual settings | doctor reports a configured model, not a run record | No | unavailable/unproven in current evidence |
| JSONL runtime events | external reviewer probes emitted four valid lifecycle/item events | Yes, but no command item | partial interface/runtime evidence |
| Usage/token/cost telemetry | doctor gives no per-run values; failed run emitted none | No | unavailable |

## Completed host-startup diagnostic

The current managed host cannot run this diagnostic. The user ran it from a normal local macOS Terminal with the same installed Codex CLI and existing sign-in:

```sh
sh /Users/jeroenmol/workspace/agent-kit/experiments/check-runtime.sh
```

The script creates one fresh mode-700 `mktemp` directory, runs a 120-second bounded `codex exec` using `gpt-5.6-terra` and `model_reasoning_effort="medium"`, and asks only for `READY` without tools. It uses the advertised `--ephemeral`, `--ignore-user-config`, `--model`, `--sandbox read-only`, `--output-last-message`, and global `--ask-for-approval never` surfaces. The installed CLI's `--strict-config` validates the effort configuration key before the model call. It does not alter config, profiles, auth, sandbox policy, or the checkout; it suppresses raw process output and retains only a five-line sanitized status in the temporary fixture. Its supervisor terminates the process group after timeout or cancellation, then force-kills it if it does not exit.

`host_startup_check=started` establishes only that an ephemeral noninteractive session can initialize and return the fixed response on that host. Any other status is enough to preserve the present blocker with a small, non-sensitive artifact. Neither outcome demonstrates automatic Agent Kit activation, AC-02, AC-12, role separation, or protected-effect enforcement; those remain the later fixture probes below.

OpenAI's [model guidance](https://developers.openai.com/api/docs/guides/latest-model) confirms that reasoning effort is a request configuration, while the exact command switches and config key above are verified against the installed CLI's `codex exec --help` because CLI releases can differ from public documentation.

## Required next probe boundary

The user's external host can initialize an ephemeral session. Before another fixture, inspect a controlled app-server/protocol path or another documented runtime surface to establish effective shell-tool exposure and invocation evidence without modifying real configuration. Pin the CLI, desktop app, model, effort configuration, and permission-profile files. Only after that boundary is understood should the adapter prove each of these separately:

1. First ordinary implementation prompt activates; a question and planning prompt do not; a forked child cannot become another root.
2. Reviewer source write is denied through shell, direct executable, and a second available tool; allowed temp/cache check output succeeds.
3. Builder delivery operation and a protected network/Git/external effect are denied through the same alternate routes.
4. Requested and actual model/effort plus runtime usage are captured, or each missing field is recorded as unavailable.

Until those probes pass, there is no supported automatic activation or AC-12 enforcement path.

## Readiness and end-to-end acceptance are separate gates

The external check closes one readiness question: this user has a host that can start the selected CLI. Automatic activation cannot be tested from the raw CLI because the Agent Kit bootstrap/adapter that would recognize an ordinary implementation request has not been implemented. That is a dependency, not evidence of activation.

Proceed in two gates: first identify or build a disposable, documented policy boundary that binds each role to files and effects and can disable network; then implement the bootstrap/adapter and run the ordinary-request activation fixture plus the role-specific alternate-route fixtures. The first gate is still blocked by the documented-interface gap above; the second must remain unclaimed until the integration exists.
