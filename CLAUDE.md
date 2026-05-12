# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.

## Project-Specific Guidelines

## 1. Commands

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

## 2. Architecture

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

## 3. Testing

Uses both unit tests and property-based tests via `github.com/leanovate/gopter`.

Property tests tag format: `Feature: <spec-name>, Property N: <property_text>`

Test files live alongside source (`*_test.go` in same package). Test data in `internal/aws/tagpolicy/testdata/`
