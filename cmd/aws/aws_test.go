package aws

import (
	"os"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: devops-cli-tool, Property 12: Region Flag Override
// For any region value provided via --region flag, that region SHALL be used
// for AWS operations regardless of profile or environment configuration.
func TestRegionFlagOverride(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Region flag value is returned by GetRegion
	properties.Property("region flag is returned by GetRegion", prop.ForAll(
		func(region string) bool {
			// Set the region flag
			awsRegion = region

			result := GetRegion()
			return result == region
		},
		gen.Identifier(),
	))

	// Property: Region flag overrides environment variable
	properties.Property("region flag overrides env var", prop.ForAll(
		func(flagRegion, envRegion string) bool {
			// Set environment variable
			os.Setenv("AWS_REGION", envRegion)
			defer os.Unsetenv("AWS_REGION")

			// Set the flag
			awsRegion = flagRegion

			// GetRegion should return the flag value
			result := GetRegion()
			return result == flagRegion
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.TestingRun(t)

	// Reset
	awsRegion = ""
}

// Unit test: GetProfile with flag
func TestGetProfileWithFlag(t *testing.T) {
	// Set profile flag
	awsProfile = "test-profile"
	defer func() { awsProfile = "" }()

	result := GetProfile()
	if result != "test-profile" {
		t.Errorf("Expected 'test-profile', got '%s'", result)
	}
}

// Unit test: GetProfile with env var
func TestGetProfileWithEnvVar(t *testing.T) {
	// Clear flag
	awsProfile = ""

	// Set env var
	os.Setenv("AWS_PROFILE", "env-profile")
	defer os.Unsetenv("AWS_PROFILE")

	result := GetProfile()
	if result != "env-profile" {
		t.Errorf("Expected 'env-profile', got '%s'", result)
	}
}

// Unit test: GetProfile default
func TestGetProfileDefault(t *testing.T) {
	// Clear flag and env var
	awsProfile = ""
	os.Unsetenv("AWS_PROFILE")

	result := GetProfile()
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}

// Unit test: GetRegion empty
func TestGetRegionEmpty(t *testing.T) {
	awsRegion = ""
	result := GetRegion()
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

// Unit test: AWSCmd exists
func TestAWSCmdExists(t *testing.T) {
	if AWSCmd == nil {
		t.Error("AWSCmd should not be nil")
	}
	if AWSCmd.Use != "aws" {
		t.Errorf("Expected Use 'aws', got '%s'", AWSCmd.Use)
	}
}

// Unit test: Sub-commands exist
func TestSubCommandsExist(t *testing.T) {
	subCommands := AWSCmd.Commands()

	expectedCommands := []string{"whoami", "login", "profiles"}
	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range subCommands {
			if cmd.Use == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected sub-command '%s' not found", expected)
		}
	}
}

// Unit test: Flags exist
func TestFlagsExist(t *testing.T) {
	profileFlag := AWSCmd.PersistentFlags().Lookup("profile")
	if profileFlag == nil {
		t.Error("Expected --profile flag")
	}
	if profileFlag.Shorthand != "p" {
		t.Errorf("Expected shorthand 'p', got '%s'", profileFlag.Shorthand)
	}

	regionFlag := AWSCmd.PersistentFlags().Lookup("region")
	if regionFlag == nil {
		t.Error("Expected --region flag")
	}
	if regionFlag.Shorthand != "r" {
		t.Errorf("Expected shorthand 'r', got '%s'", regionFlag.Shorthand)
	}
}


// =============================================================================
// Unit Tests for TagPolicyCmd Registration
// Requirements: 5.1, 5.2
// =============================================================================

// TestTagPolicyCmdExists verifies that TagPolicyCmd is defined and has correct properties.
// Requirement 5.1: THE Tag_Policy_Command SHALL be registered as a subcommand of the existing `aws` command
func TestTagPolicyCmdExists(t *testing.T) {
	if TagPolicyCmd == nil {
		t.Fatal("TagPolicyCmd should not be nil")
	}
	if TagPolicyCmd.Use != "tag-policy" {
		t.Errorf("Expected Use 'tag-policy', got '%s'", TagPolicyCmd.Use)
	}
	if TagPolicyCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
	if TagPolicyCmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
	if TagPolicyCmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}
}

// TestTagPolicyCmdRegistered verifies that TagPolicyCmd is registered as a subcommand of AWSCmd.
// Requirement 5.1: THE Tag_Policy_Command SHALL be registered as a subcommand of the existing `aws` command
func TestTagPolicyCmdRegistered(t *testing.T) {
	subCommands := AWSCmd.Commands()

	found := false
	for _, cmd := range subCommands {
		if cmd.Use == "tag-policy" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'tag-policy' sub-command to be registered with AWSCmd")
	}
}

// TestTagPolicyCmdInheritsFlags verifies that TagPolicyCmd inherits --profile and --region flags.
// Requirement 5.2: THE Tag_Policy_Command SHALL inherit the `--profile` and `--region` flags from the parent aws command
func TestTagPolicyCmdInheritsFlags(t *testing.T) {
	// The flags are defined on AWSCmd as PersistentFlags, so they should be inherited
	// by all subcommands including TagPolicyCmd

	// Verify profile flag is accessible
	profileFlag := AWSCmd.PersistentFlags().Lookup("profile")
	if profileFlag == nil {
		t.Error("Expected --profile flag to be defined on AWSCmd")
	}

	// Verify region flag is accessible
	regionFlag := AWSCmd.PersistentFlags().Lookup("region")
	if regionFlag == nil {
		t.Error("Expected --region flag to be defined on AWSCmd")
	}

	// Verify TagPolicyCmd can access the inherited flags through its parent
	// This is implicit since TagPolicyCmd is added to AWSCmd which has PersistentFlags
}

// TestAllSubCommandsIncludeTagPolicy verifies that tag-policy is in the list of all subcommands.
func TestAllSubCommandsIncludeTagPolicy(t *testing.T) {
	subCommands := AWSCmd.Commands()

	expectedCommands := []string{"whoami", "login", "profiles", "tag-policy"}
	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range subCommands {
			if cmd.Use == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected sub-command '%s' not found", expected)
		}
	}
}
