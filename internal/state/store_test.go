package state_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jeroenmol/agent-kit/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreCreateLoadAndUpdate(t *testing.T) {
	t.Parallel()

	store, input, _, _ := newStore(t)
	created, err := store.CreateRun(context.Background(), input)
	require.NoError(t, err, "creating a run should persist its first record")
	assert.Equal(t, uint64(1), created.Generation, "new run should start at generation one")
	assert.Equal(t, state.RunStateActive, created.State, "new run should be active")

	lease, err := store.AcquireOwner(context.Background(), input.RunID, "owner_12345678")
	require.NoError(t, err, "one owner should acquire the run")
	t.Cleanup(func() { require.NoError(t, lease.Release(), "owner should release its lock") })

	owned, err := store.LoadRun(context.Background(), input.RunID)
	require.NoError(t, err, "owned run should remain readable")
	updated, err := lease.Update(context.Background(), state.Update{
		ExpectedGeneration: owned.Generation,
		State:              state.RunStateWaiting,
		Operation: state.Operation{
			ID:                 "op_12345678",
			Intent:             "record synthetic wait",
			ExpectedGeneration: owned.Generation,
			ExpectedFence:      owned.Owner.Fence,
		},
	})
	require.NoError(t, err, "fenced update should atomically replace the record")
	assert.Equal(t, uint64(3), updated.Generation, "ownership and update should each advance generation")
	assert.Equal(t, state.RunStateWaiting, updated.State, "update should persist requested state")
	assert.Equal(t, "op_12345678", updated.Operation.ID, "operation identity should be retained")

	loaded, err := store.LoadRun(context.Background(), input.RunID)
	require.NoError(t, err, "updated run should reload")
	assert.Equal(t, updated, loaded, "reload should return the authoritative replacement")
}

func TestStorePreservesUnknownOptionalFields(t *testing.T) {
	t.Parallel()

	store, input, root, _ := newStore(t)
	created, err := store.CreateRun(context.Background(), input)
	require.NoError(t, err, "fixture should create run")
	bytes, err := json.Marshal(created)
	require.NoError(t, err, "fixture should encode run")
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(bytes, &raw), "fixture should decode run object")
	raw["future_optional"] = json.RawMessage(`{"value":true}`)
	bytes, err = json.Marshal(raw)
	require.NoError(t, err, "fixture should encode extended run")
	path := filepath.Join(root, "runs", input.RunID, "state.json")
	require.NoError(t, os.WriteFile(path, bytes, 0o600), "fixture should write extended run")

	lease, err := store.AcquireOwner(context.Background(), input.RunID, "owner_12345678")
	require.NoError(t, err, "compatible reader should acquire extended record")
	t.Cleanup(func() { require.NoError(t, lease.Release(), "owner should release lock") })
	owned, err := store.LoadRun(context.Background(), input.RunID)
	require.NoError(t, err, "extended record should load")
	_, err = lease.Update(context.Background(), updateFor(owned))
	require.NoError(t, err, "extended record should update")
	bytes, err = os.ReadFile(path)
	require.NoError(t, err, "updated bytes should remain readable")
	require.NoError(t, json.Unmarshal(bytes, &raw), "updated bytes should remain an object")
	assert.Contains(t, raw, "future_optional", "update must retain unknown optional fields")
}

