package state

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	// ErrNotFound indicates that no record exists for a valid run ID.
	ErrNotFound = errors.New("state: run not found")
	// ErrOwned indicates a live or unverified prior owner. Lease expiry never
	// grants takeover; explicit recovery belongs to the recovery workflow.
	ErrOwned = errors.New("state: run is already owned")
	// ErrStaleFence indicates a caller no longer owns the recorded fence.
	ErrStaleFence = errors.New("state: stale owner fence")
	// ErrStaleGeneration indicates the record changed since the caller read it.
	ErrStaleGeneration = errors.New("state: stale generation")
	// ErrUnsupportedSchema indicates bytes that this version must preserve.
	ErrUnsupportedSchema = errors.New("state: unsupported schema version")
	// ErrCorrupt indicates malformed or invariant-breaking persisted data.
	ErrCorrupt = errors.New("state: corrupt record")
	// ErrDurabilityUncertain indicates rename completed but the required parent
	// directory sync did not. The lease cannot make another mutation.
	ErrDurabilityUncertain = errors.New("state: durability uncertain")
)

var (
	opaqueID = regexp.MustCompile(`^[a-z]+_[A-Za-z0-9][A-Za-z0-9_-]{7,127}$`)
	idPrefix = regexp.MustCompile(`^[a-z]+_$`)
)

// AtomicHooks provide named fault-injection boundaries for persistence tests.
// Production construction leaves all hooks nil.
type AtomicHooks struct {
	BeforeRename   func() error
	AfterRename    func() error
	BeforeFileSync func() error
	BeforeDirSync  func() error
	BeforeRootSync func() error
	BeforeRunsSync func() error
}

// Option configures a Store.
type Option func(*Store)

// WithAtomicHooks installs named persistence boundaries, primarily for tests.
func WithAtomicHooks(hooks AtomicHooks) Option {
	return func(s *Store) { s.hooks = hooks }
}

// WithClock supplies a clock for deterministic tests.
func WithClock(now func() time.Time) Option {
	return func(s *Store) { s.now = now }
}

// Store owns one canonical external state root. projectRoot is required so a
// state root cannot be placed inside a target checkout.
type Store struct {
	root  string
	now   func() time.Time
	hooks AtomicHooks
}

// NewStore validates and canonicalizes a state root outside projectRoot.
func NewStore(stateRoot, projectRoot string, options ...Option) (*Store, error) {
	project, err := canonicalDirectory(projectRoot, false)
	if err != nil {
		return nil, fmt.Errorf("canonicalizing project root: %w", err)
	}
	root, err := canonicalDirectory(stateRoot, true)
	if err != nil {
		return nil, fmt.Errorf("canonicalizing state root: %w", err)
	}
	if containsPath(project, root) {
		return nil, fmt.Errorf("state root is inside project root: %w", ErrCorrupt)
	}
	s := &Store{root: root, now: time.Now}
	for _, option := range options {
		option(s)
	}
	if s.now == nil {
		return nil, fmt.Errorf("store clock is nil: %w", ErrCorrupt)
	}
	return s, nil
}

// CreateRun creates a new opaque run directory and its first durable record.
func (s *Store) CreateRun(ctx context.Context, input CreateRunInput) (*Run, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateCreateInput(input); err != nil {
		return nil, err
	}
	directory, err := s.runDirectory(input.RunID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureRunsDirectory(); err != nil {
		return nil, err
	}
	if err := os.Mkdir(directory, 0o700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("creating run: %w", ErrCorrupt)
		}
		return nil, fmt.Errorf("creating run directory: %w", err)
	}
	if err := syncDirectory(filepath.Dir(directory), s.hooks.BeforeRunsSync); err != nil {
		return nil, fmt.Errorf("syncing run directory creation: %w", errors.Join(ErrDurabilityUncertain, err))
	}
	now := s.now().UTC()
	run := &Run{
		SchemaVersion: schemaVersion,
		RunID:         input.RunID,
		RequestID:     input.RequestID,
		RepositoryID:  input.RepositoryID,
		RerunOfRunID:  input.RerunOfRunID,
		CreatedAt:     now,
		State:         RunStateActive,
		Generation:    1,
	}
	if err := s.replaceState(directory, run); err != nil {
		return nil, err
	}
	return cloneRun(run), nil
}

