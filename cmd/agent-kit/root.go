package main

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const unavailableExitCode = 69

var (
	version = "devel"
	commit  = ""
	dirty   = ""
)

type unavailableError struct {
	command string
}

func (err unavailableError) Error() string {
	return fmt.Sprintf("agent-kit %s is unavailable: runtime integration has not been validated on this host", err.command)
}

func newRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:           "agent-kit",
		Short:         "Coordinate an evidence-backed local engineering workflow",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       buildIdentity(),
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	for _, name := range []string{"setup", "doctor", "init", "update"} {
		command.AddCommand(newUnavailableCommand(name))
	}
	command.AddCommand(newVersionCommand())

	return command
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "version",
		Short:         "Print build identity",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), buildIdentity())
			return err
		},
	}
}

func newUnavailableCommand(name string) *cobra.Command {
	return &cobra.Command{
		Use:           name,
		Short:         "Reserved lifecycle command",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return unavailableError{command: name}
		},
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}

	if errors.Is(err, pflag.ErrHelp) {
		return 0
	}

	var unavailable unavailableError
	if errors.As(err, &unavailable) {
		return unavailableExitCode
	}

	return 2
}

func buildIdentity() string {
	return buildIdentityFrom(version, commit, dirty, buildSettings())
}

func buildIdentityFrom(buildVersion, buildCommit, buildDirty string, settings map[string]string) string {
	if buildVersion == "" {
		buildVersion = "devel"
	}
	if buildCommit == "" {
		buildCommit = settings["vcs.revision"]
	}
	if buildDirty == "" {
		buildDirty = settings["vcs.modified"]
	}
	if buildCommit == "" {
		buildCommit = "unknown"
	}
	if buildDirty == "" {
		buildDirty = "unknown"
	}

	return fmt.Sprintf("%s (commit %s; dirty %s)", buildVersion, buildCommit, buildDirty)
}

func buildSettings() map[string]string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}

	settings := make(map[string]string, len(info.Settings))
	for _, setting := range info.Settings {
		if strings.HasPrefix(setting.Key, "vcs.") {
			settings[setting.Key] = setting.Value
		}
	}

	return settings
}