func TestStoreRejectsUnsafePathsAndRecords(t *testing.T) {
	t.Parallel()

	t.Run("state root inside project", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		_, err := state.NewStore(filepath.Join(project, "state"), project)
		require.ErrorIs(t, err, state.ErrCorrupt, "target checkout must not contain durable state")
	})

	t.Run("run directory symlink", func(t *testing.T) {
		t.Parallel()

		store, input, root, _ := newStore(t)
		runs := filepath.Join(root, "runs")
		require.NoError(t, os.MkdirAll(runs, 0o700), "fixture should create runs directory")
		require.NoError(t, os.Symlink(t.TempDir(), filepath.Join(runs, input.RunID)), "fixture should create escaping link")

		_, err := store.LoadRun(context.Background(), input.RunID)
		require.ErrorIs(t, err, state.ErrCorrupt, "symlinked run directory must be rejected")
	})

	t.Run("unsupported and malformed state", func(t *testing.T) {
		t.Parallel()

		store, input, root, _ := newStore(t)
		_, err := store.CreateRun(context.Background(), input)
		require.NoError(t, err, "fixture should create run")
		path := filepath.Join(root, "runs", input.RunID, "state.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"schema_version":2}`), 0o600), "fixture should write newer schema")

		_, err = store.LoadRun(context.Background(), input.RunID)
		require.ErrorIs(t, err, state.ErrUnsupportedSchema, "unknown schema must remain visible")
		require.NoError(t, os.WriteFile(path, []byte(`{`), 0o600), "fixture should write malformed json")

		_, err = store.LoadRun(context.Background(), input.RunID)
		require.ErrorIs(t, err, state.ErrCorrupt, "malformed state must block reads")
	})
}

func TestStoreCreateRunSyncFailureIsUncertain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		hooks state.AtomicHooks
	}{
		{name: "root parent", hooks: state.AtomicHooks{BeforeRootSync: func() error { return errors.New("injected root sync failure") }}},
		{name: "runs parent", hooks: state.AtomicHooks{BeforeRunsSync: func() error { return errors.New("injected runs sync failure") }}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := filepath.Join(t.TempDir(), "state")
			project := filepath.Join(t.TempDir(), "project")
			require.NoError(t, os.MkdirAll(project, 0o700), "fixture should create project root")
			store, err := state.NewStore(root, project, state.WithAtomicHooks(tt.hooks))
			require.NoError(t, err, "store should initialize")
			_, err = store.CreateRun(context.Background(), state.CreateRunInput{
				RunID:        "run_12345678",
				RequestID:    "req_12345678",
				RepositoryID: "repo_12345678",
			})
			require.ErrorIs(t, err, state.ErrDurabilityUncertain, "post-creation parent sync failure must remain uncertain")
			_, statErr := os.Stat(filepath.Join(root, "runs"))
			require.NoError(t, statErr, "created directory must remain for recovery inspection")
		})
	}
}

func TestStoreRejectsPartialOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation json.RawMessage
	}{
		{name: "intent without id", operation: json.RawMessage(`{"intent":"prepared"}`)},
		{name: "preconditions without id", operation: json.RawMessage(`{"expected_fence":1}`)},
		{name: "artifact refs without id", operation: json.RawMessage(`{"artifact_refs":[{}]}`)},
		{name: "event range without id", operation: json.RawMessage(`{"event_range":{"relative_path":"events.jsonl"}}`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store, input, root, _ := newStore(t)
			created, err := store.CreateRun(context.Background(), input)
			require.NoError(t, err, "fixture should create run")
			bytes, err := json.Marshal(created)
			require.NoError(t, err, "fixture should encode run")
			var raw map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(bytes, &raw), "fixture should decode run")
			raw["operation"] = tt.operation
			bytes, err = json.Marshal(raw)
			require.NoError(t, err, "fixture should encode invalid run")
			require.NoError(t, os.WriteFile(filepath.Join(root, "runs", input.RunID, "state.json"), bytes, 0o600), "fixture should write invalid run")
			_, err = store.LoadRun(context.Background(), input.RunID)
			require.ErrorIs(t, err, state.ErrCorrupt, "partial operation must block reads")
		})
	}
}