// LoadRun reads and validates the authoritative record for runID.
func (s *Store) LoadRun(ctx context.Context, runID string) (*Run, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	directory, err := s.runDirectory(runID)
	if err != nil {
		return nil, err
	}
	if err := s.validateRunsDirectory(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := rejectSymlink(directory); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	run, err := readRun(filepath.Join(directory, "state.json"))
	if err != nil {
		return nil, err
	}
	if run.RunID != runID {
		return nil, fmt.Errorf("run id does not match directory: %w", ErrCorrupt)
	}
	return run, nil
}

// AcquireOwner exclusively acquires a run owner lease. It never reclaims an
// existing lock: a possibly-live writer remains a safety blocker.
func (s *Store) AcquireOwner(ctx context.Context, runID, ownerID string) (*Lease, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !validID(runID, "run_") || !validID(ownerID, "owner_") {
		return nil, fmt.Errorf("invalid run or owner id: %w", ErrCorrupt)
	}
	directory, err := s.runDirectory(runID)
	if err != nil {
		return nil, err
	}
	if _, err := s.LoadRun(ctx, runID); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	lockPath := filepath.Join(directory, "lock")
	file, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, ErrOwned
		}
		return nil, fmt.Errorf("creating run lock: %w", err)
	}
	run, err := s.LoadRun(ctx, runID)
	if err != nil {
		_ = file.Close()
		_ = os.Remove(lockPath)
		return nil, err
	}
	now := s.now().UTC()
	owner := Owner{ID: ownerID, Fence: run.Owner.Fence + 1, AcquiredAt: now, HeartbeatAt: now}
	if err := json.NewEncoder(file).Encode(owner); err != nil {
		_ = file.Close()
		_ = os.Remove(lockPath)
		return nil, fmt.Errorf("writing run lock: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(lockPath)
		return nil, fmt.Errorf("syncing run lock: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(lockPath)
		return nil, fmt.Errorf("closing run lock: %w", err)
	}
	run.Owner = owner
	run.Generation++
	if err := s.replaceState(directory, run); err != nil {
		if errors.Is(err, ErrDurabilityUncertain) {
			return nil, err
		}
		_ = os.Remove(lockPath)
		return nil, err
	}
	return &Lease{store: s, runID: runID, owner: owner}, nil
}

// RunDirectory returns the validated canonical directory for a run. Writers
// may publish only relative paths contained within this directory.
func (s *Store) RunDirectory(runID string) (string, error) {
	directory, err := s.runDirectory(runID)
	if err != nil {
		return "", err
	}
	if err := s.validateRunsDirectory(); err != nil {
		return "", err
	}
	if err := rejectSymlink(directory); err != nil {
		return "", err
	}
	return directory, nil
}

// Lease owns state mutations until Release. It is safe for concurrent calls;
// updates are serialized and every one validates the persisted fence.
type Lease struct {
	store     *Store
	runID     string
	owner     Owner
	mu        sync.Mutex
	closed    bool
	uncertain bool
}

// Fence returns the fencing token required by an operation prepared by this
// lease. The token becomes stale when ownership is replaced.
func (l *Lease) Fence() uint64 {
	return l.owner.Fence
}

// RunDirectory returns the canonical publication directory owned by the lease.
func (l *Lease) RunDirectory() (string, error) {
	return l.store.RunDirectory(l.runID)
}

// Update atomically replaces the record after validating the caller's observed
// generation, current owner fence, and operation preconditions.
func (l *Lease) Update(ctx context.Context, update Update) (*Run, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.updateLocked(ctx, update)
}

func (l *Lease) updateLocked(ctx context.Context, update Update) (*Run, error) {
	if l.closed {
		return nil, ErrStaleFence
	}
	if l.uncertain {
		return nil, ErrDurabilityUncertain
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateUpdate(update, l.owner.Fence); err != nil {
		return nil, err
	}
	if err := l.verifyLock(); err != nil {
		return nil, err
	}
	run, err := l.store.LoadRun(ctx, l.runID)
	if err != nil {
		return nil, err
	}
	if run.Owner.Fence != l.owner.Fence || run.Owner.ID != l.owner.ID {
		return nil, ErrStaleFence
	}
	if run.Generation != update.ExpectedGeneration {
		return nil, ErrStaleGeneration
	}
	if run.Generation == ^uint64(0) {
		return nil, fmt.Errorf("generation overflow: %w", ErrCorrupt)
	}
	run.State = update.State
	run.Generation++
	run.Owner.HeartbeatAt = l.store.now().UTC()
	run.Operation = update.Operation
	directory, err := l.store.runDirectory(l.runID)
	if err != nil {
		return nil, err
	}
	if err := l.store.replaceState(directory, run); err != nil {
		if errors.Is(err, ErrDurabilityUncertain) {
			l.uncertain = true
		}
		return nil, err
	}
	return cloneRun(run), nil
}

// Release relinquishes the lock only when it still contains this exact owner.
func (l *Lease) Release() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	if l.uncertain {
		return ErrDurabilityUncertain
	}
	path, err := l.lockPath()
	if err != nil {
		return err
	}
	owner, err := readOwner(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrStaleFence
		}
		return err
	}
	if owner.ID != l.owner.ID || owner.Fence != l.owner.Fence {
		return ErrStaleFence
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("removing run lock: %w", err)
	}
	l.closed = true
	return nil
}

