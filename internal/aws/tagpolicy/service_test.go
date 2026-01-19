package tagpolicy

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/smithy-go"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	infraerrors "github.com/user/infra-cli/internal/errors"
)

// =============================================================================
// Mock Client for Testing
// =============================================================================

// mockOrganizationsClient is a mock implementation of OrganizationsClient for testing.
type mockOrganizationsClient struct{}

// DescribeEffectivePolicy is a mock implementation that returns nil.
// This is sufficient for testing configuration flag pass-through since we only
// need to verify that the service is initialized with the correct values.
func (m *mockOrganizationsClient) DescribeEffectivePolicy(ctx context.Context, params *organizations.DescribeEffectivePolicyInput, optFns ...func(*organizations.Options)) (*organizations.DescribeEffectivePolicyOutput, error) {
	return nil, nil
}

// =============================================================================
// Property-Based Tests
// =============================================================================

// genProfileName generates valid AWS profile name strings.
// AWS profile names can contain alphanumeric characters, hyphens, and underscores.
func genProfileName() gopter.Gen {
	return gen.OneGenOf(
		// Empty profile (uses default)
		gen.Const(""),
		// Simple alphanumeric profile names
		gen.Identifier().SuchThat(func(s string) bool {
			return len(s) > 0 && len(s) <= 64
		}),
		// Profile names with hyphens
		gen.Identifier().Map(func(s string) string {
			if len(s) > 30 {
				s = s[:30]
			}
			return s + "-profile"
		}),
		// Profile names with underscores
		gen.Identifier().Map(func(s string) string {
			if len(s) > 30 {
				s = s[:30]
			}
			return s + "_profile"
		}),
		// Common profile name patterns
		gen.OneConstOf(
			"default",
			"dev",
			"staging",
			"production",
			"admin",
			"readonly",
			"power-user",
			"my_company_profile",
		),
	)
}

// genRegionName generates valid AWS region name strings.
// AWS regions follow the pattern: area-location-number (e.g., us-east-1).
func genRegionName() gopter.Gen {
	return gen.OneGenOf(
		// Empty region (uses default)
		gen.Const(""),
		// Standard AWS region patterns
		gen.OneConstOf(
			"us-east-1",
			"us-east-2",
			"us-west-1",
			"us-west-2",
			"eu-west-1",
			"eu-west-2",
			"eu-west-3",
			"eu-central-1",
			"eu-north-1",
			"ap-northeast-1",
			"ap-northeast-2",
			"ap-northeast-3",
			"ap-southeast-1",
			"ap-southeast-2",
			"ap-south-1",
			"sa-east-1",
			"ca-central-1",
			"me-south-1",
			"af-south-1",
		),
		// Generated region-like patterns
		gen.Identifier().Map(func(s string) string {
			if len(s) > 10 {
				s = s[:10]
			}
			return s + "-region-1"
		}),
	)
}

