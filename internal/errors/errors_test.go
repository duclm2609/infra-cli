package errors

import (
	"errors"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: devops-cli-tool, Property 6: Error Exit Codes
// For any operation that results in an error, the CLI SHALL exit with a non-zero exit code.
func TestErrorExitCodes(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Auth errors have non-zero exit code
	properties.Property("auth errors have non-zero exit code", prop.ForAll(
		func(profile, message string) bool {
			err := NewAuthError("TEST", message, profile, nil)
			return err.ExitCode() != ExitSuccess
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	// Property: Config errors have non-zero exit code
	properties.Property("config errors have non-zero exit code", prop.ForAll(
		func(message, source string) bool {
			err := NewConfigError("TEST", message, source, nil)
			return err.ExitCode() != ExitSuccess
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	// Property: AWS API errors have non-zero exit code
	properties.Property("AWS API errors have non-zero exit code", prop.ForAll(
		func(code, message string) bool {
			err := NewAWSAPIError(code, message, "TestService", "TestOp", nil)
			return err.ExitCode() != ExitSuccess
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	// Property: Input errors have non-zero exit code
	properties.Property("input errors have non-zero exit code", prop.ForAll(
		func(field, value, message string) bool {
			err := NewInputError(field, value, message)
			return err.ExitCode() != ExitSuccess
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Identifier(),
	))

	// Property: Internal errors have non-zero exit code
	properties.Property("internal errors have non-zero exit code", prop.ForAll(
		func(message string) bool {
			err := NewInternalError(message, nil)
			return err.ExitCode() != ExitSuccess
		},
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// Feature: devops-cli-tool, Property 7: Success Exit Codes
// For any operation that completes successfully, the CLI SHALL exit with exit code 0.
// (This is tested implicitly - ExitSuccess constant is 0)
func TestSuccessExitCode(t *testing.T) {
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess should be 0, got %d", ExitSuccess)
	}
}

// Feature: devops-cli-tool, Property 11: AWS Error Message Clarity
// For any AWS API error, the formatted error message SHALL contain both
// the AWS error code and a human-readable description.
func TestAWSErrorMessageClarity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("AWS error contains code and message", prop.ForAll(
		func(code, message string) bool {
			err := NewAWSAPIError(code, message, "TestService", "TestOp", nil)
			formatted := err.Format(false)

			// Should contain the error code
			containsCode := strings.Contains(formatted, code)
			// Should contain the message
			containsMessage := strings.Contains(formatted, message)

			return containsCode && containsMessage
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// Unit tests
func TestAuthErrorCreation(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewAuthError("AUTH_FAILED", "Authentication failed", "my-profile", cause)

	if err.Category != CategoryAuth {
		t.Errorf("Expected category %s, got %s", CategoryAuth, err.Category)
	}
	if err.Code != "AUTH_FAILED" {
		t.Errorf("Expected code 'AUTH_FAILED', got '%s'", err.Code)
	}
	if err.Profile != "my-profile" {
		t.Errorf("Expected profile 'my-profile', got '%s'", err.Profile)
	}
	if err.ExitCode() != ExitAuthError {
		t.Errorf("Expected exit code %d, got %d", ExitAuthError, err.ExitCode())
	}
}

func TestSSOExpiredError(t *testing.T) {
	err := NewSSOExpiredError("test-profile", nil)

	if err.Code != "SSO_EXPIRED" {
		t.Errorf("Expected code 'SSO_EXPIRED', got '%s'", err.Code)
	}
	if !strings.Contains(err.Details, "test-profile") {
		t.Error("Details should contain profile name")
	}
	if !strings.Contains(err.Suggestion, "aws sso login") {
		t.Error("Suggestion should mention aws sso login")
	}
}

func TestConfigErrorCreation(t *testing.T) {
	err := NewConfigError("CONFIG_INVALID", "Invalid config", "file", nil)

	if err.Category != CategoryConfig {
		t.Errorf("Expected category %s, got %s", CategoryConfig, err.Category)
	}
	if err.Source != "file" {
		t.Errorf("Expected source 'file', got '%s'", err.Source)
	}
	if err.ExitCode() != ExitConfigError {
		t.Errorf("Expected exit code %d, got %d", ExitConfigError, err.ExitCode())
	}
}

func TestProfileNotFoundError(t *testing.T) {
	err := NewProfileNotFoundError("nonexistent", nil)

	if err.Code != "PROFILE_NOT_FOUND" {
		t.Errorf("Expected code 'PROFILE_NOT_FOUND', got '%s'", err.Code)
	}
	if !strings.Contains(err.Details, "nonexistent") {
		t.Error("Details should contain profile name")
	}
}

func TestAWSAPIErrorCreation(t *testing.T) {
	err := NewAWSAPIError("AccessDenied", "Access denied", "S3", "GetObject", nil)

	if err.Category != CategoryAWSAPI {
		t.Errorf("Expected category %s, got %s", CategoryAWSAPI, err.Category)
	}
	if err.AWSErrorCode != "AccessDenied" {
		t.Errorf("Expected AWS error code 'AccessDenied', got '%s'", err.AWSErrorCode)
	}
	if err.Service != "S3" {
		t.Errorf("Expected service 'S3', got '%s'", err.Service)
	}
	if err.ExitCode() != ExitAWSAPIError {
		t.Errorf("Expected exit code %d, got %d", ExitAWSAPIError, err.ExitCode())
	}
}

func TestInputErrorCreation(t *testing.T) {
	err := NewInputError("region", "invalid-region", "Invalid region")

	if err.Category != CategoryInput {
		t.Errorf("Expected category %s, got %s", CategoryInput, err.Category)
	}
	if err.Field != "region" {
		t.Errorf("Expected field 'region', got '%s'", err.Field)
	}
	if err.ExitCode() != ExitInputError {
		t.Errorf("Expected exit code %d, got %d", ExitInputError, err.ExitCode())
	}
}

func TestInvalidRegionError(t *testing.T) {
	err := NewInvalidRegionError("bad-region")

	if err.Code != "INVALID_REGION" {
		t.Errorf("Expected code 'INVALID_REGION', got '%s'", err.Code)
	}
	if !strings.Contains(err.Suggestion, "us-east-1") {
		t.Error("Suggestion should contain example regions")
	}
}

func TestInternalErrorCreation(t *testing.T) {
	cause := errors.New("panic")
	err := NewInternalError("Something went wrong", cause)

	if err.Category != CategoryInternal {
		t.Errorf("Expected category %s, got %s", CategoryInternal, err.Category)
	}
	if err.ExitCode() != ExitInternalError {
		t.Errorf("Expected exit code %d, got %d", ExitInternalError, err.ExitCode())
	}
}

func TestErrorFormatting(t *testing.T) {
	err := NewAuthError("TEST", "Test error", "profile", errors.New("cause"))

	// Non-verbose format
	formatted := err.Format(false)
	if !strings.Contains(formatted, "Error:") {
		t.Error("Formatted error should contain 'Error:'")
	}
	if !strings.Contains(formatted, "--verbose") {
		t.Error("Non-verbose format should mention --verbose flag")
	}

	// Verbose format
	verboseFormatted := err.Format(true)
	if !strings.Contains(verboseFormatted, "Cause:") {
		t.Error("Verbose format should contain cause")
	}
}

func TestIsErrorFunctions(t *testing.T) {
	authErr := NewAuthError("TEST", "test", "profile", nil)
	configErr := NewConfigError("TEST", "test", "file", nil)
	awsErr := NewAWSAPIError("TEST", "test", "S3", "Get", nil)

	if !IsAuthError(authErr) {
		t.Error("IsAuthError should return true for AuthError")
	}
	if IsAuthError(configErr) {
		t.Error("IsAuthError should return false for ConfigError")
	}

	if !IsConfigError(configErr) {
		t.Error("IsConfigError should return true for ConfigError")
	}
	if IsConfigError(authErr) {
		t.Error("IsConfigError should return false for AuthError")
	}

	if !IsAWSAPIError(awsErr) {
		t.Error("IsAWSAPIError should return true for AWSAPIError")
	}
	if IsAWSAPIError(authErr) {
		t.Error("IsAWSAPIError should return false for AuthError")
	}
}

func TestGetExitCode(t *testing.T) {
	authErr := NewAuthError("TEST", "test", "profile", nil)
	if GetExitCode(authErr) != ExitAuthError {
		t.Errorf("Expected %d, got %d", ExitAuthError, GetExitCode(authErr))
	}

	configErr := NewConfigError("TEST", "test", "file", nil)
	if GetExitCode(configErr) != ExitConfigError {
		t.Errorf("Expected %d, got %d", ExitConfigError, GetExitCode(configErr))
	}

	// Standard error should return internal error code
	stdErr := errors.New("standard error")
	if GetExitCode(stdErr) != ExitInternalError {
		t.Errorf("Expected %d for standard error, got %d", ExitInternalError, GetExitCode(stdErr))
	}
}
