# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build          # Build binary for current platform → bin/infra
make build-all      # Cross-compile for darwin/linux/windows
make test           # Run all tests with verbose output
make test-coverage  # Run tests + generate coverage.html
make lint           # Run golangci-lint (must be installed)
make deps           # Download and tidy dependencies
make clean          # Remove bin/ and coverage artifacts
```

Single test:
```bash
go test -v -run TestFunctionName ./internal/aws/tagpolicy/
```

Version is injected at build time via ldflags: `-X github.com/user/infra-cli/cmd/infra.Version=$(VERSION)`

## Architecture

`main.go` → `cmd/infra.Execute()` → Cobra root command → sub-commands

**Layer structure:**
- `cmd/infra/` — root command, global flags (`--verbose`, `--quiet`, `--output`), initializes `config.Manager` and `output.Formatter`
- `cmd/aws/` — AWS sub-commands (`whoami`, `login`, `profiles`, `tag-policy`); owns `--profile` and `--region` flags
- `internal/aws/auth/` — SSO authenticator wrapping AWS SDK v2; checks cached credentials, initiates browser-based SSO login
- `internal/aws/profile/` — reads `~/.aws/config`, resolves profile precedence (flag > `AWS_PROFILE` env > "default")
- `internal/aws/tagpolicy/` — Organizations API client, JSON policy parser, terminal TUI (keyboard-driven, no external TUI library)
- `internal/config/` — loads `~/.config/infra/config.yaml` (Linux), `~/Library/Application Support/infra/config.yaml` (macOS); config precedence: flags > env vars > config file > defaults
- `internal/output/` — unified formatter for `json`, `yaml`, `table` output formats
- `internal/errors/` — typed errors (`AuthError`, `ConfigError`, `AWSAPIError`, `InfraError`) with exit codes (1=auth, 2=config, 3=aws-api, 4=input, 5=internal)

**Adding a new top-level sub-command:** create `cmd/<name>/`, define a `*cobra.Command`, register it in `cmd/infra/root.go`'s `init()`.

**Adding a new AWS sub-command:** add to `cmd/aws/`, register in `cmd/aws/aws.go`'s `init()`.

## Testing

Uses both unit tests and property-based tests via `github.com/leanovate/gopter`.

Property tests tag format: `Feature: <spec-name>, Property N: <property_text>`

Test files live alongside source (`*_test.go` in same package). Test data in `internal/aws/tagpolicy/testdata/`.

## Specs

`.kiro/specs/` contains design documents with correctness properties, component interfaces, and error handling tables. Read these before implementing new features in their respective areas:
- `devops-cli-tool/` — overall CLI architecture, output formatting, profile resolution
- `aws-tag-policy/` — tag-policy TUI command design and property tests
