package errors

import (
	"fmt"
	"strings"
)

// Exit codes for different error categories
const (
	ExitSuccess          = 0
	ExitAuthError        = 1
	ExitConfigError      = 2
	ExitAWSAPIError      = 3
	ExitInputError       = 4
	ExitInternalError    = 5
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	CategoryAuth     ErrorCategory = "Authentication"
	CategoryConfig   ErrorCategory = "Configuration"
	CategoryAWSAPI   ErrorCategory = "AWS API"
	CategoryInput    ErrorCategory = "Input Validation"
	CategoryInternal ErrorCategory = "Internal"
)

// InfraError is the base error type for all CLI errors
type InfraError struct {
	Category   ErrorCategory
	Code       string
	Message    string
	Details    string
	Suggestion string
	Cause      error
}

func (e *InfraError) Error() string {
	return fmt.Sprintf("%s: %s", e.Category, e.Message)
}

func (e *InfraError) Unwrap() error {
	return e.Cause
}

// ExitCode returns the appropriate exit code for this error
func (e *InfraError) ExitCode() int {
	switch e.Category {
	case CategoryAuth:
		return ExitAuthError
	case CategoryConfig:
		return ExitConfigError
	case CategoryAWSAPI:
		return ExitAWSAPIError
	case CategoryInput:
		return ExitInputError
	default:
		return ExitInternalError
	}
}

// Format returns a formatted error message
func (e *InfraError) Format(verbose bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Error: %s - %s\n", e.Category, e.Message))

	if e.Details != "" {
		sb.WriteString(fmt.Sprintf("\nDetails: %s\n", e.Details))
	}

	if e.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("\nSuggestion: %s\n", e.Suggestion))
	}

	if verbose && e.Cause != nil {
		sb.WriteString(fmt.Sprintf("\nCause: %v\n", e.Cause))
	}

	if !verbose {
		sb.WriteString("\nFor more information, run with --verbose flag.\n")
	}

	return sb.String()
}

// AuthError represents authentication-related errors
type AuthError struct {
	InfraError
	Profile string
}

// NewAuthError creates a new authentication error
func NewAuthError(code, message string, profile string, cause error) *AuthError {
	return &AuthError{
		InfraError: InfraError{
			Category: CategoryAuth,
			Code:     code,
			Message:  message,
			Cause:    cause,
		},
		Profile: profile,
	}
}

// NewSSOExpiredError creates an error for expired SSO session
func NewSSOExpiredError(profile string, cause error) *AuthError {
	return &AuthError{
		InfraError: InfraError{
			Category:   CategoryAuth,
			Code:       "SSO_EXPIRED",
			Message:    "SSO session expired",
			Details:    fmt.Sprintf("Your AWS SSO session for profile '%s' has expired.", profile),
			Suggestion: fmt.Sprintf("Run 'aws sso login --profile %s' to refresh your session, or let infra handle it automatically by running your command again.", profile),
			Cause:      cause,
		},
		Profile: profile,
	}
}

// NewCredentialsNotFoundError creates an error for missing credentials
func NewCredentialsNotFoundError(profile string, cause error) *AuthError {
	return &AuthError{
		InfraError: InfraError{
			Category:   CategoryAuth,
			Code:       "CREDENTIALS_NOT_FOUND",
			Message:    "AWS credentials not found",
			Details:    fmt.Sprintf("No valid AWS credentials found for profile '%s'.", profile),
			Suggestion: "Ensure your AWS credentials are configured. For SSO profiles, run 'aws sso login'. For IAM users, check your ~/.aws/credentials file.",
			Cause:      cause,
		},
		Profile: profile,
	}
}

// ConfigError represents configuration-related errors
type ConfigError struct {
	InfraError
	Source string // "file", "env", "flag"
}

// NewConfigError creates a new configuration error
func NewConfigError(code, message, source string, cause error) *ConfigError {
	return &ConfigError{
		InfraError: InfraError{
			Category: CategoryConfig,
			Code:     code,
			Message:  message,
			Cause:    cause,
		},
		Source: source,
	}
}

// NewProfileNotFoundError creates an error for missing profile
func NewProfileNotFoundError(profile string, cause error) *ConfigError {
	return &ConfigError{
		InfraError: InfraError{
			Category:   CategoryConfig,
			Code:       "PROFILE_NOT_FOUND",
			Message:    "Profile not found",
			Details:    fmt.Sprintf("The profile '%s' was not found in ~/.aws/config.", profile),
			Suggestion: "Check your AWS configuration file or use --profile to specify a valid profile. Run 'infra aws profiles' to list available profiles.",
			Cause:      cause,
		},
		Source: "file",
	}
}

// NewInvalidConfigError creates an error for invalid configuration
func NewInvalidConfigError(key, value, source string, cause error) *ConfigError {
	return &ConfigError{
		InfraError: InfraError{
			Category:   CategoryConfig,
			Code:       "INVALID_CONFIG",
			Message:    fmt.Sprintf("Invalid configuration value for '%s'", key),
			Details:    fmt.Sprintf("The value '%s' is not valid for configuration key '%s' (source: %s).", value, key, source),
			Suggestion: "Check the configuration documentation for valid values.",
			Cause:      cause,
		},
		Source: source,
	}
}