func TestStoreExclusiveOwnershipAndFencing(t *testing.T) {
	t.Parallel()

	store, input, _, _ := newStore(t)
	_, err := store.CreateRun(context.Background(), input)
	require.NoError(t, err, "fixture should create run")

	const contenders = 12
	var successes atomic.Int32
	var lease *state.Lease
	var leaseMu sync.Mutex
	var group sync.WaitGroup
	for range contenders {
		group.Add(1)
		go func() {
			defer group.Done()
			candidate, acquireErr := store.AcquireOwner(context.Background(), input.RunID, "owner_12345678")
			if acquireErr == nil {
				leaseMu.Lock()
				lease = candidate
				leaseMu.Unlock()
				successes.Add(1)
				return
			}
			assert.ErrorIs(t, acquireErr, state.ErrOwned, "concurrent owner should fail safely")
		}()
	}
	group.Wait()
	require.Equal(t, int32(1), successes.Load(), "exactly one owner should acquire the run")
	require.NotNil(t, lease, "successful contender should return a lease")
	require.NoError(t, lease.Release(), "winning owner should release lock")

	next, err := store.AcquireOwner(context.Background(), input.RunID, "owner_87654321")
	require.NoError(t, err, "released owner may be replaced")
	t.Cleanup(func() { require.NoError(t, next.Release(), "replacement owner should release lock") })
	current, err := store.LoadRun(context.Background(), input.RunID)
	require.NoError(t, err, "replacement owner should persist")
	assert.Equal(t, uint64(2), current.Owner.Fence, "replacement owner must receive a newer fence")
}

func TestStoreAtomicReplacementFailures(t *testing.T) {
	t.Parallel()

	t.Run("before rename preserves previous record", func(t *testing.T) {
		t.Parallel()

		store, input, root, project := newStore(t)
		_, err := store.CreateRun(context.Background(), input)
		require.NoError(t, err, "fixture should create run")
		failing, err := state.NewStore(root, project, state.WithAtomicHooks(state.AtomicHooks{
			BeforeRename: func() error { return errors.New("injected pre-rename failure") },
		}))
		require.NoError(t, err, "failing store should initialize")
		lease, err := failing.AcquireOwner(context.Background(), input.RunID, "owner_12345678")
		require.Error(t, err, "ownership write should fail before rename")
		assert.Nil(t, lease, "failed ownership should not return a lease")
		loaded, loadErr := store.LoadRun(context.Background(), input.RunID)
		require.NoError(t, loadErr, "previous record should remain readable")
		assert.Equal(t, uint64(1), loaded.Generation, "pre-rename failure must preserve generation")
	})

	t.Run("after rename blocks further mutation", func(t *testing.T) {
		t.Parallel()

		store, input, root, project := newStore(t)
		_, err := store.CreateRun(context.Background(), input)
		require.NoError(t, err, "fixture should create run")
		var fail atomic.Bool
		failing, err := state.NewStore(root, project, state.WithAtomicHooks(state.AtomicHooks{
			AfterRename: func() error {
				if fail.Load() {
					return errors.New("injected post-rename failure")
				}
				return nil
			},
		}))
		require.NoError(t, err, "failing store should initialize")
		lease, err := failing.AcquireOwner(context.Background(), input.RunID, "owner_12345678")
		require.NoError(t, err, "fixture should acquire owner")
		fail.Store(true)
		owned, err := store.LoadRun(context.Background(), input.RunID)
		require.NoError(t, err, "fixture should load owned run")
		_, err = lease.Update(context.Background(), updateFor(owned))
		require.ErrorIs(t, err, state.ErrDurabilityUncertain, "post-rename sync failure must be uncertain")
		_, err = lease.Update(context.Background(), updateFor(owned))
		require.ErrorIs(t, err, state.ErrDurabilityUncertain, "uncertain lease must stop further writes")
		called := false
		_, err = lease.Commit(context.Background(), updateFor(owned), func(string) ([]state.ArtifactRef, *state.EventRange, error) {
			called = true
			return nil, nil, nil
		})
		require.ErrorIs(t, err, state.ErrDurabilityUncertain, "uncertain lease must reject commit")
		assert.False(t, called, "uncertain lease must not invoke preparation")
		require.ErrorIs(t, lease.Release(), state.ErrDurabilityUncertain, "uncertain owner must retain its lock")
		_, err = store.AcquireOwner(context.Background(), input.RunID, "owner_87654321")
		require.ErrorIs(t, err, state.ErrOwned, "post-rename uncertainty must block takeover")
		loaded, err := store.LoadRun(context.Background(), input.RunID)
		require.NoError(t, err, "physically replaced state should remain inspectable")
		assert.Equal(t, uint64(3), loaded.Generation, "post-rename state may be the new generation")
	})
}

