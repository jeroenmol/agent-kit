package evidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jeroenmol/agent-kit/internal/state"
)

const (
	schemaVersion  = 1
	maxOutputBytes = 4096
)

var (
	opaqueID = regexp.MustCompile(`^[a-z]+_[A-Za-z0-9][A-Za-z0-9_-]{7,127}$`)
	// ErrPartialEventTail identifies an interrupted append that recovery must
	// inspect before another event is added.
	ErrPartialEventTail = errors.New("evidence: partial event tail")
	// ErrCorruptEvent identifies a complete event that cannot satisfy the
	// supported required envelope.
	ErrCorruptEvent = errors.New("evidence: corrupt event")
	// ErrEssentialPublication prevents callers from claiming a completed run.
	ErrEssentialPublication = errors.New("evidence: essential publication failed")
)

// OutcomeState is the factual result of this attempted run.
type OutcomeState string

const (
	OutcomeSucceeded OutcomeState = "succeeded"
	OutcomeFailed    OutcomeState = "failed"
	OutcomeCancelled OutcomeState = "cancelled"
	OutcomeBlocked   OutcomeState = "blocked"
)

// Value records an optional runtime-reported measurement. An unavailable value
// is explicit, avoiding an invented zero.
type Value struct {
	Status string `json:"status"`
	Value  *int64 `json:"value,omitempty"`
}

// Provenance captures only the identifying, non-secret context of an attempt.
type Provenance struct {
	ToolkitContentSHA256  string            `json:"toolkit_content_sha256"`
	ToolkitDirty          bool              `json:"toolkit_dirty"`
	ToolkitRevision       string            `json:"toolkit_revision,omitempty"`
	Definitions           map[string]string `json:"definitions"`
	RuntimeVersion        string            `json:"runtime_version,omitempty"`
	AdapterVersion        string            `json:"adapter_version,omitempty"`
	EffectiveConfigSHA256 string            `json:"effective_config_sha256,omitempty"`
	BaseSHA               string            `json:"base_sha,omitempty"`
	RequestedModel        string            `json:"requested_model,omitempty"`
	ActualModel           string            `json:"actual_model,omitempty"`
	RequestedEffort       string            `json:"requested_effort,omitempty"`
	ActualEffort          string            `json:"actual_effort,omitempty"`
	Usage                 Value             `json:"usage"`
}

// OutcomeFacts keeps outcome claims structured so unavailable or not-reached
// workflow stages are visible rather than inferred from narrative prose.
type OutcomeFacts struct {
	AchievedCriteria []string `json:"achieved_criteria"`
	ValidationRefs   []string `json:"validation_refs"`
	ReviewRefs       []string `json:"review_refs"`
	CurrentRevision  string   `json:"current_revision"`
	Checkpoint       string   `json:"checkpoint"`
	Branch           string   `json:"branch"`
	LocalCompletion  string   `json:"local_completion"`
	DeliveryState    string   `json:"delivery_state"`
	Blockers         []string `json:"blockers"`
	Cancellation     string   `json:"cancellation"`
	ProbeRefs        []string `json:"probe_refs"`
	CapabilityRefs   []string `json:"capability_refs"`
}

// PublishInput is a synthetic run result. Its IDs are supplied by the caller
// so a retry writes identical artifact and event identities.
type PublishInput struct {
	RunID           string
	OutcomeID       string
	OperationID     string
	StateGeneration uint64
	OwnerFence      uint64
	Timestamp       time.Time
	Actor           string
	State           OutcomeState
	Summary         string
	Output          string
	Provenance      Provenance
	Facts           OutcomeFacts
}

// ArtifactRef identifies the immutable outcome bytes published by a call.
type ArtifactRef struct {
	RelativePath string
	SHA256       string
}

// EventRange identifies the exact prepared event record.
type EventRange struct {
	StartOffset int64
	EndOffset   int64
	SHA256      string
}

