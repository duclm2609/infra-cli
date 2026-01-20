# Implementation Plan: Infra CLI Tool

## Overview

This implementation plan builds the "infra" CLI tool incrementally, starting with project setup and core infrastructure, then adding AWS SSO authentication support. Each task builds on previous work, ensuring no orphaned code.

## Tasks

- [ ] 1. Project Setup and Core Structure
  - [x] 1.1 Initialize Go module and project structure
    - Create `go.mod` with module name `github.com/user/infra-cli`
    - Set up directory structure: `cmd/`, `internal/`, `pkg/`
    - Add `.gitignore` for Go projects
    - _Requirements: 2.1, 2.2, 2.4_

  - [x] 1.2 Add core dependencies
    - Add Cobra framework: `github.com/spf13/cobra`
    - Add Viper for config: `github.com/spf13/viper`
    - Add AWS SDK v2 packages
    - Add YAML library: `gopkg.in/yaml.v3`
    - Add table writer: `github.com/olekukonko/tablewriter`
    - _Requirements: 2.3, 4.1_

  - [x] 1.3 Create Makefile for build tasks
    - Add `build` target for local builds
    - Add `build-all` target for cross-platform builds (darwin/linux/windows)
    - Add `test` target for running tests
    - Add `clean` target
    - _Requirements: 2.5, 1.1, 1.2, 1.3_

- [ ] 2. Root Command and Global Flags
  - [x] 2.1 Implement root command
    - Create `cmd/root.go` with Cobra root command
    - Add --version flag with version info
    - Add --help flag (auto-provided by Cobra)
    - Display usage when run without arguments
    - _Requirements: 5.1, 5.3, 5.4, 5.5_

  - [x] 2.2 Add global flags
    - Add --verbose (-v) flag for detailed output
    - Add --quiet (-q) flag for minimal output
    - Add --output (-o) flag with choices: json, yaml, table
    - Set default output format to "table"
    - _Requirements: 8.3, 8.4, 9.4, 9.5_

  - [x] 2.3 Create main.go entry point
    - Create `main.go` that executes root command
    - Handle exit codes properly (0 for success, non-zero for errors)
    - _Requirements: 8.2, 8.5_

- [ ] 3. Configuration Management
  - [x] 3.1 Implement Config Manager
    - Create `internal/config/config.go`
    - Implement OS-appropriate config directory detection
    - Support loading from config file (~/.config/infra/config.yaml)
    - Support environment variables (INFRA_*)
    - _Requirements: 7.1, 7.2, 1.5_

  - [x] 3.2 Implement configuration precedence
    - Flags override environment variables
    - Environment variables override config file
    - Config file overrides defaults
    - _Requirements: 7.3, 7.4_

  - [x] 3.3 Write property test for configuration precedence
    - **Property 2: Configuration Precedence**
    - **Validates: Requirements 7.4**

- [ ] 4. Output Formatter
  - [x] 4.1 Implement Output Formatter interface
    - Create `internal/output/formatter.go`
    - Define OutputFormatter interface
    - Implement format selection logic
    - _Requirements: 9.1, 9.2, 9.3_

  - [x] 4.2 Implement JSON formatter
    - Format data as indented JSON
    - Ensure valid, parseable output
    - _Requirements: 9.1, 9.6_

  - [x] 4.3 Implement YAML formatter
    - Format data as YAML
    - Ensure valid, parseable output
    - _Requirements: 9.2, 9.7_

  - [x] 4.4 Implement Table formatter
    - Format data as aligned table
    - Support dynamic column widths
    - Align columns properly
    - _Requirements: 9.3, 9.8_

  - [x] 4.5 Write property tests for output formatters
    - **Property 3: JSON Output Round-Trip**
    - **Property 4: YAML Output Round-Trip**
    - **Property 5: Table Column Alignment**
    - **Validates: Requirements 9.6, 9.7, 9.8**

