package evidence_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/jeroenmol/agent-kit/internal/evidence"
	"github.com/jeroenmol/agent-kit/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriterPublishesSanitizedOutcomeForEveryTerminalResult(t *testing.T) {
	t.Parallel()

	for _, state := range []evidence.OutcomeState{evidence.OutcomeSucceeded, evidence.OutcomeFailed, evidence.OutcomeCancelled, evidence.OutcomeBlocked} {
		t.Run(string(state), func(t *testing.T) {
			t.Parallel()
			runDirectory, lease, generation := newLease(t)
			writer, err := evidence.NewWriter(runDirectory, evidence.NewRedactor("seeded-secret"))
			require.NoError(t, err, "fixture writer should initialize")

			publication, err := writer.Publish(context.Background(), lease, generation, input(state))
			require.NoError(t, err, "terminal outcome should publish required evidence")
			assert.NotEmpty(t, publication.Artifact.SHA256, "outcome should have a digest")
			assert.NoError(t, evidence.InspectEvents(filepath.Join(runDirectory, "events.jsonl")), "events should remain parseable")

			persisted, readErr := os.ReadFile(filepath.Join(runDirectory, publication.Artifact.RelativePath))
			require.NoError(t, readErr, "outcome artifact should be readable")
			assert.NotContains(t, string(persisted), "seeded-secret", "artifact must not retain configured secrets")
			assert.Contains(t, string(persisted), "Usage: unavailable", "missing telemetry must stay unavailable")
			assertPersistedFilesDoNotContain(t, runDirectory, "seeded-secret")
		})
	}
}

func assertPersistedFilesDoNotContain(t *testing.T, directory, secret string) {
	t.Helper()
	err := filepath.Walk(directory, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(contents), secret) {
			return errors.New("persisted secret")
		}
		return nil
	})
	assert.NoError(t, err, "no persisted evidence channel should retain a configured secret")
}

func TestWriterBlocksOutcomeWhenStateCommitFails(t *testing.T) {
	t.Parallel()

	directory, lease, generation := newLease(t)
	writer, err := evidence.NewWriter(directory, evidence.NewRedactor())
	require.NoError(t, err, "fixture writer should initialize")
	_, err = writer.Publish(context.Background(), lease, generation-1, input(evidence.OutcomeSucceeded))
	require.ErrorIs(t, err, evidence.ErrEssentialPublication, "state reference failure must prevent a completion claim")
}

func TestWriterRejectsRunIDThatDoesNotMatchLease(t *testing.T) {
	t.Parallel()

	directory, lease, generation := newLease(t)
	writer, err := evidence.NewWriter(directory, evidence.NewRedactor())
	require.NoError(t, err, "fixture writer should initialize")
	input := input(evidence.OutcomeFailed)
	input.RunID = "run_87654321"
	_, err = writer.Publish(context.Background(), lease, generation, input)
	require.ErrorIs(t, err, evidence.ErrEssentialPublication, "mismatched run must be rejected before publication")
	_, statErr := os.Stat(filepath.Join(directory, "events.jsonl"))
	assert.ErrorIs(t, statErr, os.ErrNotExist, "rejected run must not append events")
}

func TestWriterRejectsStaleGenerationWithoutPublishing(t *testing.T) {
	t.Parallel()

	directory, lease, generation := newLease(t)
	writer, err := evidence.NewWriter(directory, evidence.NewRedactor())
	require.NoError(t, err, "fixture writer should initialize")
	_, err = writer.Publish(context.Background(), lease, generation, input(evidence.OutcomeFailed))
	require.NoError(t, err, "first attempt should publish")
	_, err = writer.Publish(context.Background(), lease, generation, input(evidence.OutcomeFailed))
	require.ErrorIs(t, err, evidence.ErrEssentialPublication, "stale operation must not publish")
	events, err := os.ReadFile(filepath.Join(directory, "events.jsonl"))
	require.NoError(t, err, "events should be readable")
	assert.Equal(t, 1, strings.Count(string(events), "\n"), "stale operation must not append an event")
}