func TestLeaseCommitSerializesPreparation(t *testing.T) {
	t.Parallel()

	store, input, _, _ := newStore(t)
	_, err := store.CreateRun(context.Background(), input)
	require.NoError(t, err, "fixture should create run")
	lease, err := store.AcquireOwner(context.Background(), input.RunID, "owner_12345678")
	require.NoError(t, err, "fixture should acquire owner")
	t.Cleanup(func() { require.NoError(t, lease.Release(), "owner should release lock") })
	owned, err := store.LoadRun(context.Background(), input.RunID)
	require.NoError(t, err, "fixture should load run")
	started := make(chan struct{})
	continuePrepare := make(chan struct{})
	committed := make(chan error, 1)
	go func() {
		_, commitErr := lease.Commit(context.Background(), updateFor(owned), func(string) ([]state.ArtifactRef, *state.EventRange, error) {
			close(started)
			<-continuePrepare
			return nil, nil, nil
		})
		committed <- commitErr
	}()
	<-started
	updated := make(chan error, 1)
	go func() {
		_, updateErr := lease.Update(context.Background(), updateFor(owned))
		updated <- updateErr
	}()
	select {
	case err := <-updated:
		t.Fatalf("update completed while preparation held lease: %v", err)
	default:
	}
	close(continuePrepare)
	require.NoError(t, <-committed, "prepared publication should commit")
	require.ErrorIs(t, <-updated, state.ErrStaleGeneration, "later update must observe committed generation")
}

func TestLeaseSnapshot(t *testing.T) {
	t.Parallel()

	t.Run("valid owner", func(t *testing.T) {
		t.Parallel()
		store, input, _, _ := newStore(t)
		_, err := store.CreateRun(context.Background(), input)
		require.NoError(t, err, "fixture should create run")
		lease, err := store.AcquireOwner(context.Background(), input.RunID, "owner_12345678")
		require.NoError(t, err, "fixture should acquire owner")
		t.Cleanup(func() { require.NoError(t, lease.Release(), "owner should release lock") })
		snapshot, err := lease.Snapshot(context.Background())
		require.NoError(t, err, "live owner should read snapshot")
		assert.Equal(t, input.RunID, snapshot.RunID, "snapshot should identify current run")
		assert.Equal(t, lease.Fence(), snapshot.Owner.Fence, "snapshot should retain lease fence")
	})

	t.Run("closed owner", func(t *testing.T) {
		t.Parallel()
		store, input, _, _ := newStore(t)
		_, err := store.CreateRun(context.Background(), input)
		require.NoError(t, err, "fixture should create run")
		lease, err := store.AcquireOwner(context.Background(), input.RunID, "owner_12345678")
		require.NoError(t, err, "fixture should acquire owner")
		require.NoError(t, lease.Release(), "owner should release lock")
		_, err = lease.Snapshot(context.Background())
		require.ErrorIs(t, err, state.ErrStaleFence, "closed lease must not return a snapshot")
	})

	t.Run("stale lock", func(t *testing.T) {
		t.Parallel()
		store, input, root, _ := newStore(t)
		_, err := store.CreateRun(context.Background(), input)
		require.NoError(t, err, "fixture should create run")
		lease, err := store.AcquireOwner(context.Background(), input.RunID, "owner_12345678")
		require.NoError(t, err, "fixture should acquire owner")
		require.NoError(t, os.Remove(filepath.Join(root, "runs", input.RunID, "lock")), "fixture should remove lock")
		_, err = lease.Snapshot(context.Background())
		require.ErrorIs(t, err, state.ErrStaleFence, "missing lock must reject snapshot")
	})
}

