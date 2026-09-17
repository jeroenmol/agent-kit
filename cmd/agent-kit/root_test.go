package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// package main is required because a command package cannot be imported by an
// external test package. The tests exercise the command's public CLI behavior.
func TestRootCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		arguments    []string
		wantStdout   string
		wantErr      string
		wantExitCode int
	}{
		{
			name:         "help",
			arguments:    []string{"--help"},
			wantStdout:   "Coordinate an evidence-backed local engineering workflow",
			wantExitCode: 0,
		},
		{
			name:         "version command",
			arguments:    []string{"version"},
			wantStdout:   "devel (commit",
			wantExitCode: 0,
		},
		{
			name:         "version flag",
			arguments:    []string{"--version"},
			wantStdout:   "agent-kit version devel",
			wantExitCode: 0,
		},
		{
			name:         "unavailable lifecycle command",
			arguments:    []string{"doctor"},
			wantErr:      "agent-kit doctor is unavailable",
			wantExitCode: unavailableExitCode,
		},
		{
			name:         "unknown command",
			arguments:    []string{"missing"},
			wantErr:      "unknown command \"missing\"",
			wantExitCode: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			command := newRootCommand()
			var stdout bytes.Buffer
			command.SetArgs(tt.arguments)
			command.SetOut(&stdout)
			command.SetErr(&stdout)

			err := command.Execute()
			if tt.wantErr != "" {
				require.Error(t, err, "command should return an error")
				assert.ErrorContains(t, err, tt.wantErr, "command should describe the failure")
			} else {
				require.NoError(t, err, "command should succeed")
			}
			assert.Equal(t, tt.wantExitCode, exitCode(err), "command should map to the documented exit code")
			assert.Contains(t, stdout.String(), tt.wantStdout, "command should print its expected output")
		})
	}
}

func TestUnavailableCommands(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"setup", "doctor", "init", "update"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			command := newRootCommand()
			command.SetArgs([]string{name})

			err := command.Execute()
			var unavailable unavailableError
			require.ErrorAs(t, err, &unavailable, "lifecycle command should remain unavailable")
			assert.Equal(t, name, unavailable.command, "unavailable result should name the invoked command")
		})
	}
}

func TestBuildIdentityFrom(t *testing.T) {
	t.Parallel()

	identity := buildIdentityFrom("", "", "", map[string]string{
		"vcs.revision": "abc123",
		"vcs.modified": "true",
	})

	assert.Equal(t, "devel (commit abc123; dirty true)", identity, "build settings should provide fallback provenance")
	assert.Equal(t, 0, exitCode(nil), "successful commands should have a zero exit code")
	assert.Equal(t, unavailableExitCode, exitCode(unavailableError{}), "unavailable errors should have a distinct exit code")
	assert.Equal(t, 2, exitCode(errors.New("usage error")), "other command errors should be usage errors")
}