func TestInspectEventsDetectsInterruptedTail(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	events := filepath.Join(directory, "events.jsonl")
	require.NoError(t, os.WriteFile(events, []byte("{\"event\":\"valid\"}\n{\"event\":\"seeded-secret"), 0o600), "fixture should create interrupted append")
	require.ErrorIs(t, evidence.InspectEvents(events), evidence.ErrCorruptEvent, "complete malformed prefix must block recovery")

}

func TestRecoverPartialTailUsesCommittedStateRange(t *testing.T) {
	t.Parallel()

	directory, lease, generation := newLease(t)
	writer, err := evidence.NewWriter(directory, evidence.NewRedactor("seeded-secret"))
	require.NoError(t, err, "fixture writer should initialize")
	_, err = writer.Publish(context.Background(), lease, generation, input(evidence.OutcomeBlocked))
	require.NoError(t, err, "fixture should establish a committed event range")
	events := filepath.Join(directory, "events.jsonl")
	file, err := os.OpenFile(events, os.O_WRONLY|os.O_APPEND, 0o600)
	require.NoError(t, err, "fixture should open events")
	_, err = file.WriteString("{\"secret\":\"seeded-secret")
	require.NoError(t, err, "fixture should append partial event")
	require.NoError(t, file.Close(), "fixture should close events")

	err = writer.RecoverPartialTail(context.Background(), lease, "op_87654321", "recovery actor", time.Date(2026, time.September, 17, 12, 1, 0, 0, time.UTC))
	require.NoError(t, err, "recovery should commit a replacement event range")
	assert.NoError(t, evidence.InspectEvents(events), "recovered event stream should be valid")
	quarantine, err := os.ReadFile(filepath.Join(directory, "quarantine", "op_87654321.jsonl"))
	require.NoError(t, err, "partial bytes should be quarantined")
	assert.NotContains(t, string(quarantine), "seeded-secret", "quarantine must be sanitized")
}

func TestRecoverPartialTailRejectsCorruptCommittedPrefix(t *testing.T) {
	t.Parallel()

	directory, lease, generation := newLease(t)
	writer, err := evidence.NewWriter(directory, evidence.NewRedactor())
	require.NoError(t, err, "fixture writer should initialize")
	_, err = writer.Publish(context.Background(), lease, generation, input(evidence.OutcomeBlocked))
	require.NoError(t, err, "fixture should establish committed evidence")
	events := filepath.Join(directory, "events.jsonl")
	contents, err := os.ReadFile(events)
	require.NoError(t, err, "fixture events should be readable")
	contents[0] = 'X'
	contents = append(contents, []byte("{\"partial\":")...)
	require.NoError(t, os.WriteFile(events, contents, 0o600), "fixture should corrupt committed prefix")

	err = writer.RecoverPartialTail(context.Background(), lease, "op_87654321", "recovery actor", time.Date(2026, time.September, 17, 12, 1, 0, 0, time.UTC))
	require.ErrorIs(t, err, evidence.ErrEssentialPublication, "corrupt committed prefix must block recovery")
	actual, readErr := os.ReadFile(events)
	require.NoError(t, readErr, "blocked recovery should preserve bytes")
	assert.Equal(t, contents, actual, "blocked recovery must not truncate corrupt evidence")
}

func TestRecoverPartialTailRejectsCommittedDigestMismatch(t *testing.T) {
	t.Parallel()

	directory, lease, generation := newLease(t)
	writer, err := evidence.NewWriter(directory, evidence.NewRedactor())
	require.NoError(t, err, "fixture writer should initialize")
	_, err = writer.Publish(context.Background(), lease, generation, input(evidence.OutcomeBlocked))
	require.NoError(t, err, "fixture should establish committed evidence")
	events := filepath.Join(directory, "events.jsonl")
	contents, err := os.ReadFile(events)
	require.NoError(t, err, "fixture events should be readable")
	changed := strings.Replace(string(contents), "2026-09-17", "2027-09-17", 1)
	require.NotEqual(t, string(contents), changed, "fixture must alter a valid committed event field")
	contents = []byte(changed + `{"partial":`)
	require.NoError(t, os.WriteFile(events, contents, 0o600), "fixture should alter committed event bytes")

	err = writer.RecoverPartialTail(context.Background(), lease, "op_87654321", "recovery actor", time.Date(2026, time.September, 17, 12, 1, 0, 0, time.UTC))
	require.ErrorIs(t, err, evidence.ErrEssentialPublication, "digest mismatch must block recovery")
	actual, readErr := os.ReadFile(events)
	require.NoError(t, readErr, "blocked recovery should preserve bytes")
	assert.Equal(t, contents, actual, "digest mismatch must not truncate evidence")
}