// AWSAPIError represents AWS API errors
type AWSAPIError struct {
	InfraError
	AWSErrorCode    string
	AWSErrorMessage string
	Service         string
	Operation       string
}

// NewAWSAPIError creates a new AWS API error
func NewAWSAPIError(awsCode, awsMessage, service, operation string, cause error) *AWSAPIError {
	return &AWSAPIError{
		InfraError: InfraError{
			Category:   CategoryAWSAPI,
			Code:       awsCode,
			Message:    fmt.Sprintf("%s: %s", awsCode, awsMessage),
			Details:    fmt.Sprintf("AWS %s operation '%s' failed.", service, operation),
			Suggestion: "Check your AWS permissions and ensure the resource exists.",
			Cause:      cause,
		},
		AWSErrorCode:    awsCode,
		AWSErrorMessage: awsMessage,
		Service:         service,
		Operation:       operation,
	}
}

// InputError represents input validation errors
type InputError struct {
	InfraError
	Field string
	Value string
}

// NewInputError creates a new input validation error
func NewInputError(field, value, message string) *InputError {
	return &InputError{
		InfraError: InfraError{
			Category:   CategoryInput,
			Code:       "INVALID_INPUT",
			Message:    message,
			Details:    fmt.Sprintf("Invalid value '%s' for field '%s'.", value, field),
			Suggestion: "Check the command help for valid input values.",
		},
		Field: field,
		Value: value,
	}
}

// NewInvalidRegionError creates an error for invalid AWS region
func NewInvalidRegionError(region string) *InputError {
	return &InputError{
		InfraError: InfraError{
			Category:   CategoryInput,
			Code:       "INVALID_REGION",
			Message:    fmt.Sprintf("Invalid AWS region: %s", region),
			Details:    fmt.Sprintf("The region '%s' is not a valid AWS region.", region),
			Suggestion: "Use a valid AWS region like 'us-east-1', 'us-west-2', 'eu-west-1', etc.",
		},
		Field: "region",
		Value: region,
	}
}

// NewInvalidOutputFormatError creates an error for invalid output format
func NewInvalidOutputFormatError(format string) *InputError {
	return &InputError{
		InfraError: InfraError{
			Category:   CategoryInput,
			Code:       "INVALID_OUTPUT_FORMAT",
			Message:    fmt.Sprintf("Invalid output format: %s", format),
			Details:    fmt.Sprintf("The output format '%s' is not supported.", format),
			Suggestion: "Use one of: json, yaml, table",
		},
		Field: "output",
		Value: format,
	}
}

// InternalError represents unexpected internal errors
type InternalError struct {
	InfraError
}

// NewInternalError creates a new internal error
func NewInternalError(message string, cause error) *InternalError {
	return &InternalError{
		InfraError: InfraError{
			Category:   CategoryInternal,
			Code:       "INTERNAL_ERROR",
			Message:    message,
			Details:    "An unexpected error occurred.",
			Suggestion: "Please report this issue with the --verbose output.",
			Cause:      cause,
		},
	}
}

// IsAuthError checks if an error is an authentication error
func IsAuthError(err error) bool {
	_, ok := err.(*AuthError)
	return ok
}

// IsConfigError checks if an error is a configuration error
func IsConfigError(err error) bool {
	_, ok := err.(*ConfigError)
	return ok
}

// IsAWSAPIError checks if an error is an AWS API error
func IsAWSAPIError(err error) bool {
	_, ok := err.(*AWSAPIError)
	return ok
}

// GetExitCode returns the exit code for an error
func GetExitCode(err error) int {
	if infraErr, ok := err.(*InfraError); ok {
		return infraErr.ExitCode()
	}
	if authErr, ok := err.(*AuthError); ok {
		return authErr.ExitCode()
	}
	if configErr, ok := err.(*ConfigError); ok {
		return configErr.ExitCode()
	}
	if awsErr, ok := err.(*AWSAPIError); ok {
		return awsErr.ExitCode()
	}
	if inputErr, ok := err.(*InputError); ok {
		return inputErr.ExitCode()
	}
	if internalErr, ok := err.(*InternalError); ok {
		return internalErr.ExitCode()
	}
	return ExitInternalError
}

// FormatError formats an error for display
func FormatError(err error, verbose bool) string {
	if infraErr, ok := err.(*InfraError); ok {
		return infraErr.Format(verbose)
	}
	if authErr, ok := err.(*AuthError); ok {
		return authErr.Format(verbose)
	}
	if configErr, ok := err.(*ConfigError); ok {
		return configErr.Format(verbose)
	}
	if awsErr, ok := err.(*AWSAPIError); ok {
		return awsErr.Format(verbose)
	}
	if inputErr, ok := err.(*InputError); ok {
		return inputErr.Format(verbose)
	}
	if internalErr, ok := err.(*InternalError); ok {
		return internalErr.Format(verbose)
	}
	return fmt.Sprintf("Error: %v\n", err)
}