// Publication identifies the evidence committed by the state index.
type Publication struct {
	Artifact ArtifactRef
	Events   EventRange
}

// Writer publishes immutable Markdown and prepared JSONL into one run
// directory. It does not mutate state.json.
type Writer struct {
	runDirectory string
	redactor     Redactor
	hooks        AtomicHooks
}

// AtomicHooks expose durability boundaries for deterministic failure tests.
type AtomicHooks struct {
	BeforeDirSync func() error
}

// Option configures a Writer.
type Option func(*Writer)

// WithAtomicHooks installs test-only durability hooks.
func WithAtomicHooks(hooks AtomicHooks) Option {
	return func(writer *Writer) { writer.hooks = hooks }
}

// NewWriter validates a run directory supplied by the state layer.
func NewWriter(runDirectory string, redactor Redactor, options ...Option) (*Writer, error) {
	if strings.TrimSpace(runDirectory) == "" {
		return nil, fmt.Errorf("empty run directory: %w", ErrEssentialPublication)
	}
	info, err := os.Lstat(runDirectory)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil, fmt.Errorf("unsafe run directory: %w", ErrEssentialPublication)
	}
	writer := &Writer{runDirectory: runDirectory, redactor: redactor}
	for _, option := range options {
		option(writer)
	}
	return writer, nil
}

// Publish writes prepared evidence and commits its references with one fenced
// state operation. A failure leaves any prepared bytes non-authoritative.
func (w *Writer) Publish(ctx context.Context, lease *state.Lease, expectedGeneration uint64, input PublishInput) (Publication, error) {
	if err := validateInput(input); err != nil {
		return Publication{}, err
	}
	if lease == nil || input.StateGeneration != expectedGeneration+1 || input.OwnerFence != lease.Fence() {
		return Publication{}, fmt.Errorf("invalid state publication preconditions: %w", ErrEssentialPublication)
	}
	run, err := lease.Snapshot(ctx)
	if err != nil {
		return Publication{}, fmt.Errorf("reading state publication preconditions: %w", errors.Join(ErrEssentialPublication, err))
	}
	if run.RunID != input.RunID || run.Generation != expectedGeneration {
		return Publication{}, fmt.Errorf("run does not match publication preconditions: %w", ErrEssentialPublication)
	}
	var publication Publication
	_, err = lease.Commit(ctx, state.Update{
		ExpectedGeneration: expectedGeneration,
		State:              runState(input.State),
		Operation: state.Operation{
			ID:                 input.OperationID,
			Intent:             "publish run outcome",
			ExpectedGeneration: expectedGeneration,
			ExpectedFence:      input.OwnerFence,
		},
	}, func(runDirectory string) ([]state.ArtifactRef, *state.EventRange, error) {
		if runDirectory != w.runDirectory {
			return nil, nil, errors.New("writer directory does not match lease")
		}
		artifact, err := w.writeOutcome(input)
		if err != nil {
			return nil, nil, err
		}
		events, err := w.appendPreparedEvent(input, artifact)
		if err != nil {
			return nil, nil, err
		}
		publication = Publication{Artifact: artifact, Events: events}
		return []state.ArtifactRef{{
				Kind:          "outcome",
				ID:            input.OutcomeID,
				RelativePath:  artifact.RelativePath,
				SchemaVersion: schemaVersion,
				ContentSHA256: artifact.SHA256,
			}}, &state.EventRange{
				RelativePath: "events.jsonl",
				StartOffset:  events.StartOffset,
				EndOffset:    events.EndOffset,
				Digest:       events.SHA256,
			}, nil
	})
	if err != nil {
		return Publication{}, fmt.Errorf("publishing evidence: %w", errors.Join(ErrEssentialPublication, err))
	}
	return publication, nil
}

func runState(outcome OutcomeState) state.RunState {
	switch outcome {
	case OutcomeSucceeded:
		return state.RunStateCompleted
	case OutcomeFailed:
		return state.RunStateFailed
	case OutcomeCancelled:
		return state.RunStateCancelled
	default:
		return state.RunStateWaiting
	}
}