func TestWriterBoundsOutput(t *testing.T) {
	t.Parallel()

	directory, lease, generation := newLease(t)
	writer, err := evidence.NewWriter(directory, evidence.NewRedactor())
	require.NoError(t, err, "fixture writer should initialize")
	input := input(evidence.OutcomeFailed)
	input.Output = strings.Repeat("x", 6000)
	publication, err := writer.Publish(context.Background(), lease, generation, input)
	require.NoError(t, err, "bounded output should publish")
	persisted, err := os.ReadFile(filepath.Join(directory, publication.Artifact.RelativePath))
	require.NoError(t, err, "outcome artifact should be readable")
	assert.Contains(t, string(persisted), "[TRUNCATED]", "oversize output must be marked")
}

func TestWriterBoundsUTF8Output(t *testing.T) {
	t.Parallel()

	directory, lease, generation := newLease(t)
	writer, err := evidence.NewWriter(directory, evidence.NewRedactor())
	require.NoError(t, err, "fixture writer should initialize")
	input := input(evidence.OutcomeFailed)
	input.Output = strings.Repeat("€", 2000)
	publication, err := writer.Publish(context.Background(), lease, generation, input)
	require.NoError(t, err, "UTF-8 output should publish")
	persisted, err := os.ReadFile(filepath.Join(directory, publication.Artifact.RelativePath))
	require.NoError(t, err, "outcome artifact should be readable")
	assert.True(t, utf8.Valid(persisted), "bounded artifact must remain valid UTF-8")
}

func newLease(t *testing.T) (string, *state.Lease, uint64) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "state")
	project := t.TempDir()
	store, err := state.NewStore(root, project)
	require.NoError(t, err, "fixture store should initialize")
	created, err := store.CreateRun(context.Background(), state.CreateRunInput{RunID: "run_12345678", RequestID: "req_12345678", RepositoryID: "repo_12345678"})
	require.NoError(t, err, "fixture run should initialize")
	lease, err := store.AcquireOwner(context.Background(), created.RunID, "owner_12345678")
	require.NoError(t, err, "fixture owner should acquire run")
	t.Cleanup(func() { require.NoError(t, lease.Release(), "fixture owner should release run") })
	directory, err := lease.RunDirectory()
	require.NoError(t, err, "fixture run directory should resolve")
	return directory, lease, created.Generation + 1
}

func input(state evidence.OutcomeState) evidence.PublishInput {
	return evidence.PublishInput{
		RunID:           "run_12345678",
		OutcomeID:       "outcome_12345678",
		OperationID:     "op_12345678",
		StateGeneration: 3,
		OwnerFence:      1,
		Timestamp:       time.Date(2026, time.September, 17, 12, 0, 0, 0, time.UTC),
		Actor:           "synthetic actor seeded-secret",
		State:           state,
		Summary:         "synthetic result seeded-secret",
		Output:          "output seeded-secret",
		Provenance: evidence.Provenance{
			ToolkitContentSHA256: "abcdef",
			ToolkitDirty:         true,
			Definitions:          map[string]string{"definition": "abcdef"},
			RequestedModel:       "requested",
			ActualModel:          "actual",
			RequestedEffort:      "medium",
			ActualEffort:         "unavailable",
			Usage:                evidence.Value{Status: "unavailable"},
		},
	}
}
