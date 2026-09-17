package state

import (
	"context"
	"fmt"
)

// Prepare publishes immutable artifacts and complete prepared event bytes under
// the lease. It must return the references that the following state replacement
// commits. A preparation error leaves any published bytes inspectable and does
// not change authoritative state.
type Prepare func(runDirectory string) ([]ArtifactRef, *EventRange, error)

// Snapshot returns the current authoritative record while verifying that this
// lease still owns its fence. Callers use its generation as Commit's expected
// generation; a later change prevents preparation from running.
func (l *Lease) Snapshot(ctx context.Context) (*Run, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil, ErrStaleFence
	}
	if l.uncertain {
		return nil, ErrDurabilityUncertain
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := l.verifyLock(); err != nil {
		return nil, err
	}
	run, err := l.store.LoadRun(ctx, l.runID)
	if err != nil {
		return nil, err
	}
	if run.Owner.ID != l.owner.ID || run.Owner.Fence != l.owner.Fence {
		return nil, ErrStaleFence
	}
	return cloneRun(run), nil
}

// Commit serializes preparation and the authoritative state replacement under
// one owner lease. It prevents another writer from appending prepared evidence
// between this operation's preparation and commit.
func (l *Lease) Commit(ctx context.Context, update Update, prepare Prepare) (*Run, error) {
	if prepare == nil {
		return nil, fmt.Errorf("nil prepare function: %w", ErrCorrupt)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed || l.uncertain {
		return l.updateLocked(ctx, update)
	}
	if err := validateUpdate(update, l.owner.Fence); err != nil {
		return nil, err
	}
	if len(update.Operation.ArtifactRefs) != 0 || update.Operation.EventRange != nil {
		return nil, fmt.Errorf("commit operation already has publication references: %w", ErrCorrupt)
	}
	if err := l.verifyLock(); err != nil {
		return nil, err
	}
	run, err := l.store.LoadRun(ctx, l.runID)
	if err != nil {
		return nil, err
	}
	if run.Owner.ID != l.owner.ID || run.Owner.Fence != l.owner.Fence {
		return nil, ErrStaleFence
	}
	if run.Generation != update.ExpectedGeneration {
		return nil, ErrStaleGeneration
	}
	directory, err := l.RunDirectory()
	if err != nil {
		return nil, err
	}
	if err := writeOperationIntent(directory, update); err != nil {
		return nil, err
	}
	refs, eventRange, err := prepare(directory)
	if err != nil {
		return nil, err
	}
	update.Operation.ArtifactRefs = refs
	update.Operation.EventRange = eventRange
	return l.updateLocked(ctx, update)
}