func (w *Writer) writeOutcome(input PublishInput) (ArtifactRef, error) {
	path := filepath.Join(w.runDirectory, "artifacts", "outcome", input.OutcomeID+".md")
	if err := w.ensureDirectory(filepath.Join(w.runDirectory, "artifacts", "outcome")); err != nil {
		return ArtifactRef{}, err
	}
	contents := w.outcomeMarkdown(input)
	contentDigest := digest([]byte(strings.Replace(contents, "content_sha256: <content-sha256>\n", "", 1)))
	contents = strings.Replace(contents, "<content-sha256>", contentDigest, 1)
	if existing, err := os.ReadFile(path); err == nil {
		if digestBytes(existing) != digest([]byte(contents)) {
			return ArtifactRef{}, errors.New("immutable outcome already differs")
		}
		return ArtifactRef{RelativePath: filepath.ToSlash(filepath.Join("artifacts", "outcome", input.OutcomeID+".md")), SHA256: contentDigest}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return ArtifactRef{}, err
	}
	if err := atomicWrite(path, []byte(contents)); err != nil {
		return ArtifactRef{}, err
	}
	return ArtifactRef{RelativePath: filepath.ToSlash(filepath.Join("artifacts", "outcome", input.OutcomeID+".md")), SHA256: contentDigest}, nil
}

func (w *Writer) appendPreparedEvent(input PublishInput, artifact ArtifactRef) (EventRange, error) {
	path := filepath.Join(w.runDirectory, "events.jsonl")
	if err := rejectNonRegular(path); err != nil {
		return EventRange{}, err
	}
	if err := InspectEvents(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return EventRange{}, err
	}
	event := map[string]any{
		"schema_version":   schemaVersion,
		"event_id":         input.OperationID + ":outcome.prepared:1",
		"timestamp":        input.Timestamp.UTC().Format(time.RFC3339Nano),
		"event":            "outcome.prepared",
		"phase":            "prepared",
		"run_id":           input.RunID,
		"operation_id":     input.OperationID,
		"state_generation": input.StateGeneration,
		"owner_fence":      input.OwnerFence,
		"actor":            w.redactor.Redact(input.Actor),
		"outcome_id":       input.OutcomeID,
		"outcome_state":    input.State,
		"artifact":         artifact,
	}
	bytes, err := json.Marshal(event)
	if err != nil {
		return EventRange{}, err
	}
	bytes = append(bytes, '\n')
	if existing, err := os.ReadFile(path); err == nil {
		if eventRange, found, err := matchingEvent(existing, event["event_id"].(string), bytes); err != nil {
			return EventRange{}, err
		} else if found {
			return eventRange, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return EventRange{}, err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return EventRange{}, err
	}
	defer func() { _ = file.Close() }()
	start, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return EventRange{}, err
	}
	if _, err := file.Write(bytes); err != nil {
		return EventRange{}, err
	}
	if err := file.Sync(); err != nil {
		return EventRange{}, err
	}
	return EventRange{StartOffset: start, EndOffset: start + int64(len(bytes)), SHA256: digest(bytes)}, nil
}