// TestConfigurationFlagsPassedThrough is a property-based test that verifies:
// For any profile name and region value provided via flags, the TagPolicyService
// SHALL be initialized with those exact values for AWS API calls.
//
// Feature: aws-tag-policy, Property 3: Configuration Flags Passed Through
// **Validates: Requirements 1.2, 1.3**
func TestConfigurationFlagsPassedThrough(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Profile value is preserved exactly
	properties.Property("profile value is preserved exactly", prop.ForAll(
		func(profile string) bool {
			mockClient := &mockOrganizationsClient{}
			service := NewTagPolicyServiceWithClient(mockClient, profile, "us-east-1")

			actualProfile := service.GetProfile()
			if actualProfile != profile {
				t.Logf("Profile mismatch: expected %q, got %q", profile, actualProfile)
				return false
			}
			return true
		},
		genProfileName(),
	))

	// Property: Region value is preserved exactly
	properties.Property("region value is preserved exactly", prop.ForAll(
		func(region string) bool {
			mockClient := &mockOrganizationsClient{}
			service := NewTagPolicyServiceWithClient(mockClient, "default", region)

			actualRegion := service.GetRegion()
			if actualRegion != region {
				t.Logf("Region mismatch: expected %q, got %q", region, actualRegion)
				return false
			}
			return true
		},
		genRegionName(),
	))

	// Property: Both profile and region values are preserved exactly
	properties.Property("both profile and region values are preserved exactly", prop.ForAll(
		func(profile, region string) bool {
			mockClient := &mockOrganizationsClient{}
			service := NewTagPolicyServiceWithClient(mockClient, profile, region)

			actualProfile := service.GetProfile()
			actualRegion := service.GetRegion()

			if actualProfile != profile {
				t.Logf("Profile mismatch: expected %q, got %q", profile, actualProfile)
				return false
			}
			if actualRegion != region {
				t.Logf("Region mismatch: expected %q, got %q", region, actualRegion)
				return false
			}
			return true
		},
		genProfileName(),
		genRegionName(),
	))

	// Property: Empty values are preserved (not replaced with defaults)
	properties.Property("empty values are preserved as empty strings", prop.ForAll(
		func(useEmptyProfile, useEmptyRegion bool) bool {
			mockClient := &mockOrganizationsClient{}

			profile := ""
			region := ""
			if !useEmptyProfile {
				profile = "test-profile"
			}
			if !useEmptyRegion {
				region = "us-west-2"
			}

			service := NewTagPolicyServiceWithClient(mockClient, profile, region)

			actualProfile := service.GetProfile()
			actualRegion := service.GetRegion()

			if actualProfile != profile {
				t.Logf("Profile mismatch: expected %q, got %q", profile, actualProfile)
				return false
			}
			if actualRegion != region {
				t.Logf("Region mismatch: expected %q, got %q", region, actualRegion)
				return false
			}
			return true
		},
		gen.Bool(),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Unit Tests for Service Configuration
// =============================================================================

func TestNewTagPolicyServiceWithClient_StoresProfile(t *testing.T) {
	testCases := []struct {
		name     string
		profile  string
		expected string
	}{
		{"empty profile", "", ""},
		{"default profile", "default", "default"},
		{"custom profile", "my-custom-profile", "my-custom-profile"},
		{"profile with underscore", "my_profile", "my_profile"},
		{"profile with numbers", "profile123", "profile123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockOrganizationsClient{}
			service := NewTagPolicyServiceWithClient(mockClient, tc.profile, "us-east-1")

			actual := service.GetProfile()
			if actual != tc.expected {
				t.Errorf("GetProfile() = %q, want %q", actual, tc.expected)
			}
		})
	}
}

func TestNewTagPolicyServiceWithClient_StoresRegion(t *testing.T) {
	testCases := []struct {
		name     string
		region   string
		expected string
	}{
		{"empty region", "", ""},
		{"us-east-1", "us-east-1", "us-east-1"},
		{"eu-west-1", "eu-west-1", "eu-west-1"},
		{"ap-southeast-2", "ap-southeast-2", "ap-southeast-2"},
		{"custom region", "custom-region-1", "custom-region-1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockOrganizationsClient{}
			service := NewTagPolicyServiceWithClient(mockClient, "default", tc.region)

			actual := service.GetRegion()
			if actual != tc.expected {
				t.Errorf("GetRegion() = %q, want %q", actual, tc.expected)
			}
		})
	}
}

func TestNewTagPolicyServiceWithClient_StoresBothValues(t *testing.T) {
	mockClient := &mockOrganizationsClient{}
	profile := "production"
	region := "eu-central-1"

	service := NewTagPolicyServiceWithClient(mockClient, profile, region)

	if service.GetProfile() != profile {
		t.Errorf("GetProfile() = %q, want %q", service.GetProfile(), profile)
	}
	if service.GetRegion() != region {
		t.Errorf("GetRegion() = %q, want %q", service.GetRegion(), region)
	}
}


// =============================================================================
// Mock AWS API Error for Testing
// =============================================================================

// mockAPIError implements smithy.APIError interface for testing error mapping.
type mockAPIError struct {
	code    string
	message string
}

func (e *mockAPIError) ErrorCode() string               { return e.code }
func (e *mockAPIError) ErrorMessage() string            { return e.message }
func (e *mockAPIError) ErrorFault() smithy.ErrorFault   { return smithy.FaultUnknown }
func (e *mockAPIError) Error() string                   { return e.code + ": " + e.message }

// =============================================================================
// Unit Tests for Error Mapping
// =============================================================================

// TestMapAWSError_EffectivePolicyNotFoundException tests that EffectivePolicyNotFoundException
// is mapped to a NoTagPolicyError with code "NO_TAG_POLICY".
// Requirements: 4.1, 4.2, 4.3
func TestMapAWSError_EffectivePolicyNotFoundException(t *testing.T) {
	awsErr := &mockAPIError{
		code:    "EffectivePolicyNotFoundException",
		message: "The effective policy for the specified target was not found.",
	}

	result := mapAWSError(awsErr)

	// Verify it returns an AWSAPIError
	apiErr, ok := result.(*infraerrors.AWSAPIError)
	if !ok {
		t.Fatalf("Expected *infraerrors.AWSAPIError, got %T", result)
	}

	// Verify the error code is "NO_TAG_POLICY"
	if apiErr.Code != "NO_TAG_POLICY" {
		t.Errorf("Expected Code 'NO_TAG_POLICY', got %q", apiErr.Code)
	}

	// Verify the message is appropriate
	if apiErr.Message != "No tag policy in effect" {
		t.Errorf("Expected Message 'No tag policy in effect', got %q", apiErr.Message)
	}

	// Verify the service and operation are set correctly
	if apiErr.Service != "organizations" {
		t.Errorf("Expected Service 'organizations', got %q", apiErr.Service)
	}
	if apiErr.Operation != "DescribeEffectivePolicy" {
		t.Errorf("Expected Operation 'DescribeEffectivePolicy', got %q", apiErr.Operation)
	}

	// Verify the suggestion is helpful
	expectedSuggestion := "Tag policies are managed through AWS Organizations. Contact your organization administrator."
	if apiErr.Suggestion != expectedSuggestion {
		t.Errorf("Expected Suggestion %q, got %q", expectedSuggestion, apiErr.Suggestion)
	}
}

// TestMapAWSError_AWSOrganizationsNotInUseException tests that AWSOrganizationsNotInUseException
// is mapped to a NotInOrganizationError with code "NOT_IN_ORGANIZATION".
// Requirements: 4.1, 4.2, 4.3
func TestMapAWSError_AWSOrganizationsNotInUseException(t *testing.T) {
	awsErr := &mockAPIError{
		code:    "AWSOrganizationsNotInUseException",
		message: "Your account is not a member of an organization.",
	}

	result := mapAWSError(awsErr)

	// Verify it returns an AWSAPIError
	apiErr, ok := result.(*infraerrors.AWSAPIError)
	if !ok {
		t.Fatalf("Expected *infraerrors.AWSAPIError, got %T", result)
	}

	// Verify the error code is "NOT_IN_ORGANIZATION"
	if apiErr.Code != "NOT_IN_ORGANIZATION" {
		t.Errorf("Expected Code 'NOT_IN_ORGANIZATION', got %q", apiErr.Code)
	}

	// Verify the message is appropriate
	if apiErr.Message != "Account is not part of an AWS Organization" {
		t.Errorf("Expected Message 'Account is not part of an AWS Organization', got %q", apiErr.Message)
	}

	// Verify the service and operation are set correctly
	if apiErr.Service != "organizations" {
		t.Errorf("Expected Service 'organizations', got %q", apiErr.Service)
	}
	if apiErr.Operation != "DescribeEffectivePolicy" {
		t.Errorf("Expected Operation 'DescribeEffectivePolicy', got %q", apiErr.Operation)
	}

	// Verify the suggestion is helpful
	expectedSuggestion := "Join an AWS Organization or create one to use tag policies."
	if apiErr.Suggestion != expectedSuggestion {
		t.Errorf("Expected Suggestion %q, got %q", expectedSuggestion, apiErr.Suggestion)
	}

	// Verify the original error is preserved as the cause
	if apiErr.Cause == nil {
		t.Error("Expected Cause to be set, got nil")
	}
}

// TestMapAWSError_AccessDeniedException tests that AccessDeniedException
// is mapped to an AWSAPIError with code "AccessDenied".
// Requirements: 4.1, 4.2, 4.3
func TestMapAWSError_AccessDeniedException(t *testing.T) {
	awsErr := &mockAPIError{
		code:    "AccessDeniedException",
		message: "User is not authorized to perform organizations:DescribeEffectivePolicy",
	}

	result := mapAWSError(awsErr)

	// Verify it returns an AWSAPIError
	apiErr, ok := result.(*infraerrors.AWSAPIError)
	if !ok {
		t.Fatalf("Expected *infraerrors.AWSAPIError, got %T", result)
	}

	// Verify the error code is "AccessDenied"
	if apiErr.Code != "AccessDenied" {
		t.Errorf("Expected Code 'AccessDenied', got %q", apiErr.Code)
	}

	// Verify the message contains permission information
	expectedMessage := "AccessDenied: Insufficient permissions to describe tag policy"
	if apiErr.Message != expectedMessage {
		t.Errorf("Expected Message %q, got %q", expectedMessage, apiErr.Message)
	}

	// Verify the service and operation are set correctly
	if apiErr.Service != "organizations" {
		t.Errorf("Expected Service 'organizations', got %q", apiErr.Service)
	}
	if apiErr.Operation != "DescribeEffectivePolicy" {
		t.Errorf("Expected Operation 'DescribeEffectivePolicy', got %q", apiErr.Operation)
	}

	// Verify the original error is preserved as the cause
	if apiErr.Cause == nil {
		t.Error("Expected Cause to be set, got nil")
	}
}

// TestMapAWSError_OtherAWSAPIErrors tests that other AWS API errors
// are mapped to an AWSAPIError with the original error code.
// Requirements: 4.1, 4.2, 4.3
func TestMapAWSError_OtherAWSAPIErrors(t *testing.T) {
	testCases := []struct {
		name            string
		code            string
		message         string
		expectedCode    string
		expectedMessage string
	}{
		{
			name:            "ThrottlingException",
			code:            "ThrottlingException",
			message:         "Rate exceeded",
			expectedCode:    "ThrottlingException",
			expectedMessage: "ThrottlingException: Rate exceeded",
		},
		{
			name:            "ServiceException",
			code:            "ServiceException",
			message:         "Internal server error",
			expectedCode:    "ServiceException",
			expectedMessage: "ServiceException: Internal server error",
		},
		{
			name:            "InvalidInputException",
			code:            "InvalidInputException",
			message:         "Invalid target ID",
			expectedCode:    "InvalidInputException",
			expectedMessage: "InvalidInputException: Invalid target ID",
		},
		{
			name:            "TargetNotFoundException",
			code:            "TargetNotFoundException",
			message:         "The specified target was not found",
			expectedCode:    "TargetNotFoundException",
			expectedMessage: "TargetNotFoundException: The specified target was not found",
		},
		{
			name:            "TooManyRequestsException",
			code:            "TooManyRequestsException",
			message:         "Too many requests",
			expectedCode:    "TooManyRequestsException",
			expectedMessage: "TooManyRequestsException: Too many requests",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			awsErr := &mockAPIError{
				code:    tc.code,
				message: tc.message,
			}

			result := mapAWSError(awsErr)

			// Verify it returns an AWSAPIError
			apiErr, ok := result.(*infraerrors.AWSAPIError)
			if !ok {
				t.Fatalf("Expected *infraerrors.AWSAPIError, got %T", result)
			}

			// Verify the error code matches the original AWS error code
			if apiErr.Code != tc.expectedCode {
				t.Errorf("Expected Code %q, got %q", tc.expectedCode, apiErr.Code)
			}

			// Verify the message contains the original AWS error message
			if apiErr.Message != tc.expectedMessage {
				t.Errorf("Expected Message %q, got %q", tc.expectedMessage, apiErr.Message)
			}

			// Verify the service and operation are set correctly
			if apiErr.Service != "organizations" {
				t.Errorf("Expected Service 'organizations', got %q", apiErr.Service)
			}
			if apiErr.Operation != "DescribeEffectivePolicy" {
				t.Errorf("Expected Operation 'DescribeEffectivePolicy', got %q", apiErr.Operation)
			}

			// Verify the original error is preserved as the cause
			if apiErr.Cause == nil {
				t.Error("Expected Cause to be set, got nil")
			}
		})
	}
}