func (l *Lease) verifyLock() error {
	path, err := l.lockPath()
	if err != nil {
		return err
	}
	owner, err := readOwner(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrStaleFence
		}
		return err
	}
	if owner.ID != l.owner.ID || owner.Fence != l.owner.Fence {
		return ErrStaleFence
	}
	return nil
}

func (l *Lease) lockPath() (string, error) {
	directory, err := l.store.runDirectory(l.runID)
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "lock"), nil
}

func (s *Store) runDirectory(runID string) (string, error) {
	if !validID(runID, "run_") {
		return "", fmt.Errorf("invalid run id: %w", ErrCorrupt)
	}
	path := filepath.Join(s.root, "runs", runID)
	if !containsPath(s.root, path) {
		return "", fmt.Errorf("run path escapes state root: %w", ErrCorrupt)
	}
	return path, nil
}

func (s *Store) ensureRunsDirectory() error {
	path := filepath.Join(s.root, "runs")
	if !containsPath(s.root, path) {
		return fmt.Errorf("runs path escapes state root: %w", ErrCorrupt)
	}
	created := false
	if err := os.Mkdir(path, 0o700); err == nil {
		created = true
	} else if !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("creating runs directory: %w", err)
	}
	if err := rejectSymlink(path); err != nil {
		return fmt.Errorf("validating runs directory: %w", err)
	}
	if created {
		if err := syncDirectory(s.root, s.hooks.BeforeRootSync); err != nil {
			return fmt.Errorf("syncing runs directory creation: %w", errors.Join(ErrDurabilityUncertain, err))
		}
	}
	return nil
}

func (s *Store) validateRunsDirectory() error {
	path := filepath.Join(s.root, "runs")
	if !containsPath(s.root, path) {
		return fmt.Errorf("runs path escapes state root: %w", ErrCorrupt)
	}
	if err := rejectSymlink(path); err != nil {
		return fmt.Errorf("validating runs directory: %w", err)
	}
	return nil
}

