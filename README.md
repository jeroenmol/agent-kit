# Agent Kit

Agent Kit is a local workflow coordinator. The first executable is deliberately a CLI skeleton: it provides stable command grammar and build provenance while the runtime integration remains blocked on actual enforcement evidence.

## Commands

`agent-kit version` and `agent-kit --version` print the version, commit, and dirty-state identity embedded at build time. When built outside the Makefile, Go build metadata supplies Git identity when available.

`agent-kit setup`, `agent-kit doctor`, `agent-kit init`, and `agent-kit update` are reserved lifecycle commands. They currently return exit status 69 (`EX_UNAVAILABLE`) and do not execute runtime integration, install anything, write configuration, or modify a repository. `status` and `resume` are intentionally not registered yet.

Invalid command syntax returns exit status 2. Command results belong on standard output; diagnostics belong on standard error when the binary is invoked.

## Implemented foundations

The internal state package provides external run storage, exclusive ownership, fenced publication transactions, durable operation intents, and summary integrity checks. The evidence package publishes redacted outcomes and typed events, and quarantines interrupted uncommitted event tails after validating committed evidence. These packages are covered by synthetic failure and recovery tests; lifecycle CLI commands are not wired to them yet. Runtime activation and role isolation remain unproven.

## Development

Use these exact commands from the repository root:

```sh
make focused-test  # cmd/agent-kit tests
make test          # all tests
make fmt           # apply gofmt
make build         # build bin/agent-kit with Git provenance
make check         # verify formatting, vet, build, and test
make lint          # run golangci-lint when it is installed
```

The Makefile stores Go build and module caches under `.cache/`, so these commands do not depend on or alter the user's home-directory Go caches. Formatting uses the Go-distributed `gofmt`; linting remains an explicit optional command until its version is provisioned by setup. It uses Cobra for command parsing and help. Viper is intentionally absent because this skeleton has no configuration behavior to layer yet.
