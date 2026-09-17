package state

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type operationRecord struct {
	SchemaVersion      int      `json:"schema_version"`
	ID                 string   `json:"id"`
	RunID              string   `json:"run_id"`
	Intent             string   `json:"intent"`
	ExpectedGeneration uint64   `json:"expected_generation"`
	ExpectedFence      uint64   `json:"expected_fence"`
	State              RunState `json:"state"`
	ContentSHA256      string   `json:"content_sha256"`
}

// canonicalOperationBytes is deterministic, lexically key-sorted JSON for an
// operation record with content_sha256 omitted entirely and no trailing newline.
func canonicalOperationBytes(record operationRecord) ([]byte, error) {
	return json.Marshal(map[string]any{
		"expected_fence":      record.ExpectedFence,
		"expected_generation": record.ExpectedGeneration,
		"id":                  record.ID,
		"intent":              record.Intent,
		"run_id":              record.RunID,
		"schema_version":      record.SchemaVersion,
		"state":               record.State,
	})
}

func writeOperationIntent(directory string, update Update) error {
	record := operationRecord{
		SchemaVersion:      schemaVersion,
		ID:                 update.Operation.ID,
		RunID:              filepath.Base(directory),
		Intent:             update.Operation.Intent,
		ExpectedGeneration: update.ExpectedGeneration,
		ExpectedFence:      update.Operation.ExpectedFence,
		State:              update.State,
	}
	canonical, err := canonicalOperationBytes(record)
	if err != nil {
		return fmt.Errorf("encoding operation intent: %w", err)
	}
	digest := sha256.Sum256(canonical)
	record.ContentSHA256 = hex.EncodeToString(digest[:])
	contents, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encoding operation intent digest: %w", err)
	}
	contents = append(contents, '\n')
	operations, err := operationDirectory(directory)
	if err != nil {
		return err
	}
	return createImmutable(filepath.Join(operations, record.ID+".json"), contents)
}

func operationDirectory(directory string) (string, error) {
	artifacts, err := createContainedDirectory(directory, "artifacts")
	if err != nil {
		return "", err
	}
	return createContainedDirectory(artifacts, "operations")
}

func createContainedDirectory(parent, name string) (string, error) {
	path := filepath.Join(parent, name)
	if !containsPath(parent, path) {
		return "", fmt.Errorf("operation path escapes run: %w", ErrCorrupt)
	}
	created := false
	if err := os.Mkdir(path, 0o700); err == nil {
		created = true
	} else if !errors.Is(err, os.ErrExist) {
		return "", fmt.Errorf("creating operation directory: %w", err)
	}
	if err := rejectSymlink(path); err != nil {
		return "", err
	}
	if created {
		if err := syncDirectory(parent, nil); err != nil {
			return "", fmt.Errorf("syncing operation directory: %w", err)
		}
	}
	return path, nil
}

func createImmutable(path string, contents []byte) error {
	if _, err := os.Lstat(path); err == nil {
		if err := rejectNonRegular(path); err != nil {
			return err
		}
		existing, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("reading operation intent: %w", readErr)
		}
		if string(existing) != string(contents) {
			return fmt.Errorf("operation id collision: %w", ErrCorrupt)
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking operation intent: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".operation-*")
	if err != nil {
		return fmt.Errorf("creating operation temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Link(temporaryPath, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return createImmutable(path, contents)
		}
		return fmt.Errorf("publishing operation intent: %w", err)
	}
	if err := syncDirectory(filepath.Dir(path), nil); err != nil {
		return fmt.Errorf("syncing operation intent: %w", err)
	}
	return nil
}