func (s *Store) replaceState(directory string, run *Run) error {
	if err := validateRun(run); err != nil {
		return err
	}
	if err := rejectSymlink(directory); err != nil {
		return err
	}
	path := filepath.Join(directory, "state.json")
	if err := rejectNonRegular(path); err != nil {
		return err
	}
	contents, err := json.Marshal(run)
	if err != nil {
		return fmt.Errorf("encoding run state: %w", err)
	}
	contents = append(contents, '\n')
	temporary, err := os.CreateTemp(directory, ".state-*")
	if err != nil {
		return fmt.Errorf("creating state temporary file: %w", err)
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("setting state file permissions: %w", err)
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("writing state file: %w", err)
	}
	if s.hooks.BeforeFileSync != nil {
		if err := s.hooks.BeforeFileSync(); err != nil {
			_ = temporary.Close()
			return fmt.Errorf("before state file sync: %w", err)
		}
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("syncing state file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("closing state file: %w", err)
	}
	if s.hooks.BeforeRename != nil {
		if err := s.hooks.BeforeRename(); err != nil {
			return fmt.Errorf("before state rename: %w", err)
		}
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return fmt.Errorf("replacing state file: %w", err)
	}
	if s.hooks.AfterRename != nil {
		if err := s.hooks.AfterRename(); err != nil {
			return fmt.Errorf("after state rename: %w", errors.Join(ErrDurabilityUncertain, err))
		}
	}
	if s.hooks.BeforeDirSync != nil {
		if err := s.hooks.BeforeDirSync(); err != nil {
			return fmt.Errorf("before state directory sync: %w", errors.Join(ErrDurabilityUncertain, err))
		}
	}
	if err := syncDirectory(directory, nil); err != nil {
		return fmt.Errorf("syncing state directory: %w", errors.Join(ErrDurabilityUncertain, err))
	}
	return nil
}

func readRun(path string) (*Run, error) {
	if err := rejectNonRegular(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("opening state file: %w", err)
	}
	defer func() { _ = file.Close() }()
	var run Run
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&run); err != nil {
		return nil, fmt.Errorf("decoding state file: %w", errors.Join(ErrCorrupt, err))
	}
	if err := ensureEOF(decoder); err != nil {
		return nil, err
	}
	if err := validateRun(&run); err != nil {
		return nil, err
	}
	return &run, nil
}

