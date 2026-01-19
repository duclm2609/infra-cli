package infra

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/infra-cli/cmd/aws"
	"github.com/user/infra-cli/internal/config"
	infraerrors "github.com/user/infra-cli/internal/errors"
	infraoutput "github.com/user/infra-cli/internal/output"
)

// Version is set at build time via ldflags
var Version = "dev"

// Global flags
var (
	verbose bool
	quiet   bool
	output  string
)

// Global components
var (
	configManager   *config.Manager
	outputFormatter *infraoutput.Formatter
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "infra",
	Short: "A CLI tool for DevOps/CloudOps automation tasks",
	Long: `Infra is a cross-platform CLI tool designed to assist DevOps/CloudOps 
engineers with daily automation tasks.

It provides various sub-commands for interacting with cloud services,
starting with AWS SSO authentication and expanding to other operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Display help when run without arguments
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Initialize components
	initComponents()

	if err := rootCmd.Execute(); err != nil {
		// Format error based on type
		exitCode := infraerrors.ExitInternalError
		if infraErr, ok := err.(*infraerrors.InfraError); ok {
			fmt.Fprintln(os.Stderr, infraErr.Format(verbose))
			exitCode = infraErr.ExitCode()
		} else {
			fmt.Fprintln(os.Stderr, infraerrors.FormatError(err, verbose))
			exitCode = infraerrors.GetExitCode(err)
		}
		os.Exit(exitCode)
	}
}

// initComponents initializes global components
func initComponents() {
	// Initialize config manager
	configManager = config.NewManager()
	cfg, err := configManager.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
	}

	// Apply config defaults if not overridden by flags
	if cfg != nil {
		if output == "table" && cfg.DefaultOutput != "" {
			output = cfg.DefaultOutput
		}
		if !verbose && cfg.Verbose {
			verbose = cfg.Verbose
		}
	}

	// Initialize output formatter
	outputFormatter = infraoutput.NewFormatter(output)
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "Output format (json, yaml, table)")

	// Version flag
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("infra version {{.Version}}\n")

	// Add sub-commands
	rootCmd.AddCommand(aws.AWSCmd)
}

// GetVerbose returns the verbose flag value
func GetVerbose() bool {
	return verbose
}

// GetQuiet returns the quiet flag value
func GetQuiet() bool {
	return quiet
}

// GetOutput returns the output format
func GetOutput() string {
	return output
}

// GetConfigManager returns the config manager
func GetConfigManager() *config.Manager {
	return configManager
}

// GetOutputFormatter returns the output formatter
func GetOutputFormatter() *infraoutput.Formatter {
	return outputFormatter
}
