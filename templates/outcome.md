---
schema_version: 1
kind: outcome
id: outcome_<opaque-id>
run_id: run_<opaque-id>
created_at: "<rfc3339-utc>"
published_generation: <generation>
content_sha256: <sha256-without-this-field>
---

# Outcome <outcome-id>

State: <succeeded|failed|cancelled|blocked>

Record achieved work, factual validation and delivery state, material blockers,
and known limitations. Record toolkit revision and dirty-content SHA-256,
loaded-definition hashes, runtime/adapter versions, effective-config and base
identities, requested and actual settings, and unavailable telemetry explicitly.
Record achieved criteria, validation/review references, checkpoint and branch,
local-completion and delivery facts separately, plus cancellation, blockers,
and adapter probe/capability references where applicable.
State remains authoritative; this artifact is immutable.
