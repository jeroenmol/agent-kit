package evidence_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeroenmol/agent-kit/internal/evidence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspectEventsRejectsInvalidCompleteEnvelopes(t *testing.T) {
	t.Parallel()

	valid := `{"schema_version":1,"event_id":"evt","timestamp":"2026-09-17T12:00:00Z","event":"outcome.prepared","phase":"prepared","run_id":"run_12345678","operation_id":"op_12345678","state_generation":1,"owner_fence":1,"actor":"actor"}`
	for _, invalid := range []string{
		`{}`,
		strings.Replace(valid, `"phase":"prepared"`, `"phase":"invalid"`, 1),
		strings.Replace(valid, `"owner_fence":1`, `"owner_fence":"1"`, 1),
		strings.Replace(valid, `"state_generation":1`, `"state_generation":0`, 1),
		strings.Replace(valid, `2026-09-17T12:00:00Z`, `invalid`, 1),
		strings.Replace(valid, `2026-09-17T12:00:00Z`, `2026-09-17T13:00:00+01:00`, 1),
		strings.Replace(valid, `run_12345678`, `run_bad`, 1),
		strings.Replace(valid, `"actor":"actor"`, `"actor":null`, 1),
	} {
		directory := t.TempDir()
		path := filepath.Join(directory, "events.jsonl")
		require.NoError(t, os.WriteFile(path, []byte(invalid+"\n"), 0o600), "fixture should write event")
		assert.ErrorIs(t, evidence.InspectEvents(path), evidence.ErrCorruptEvent, "complete invalid envelope must be corruption")
	}
}

func TestOutcomeReferenceUsesCanonicalDigest(t *testing.T) {
	t.Parallel()

	directory, lease, generation := newLease(t)
	writer, err := evidence.NewWriter(directory, evidence.NewRedactor())
	require.NoError(t, err, "fixture writer should initialize")
	publication, err := writer.Publish(context.Background(), lease, generation, input(evidence.OutcomeSucceeded))
	require.NoError(t, err, "outcome should publish")
	bytes, err := os.ReadFile(filepath.Join(directory, publication.Artifact.RelativePath))
	require.NoError(t, err, "artifact should be readable")
	line := "content_sha256: " + publication.Artifact.SHA256 + "\n"
	canonical := strings.Replace(string(bytes), line, "", 1)
	sum := sha256.Sum256([]byte(canonical))
	assert.Equal(t, hex.EncodeToString(sum[:]), publication.Artifact.SHA256, "reference must use canonical bytes excluding digest field")
	assert.NotEqual(t, hex.EncodeToString(sha256Bytes(bytes)), publication.Artifact.SHA256, "complete file digest must not stand in for canonical digest")
}

func TestOutcomeFactsAreExplicitAndRedacted(t *testing.T) {
	t.Parallel()

	directory, lease, generation := newLease(t)
	writer, err := evidence.NewWriter(directory, evidence.NewRedactor("seeded-secret"))
	require.NoError(t, err, "fixture writer should initialize")
	in := input(evidence.OutcomeBlocked)
	in.Facts.CurrentRevision = "rev_12345678"
	in.Facts.ValidationRefs = []string{"check_seeded-secret"}
	publication, err := writer.Publish(context.Background(), lease, generation, in)
	require.NoError(t, err, "outcome should publish")
	bytes, err := os.ReadFile(filepath.Join(directory, publication.Artifact.RelativePath))
	require.NoError(t, err, "artifact should be readable")
	assert.Contains(t, string(bytes), `"current_revision":"rev_12345678"`, "present fact should persist")
	assert.Contains(t, string(bytes), `"review_refs":"not_reached"`, "absent lists should be explicit")
	assert.Contains(t, string(bytes), `"delivery_state":"unavailable"`, "absent scalar should be explicit")
	assert.NotContains(t, string(bytes), "seeded-secret", "fact references must be redacted")
}

func TestWriterStopsWhenCreatingArtifactAncestorCannotSync(t *testing.T) {
	t.Parallel()

	directory, lease, generation := newLease(t)
	writer, err := evidence.NewWriter(directory, evidence.NewRedactor(), evidence.WithAtomicHooks(evidence.AtomicHooks{
		BeforeDirSync: func() error { return errors.New("injected directory sync failure") },
	}))
	require.NoError(t, err, "fixture writer should initialize")
	_, err = writer.Publish(context.Background(), lease, generation, input(evidence.OutcomeFailed))
	require.ErrorIs(t, err, evidence.ErrEssentialPublication, "ancestor durability failure must block publication")
	_, statErr := os.Stat(filepath.Join(directory, "events.jsonl"))
	assert.ErrorIs(t, statErr, os.ErrNotExist, "failed artifact preparation must not append events")
}

func sha256Bytes(value []byte) []byte {
	sum := sha256.Sum256(value)
	return sum[:]
}