func (w *Writer) outcomeMarkdown(input PublishInput) string {
	output := bound(w.redactor.Redact(input.Output))
	summary := w.redactor.Redact(input.Summary)
	definitions, _ := json.Marshal(input.Provenance.Definitions)
	facts, _ := json.Marshal(normalizeFacts(input.Facts))
	return fmt.Sprintf("---\nschema_version: 1\nkind: outcome\nid: %s\nrun_id: %s\ncreated_at: %q\npublished_generation: %d\ncontent_sha256: <content-sha256>\n---\n\n# Outcome %s\n\nState: %s\n\n%s\n\n## Outcome facts\n\n%s\n\n## Provenance\n\n- Toolkit content SHA-256: %s\n- Toolkit dirty: %t\n- Toolkit revision: %s\n- Loaded definitions: %s\n- Runtime version: %s\n- Adapter version: %s\n- Effective config SHA-256: %s\n- Base SHA: %s\n- Requested model: %s\n- Actual model: %s\n- Requested effort: %s\n- Actual effort: %s\n- Usage: %s\n\n## Bounded output\n\n```text\n%s\n```\n", input.OutcomeID, input.RunID, input.Timestamp.UTC().Format(time.RFC3339Nano), input.StateGeneration, input.OutcomeID, input.State, summary, w.redactor.Redact(string(facts)), unavailable(w.redactor.Redact(input.Provenance.ToolkitContentSHA256)), input.Provenance.ToolkitDirty, unavailable(w.redactor.Redact(input.Provenance.ToolkitRevision)), w.redactor.Redact(string(definitions)), unavailable(w.redactor.Redact(input.Provenance.RuntimeVersion)), unavailable(w.redactor.Redact(input.Provenance.AdapterVersion)), unavailable(w.redactor.Redact(input.Provenance.EffectiveConfigSHA256)), unavailable(w.redactor.Redact(input.Provenance.BaseSHA)), unavailable(w.redactor.Redact(input.Provenance.RequestedModel)), unavailable(w.redactor.Redact(input.Provenance.ActualModel)), unavailable(w.redactor.Redact(input.Provenance.RequestedEffort)), unavailable(w.redactor.Redact(input.Provenance.ActualEffort)), valueText(input.Provenance.Usage), output)
}

func normalizeFacts(facts OutcomeFacts) map[string]any {
	return map[string]any{
		"achieved_criteria": factList(facts.AchievedCriteria), "validation_refs": factList(facts.ValidationRefs), "review_refs": factList(facts.ReviewRefs), "current_revision": unavailable(facts.CurrentRevision), "checkpoint": unavailable(facts.Checkpoint), "branch": unavailable(facts.Branch), "local_completion": unavailable(facts.LocalCompletion), "delivery_state": unavailable(facts.DeliveryState), "blockers": factList(facts.Blockers), "cancellation": unavailable(facts.Cancellation), "probe_refs": factList(facts.ProbeRefs), "capability_refs": factList(facts.CapabilityRefs),
	}
}

func factList(values []string) any {
	if len(values) == 0 {
		return "not_reached"
	}
	return values
}

func unavailable(value string) string {
	if value == "" {
		return "unavailable"
	}
	return value
}

func valueText(value Value) string {
	if value.Status != "available" || value.Value == nil {
		return "unavailable"
	}
	return fmt.Sprintf("%d", *value.Value)
}

func validateInput(input PublishInput) error {
	if !validID(input.RunID, "run_") || !validID(input.OutcomeID, "outcome_") || !validID(input.OperationID, "op_") || input.Actor == "" || input.Timestamp.IsZero() || input.StateGeneration == 0 || input.OwnerFence == 0 {
		return fmt.Errorf("incomplete evidence input: %w", ErrEssentialPublication)
	}
	switch input.State {
	case OutcomeSucceeded, OutcomeFailed, OutcomeCancelled, OutcomeBlocked:
		return nil
	default:
		return fmt.Errorf("invalid outcome state: %w", ErrEssentialPublication)
	}
}

func validID(value, prefix string) bool {
	return strings.HasPrefix(value, prefix) && opaqueID.MatchString(value)
}

// InspectEvents validates complete JSONL records and reports a partial tail.
func InspectEvents(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	complete := bytes
	partial := len(bytes) > 0 && bytes[len(bytes)-1] != '\n'
	if partial {
		complete = bytes[:strings.LastIndex(string(bytes), "\n")+1]
	}
	if err := validateCompleteEvents(complete); err != nil {
		return err
	}
	if partial {
		return ErrPartialEventTail
	}
	return nil
}

