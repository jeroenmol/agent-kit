# Codex runtime probe results

**Status:** blocked; no runtime AC-02 or AC-12 criterion passed. Blocker and interface evidence were collected; Task 01 remains incomplete.

**Probe date:** 2026-09-17

**Evidence source:** [runtime probe](runtime-probe.md), command help, and sanitized excerpts of redacted `codex doctor --json` output retained in that probe document.

## Result summary

The installed target is a real Codex standalone CLI (`0.154.0`) on macOS arm64 and exposes a documented local interface: noninteractive execution, session fork, sandbox modes, JSON events, and an experimental app-server with TypeScript binding/schema generation. Doctor successfully handshook with the already-running Desktop app-server, while a nested standalone CLI execution failed to initialize its own in-process app-server client under the host sandbox. No model ran and no disposable source file was created.

The direct sandbox command also did not execute because it requires an existing permission profile. No documented isolated profile-root route was found in the inspected help, so no profile was created. These failures are useful evidence of missing host capability; they do not prove a security denial.

The user subsequently ran the bounded external startup diagnostic in a normal local Terminal. Its sanitized status was `host_startup_check=started`, `process_exit=0`, and `diagnostic=ready_response`, with requested `gpt-5.6-terra` and medium effort. This proves that this external host can initialize an ephemeral noninteractive session. It does not prove provider-resolved model or effort, Agent Kit automatic activation, or any permission outcome.

## Enforcement evidence by protected effect

| Required effect | Candidate mediator inspected | Actor/resource binding proven? | Alternate shell/tool result | Verdict |
|---|---|---:|---|---|
| Reviewer writes reviewed source | CLI read-only sandbox / permission profile | No role binding seen | two initial prompts emitted no command item; revision 2 emitted `tool_refusal` and no command item | unproven; sandbox never exercised |
| Builder delivers a change | CLI approval and sandbox options | No delivery-specific control found in inspected help | no actual builder/delivery probe | unproven |
| Network or external protected effect | sandbox modes / host network sandbox | No operation-specific grant executed | no role- or network-bound probe | blocked |
| Git history or remote mutation | CLI options | No Git-specific role control found in inspected help | no actual Git effect probe | unproven |

An Agent Kit adapter may use its own preflight policy to refuse an assignment it cannot mediate. It must treat arbitrary `process.execute` as denied for an actor whose possible effects cannot be bounded. That preflight is not evidence that Codex itself blocks an alternate route.

The installed CLI still exposes no documented actor- or resource-scoped permission control. `exec --sandbox read-only` is a shell-command sandbox mode, while direct `codex sandbox` requires either a named active-stack permission profile or a caller-supplied sandbox-state JSON. The latter can disable network, but the inspected interface has no documented disposable state/profile creation path. A prompt-driven reviewer write attempt would therefore remain cooperative and could not safely establish denial of network, delivery, or alternate tool routes.

For an incremental source-write check, [check-readonly-reviewer.sh](check-readonly-reviewer.sh) is ready for the user's external host. It requests actual shell and Python write commands against a fresh sentinel file under `exec --sandbox read-only`, consumes JSONL command-execution evidence without retaining raw logs, and reports the independent event, exit, recognized denial, and sentinel results. It reports `narrow_readonly_source_write=observed` only when both expected write commands have nonzero tool exits with recognized permission-denial output and the sentinel remains unchanged. It cannot satisfy AC-12 by itself: source-write behavior is one narrow feasibility result, and the cache command is only an allowed-temp-write observation.

Its first external run was inconclusive: process completion and an unchanged sentinel were observed, but no recognized route events, exits, denials, or cache event were present. Because raw JSONL was intentionally discarded, this cannot determine whether no tool was called, the model refused, or `exec --json` used a different event shape. The installed CLI's generated experimental app-server schema defines `method: item/completed` with `params.item.type: commandExecution` and camel-case command fields; the probe now also supports that exact schema variant alongside its legacy snake-case item-completed envelope. It reports only allowlisted schema identifiers and a classified agent-message disposition on a future run so that an unsupported format is diagnosable without exposing a transcript. This result is not a permission outcome.

The targeted diagnostic rerun emitted four valid legacy-envelope events and only an `agent_message` item, classified as `completion_claim`; no command item was present. The local CLI inventory lists stable `shell_tool` and `unified_exec` features. Its help documents that `--ignore-user-config` skips the user config, but does not state that it hides tools. Neither fact establishes why the model did not call one. The original wording may have been ambiguous, but that remained a hypothesis.

Prompt revision 2 explicitly required the three shell calls and specified an inability response. It emitted the same four valid events and only an `agent_message`, classified as `tool_refusal`; source remained unchanged and no command item, exit, cache event, or denial was observed. This means the model refused or could not make a tool call for an unknown reason. It does not demonstrate a read-only denial, tool absence, or unsupported runtime. The sandbox was never exercised, and no further prompt-only rerun is justified. The next investigation, if pursued, needs direct controlled app-server/protocol evidence of effective tool exposure and invocation before any new source-write fixture.

## Telemetry result

`codex exec --json` indicates an event-output option, but the failed invocation emitted no JSONL session events. `codex doctor --json` supplies installation and configuration diagnostics and reports the configured model, but no per-run token, tool-call, context, cached-token, or cost values. The telemetry fields required by TE-02/03 are therefore **unavailable** for the selected runtime in this host evidence. They must never be represented as zero or estimated without a documented estimate.

## Decision impact

The investigation selects a narrow candidate target for future validation: native standalone Codex CLI 0.154.0 on macOS arm64, with its app-server protocol considered experimental. It does not satisfy IN-04 or AC-12 and does not establish a cross-platform claim. Runtime-independent foundations (Tasks 04–06) are separately authorized by the plan continuation decision; Task 04 is implemented, but that does not enable runtime-dependent setup, routing, assignment, or enforcement work.

The next decision gate is a documented disposable policy boundary that can be exercised on the user's external host without changing a real profile. It must bind roles to resources and effects, including network, before the exact permission fixtures can run. Do not mark Task 01 complete before those fixtures pass.

Host readiness and end-to-end acceptance now have distinct states. Host readiness passed for the user's normal Terminal. End-to-end activation remains blocked because the Agent Kit bootstrap/adapter has not been built, and permission acceptance remains blocked because a safe documented role/effect boundary has not been identified. Building that bootstrap may enable the activation fixture later; it cannot retroactively make the startup response an activation result.

## Completed external startup diagnostic

The user already ran `sh /Users/jeroenmol/workspace/agent-kit/experiments/check-runtime.sh` from a normal local macOS Terminal outside this managed host. It made a fresh temporary fixture, used explicit requested `gpt-5.6-terra` and medium effort, and returned the recorded `started` status without retaining raw model output.

That result resolved only the external-host session-initialization question. It is not evidence for automatic activation or any AC-12 permission claim.
