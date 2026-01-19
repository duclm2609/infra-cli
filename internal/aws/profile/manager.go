package profile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SSOConfig holds SSO configuration from AWS config file
type SSOConfig struct {
	SSOStartURL  string
	SSORegion    string
	SSOAccountID string
	SSORoleName  string
}

// ProfileConfig holds AWS profile configuration
type ProfileConfig struct {
	Name          string
	Region        string
	SSOConfig     *SSOConfig
	RoleARN       string
	SourceProfile string
}

// Manager handles AWS profile operations
type Manager struct {
	configPath string
}

// NewManager creates a new profile manager
func NewManager() *Manager {
	return &Manager{
		configPath: getAWSConfigPath(),
	}
}

// NewManagerWithPath creates a profile manager with custom config path
func NewManagerWithPath(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// ResolveProfile determines which profile to use based on precedence:
// 1. Flag value (if provided)
// 2. AWS_PROFILE environment variable
// 3. "default" profile
func (m *Manager) ResolveProfile(flagProfile string) string {
	// Flag takes highest precedence
	if flagProfile != "" {
		return flagProfile
	}

	// Check environment variable
	if envProfile := os.Getenv("AWS_PROFILE"); envProfile != "" {
		return envProfile
	}

	// Default to "default"
	return "default"
}

// GetProfileConfig returns configuration for a profile
func (m *Manager) GetProfileConfig(profileName string) (*ProfileConfig, error) {
	profiles, err := m.parseConfigFile()
	if err != nil {
		return nil, err
	}

	config, exists := profiles[profileName]
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found in AWS config", profileName)
	}

	return config, nil
}

// ListProfiles returns all available profiles
func (m *Manager) ListProfiles() ([]string, error) {
	profiles, err := m.parseConfigFile()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}

	return names, nil
}

// parseConfigFile parses the AWS config file
func (m *Manager) parseConfigFile() (map[string]*ProfileConfig, error) {
	file, err := os.Open(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*ProfileConfig), nil
		}
		return nil, fmt.Errorf("failed to open AWS config file: %w", err)
	}
	defer file.Close()

	profiles := make(map[string]*ProfileConfig)
	var currentProfile *ProfileConfig
	var currentProfileName string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for profile section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous profile
			if currentProfile != nil {
				profiles[currentProfileName] = currentProfile
			}

			// Parse profile name
			section := strings.TrimPrefix(strings.TrimSuffix(line, "]"), "[")
			if strings.HasPrefix(section, "profile ") {
				currentProfileName = strings.TrimPrefix(section, "profile ")
			} else if section == "default" {
				currentProfileName = "default"
			} else {
				currentProfileName = section
			}

			currentProfile = &ProfileConfig{
				Name:      currentProfileName,
				SSOConfig: &SSOConfig{},
			}
			continue
		}

		// Parse key-value pairs
		if currentProfile != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "region":
					currentProfile.Region = value
				case "sso_start_url":
					currentProfile.SSOConfig.SSOStartURL = value
				case "sso_region":
					currentProfile.SSOConfig.SSORegion = value
				case "sso_account_id":
					currentProfile.SSOConfig.SSOAccountID = value
				case "sso_role_name":
					currentProfile.SSOConfig.SSORoleName = value
				case "role_arn":
					currentProfile.RoleARN = value
				case "source_profile":
					currentProfile.SourceProfile = value
				}
			}
		}
	}

	// Save last profile
	if currentProfile != nil {
		profiles[currentProfileName] = currentProfile
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading AWS config file: %w", err)
	}

	return profiles, nil
}

// IsSSO checks if a profile is configured for SSO
func (p *ProfileConfig) IsSSO() bool {
	return p.SSOConfig != nil && p.SSOConfig.SSOStartURL != ""
}

// getAWSConfigPath returns the path to the AWS config file
func getAWSConfigPath() string {
	// Check for custom config file path
	if configFile := os.Getenv("AWS_CONFIG_FILE"); configFile != "" {
		return configFile
	}

	// Default to ~/.aws/config
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "config")
}

// GetAWSConfigPath returns the AWS config file path
func GetAWSConfigPath() string {
	return getAWSConfigPath()
}