// TestMapAWSError_NonAPIError tests that non-API errors (errors that don't implement
// smithy.APIError) are mapped to an InternalError.
// Requirements: 4.1, 4.2, 4.3
func TestMapAWSError_NonAPIError(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{
			name: "simple error",
			err:  errors.New("connection refused"),
		},
		{
			name: "wrapped error",
			err:  errors.New("network timeout"),
		},
		{
			name: "context error",
			err:  errors.New("context deadline exceeded"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapAWSError(tc.err)

			// Verify it returns an InternalError
			internalErr, ok := result.(*infraerrors.InternalError)
			if !ok {
				t.Fatalf("Expected *infraerrors.InternalError, got %T", result)
			}

			// Verify the error code is "INTERNAL_ERROR"
			if internalErr.Code != "INTERNAL_ERROR" {
				t.Errorf("Expected Code 'INTERNAL_ERROR', got %q", internalErr.Code)
			}

			// Verify the message indicates unexpected error
			expectedMessage := "unexpected error from AWS Organizations"
			if internalErr.Message != expectedMessage {
				t.Errorf("Expected Message %q, got %q", expectedMessage, internalErr.Message)
			}

			// Verify the original error is preserved as the cause
			if internalErr.Cause == nil {
				t.Error("Expected Cause to be set, got nil")
			}
		})
	}
}

