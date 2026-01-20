package util

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: devops-cli-tool, Property 13: OS-Agnostic Path Handling
// For any file path constructed by the CLI, the path SHALL use the correct
// path separator for the current operating system and SHALL not contain
// invalid path characters.
func TestOSAgnosticPathHandling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: JoinPath uses correct separator
	properties.Property("JoinPath uses OS separator", prop.ForAll(
		func(parts []string) bool {
			if len(parts) == 0 {
				return true
			}

			// Filter out empty parts
			validParts := make([]string, 0)
			for _, p := range parts {
				if p != "" {
					validParts = append(validParts, p)
				}
			}
			if len(validParts) == 0 {
				return true
			}

			result := JoinPath(validParts...)

			// Result should use OS separator (or be a single element)
			if len(validParts) > 1 {
				expectedSep := string(filepath.Separator)
				// The path should contain the OS separator or be cleaned
				return strings.Contains(result, expectedSep) || !strings.Contains(result, "/") && !strings.Contains(result, "\\")
			}
			return true
		},
		gen.SliceOfN(3, gen.Identifier()),
	))

	// Property: NormalizePath produces valid paths
	properties.Property("NormalizePath produces valid paths", prop.ForAll(
		func(path string) bool {
			if path == "" {
				return true
			}

			result := NormalizePath(path)

			// Result should not contain null bytes
			if strings.Contains(result, "\x00") {
				return false
			}

			return true
		},
		gen.Identifier(),
	))

	// Property: CleanPath removes redundant separators
	properties.Property("CleanPath removes redundant separators", prop.ForAll(
		func(path string) bool {
			if path == "" {
				return true
			}

			result := CleanPath(path)

			// Result should not have consecutive separators
			doubleSep := string(filepath.Separator) + string(filepath.Separator)
			return !strings.Contains(result, doubleSep)
		},
		gen.Identifier(),
	))

	// Property: IsValidPath rejects null bytes
	properties.Property("IsValidPath rejects null bytes", prop.ForAll(
		func(path string) bool {
			pathWithNull := path + "\x00" + path
			return !IsValidPath(pathWithNull)
		},
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// Unit tests
func TestExpandHome(t *testing.T) {
	// Test ~ expansion
	result := ExpandHome("~")
	if result == "~" {
		t.Error("~ should be expanded to home directory")
	}

	// Test ~/path expansion
	result = ExpandHome("~/test")
	if strings.HasPrefix(result, "~") {
		t.Error("~/test should be expanded")
	}
	if !strings.HasSuffix(result, "test") {
		t.Error("Path should end with 'test'")
	}

	// Test non-home path
	result = ExpandHome("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("Non-home path should not be modified, got '%s'", result)
	}
}

func TestJoinPath(t *testing.T) {
	result := JoinPath("a", "b", "c")
	expected := filepath.Join("a", "b", "c")
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestNormalizePath(t *testing.T) {
	// Test forward slash conversion
	result := NormalizePath("a/b/c")
	expected := filepath.FromSlash("a/b/c")
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestToSlash(t *testing.T) {
	// Create a path with OS separator
	osPath := filepath.Join("a", "b", "c")
	result := ToSlash(osPath)

	// Should use forward slashes
	if strings.Contains(result, "\\") {
		t.Error("ToSlash should convert to forward slashes")
	}
}

func TestIsAbsPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		if !IsAbsPath("C:\\test") {
			t.Error("C:\\test should be absolute on Windows")
		}
	} else {
		if !IsAbsPath("/test") {
			t.Error("/test should be absolute on Unix")
		}
	}

	if IsAbsPath("relative/path") {
		t.Error("relative/path should not be absolute")
	}
}

func TestGetConfigDir(t *testing.T) {
	result := GetConfigDir("testapp")
	if result == "" {
		t.Error("Config dir should not be empty")
	}
	if !strings.Contains(result, "testapp") {
		t.Error("Config dir should contain app name")
	}
}

func TestGetDataDir(t *testing.T) {
	result := GetDataDir("testapp")
	if result == "" {
		t.Error("Data dir should not be empty")
	}
	if !strings.Contains(result, "testapp") {
		t.Error("Data dir should contain app name")
	}
}

func TestGetCacheDir(t *testing.T) {
	result := GetCacheDir("testapp")
	if result == "" {
		t.Error("Cache dir should not be empty")
	}
	if !strings.Contains(result, "testapp") {
		t.Error("Cache dir should contain app name")
	}
}

func TestGetPathSeparator(t *testing.T) {
	sep := GetPathSeparator()
	expected := string(filepath.Separator)
	if sep != expected {
		t.Errorf("Expected '%s', got '%s'", expected, sep)
	}
}

func TestIsValidPath(t *testing.T) {
	// Empty path is invalid
	if IsValidPath("") {
		t.Error("Empty path should be invalid")
	}

	// Path with null byte is invalid
	if IsValidPath("test\x00path") {
		t.Error("Path with null byte should be invalid")
	}

	// Normal path is valid
	if !IsValidPath("normal/path") {
		t.Error("Normal path should be valid")
	}

	// Windows-specific tests
	if runtime.GOOS == "windows" {
		if IsValidPath("test<path") {
			t.Error("Path with < should be invalid on Windows")
		}
		if IsValidPath("test>path") {
			t.Error("Path with > should be invalid on Windows")
		}
		if IsValidPath("test|path") {
			t.Error("Path with | should be invalid on Windows")
		}
		// C: should be valid
		if !IsValidPath("C:\\test") {
			t.Error("C:\\test should be valid on Windows")
		}
	}
}

func TestCleanPath(t *testing.T) {
	// Test removing ..
	result := CleanPath("a/b/../c")
	if strings.Contains(result, "..") {
		t.Error("CleanPath should resolve ..")
	}

	// Test removing .
	result = CleanPath("a/./b")
	if strings.Contains(result, "/.") || strings.Contains(result, "\\.") {
		t.Error("CleanPath should remove .")
	}
}