- [ ] 5. Checkpoint - Core Infrastructure
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 6. AWS Profile Manager
  - [x] 6.1 Implement Profile Manager
    - Create `internal/aws/profile/manager.go`
    - Read profiles from ~/.aws/config
    - Parse SSO configuration from profiles
    - _Requirements: 3.7_

  - [x] 6.2 Implement profile resolution logic
    - Check --profile flag first
    - Fall back to AWS_PROFILE env var
    - Default to "default" profile
    - _Requirements: 3.1, 3.2, 3.3_

  - [x] 6.3 Write property test for profile resolution
    - **Property 1: Profile Resolution Chain**
    - **Validates: Requirements 3.1, 3.2, 3.3**

- [ ] 7. AWS SSO Authenticator
  - [x] 7.1 Implement SSO Authenticator interface
    - Create `internal/aws/auth/sso.go`
    - Define SSOAuthenticator interface
    - Implement credential caching check
    - _Requirements: 3.4, 4.2_

  - [x] 7.2 Implement SSO login flow
    - Use AWS SDK v2 SSO OIDC client
    - Open browser for authentication
    - Handle token exchange
    - _Requirements: 3.4, 3.5_

  - [x] 7.3 Implement credential validation
    - Check if credentials are valid before operations
    - Auto-refresh expired SSO credentials
    - _Requirements: 6.5_

  - [x] 7.4 Write property test for credential validation
    - **Property 9: Credential Validation Before Operations**
    - **Validates: Requirements 6.5**

- [ ] 8. Error Handling
  - [x] 8.1 Implement error types
    - Create `internal/errors/errors.go`
    - Define AuthError, ConfigError types
    - Include error codes and remediation suggestions
    - _Requirements: 3.6, 8.1_

  - [x] 8.2 Implement error formatting
    - Format errors with category, details, and suggestions
    - Support verbose mode for stack traces
    - _Requirements: 8.1, 4.4_

  - [x] 8.3 Write property tests for error handling
    - **Property 6: Error Exit Codes**
    - **Property 7: Success Exit Codes**
    - **Property 11: AWS Error Message Clarity**
    - **Validates: Requirements 8.2, 8.5, 4.4**

- [ ] 9. Checkpoint - AWS Foundation
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 10. AWS Sub-Command
  - [x] 10.1 Implement AWS parent command
    - Create `cmd/aws/aws.go`
    - Add --profile flag
    - Add --region flag
    - Display available sub-commands when run alone
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

  - [x] 10.2 Wire AWS command to root
    - Register aws command as sub-command of root
    - Ensure nested command structure works
    - _Requirements: 5.2, 5.6_

  - [x] 10.3 Write property tests for AWS command flags
    - **Property 12: Region Flag Override**
    - **Validates: Requirements 4.3, 6.4**

- [ ] 11. CLI Behavior Properties
  - [x] 11.1 Write property test for help flag
    - **Property 8: Help Flag Availability**
    - **Validates: Requirements 5.4**

  - [x] 11.2 Write property test for output verbosity
    - **Property 10: Output Verbosity Control**
    - **Validates: Requirements 8.3, 8.4**

- [ ] 12. Path Handling
  - [x] 12.1 Implement OS-agnostic path utilities
    - Create `internal/util/path.go`
    - Use filepath.Join for path construction
    - Handle home directory expansion cross-platform
    - _Requirements: 1.4_

  - [x] 12.2 Write property test for path handling
    - **Property 13: OS-Agnostic Path Handling**
    - **Validates: Requirements 1.4**

- [ ] 13. Final Integration and Wiring
  - [x] 13.1 Wire all components together
    - Initialize config manager in root command
    - Pass output formatter to all commands
    - Set up SSO authenticator for AWS commands
    - _Requirements: 5.1, 5.2_

  - [x] 13.2 Add progress indicators
    - Show spinner for long-running operations
    - Display status updates during SSO login
    - _Requirements: 8.6_

- [x] 14. Final Checkpoint
  - Ensure all tests pass, ask the user if questions arise.
  - Verify cross-platform build works
  - Test SSO authentication flow manually

## Notes

- All tasks are required for comprehensive testing from the start
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- The implementation uses Go 1.24+ with modules
- Property-based testing uses `github.com/leanovate/gopter`
