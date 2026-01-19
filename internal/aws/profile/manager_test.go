package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: devops-cli-tool, Property 1: Profile Resolution Chain
// For any combination of profile flag value, AWS_PROFILE environment variable,
// and default profile, the Profile_Manager SHALL resolve to the correct profile
// following the precedence: flag > environment variable > "default".
func TestProfileResolutionChain(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Flag profile takes precedence over everything
	properties.Property("flag profile takes precedence", prop.ForAll(
		func(flagProfile, envProfile string) bool {
			// Set environment variable
			os.Setenv("AWS_PROFILE", envProfile)
			defer os.Unsetenv("AWS_PROFILE")

			m := NewManager()
			resolved := m.ResolveProfile(flagProfile)

			return resolved == flagProfile
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	// Property: Environment variable takes precedence when no flag
	properties.Property("env var takes precedence when no flag", prop.ForAll(
		func(envProfile string) bool {
			os.Setenv("AWS_PROFILE", envProfile)
			defer os.Unsetenv("AWS_PROFILE")

			m := NewManager()
			resolved := m.ResolveProfile("")

			return resolved == envProfile
		},
		gen.Identifier(),
	))

	// Property: Default is used when no flag and no env var
	properties.Property("default used when no flag and no env var", prop.ForAll(
		func(_ int) bool {
			os.Unsetenv("AWS_PROFILE")

			m := NewManager()
			resolved := m.ResolveProfile("")

			return resolved == "default"
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

// Unit test: Profile parsing
func TestParseConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	configContent := `[default]
region = us-east-1

[profile my-sso-profile]
sso_start_url = https://my-org.awsapps.com/start
sso_region = us-east-1
sso_account_id = 123456789012
sso_role_name = AdministratorAccess
region = us-west-2

[profile role-profile]
role_arn = arn:aws:iam::123456789012:role/MyRole
source_profile = my-sso-profile
region = eu-west-1
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	m := NewManagerWithPath(configPath)

	// Test default profile
	defaultProfile, err := m.GetProfileConfig("default")
	if err != nil {
		t.Fatalf("Failed to get default profile: %v", err)
	}
	if defaultProfile.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got '%s'", defaultProfile.Region)
	}

	// Test SSO profile
	ssoProfile, err := m.GetProfileConfig("my-sso-profile")
	if err != nil {
		t.Fatalf("Failed to get SSO profile: %v", err)
	}
	if ssoProfile.SSOConfig.SSOStartURL != "https://my-org.awsapps.com/start" {
		t.Errorf("Expected SSO start URL, got '%s'", ssoProfile.SSOConfig.SSOStartURL)
	}
	if !ssoProfile.IsSSO() {
		t.Error("Expected profile to be SSO")
	}

	// Test role profile
	roleProfile, err := m.GetProfileConfig("role-profile")
	if err != nil {
		t.Fatalf("Failed to get role profile: %v", err)
	}
	if roleProfile.RoleARN != "arn:aws:iam::123456789012:role/MyRole" {
		t.Errorf("Expected role ARN, got '%s'", roleProfile.RoleARN)
	}
	if roleProfile.SourceProfile != "my-sso-profile" {
		t.Errorf("Expected source profile 'my-sso-profile', got '%s'", roleProfile.SourceProfile)
	}
}

// Unit test: List profiles
func TestListProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	configContent := `[default]
region = us-east-1

[profile profile1]
region = us-west-1

[profile profile2]
region = eu-west-1
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	m := NewManagerWithPath(configPath)
	profiles, err := m.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(profiles))
	}
}

// Unit test: Profile not found
func TestProfileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	configContent := `[default]
region = us-east-1
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	m := NewManagerWithPath(configPath)
	_, err = m.GetProfileConfig("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent profile")
	}
}

// Unit test: Missing config file
func TestMissingConfigFile(t *testing.T) {
	m := NewManagerWithPath("/nonexistent/path/config")
	profiles, err := m.ListProfiles()
	if err != nil {
		t.Fatalf("Expected no error for missing config file, got: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("Expected empty profiles for missing config file, got %d", len(profiles))
	}
}
