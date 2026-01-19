# Implementation Plan: AWS Tag Policy Command

## Overview

This plan implements the `infra aws tag-policy` subcommand following the existing patterns in the infra-cli project. The implementation is structured to build incrementally, with each task building on previous work and integrating immediately to avoid orphaned code.

## Tasks

- [x] 1. Create tag policy data models and parser
  - [x] 1.1 Create data models for TagPolicy and TagKey
    - Create `internal/aws/tagpolicy/models.go`
    - Define `TagPolicy` struct with `Tags []TagKey`
    - Define `TagKey` struct with `Name`, `Values`, and `EnforcedFor` fields
    - _Requirements: 2.1, 2.2, 2.3, 2.5_

  - [x] 1.2 Implement policy parser
    - Create `internal/aws/tagpolicy/parser.go`
    - Implement `PolicyParser` struct with `Parse(policyContent string) (*TagPolicy, error)`
    - Handle AWS tag policy JSON structure with `@@assign` operators
    - Return descriptive errors for malformed JSON
    - _Requirements: 2.1, 2.4_

  - [x] 1.3 Write property test for policy parsing round-trip
    - **Property 1: Policy Parsing Round-Trip**
    - Create `internal/aws/tagpolicy/parser_test.go`
    - Generate random TagPolicy structures
    - Serialize to AWS JSON format, parse back, verify equivalence
    - **Validates: Requirements 2.1, 2.2, 2.3, 2.5**

  - [x] 1.4 Write property test for malformed JSON errors
    - **Property 2: Malformed JSON Produces Errors**
    - Generate invalid JSON strings and non-conforming structures
    - Verify parser returns errors, not partial results
    - **Validates: Requirements 2.4**

- [x] 2. Implement tag policy service layer
  - [x] 2.1 Create tag policy service
    - Create `internal/aws/tagpolicy/service.go`
    - Implement `TagPolicyService` struct with Organizations client
    - Implement `NewTagPolicyService(profile, region string) (*TagPolicyService, error)`
    - Implement `GetEffectiveTagPolicy(ctx context.Context) (*TagPolicy, error)`
    - Use `organizations.DescribeEffectivePolicy` API with `TAG_POLICY` type
    - _Requirements: 1.1, 1.2, 1.3_

  - [x] 2.2 Implement AWS error mapping
    - Add error mapping function to convert AWS errors to infraerrors types
    - Handle `EffectivePolicyNotFoundException` → NoTagPolicyError
    - Handle `AWSOrganizationsNotInUseException` → NotInOrganizationError
    - Handle `AccessDeniedException` → AWSAPIError with permission message
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [x] 2.3 Write property test for configuration flags
    - **Property 3: Configuration Flags Passed Through**
    - Generate random profile and region values
    - Verify service is initialized with exact values
    - **Validates: Requirements 1.2, 1.3**

  - [x] 2.4 Write unit tests for error mapping
    - Test each AWS error code maps to correct infraerror type
    - Test error messages and suggestions are appropriate
    - _Requirements: 4.1, 4.2, 4.3_

- [x] 3. Checkpoint - Ensure service layer tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Implement TUI view layer
  - [x] 4.1 Create view model and state management
    - Create `internal/aws/tagpolicy/view.go`
    - Implement `ViewModel` struct with `TagKeys`, `SelectedIndex`, `ExpandedKeys`
    - Implement state initialization from `TagPolicy`
    - _Requirements: 3.1, 3.6, 3.7_

  - [x] 4.2 Implement keyboard navigation
    - Implement `HandleKeyPress(key rune) (quit bool)` method
    - Handle Up arrow: decrement selected index (bounded at 0)
    - Handle Down arrow: increment selected index (bounded at len-1)
    - Handle Enter/Space: toggle expanded state for selected key
    - Handle 'q': return quit=true
    - _Requirements: 3.2, 3.3, 3.5_

  - [x] 4.3 Implement render function
    - Implement `Render() string` method
    - Display all tag keys with selection indicator (e.g., `>` or highlight)
    - Display expansion indicator (e.g., `▶` collapsed, `▼` expanded)
    - When expanded, display values indented under the tag key
    - _Requirements: 3.1, 3.4, 3.6, 3.7_

  - [x] 4.4 Implement TUI run loop
    - Implement `Run() error` method
    - Set up terminal raw mode for keyboard input
    - Loop: render, read key, handle key, check quit
    - Restore terminal on exit
    - _Requirements: 3.1, 3.5_

  - [x] 4.5 Write property test for render contains all tag keys
    - **Property 4: Render Output Contains All Tag Keys**
    - Generate random TagPolicy with N keys
    - Verify rendered output contains all N key names
    - **Validates: Requirements 3.1**

  - [x] 4.6 Write property test for navigation state changes
    - **Property 5: Navigation State Changes Correctly**
    - Generate random view state with index I and N keys
    - Verify Down results in min(I+1, N-1)
    - Verify Up results in max(I-1, 0)
    - **Validates: Requirements 3.2**

  - [x] 4.7 Write property test for toggle reversibility
    - **Property 6: Toggle Expand/Collapse is Reversible**
    - Generate random view state
    - Toggle twice, verify state returns to original
    - **Validates: Requirements 3.3**

  - [x] 4.8 Write property test for expanded values display
    - **Property 7: Expanded Keys Show All Values**
    - Generate tag key with M values, set expanded
    - Verify rendered output contains all M values
    - **Validates: Requirements 3.4**

  - [x] 4.9 Write property test for render reflects view state
    - **Property 8: Render Reflects View State**
    - Generate random view state
    - Verify selection indicator at selected index
    - Verify expansion indicators match expanded state
    - **Validates: Requirements 3.6, 3.7**

- [x] 5. Checkpoint - Ensure TUI tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Create and wire tag-policy command
  - [x] 6.1 Create tag-policy command
    - Create `cmd/aws/tagpolicy.go`
    - Define `TagPolicyCmd` with Use="tag-policy", Short description, Long description
    - Implement `RunE` function that orchestrates service and view
    - _Requirements: 5.1, 5.5_

  - [x] 6.2 Implement command execution flow
    - Call `ValidateCredentials()` before API calls
    - Create `TagPolicyService` with `GetProfile()` and `GetRegion()`
    - Call `GetEffectiveTagPolicy()` to retrieve policy
    - Handle no-policy case with appropriate message
    - Create `TagPolicyView` and call `Run()` for interactive display
    - _Requirements: 1.1, 1.4, 5.3, 5.4_

  - [x] 6.3 Register command with parent aws command
    - Add `AWSCmd.AddCommand(TagPolicyCmd)` in `init()` function
    - Verify command inherits `--profile` and `--region` flags
    - _Requirements: 5.1, 5.2_

  - [x] 6.4 Write unit tests for command registration
    - Verify TagPolicyCmd is in AWSCmd.Commands()
    - Verify command has correct Use and Short values
    - Verify flags are accessible
    - _Requirements: 5.1, 5.2_

- [x] 7. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- All tasks are required for comprehensive implementation
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties using gopter
- Unit tests validate specific examples and edge cases
- The TUI implementation uses raw terminal mode for keyboard input; consider using a library like `golang.org/x/term` for cross-platform support
