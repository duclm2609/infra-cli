# Requirements Document

## Introduction

This document specifies the requirements for a new `tag-policy` subcommand under the existing `infra aws` command. The command enables users to view effective tag policies for their AWS account, displaying tag keys and their allowed values through an interactive terminal user interface (TUI) with keyboard navigation.

## Glossary

- **Tag_Policy_Command**: The CLI subcommand `infra aws tag-policy` that displays effective tag policies
- **Tag_Policy**: An AWS Organizations policy that defines standardized tags and their allowed values for resources
- **Tag_Key**: A label or category name in a tag policy (e.g., "Environment", "CostCenter")
- **Tag_Value**: An allowed value for a specific tag key (e.g., "Production", "Development")
- **TUI**: Terminal User Interface - an interactive text-based interface for keyboard navigation
- **Effective_Policy**: The combined tag policy that applies to an account after inheritance from organizational units
- **Organizations_Client**: The AWS SDK client for interacting with AWS Organizations service

## Requirements

### Requirement 1: Tag Policy Retrieval

**User Story:** As a DevOps engineer, I want to retrieve the effective tag policy for my AWS account, so that I can understand which tags are required and their allowed values.

#### Acceptance Criteria

1. WHEN the user executes `infra aws tag-policy`, THE Tag_Policy_Command SHALL retrieve the effective tag policy from AWS Organizations
2. WHEN the `--profile` flag is provided, THE Tag_Policy_Command SHALL use the specified AWS profile for authentication
3. WHEN the `--region` flag is provided, THE Tag_Policy_Command SHALL use the specified region for the AWS Organizations API call
4. IF no tag policy exists for the account, THEN THE Tag_Policy_Command SHALL display a message indicating no tag policy is in effect
5. IF AWS credentials are invalid or expired, THEN THE Tag_Policy_Command SHALL return an appropriate authentication error

### Requirement 2: Tag Policy Parsing

**User Story:** As a DevOps engineer, I want the tag policy JSON to be parsed correctly, so that I can view tag keys and their allowed values in a structured format.

#### Acceptance Criteria

1. WHEN a tag policy is retrieved, THE Tag_Policy_Command SHALL parse the JSON policy content into structured tag key and value data
2. THE Tag_Policy_Command SHALL extract all tag keys defined in the policy
3. FOR EACH tag key, THE Tag_Policy_Command SHALL extract the list of allowed values
4. IF the policy JSON is malformed, THEN THE Tag_Policy_Command SHALL return a descriptive parsing error
5. WHEN a tag key has an enforced_for constraint, THE Tag_Policy_Command SHALL preserve this information for display

### Requirement 3: Interactive TUI Display

**User Story:** As a DevOps engineer, I want to navigate through tag keys and their values using keyboard controls, so that I can efficiently explore the tag policy.

#### Acceptance Criteria

1. WHEN tag policy data is available, THE TUI SHALL display a list of all tag keys
2. THE TUI SHALL support arrow key navigation (Up/Down) to move between tag keys
3. WHEN the user presses Enter or Space on a tag key, THE TUI SHALL toggle the expanded/collapsed state to show or hide allowed values
4. WHEN a tag key is expanded, THE TUI SHALL display all allowed values indented under the tag key
5. WHEN the user presses 'q', THE TUI SHALL exit and return to the terminal
6. THE TUI SHALL visually indicate which tag key is currently selected
7. THE TUI SHALL visually indicate whether a tag key is expanded or collapsed

### Requirement 4: Error Handling

**User Story:** As a DevOps engineer, I want clear error messages when something goes wrong, so that I can troubleshoot issues effectively.

#### Acceptance Criteria

1. IF the AWS Organizations API returns an AccessDeniedException, THEN THE Tag_Policy_Command SHALL display a message indicating insufficient permissions
2. IF the account is not part of an AWS Organization, THEN THE Tag_Policy_Command SHALL display a message indicating the account is not in an organization
3. IF a network error occurs, THEN THE Tag_Policy_Command SHALL display a message indicating connectivity issues
4. WHEN an error occurs, THE Tag_Policy_Command SHALL use the existing infraerrors package for consistent error formatting

### Requirement 5: Integration with Existing AWS Command

**User Story:** As a DevOps engineer, I want the tag-policy command to work seamlessly with existing AWS command patterns, so that I have a consistent CLI experience.

#### Acceptance Criteria

1. THE Tag_Policy_Command SHALL be registered as a subcommand of the existing `aws` command
2. THE Tag_Policy_Command SHALL inherit the `--profile` and `--region` flags from the parent aws command
3. THE Tag_Policy_Command SHALL use the existing `GetProfile()` and `GetRegion()` helper functions
4. THE Tag_Policy_Command SHALL use the existing `ValidateCredentials()` function before making API calls
5. THE Tag_Policy_Command SHALL follow the existing Cobra command pattern with `RunE` for error handling