func TestRunSummaryReconciliation(t *testing.T) {
	t.Parallel()

	store, input, _, _ := newStore(t)
	_, err := store.CreateRun(context.Background(), input)
	require.NoError(t, err, "fixture should create run")
	lease, err := store.AcquireOwner(context.Background(), input.RunID, "owner_12345678")
	require.NoError(t, err, "fixture should acquire owner")
	t.Cleanup(func() { require.NoError(t, lease.Release(), "owner should release lock") })
	current, err := lease.Snapshot(context.Background())
	require.NoError(t, err, "fixture should snapshot current state")
	summary, err := state.RenderRunSummary(current)
	require.NoError(t, err, "authoritative run should render")
	status, err := store.ReconcileRunSummary(context.Background(), input.RunID, summary)
	require.NoError(t, err, "summary reconciliation should read state")
	assert.Equal(t, state.SummaryCurrent, status, "matching view should be current")
	for _, test := range []struct {
		name    string
		summary []byte
	}{
		{name: "body tamper", summary: []byte(strings.Replace(string(summary), "Current state", "Changed state", 1))},
		{name: "missing digest", summary: []byte(strings.Replace(string(summary), "content_sha256: ", "digest: ", 1))},
		{name: "bad digest", summary: []byte(strings.Replace(string(summary), "content_sha256: ", "content_sha256: deadbeef", 1))},
		{name: "duplicate field", summary: []byte(strings.Replace(string(summary), "content_sha256:", "content_sha256: duplicate\ncontent_sha256:", 1))},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result, reconcileErr := store.ReconcileRunSummary(context.Background(), input.RunID, test.summary)
			require.NoError(t, reconcileErr, "invalid summary should remain inspectable")
			assert.Equal(t, state.SummaryReconciliationRequired, result, "invalid summary must not override state")
		})
	}

	_, err = lease.Update(context.Background(), updateFor(current))
	require.NoError(t, err, "fixture should advance authoritative generation")
	status, err = store.ReconcileRunSummary(context.Background(), input.RunID, summary)
	require.NoError(t, err, "stale summary reconciliation should read state")
	assert.Equal(t, state.SummaryReconciliationRequired, status, "stale summary must not override state")
}

