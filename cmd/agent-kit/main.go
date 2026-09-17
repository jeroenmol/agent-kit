package main

import (
	"fmt"
	"os"
)

func main() {
	command := newRootCommand()
	if err := command.Execute(); err != nil {
		if exitCode(err) == 0 {
			return
		}
		_, _ = fmt.Fprintln(command.ErrOrStderr(), err)
		os.Exit(exitCode(err))
	}
}
