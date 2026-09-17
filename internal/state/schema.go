// Package state persists the authoritative current record for an Agent Kit run.
//
// It intentionally provides only the run-index and ownership kernel. Publishing
// immutable artifacts and events through the cross-artifact protocol belongs to
// the evidence layer.
package state

import (
	"encoding/json"
	"time"
)

const schemaVersion = 1

// RunState is the current lifecycle state recorded for a run. This package
// validates state names but does not expose workflow transitions; those require
// evidence checks supplied by a later layer.
type RunState string

const (
	RunStateActive    RunState = "active"
	RunStateWaiting   RunState = "waiting"
	RunStateCompleted RunState = "completed"
	RunStateFailed    RunState = "failed"
	RunStateCancelled RunState = "cancelled"
)

// Run is the authoritative versioned current record for one run.
type Run struct {
	SchemaVersion int       `json:"schema_version"`
	RunID         string    `json:"run_id"`
	RequestID     string    `json:"request_id"`
	RepositoryID  string    `json:"repository_id"`
	RerunOfRunID  string    `json:"rerun_of_run_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	State         RunState  `json:"state"`
	Generation    uint64    `json:"generation"`
	Owner         Owner     `json:"owner"`
	Operation     Operation `json:"operation"`

	extra map[string]json.RawMessage
}

// MarshalJSON retains optional fields written by a compatible newer writer so
// a fenced update does not silently discard information it cannot interpret.
func (r Run) MarshalJSON() ([]byte, error) {
	type fields Run
	bytes, err := json.Marshal(fields(r))
	if err != nil {
		return nil, err
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(bytes, &object); err != nil {
		return nil, err
	}
	for key, value := range r.extra {
		if _, exists := object[key]; !exists {
			object[key] = value
		}
	}
	return json.Marshal(object)
}

// UnmarshalJSON accepts and retains unknown optional fields for round trips.
func (r *Run) UnmarshalJSON(bytes []byte) error {
	type fields Run
	var decoded fields
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(bytes, &object); err != nil {
		return err
	}
	for _, key := range []string{"schema_version", "run_id", "request_id", "repository_id", "rerun_of_run_id", "created_at", "state", "generation", "owner", "operation"} {
		delete(object, key)
	}
	*r = Run(decoded)
	r.extra = object
	return nil
}

// Owner identifies the process that currently owns a run and its fencing token.
type Owner struct {
	ID          string    `json:"id"`
	Fence       uint64    `json:"fence"`
	AcquiredAt  time.Time `json:"acquired_at"`
	HeartbeatAt time.Time `json:"heartbeat_at"`
}

// Operation records the idempotent intent and preconditions for the mutation
// that produced the current generation. Later layers add immutable artifact and
// event references to their own operation protocol.
type Operation struct {
	ID                 string        `json:"id"`
	Intent             string        `json:"intent"`
	ExpectedGeneration uint64        `json:"expected_generation"`
	ExpectedFence      uint64        `json:"expected_fence"`
	ArtifactRefs       []ArtifactRef `json:"artifact_refs,omitempty"`
	EventRange         *EventRange   `json:"event_range,omitempty"`
}

// ArtifactRef identifies an immutable same-run artifact. The evidence layer
// verifies its bytes and digest before asking the index to reference it.
type ArtifactRef struct {
	Kind          string `json:"kind"`
	ID            string `json:"id"`
	RelativePath  string `json:"relative_path"`
	SchemaVersion int    `json:"schema_version"`
	ContentSHA256 string `json:"content_sha256"`
}

// EventRange identifies the prepared event bytes that an index generation
// commits. Event parsing and digest verification belong to the evidence layer.
type EventRange struct {
	RelativePath string `json:"relative_path"`
	StartOffset  int64  `json:"start_offset"`
	EndOffset    int64  `json:"end_offset"`
	Digest       string `json:"digest"`
}

// CreateRunInput names the stable lineage required to create a run.
type CreateRunInput struct {
	RunID        string
	RequestID    string
	RepositoryID string
	RerunOfRunID string
}

// Update describes one fenced replacement of a run record.
type Update struct {
	ExpectedGeneration uint64
	State              RunState
	Operation          Operation
}
