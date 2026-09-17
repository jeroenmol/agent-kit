package state

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SummaryStatus reports whether a current run summary agrees with state.json.
// A summary is never used to select a state when the two disagree.
type SummaryStatus string

const (
	SummaryCurrent                SummaryStatus = "current"
	SummaryReconciliationRequired SummaryStatus = "reconciliation_required"
)

// RenderRunSummary renders the current human-readable projection of a run.
// Callers may publish it as an immutable artifact; state.json remains authority.
func RenderRunSummary(run *Run) ([]byte, error) {
	if err := validateRun(run); err != nil {
		return nil, err
	}
	contents := fmt.Sprintf("---\nschema_version: 1\nkind: run\nid: %s\nrun_id: %s\ncreated_at: %q\npublished_generation: %d\nobserved_generation: %d\nstate: %s\n---\n\n# Run %s\n\nCurrent state observed from `state.json`: %s.\n", run.RunID, run.RunID, run.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"), run.Generation, run.Generation, run.State, run.RunID, run.State)
	digest := sha256.Sum256([]byte(contents))
	return []byte(strings.Replace(contents, "---\n\n# Run", "content_sha256: "+hex.EncodeToString(digest[:])+"\n---\n\n# Run", 1)), nil
}

// ReconcileRunSummary compares a current summary with the authoritative index.
func (s *Store) ReconcileRunSummary(ctx context.Context, runID string, summary []byte) (SummaryStatus, error) {
	run, err := s.LoadRun(ctx, runID)
	if err != nil {
		return "", err
	}
	fields, canonical, err := parseSummary(summary)
	if err != nil {
		return SummaryReconciliationRequired, nil
	}
	digest := sha256.Sum256(canonical)
	if fields["content_sha256"] != hex.EncodeToString(digest[:]) || fields["schema_version"] != "1" || fields["kind"] != "run" || fields["id"] != run.RunID || fields["run_id"] != run.RunID || fields["state"] != string(run.State) {
		return SummaryReconciliationRequired, nil
	}
	if _, err := time.Parse(time.RFC3339, fields["created_at"]); err != nil {
		return SummaryReconciliationRequired, nil
	}
	published, err := strconv.ParseUint(fields["published_generation"], 10, 64)
	if err != nil || published != run.Generation {
		return SummaryReconciliationRequired, nil
	}
	observed, err := strconv.ParseUint(fields["observed_generation"], 10, 64)
	if err != nil || observed != run.Generation {
		return SummaryReconciliationRequired, nil
	}
	return SummaryCurrent, nil
}

func parseSummary(summary []byte) (map[string]string, []byte, error) {
	lines := strings.Split(string(summary), "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return nil, nil, fmt.Errorf("invalid summary frontmatter")
	}
	fields := make(map[string]string)
	canonicalLines := make([]string, 0, len(lines))
	canonicalLines = append(canonicalLines, lines[0])
	for index, line := range lines[1:] {
		if line == "---" {
			canonicalLines = append(canonicalLines, line)
			canonicalLines = append(canonicalLines, lines[index+2:]...)
			if !requiredSummaryFields(fields) {
				return nil, nil, fmt.Errorf("missing summary frontmatter")
			}
			return fields, []byte(strings.Join(canonicalLines, "\n")), nil
		}
		key, value, ok := strings.Cut(line, ": ")
		if !ok || key == "" {
			return nil, nil, fmt.Errorf("invalid summary frontmatter")
		}
		if _, exists := fields[key]; exists {
			return nil, nil, fmt.Errorf("duplicate summary frontmatter")
		}
		fields[key] = strings.Trim(value, `"`)
		if key != "content_sha256" {
			canonicalLines = append(canonicalLines, line)
		}
	}
	return nil, nil, fmt.Errorf("unterminated summary frontmatter")
}

func requiredSummaryFields(fields map[string]string) bool {
	for _, key := range []string{"schema_version", "kind", "id", "run_id", "created_at", "published_generation", "observed_generation", "state", "content_sha256"} {
		if fields[key] == "" {
			return false
		}
	}
	return validDigest(fields["content_sha256"])
}
