---
schema_version: 1
kind: run
id: run_<opaque-id>
run_id: run_<opaque-id>
created_at: "<rfc3339-utc>"
published_generation: <generation>
observed_generation: <generation>
state: <run-state>
content_sha256: <sha256-without-this-field>
---

# Run <run-id>

This summary is human-readable only. `state.json` is authoritative for current
state. A reader reports `reconciliation_required` if its ID, observed generation
or state differs from the index; it never chooses the Markdown narrative.
