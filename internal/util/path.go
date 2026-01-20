package util

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ExpandHome expands the ~ prefix to the user's home directory
func ExpandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return home
	}

	// Handle ~/path or ~\path
	if len(path) > 1 && (path[1] == '/' || path[1] == '\\') {
		return filepath.Join(home, path[2:])
	}

	return path
}

// JoinPath joins path elements using the OS-appropriate separator
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// NormalizePath converts a path to use the OS-appropriate separator
func NormalizePath(path string) string {
	return filepath.FromSlash(path)
}

// ToSlash converts a path to use forward slashes (for display/storage)
func ToSlash(path string) string {
	return filepath.ToSlash(path)
}

// IsAbsPath checks if a path is absolute
func IsAbsPath(path string) bool {
	return filepath.IsAbs(path)
}

// GetConfigDir returns the OS-appropriate configuration directory
func GetConfigDir(appName string) string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, appName)
		}
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", appName)
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", appName)
	default: // Linux and others - follow XDG spec
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig != "" {
			return filepath.Join(xdgConfig, appName)
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", appName)
	}
}

// GetDataDir returns the OS-appropriate data directory
func GetDataDir(appName string) string {
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			return filepath.Join(localAppData, appName)
		}
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", appName)
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", appName)
	default: // Linux and others - follow XDG spec
		xdgData := os.Getenv("XDG_DATA_HOME")
		if xdgData != "" {
			return filepath.Join(xdgData, appName)
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", appName)
	}
}

// GetCacheDir returns the OS-appropriate cache directory
func GetCacheDir(appName string) string {
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			return filepath.Join(localAppData, appName, "Cache")
		}
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", appName, "Cache")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Caches", appName)
	default: // Linux and others - follow XDG spec
		xdgCache := os.Getenv("XDG_CACHE_HOME")
		if xdgCache != "" {
			return filepath.Join(xdgCache, appName)
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".cache", appName)
	}
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetPathSeparator returns the OS path separator
func GetPathSeparator() string {
	return string(filepath.Separator)
}

// IsValidPath checks if a path contains only valid characters
func IsValidPath(path string) bool {
	if path == "" {
		return false
	}

	// Check for null bytes (invalid on all platforms)
	if strings.Contains(path, "\x00") {
		return false
	}

	// Windows-specific invalid characters
	if runtime.GOOS == "windows" {
		invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
		for _, char := range invalidChars {
			// Allow : only as drive letter separator (e.g., C:)
			if char == ":" {
				// Check if : appears after position 1
				idx := strings.Index(path, ":")
				if idx > 1 || (idx == 1 && !isLetter(path[0])) {
					return false
				}
				continue
			}
			if strings.Contains(path, char) {
				return false
			}
		}
	}

	return true
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// CleanPath cleans a path by removing redundant separators and . and ..
func CleanPath(path string) string {
	return filepath.Clean(path)
}