// TestMapAWSError_ErrorCategoryAndExitCode tests that mapped errors have
// the correct category and exit code for proper CLI error handling.
// Requirements: 4.1, 4.2, 4.3
func TestMapAWSError_ErrorCategoryAndExitCode(t *testing.T) {
	testCases := []struct {
		name             string
		err              error
		expectedCategory infraerrors.ErrorCategory
		expectedExitCode int
	}{
		{
			name: "EffectivePolicyNotFoundException has AWS API category",
			err: &mockAPIError{
				code:    "EffectivePolicyNotFoundException",
				message: "Policy not found",
			},
			expectedCategory: infraerrors.CategoryAWSAPI,
			expectedExitCode: infraerrors.ExitAWSAPIError,
		},
		{
			name: "AWSOrganizationsNotInUseException has AWS API category",
			err: &mockAPIError{
				code:    "AWSOrganizationsNotInUseException",
				message: "Not in organization",
			},
			expectedCategory: infraerrors.CategoryAWSAPI,
			expectedExitCode: infraerrors.ExitAWSAPIError,
		},
		{
			name: "AccessDeniedException has AWS API category",
			err: &mockAPIError{
				code:    "AccessDeniedException",
				message: "Access denied",
			},
			expectedCategory: infraerrors.CategoryAWSAPI,
			expectedExitCode: infraerrors.ExitAWSAPIError,
		},
		{
			name:             "Non-API error has Internal category",
			err:              errors.New("unexpected error"),
			expectedCategory: infraerrors.CategoryInternal,
			expectedExitCode: infraerrors.ExitInternalError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapAWSError(tc.err)

			// Check category based on error type
			switch err := result.(type) {
			case *infraerrors.AWSAPIError:
				if err.Category != tc.expectedCategory {
					t.Errorf("Expected Category %q, got %q", tc.expectedCategory, err.Category)
				}
				if err.ExitCode() != tc.expectedExitCode {
					t.Errorf("Expected ExitCode %d, got %d", tc.expectedExitCode, err.ExitCode())
				}
			case *infraerrors.InternalError:
				if err.Category != tc.expectedCategory {
					t.Errorf("Expected Category %q, got %q", tc.expectedCategory, err.Category)
				}
				if err.ExitCode() != tc.expectedExitCode {
					t.Errorf("Expected ExitCode %d, got %d", tc.expectedExitCode, err.ExitCode())
				}
			default:
				t.Fatalf("Unexpected error type: %T", result)
			}
		})
	}
}
