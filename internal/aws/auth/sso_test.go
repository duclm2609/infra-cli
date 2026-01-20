package auth

import (
	"context"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: devops-cli-tool, Property 9: Credential Validation Before Operations
// For any AWS operation command, the CLI SHALL validate credentials before
// attempting the operation, and SHALL fail with an appropriate error if
// credentials are invalid or missing.
func TestCredentialValidationBeforeOperations(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Invalid profile returns error on validation
	properties.Property("invalid profile returns error", prop.ForAll(
		func(profile string) bool {
			// Use a clearly invalid profile name
			auth := NewSSOAuthenticator("nonexistent-profile-"+profile, "us-east-1")
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Validation should fail for invalid profile
			err := auth.ValidateCredentials(ctx)
			// We expect an error because the profile doesn't exist
			return err != nil
		},
		gen.Identifier(),
	))

	// Property: IsAuthenticated returns false for invalid credentials
	properties.Property("IsAuthenticated returns false for invalid profile", prop.ForAll(
		func(profile string) bool {
			auth := NewSSOAuthenticator("nonexistent-profile-"+profile, "us-east-1")
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Should return false for invalid profile
			return !auth.IsAuthenticated(ctx)
		},
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// Unit test: NewSSOAuthenticator
func TestNewSSOAuthenticator(t *testing.T) {
	auth := NewSSOAuthenticator("test-profile", "us-west-2")
	if auth == nil {
		t.Error("Expected non-nil authenticator")
	}
	if auth.profile != "test-profile" {
		t.Errorf("Expected profile 'test-profile', got '%s'", auth.profile)
	}
	if auth.region != "us-west-2" {
		t.Errorf("Expected region 'us-west-2', got '%s'", auth.region)
	}
}

// Unit test: NewAssumeRoleAuthenticator
func TestNewAssumeRoleAuthenticator(t *testing.T) {
	baseAuth := NewSSOAuthenticator("base-profile", "us-east-1")
	roleAuth := NewAssumeRoleAuthenticator(baseAuth, "arn:aws:iam::123456789012:role/TestRole")

	if roleAuth == nil {
		t.Error("Expected non-nil authenticator")
	}
	if roleAuth.roleARN != "arn:aws:iam::123456789012:role/TestRole" {
		t.Errorf("Expected role ARN, got '%s'", roleAuth.roleARN)
	}
	if roleAuth.sessionName != "infra-cli-session" {
		t.Errorf("Expected session name 'infra-cli-session', got '%s'", roleAuth.sessionName)
	}
}

// Unit test: Credentials struct
func TestCredentialsStruct(t *testing.T) {
	creds := &Credentials{
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		SessionToken:    "token123",
		Expiration:      time.Now().Add(1 * time.Hour),
		Source:          "sso",
	}

	if creds.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Error("AccessKeyID mismatch")
	}
	if creds.Source != "sso" {
		t.Error("Source mismatch")
	}
}

// Unit test: IsAuthenticated with invalid profile
func TestIsAuthenticatedInvalidProfile(t *testing.T) {
	auth := NewSSOAuthenticator("definitely-not-a-real-profile", "us-east-1")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if auth.IsAuthenticated(ctx) {
		t.Error("Expected IsAuthenticated to return false for invalid profile")
	}
}

// Unit test: GetCredentials with invalid profile
func TestGetCredentialsInvalidProfile(t *testing.T) {
	auth := NewSSOAuthenticator("definitely-not-a-real-profile", "us-east-1")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := auth.GetCredentials(ctx)
	if err == nil {
		t.Error("Expected error for invalid profile")
	}
}
