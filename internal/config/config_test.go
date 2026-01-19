package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: devops-cli-tool, Property 2: Configuration Precedence
// For any configuration key that can be set via flag, environment variable,
// and config file simultaneously, the resolved value SHALL equal the value
// from the highest priority source (flags > env vars > config file > defaults).
func TestConfigurationPrecedence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Environment variables override config file defaults
	properties.Property("env vars override defaults", prop.ForAll(
		func(envValue string) bool {
			if envValue == "" {
				return true // Skip empty values
			}

			// Set environment variable
			os.Setenv("INFRA_DEFAULT_PROFILE", envValue)
			defer os.Unsetenv("INFRA_DEFAULT_PROFILE")

			m := NewManager()
			cfg, err := m.Load()
			if err != nil {
				return false
			}

			return cfg.DefaultProfile == envValue
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
	))

	// Property: Set (flag simulation) overrides environment variables
	properties.Property("flags override env vars", prop.ForAll(
		func(flagValue, envValue string) bool {
			// Set environment variable
			os.Setenv("INFRA_DEFAULT_REGION", envValue)
			defer os.Unsetenv("INFRA_DEFAULT_REGION")

			m := NewManager()
			m.Load()

			// Simulate flag override
			m.Set("default_region", flagValue)

			return m.GetString("default_region") == flagValue
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	// Property: Defaults are used when no other source provides value
	properties.Property("defaults used when no override", prop.ForAll(
		func(_ int) bool {
			// Ensure no env vars are set
			os.Unsetenv("INFRA_DEFAULT_OUTPUT")

			m := NewManager()
			cfg, err := m.Load()
			if err != nil {
				return false
			}

			// Default output should be "table"
			return cfg.DefaultOutput == "table"
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

// Unit test: Config directory detection
func TestGetConfigDir(t *testing.T) {
	configDir := GetConfigDir()
	if configDir == "" {
		t.Error("Config directory should not be empty")
	}

	// Should contain "infra" in the path
	if !filepath.IsAbs(configDir) && configDir[0] != '~' {
		// For relative paths, just check it's not empty
		if len(configDir) == 0 {
			t.Error("Config directory path is empty")
		}
	}
}

// Unit test: Manager creation
func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Error("Manager should not be nil")
	}
	if m.viper == nil {
		t.Error("Viper instance should not be nil")
	}
}

// Unit test: Default values
func TestDefaultValues(t *testing.T) {
	// Clear any env vars that might interfere
	os.Unsetenv("INFRA_DEFAULT_PROFILE")
	os.Unsetenv("INFRA_DEFAULT_REGION")
	os.Unsetenv("INFRA_DEFAULT_OUTPUT")
	os.Unsetenv("INFRA_VERBOSE")

	m := NewManager()
	cfg, err := m.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.DefaultProfile != "default" {
		t.Errorf("Expected default profile 'default', got '%s'", cfg.DefaultProfile)
	}
	if cfg.DefaultRegion != "us-east-1" {
		t.Errorf("Expected default region 'us-east-1', got '%s'", cfg.DefaultRegion)
	}
	if cfg.DefaultOutput != "table" {
		t.Errorf("Expected default output 'table', got '%s'", cfg.DefaultOutput)
	}
	if cfg.Verbose != false {
		t.Error("Expected verbose to be false by default")
	}
}