func readOwner(path string) (Owner, error) {
	if err := rejectNonRegular(path); err != nil {
		return Owner{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		return Owner{}, fmt.Errorf("opening run lock: %w", err)
	}
	defer func() { _ = file.Close() }()
	var owner Owner
	if err := json.NewDecoder(file).Decode(&owner); err != nil {
		return Owner{}, fmt.Errorf("decoding run lock: %w", errors.Join(ErrCorrupt, err))
	}
	if !validID(owner.ID, "owner_") || owner.Fence == 0 {
		return Owner{}, fmt.Errorf("invalid run lock: %w", ErrCorrupt)
	}
	return owner, nil
}

func validateCreateInput(input CreateRunInput) error {
	if !validID(input.RunID, "run_") || !validID(input.RequestID, "req_") || !validID(input.RepositoryID, "repo_") {
		return fmt.Errorf("invalid run lineage: %w", ErrCorrupt)
	}
	if input.RerunOfRunID != "" && !validID(input.RerunOfRunID, "run_") {
		return fmt.Errorf("invalid rerun lineage: %w", ErrCorrupt)
	}
	if input.RerunOfRunID == input.RunID {
		return fmt.Errorf("run cannot rerun itself: %w", ErrCorrupt)
	}
	return nil
}

func validateUpdate(update Update, fence uint64) error {
	if !validRunState(update.State) || !validOperation(update.Operation) {
		return fmt.Errorf("invalid run update: %w", ErrCorrupt)
	}
	if update.Operation.ExpectedGeneration != update.ExpectedGeneration || update.Operation.ExpectedFence != fence {
		return fmt.Errorf("operation preconditions do not match update: %w", ErrCorrupt)
	}
	return nil
}

func validateRun(run *Run) error {
	if run.SchemaVersion != schemaVersion {
		return fmt.Errorf("got schema version %d: %w", run.SchemaVersion, ErrUnsupportedSchema)
	}
	if err := validateCreateInput(CreateRunInput{RunID: run.RunID, RequestID: run.RequestID, RepositoryID: run.RepositoryID, RerunOfRunID: run.RerunOfRunID}); err != nil {
		return err
	}
	if run.Generation == 0 || run.CreatedAt.IsZero() || !validRunState(run.State) {
		return fmt.Errorf("invalid run record: %w", ErrCorrupt)
	}
	if !zeroOperation(run.Operation) && !validOperation(run.Operation) {
		return fmt.Errorf("invalid run operation: %w", ErrCorrupt)
	}
	if run.Owner.ID == "" && run.Owner.Fence == 0 {
		return nil
	}
	if !validID(run.Owner.ID, "owner_") || run.Owner.Fence == 0 || run.Owner.AcquiredAt.IsZero() || run.Owner.HeartbeatAt.IsZero() {
		return fmt.Errorf("invalid run owner: %w", ErrCorrupt)
	}
	return nil
}

func zeroOperation(operation Operation) bool {
	return operation.ID == "" && operation.Intent == "" && operation.ExpectedGeneration == 0 && operation.ExpectedFence == 0 && len(operation.ArtifactRefs) == 0 && operation.EventRange == nil
}

func validOperation(operation Operation) bool {
	if !validID(operation.ID, "op_") || operation.ExpectedFence == 0 || strings.TrimSpace(operation.Intent) == "" {
		return false
	}
	for _, ref := range operation.ArtifactRefs {
		if !validArtifactRef(ref) {
			return false
		}
	}
	return operation.EventRange == nil || validEventRange(*operation.EventRange)
}

func validArtifactRef(ref ArtifactRef) bool {
	return strings.TrimSpace(ref.Kind) != "" && validID(ref.ID, "") && validRelativePath(ref.RelativePath, "artifacts") && ref.SchemaVersion == schemaVersion && validDigest(ref.ContentSHA256)
}

func validEventRange(eventRange EventRange) bool {
	return eventRange.RelativePath == "events.jsonl" && eventRange.StartOffset >= 0 && eventRange.EndOffset >= eventRange.StartOffset && validDigest(eventRange.Digest)
}

func validRelativePath(path, firstElement string) bool {
	if filepath.IsAbs(path) || filepath.Clean(path) != path {
		return false
	}
	parts := strings.Split(path, string(filepath.Separator))
	return len(parts) > 1 && parts[0] == firstElement && !strings.Contains(path, "..")
}

func validDigest(digest string) bool {
	if len(digest) != 64 {
		return false
	}
	_, err := hex.DecodeString(digest)
	return err == nil
}

func canonicalDirectory(path string, create bool) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("empty directory")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if create {
		if err := os.MkdirAll(abs, 0o700); err != nil {
			return "", err
		}
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("not a directory")
	}
	return resolved, nil
}

func rejectNonRegular(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("nonregular state path %q: %w", path, ErrCorrupt)
	}
	return nil
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("invalid state directory %q: %w", path, ErrCorrupt)
	}
	return nil
}

func containsPath(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}

func validID(value, prefix string) bool {
	return strings.HasPrefix(value, prefix) && opaqueID.MatchString(value)
}

func validRunState(state RunState) bool {
	switch state {
	case RunStateActive, RunStateWaiting, RunStateCompleted, RunStateFailed, RunStateCancelled:
		return true
	default:
		return false
	}
}

func ensureEOF(decoder *json.Decoder) error {
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("trailing state data: %w", ErrCorrupt)
	}
	return nil
}

func syncDirectory(path string, before func() error) error {
	if before != nil {
		if err := before(); err != nil {
			return err
		}
	}
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = directory.Close() }()
	return directory.Sync()
}

func cloneRun(run *Run) *Run {
	copy := *run
	return &copy
}

// NewID creates a cryptographically random opaque ID for callers that need a
// fresh run or operation identity.
func NewID(prefix string) (string, error) {
	if !idPrefix.MatchString(prefix) {
		return "", fmt.Errorf("invalid id prefix: %w", ErrCorrupt)
	}
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generating id: %w", err)
	}
	return prefix + hex.EncodeToString(bytes), nil
}
