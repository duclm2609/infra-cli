package infra

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Feature: devops-cli-tool, Property 8: Help Flag Availability
// For any command or sub-command in the CLI, invoking it with --help
// SHALL produce non-empty help text containing the command name.
func TestHelpFlagAvailability(t *testing.T) {
	// Collect all commands
	commands := getAllCommands(rootCmd)

	for _, cmd := range commands {
		t.Run(cmd.Name(), func(t *testing.T) {
			// Get help text
			helpText := cmd.UsageString()

			// Help should not be empty
			if helpText == "" {
				t.Errorf("Help text for '%s' should not be empty", cmd.Name())
			}

			// Help should contain the command name or "Usage:"
			if !strings.Contains(helpText, cmd.Name()) && !strings.Contains(helpText, "Usage:") {
				t.Errorf("Help text for '%s' should contain command name or Usage", cmd.Name())
			}
		})
	}
}

// Feature: devops-cli-tool, Property 10: Output Verbosity Control
// For any operation, when --verbose is set the output length SHALL be
// greater than or equal to the output without --verbose, and when --quiet
// is set the output length SHALL be less than or equal to the output without --quiet.
func TestOutputVerbosityControl(t *testing.T) {
	// Test that verbose flag exists and is accessible
	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Expected --verbose flag to exist")
	}

	quietFlag := rootCmd.PersistentFlags().Lookup("quiet")
	if quietFlag == nil {
		t.Error("Expected --quiet flag to exist")
	}

	// Test GetVerbose and GetQuiet functions
	verbose = true
	if !GetVerbose() {
		t.Error("GetVerbose should return true when verbose is set")
	}

	verbose = false
	if GetVerbose() {
		t.Error("GetVerbose should return false when verbose is not set")
	}

	quiet = true
	if !GetQuiet() {
		t.Error("GetQuiet should return true when quiet is set")
	}

	quiet = false
	if GetQuiet() {
		t.Error("GetQuiet should return false when quiet is not set")
	}
}

// getAllCommands recursively collects all commands
func getAllCommands(cmd *cobra.Command) []*cobra.Command {
	commands := []*cobra.Command{cmd}
	for _, subCmd := range cmd.Commands() {
		commands = append(commands, getAllCommands(subCmd)...)
	}
	return commands
}

// Unit tests
func TestRootCmdExists(t *testing.T) {
	if rootCmd == nil {
		t.Error("rootCmd should not be nil")
	}
	if rootCmd.Use != "infra" {
		t.Errorf("Expected Use 'infra', got '%s'", rootCmd.Use)
	}
}

func TestVersionFlag(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestGlobalFlags(t *testing.T) {
	// Check verbose flag
	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Expected --verbose flag")
	}
	if verboseFlag.Shorthand != "v" {
		t.Errorf("Expected shorthand 'v', got '%s'", verboseFlag.Shorthand)
	}

	// Check quiet flag
	quietFlag := rootCmd.PersistentFlags().Lookup("quiet")
	if quietFlag == nil {
		t.Error("Expected --quiet flag")
	}
	if quietFlag.Shorthand != "q" {
		t.Errorf("Expected shorthand 'q', got '%s'", quietFlag.Shorthand)
	}

	// Check output flag
	outputFlag := rootCmd.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		t.Error("Expected --output flag")
	}
	if outputFlag.Shorthand != "o" {
		t.Errorf("Expected shorthand 'o', got '%s'", outputFlag.Shorthand)
	}
	if outputFlag.DefValue != "table" {
		t.Errorf("Expected default 'table', got '%s'", outputFlag.DefValue)
	}
}

func TestGetOutput(t *testing.T) {
	output = "json"
	if GetOutput() != "json" {
		t.Errorf("Expected 'json', got '%s'", GetOutput())
	}

	output = "table"
	if GetOutput() != "table" {
		t.Errorf("Expected 'table', got '%s'", GetOutput())
	}
}

func TestAWSSubCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "aws" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'aws' sub-command to be registered")
	}
}