func TestLeaseCommitPreparationFailuresPreserveState(t *testing.T) {
	t.Parallel()

	store, input, _, _ := newStore(t)
	_, err := store.CreateRun(context.Background(), input)
	require.NoError(t, err, "fixture should create run")
	lease, err := store.AcquireOwner(context.Background(), input.RunID, "owner_12345678")
	require.NoError(t, err, "fixture should acquire owner")
	t.Cleanup(func() { require.NoError(t, lease.Release(), "owner should release lock") })
	owned, err := store.LoadRun(context.Background(), input.RunID)
	require.NoError(t, err, "fixture should load run")
	var sentinel string
	_, err = lease.Commit(context.Background(), updateFor(owned), func(directory string) ([]state.ArtifactRef, *state.EventRange, error) {
		intentBytes, readErr := os.ReadFile(filepath.Join(directory, "artifacts", "operations", "op_12345678.json"))
		require.NoError(t, readErr, "operation intent must exist before preparation")
		var intent map[string]any
		require.NoError(t, json.Unmarshal(intentBytes, &intent), "operation intent must be valid json")
		assert.Equal(t, "op_12345678", intent["id"], "intent should retain operation identity")
		assert.Equal(t, "test atomic replacement", intent["intent"], "intent should retain operation purpose")
		assert.NotEmpty(t, intent["content_sha256"], "intent should carry a content digest")
		digest, ok := intent["content_sha256"].(string)
		require.True(t, ok, "intent digest should be a string")
		delete(intent, "content_sha256")
		canonical, marshalErr := json.Marshal(intent)
		require.NoError(t, marshalErr, "canonical intent should encode")
		sum := sha256.Sum256(canonical)
		assert.Equal(t, hex.EncodeToString(sum[:]), digest, "intent digest must exclude its own field")
		sentinel = filepath.Join(directory, "prepared-sentinel")
		require.NoError(t, os.WriteFile(sentinel, []byte("prepared"), 0o600), "preparation should publish sentinel")
		return nil, nil, errors.New("injected preparation failure")
	})
	require.Error(t, err, "failed preparation should not commit state")
	bytes, readErr := os.ReadFile(sentinel)
	require.NoError(t, readErr, "prepared bytes should remain inspectable")
	assert.Equal(t, "prepared", string(bytes), "prepared bytes should not be discarded")
	unchanged, err := store.LoadRun(context.Background(), input.RunID)
	require.NoError(t, err, "state should remain readable")
	assert.Equal(t, owned.Generation, unchanged.Generation, "failed preparation must preserve generation")
	assert.Empty(t, unchanged.Operation.ID, "failed preparation must preserve operation")
	_, readErr = os.ReadFile(filepath.Join(filepath.Dir(sentinel), "artifacts", "operations", "op_12345678.json"))
	require.NoError(t, readErr, "failed preparation must retain operation intent")

	called := false
	wrongIntent := updateFor(unchanged)
	wrongIntent.Operation.Intent = "different operation"
	_, err = lease.Commit(context.Background(), wrongIntent, func(string) ([]state.ArtifactRef, *state.EventRange, error) {
		called = true
		return nil, nil, nil
	})
	require.ErrorIs(t, err, state.ErrCorrupt, "same operation id with different intent must be rejected")
	assert.False(t, called, "operation collision must not invoke preparation")

	_, err = lease.Commit(context.Background(), updateFor(unchanged), func(string) ([]state.ArtifactRef, *state.EventRange, error) {
		return []state.ArtifactRef{{}}, nil, nil
	})
	require.ErrorIs(t, err, state.ErrCorrupt, "invalid callback references must not replace state")
	unchanged, err = store.LoadRun(context.Background(), input.RunID)
	require.NoError(t, err, "state should remain readable after invalid references")
	assert.Equal(t, owned.Generation, unchanged.Generation, "invalid references must preserve generation")

	_, err = lease.Commit(context.Background(), updateFor(unchanged), func(string) ([]state.ArtifactRef, *state.EventRange, error) {
		return nil, nil, nil
	})
	require.NoError(t, err, "valid commit should work after failed preparations")
	require.NoError(t, lease.Release(), "owner should release before closed callback check")
	called = false
	_, err = lease.Commit(context.Background(), updateFor(unchanged), func(string) ([]state.ArtifactRef, *state.EventRange, error) {
		called = true
		return nil, nil, nil
	})
	require.ErrorIs(t, err, state.ErrStaleFence, "closed lease must reject commit")
	assert.False(t, called, "closed lease must not invoke preparation")
}

func updateFor(run *state.Run) state.Update {
	return state.Update{
		ExpectedGeneration: run.Generation,
		State:              state.RunStateWaiting,
		Operation: state.Operation{
			ID:                 "op_12345678",
			Intent:             "test atomic replacement",
			ExpectedGeneration: run.Generation,
			ExpectedFence:      run.Owner.Fence,
		},
	}
}

func newStore(t *testing.T) (*state.Store, state.CreateRunInput, string, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "state")
	project := filepath.Join(t.TempDir(), "project")
	require.NoError(t, os.MkdirAll(project, 0o700), "fixture should create project root")
	store, err := state.NewStore(root, project, state.WithClock(func() time.Time {
		return time.Date(2026, time.September, 17, 12, 0, 0, 0, time.UTC)
	}))
	require.NoError(t, err, "store should initialize outside project root")
	return store, state.CreateRunInput{
		RunID:        "run_12345678",
		RequestID:    "req_12345678",
		RepositoryID: "repo_12345678",
	}, root, project
}
