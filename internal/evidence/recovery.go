package evidence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jeroenmol/agent-kit/internal/state"
)

// RecoverPartialTail quarantines and removes only an unterminated event suffix
// beyond the event range committed by the authoritative state snapshot.
func (w *Writer) RecoverPartialTail(ctx context.Context, lease *state.Lease, operationID, actor string, timestamp time.Time) error {
	if lease == nil || !validID(operationID, "op_") || actor == "" || timestamp.IsZero() {
		return fmt.Errorf("invalid recovery input: %w", ErrEssentialPublication)
	}
	run, err := lease.Snapshot(ctx)
	if err != nil {
		return err
	}
	committedEnd := int64(0)
	if run.Operation.EventRange != nil {
		committedEnd = run.Operation.EventRange.EndOffset
	}
	return w.recover(ctx, lease, run, operationID, actor, timestamp, committedEnd)
}

func (w *Writer) recover(ctx context.Context, lease *state.Lease, run *state.Run, operationID, actor string, timestamp time.Time, committedEnd int64) error {
	_, returnError := lease.Commit(ctx, state.Update{
		ExpectedGeneration: run.Generation,
		State:              run.State,
		Operation:          state.Operation{ID: operationID, Intent: "recover interrupted event append", ExpectedGeneration: run.Generation, ExpectedFence: lease.Fence()},
	}, func(directory string) ([]state.ArtifactRef, *state.EventRange, error) {
		if directory != w.runDirectory {
			return nil, nil, errors.New("writer directory does not match lease")
		}
		eventsPath := filepath.Join(directory, "events.jsonl")
		if err := rejectNonRegular(eventsPath); err != nil {
			return nil, nil, err
		}
		bytes, err := os.ReadFile(eventsPath)
		if err != nil {
			return nil, nil, err
		}
		if len(bytes) == 0 || bytes[len(bytes)-1] == '\n' {
			return nil, nil, ErrPartialEventTail
		}
		start := strings.LastIndex(string(bytes), "\n") + 1
		if err := validateCompleteEvents(bytes[:start]); err != nil {
			return nil, nil, err
		}
		if int64(start) < committedEnd {
			return nil, nil, fmt.Errorf("partial event overlaps committed range: %w", ErrPartialEventTail)
		}
		if run.Operation.EventRange != nil {
			eventRange := run.Operation.EventRange
			if eventRange.RelativePath != "events.jsonl" || eventRange.StartOffset < 0 || eventRange.EndOffset < eventRange.StartOffset || eventRange.EndOffset > int64(len(bytes)) || digest(bytes[eventRange.StartOffset:eventRange.EndOffset]) != eventRange.Digest {
				return nil, nil, fmt.Errorf("committed event range does not match stream: %w", ErrCorruptEvent)
			}
		}
		if err := w.ensureDirectory(filepath.Join(directory, "quarantine")); err != nil {
			return nil, nil, err
		}
		quarantine := filepath.Join(directory, "quarantine", operationID+".jsonl")
		if err := atomicWrite(quarantine, []byte(w.redactor.Redact(string(bytes[start:])))); err != nil {
			return nil, nil, err
		}
		file, err := os.OpenFile(eventsPath, os.O_WRONLY, 0o600)
		if err != nil {
			return nil, nil, err
		}
		if err := file.Truncate(int64(start)); err != nil {
			_ = file.Close()
			return nil, nil, err
		}
		if err := file.Sync(); err != nil {
			_ = file.Close()
			return nil, nil, err
		}
		if err := file.Close(); err != nil {
			return nil, nil, err
		}
		return w.appendRecoveryEvent(eventsPath, run, operationID, actor, timestamp)
	})
	if returnError != nil {
		return fmt.Errorf("recovering event tail: %w", errors.Join(ErrEssentialPublication, returnError))
	}
	return nil
}

func (w *Writer) appendRecoveryEvent(path string, run *state.Run, operationID, actor string, timestamp time.Time) ([]state.ArtifactRef, *state.EventRange, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = file.Close() }()
	start, err := file.Seek(0, 2)
	if err != nil {
		return nil, nil, err
	}
	event := map[string]any{"schema_version": schemaVersion, "event_id": operationID + ":event.recovered:1", "timestamp": timestamp.UTC().Format(time.RFC3339Nano), "event": "event.recovered", "phase": "prepared", "run_id": run.RunID, "operation_id": operationID, "state_generation": run.Generation + 1, "owner_fence": run.Owner.Fence, "actor": w.redactor.Redact(actor)}
	bytes, err := json.Marshal(event)
	if err != nil {
		return nil, nil, err
	}
	bytes = append(bytes, '\n')
	if _, err := file.Write(bytes); err != nil {
		return nil, nil, err
	}
	if err := file.Sync(); err != nil {
		return nil, nil, err
	}
	return nil, &state.EventRange{RelativePath: "events.jsonl", StartOffset: start, EndOffset: start + int64(len(bytes)), Digest: digest(bytes)}, nil
}