func validateCompleteEvents(bytes []byte) error {
	for _, line := range strings.Split(strings.TrimSuffix(string(bytes), "\n"), "\n") {
		if line == "" {
			continue
		}
		var event eventEnvelope
		if err := json.Unmarshal([]byte(line), &event); err != nil || !event.valid() {
			return ErrCorruptEvent
		}
	}
	return nil
}

type eventEnvelope struct {
	SchemaVersion   int    `json:"schema_version"`
	EventID         string `json:"event_id"`
	Timestamp       string `json:"timestamp"`
	Event           string `json:"event"`
	Phase           string `json:"phase"`
	RunID           string `json:"run_id"`
	OperationID     string `json:"operation_id"`
	StateGeneration uint64 `json:"state_generation"`
	OwnerFence      uint64 `json:"owner_fence"`
	Actor           string `json:"actor"`
}

func (event eventEnvelope) valid() bool {
	if event.SchemaVersion != schemaVersion || event.EventID == "" || event.Event == "" || event.Actor == "" || !validID(event.RunID, "run_") || !validID(event.OperationID, "op_") || event.StateGeneration == 0 || event.OwnerFence == 0 {
		return false
	}
	if event.Phase != "prepared" && event.Phase != "committed" {
		return false
	}
	parsed, err := time.Parse(time.RFC3339Nano, event.Timestamp)
	if err != nil || !strings.HasSuffix(event.Timestamp, "Z") {
		return false
	}
	_, offset := parsed.Zone()
	return offset == 0
}

func bound(value string) string {
	value = strings.ToValidUTF8(value, "�")
	if len(value) <= maxOutputBytes {
		return value
	}
	end := maxOutputBytes
	for end > 0 && !utf8.RuneStart(value[end]) {
		end--
	}
	return value[:end] + "\n[TRUNCATED]"
}

func digest(value []byte) string { return digestBytes(value) }

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func (w *Writer) ensureDirectory(path string) error {
	relative, err := filepath.Rel(w.runDirectory, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return errors.New("artifact directory escapes run directory")
	}
	current := w.runDirectory
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			parent := filepath.Dir(current)
			if err := os.Mkdir(current, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
				return err
			}
			parentFile, err := os.Open(parent)
			if err != nil {
				return err
			}
			if w.hooks.BeforeDirSync != nil {
				if err := w.hooks.BeforeDirSync(); err != nil {
					_ = parentFile.Close()
					return err
				}
			}
			syncErr := parentFile.Sync()
			closeErr := parentFile.Close()
			if syncErr != nil {
				return syncErr
			}
			if closeErr != nil {
				return closeErr
			}
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return errors.New("artifact directory is not a regular directory")
		}
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("artifact directory is not a regular directory")
	}
	return nil
}

func rejectNonRegular(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return errors.New("evidence path is not a regular file")
	}
	return nil
}

func atomicWrite(path string, contents []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".outcome-*")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer func() { _ = os.Remove(temporary) }()
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(contents); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Link(temporary, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return errors.New("immutable outcome already exists")
		}
		return err
	}
	directory, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer func() { _ = directory.Close() }()
	return directory.Sync()
}

func matchingEvent(contents []byte, eventID string, expected []byte) (EventRange, bool, error) {
	var offset int64
	for _, line := range strings.SplitAfter(string(contents), "\n") {
		if line == "" {
			continue
		}
		if !strings.HasSuffix(line, "\n") {
			return EventRange{}, false, ErrPartialEventTail
		}
		var event struct {
			EventID string `json:"event_id"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return EventRange{}, false, ErrPartialEventTail
		}
		if event.EventID == eventID {
			if line != string(expected) {
				return EventRange{}, false, errors.New("stable event id already has different bytes")
			}
			return EventRange{StartOffset: offset, EndOffset: offset + int64(len(line)), SHA256: digest([]byte(line))}, true, nil
		}
		offset += int64(len(line))
	}
	return EventRange{}, false, nil
}
