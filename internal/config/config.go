package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	DefaultProfile string `mapstructure:"default_profile"`
	DefaultRegion  string `mapstructure:"default_region"`
	DefaultOutput  string `mapstructure:"default_output"`
	Verbose        bool   `mapstructure:"verbose"`
}

// Manager handles configuration loading and merging
type Manager struct {
	viper  *viper.Viper
	config *Config
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add config paths based on OS
	configDir := getConfigDir()
	v.AddConfigPath(configDir)

	// Environment variable support
	v.SetEnvPrefix("INFRA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set defaults
	v.SetDefault("default_profile", "default")
	v.SetDefault("default_region", "us-east-1")
	v.SetDefault("default_output", "table")
	v.SetDefault("verbose", false)

	return &Manager{
		viper:  v,
		config: &Config{},
	}
}

// Load reads configuration from all sources
func (m *Manager) Load() (*Config, error) {
	// Try to read config file (ignore if not found)
	_ = m.viper.ReadInConfig()

	// Unmarshal into config struct
	if err := m.viper.Unmarshal(m.config); err != nil {
		return nil, err
	}

	return m.config, nil
}

// Get returns a configuration value by key
func (m *Manager) Get(key string) interface{} {
	return m.viper.Get(key)
}

// GetString returns a string configuration value
func (m *Manager) GetString(key string) string {
	return m.viper.GetString(key)
}

// GetBool returns a boolean configuration value
func (m *Manager) GetBool(key string) bool {
	return m.viper.GetBool(key)
}

// Set sets a configuration value (for flag overrides)
func (m *Manager) Set(key string, value interface{}) {
	m.viper.Set(key, value)
}

// GetConfig returns the loaded configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// getConfigDir returns the OS-appropriate configuration directory
func getConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "infra")
		}
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "infra")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "infra")
	default: // Linux and others - follow XDG spec
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig != "" {
			return filepath.Join(xdgConfig, "infra")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "infra")
	}
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() string {
	return getConfigDir()
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configDir := getConfigDir()
	return os.MkdirAll(configDir, 0755)
}
