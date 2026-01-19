package aws

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/infra-cli/internal/aws/auth"
	"github.com/user/infra-cli/internal/aws/profile"
	infraerrors "github.com/user/infra-cli/internal/errors"
)

var (
	awsProfile string
	awsRegion  string
)

// AWSCmd represents the aws command
var AWSCmd = &cobra.Command{
	Use:   "aws",
	Short: "AWS-related commands and operations",
	Long: `AWS sub-command provides various operations for interacting with AWS services.

It supports AWS SSO authentication and allows you to specify which AWS profile
and region to use for operations.

Examples:
  infra aws --profile my-sso-profile
  infra aws --region us-west-2`,
	Run: func(cmd *cobra.Command, args []string) {
		// Display help when run without sub-commands
		cmd.Help()
	},
}

// WhoamiCmd shows the current AWS identity
var WhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current AWS identity",
	Long:  `Display information about the currently authenticated AWS identity.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWhoami()
	},
}

// LoginCmd initiates SSO login
var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to AWS SSO",
	Long:  `Initiate AWS SSO login flow for the specified profile.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogin()
	},
}

// ProfilesCmd lists available AWS profiles
var ProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List available AWS profiles",
	Long:  `List all AWS profiles configured in ~/.aws/config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProfiles()
	},
}

func init() {
	// AWS-specific flags
	AWSCmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", "", "AWS profile to use")
	AWSCmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "AWS region to use")

	// Add sub-commands
	AWSCmd.AddCommand(WhoamiCmd)
	AWSCmd.AddCommand(LoginCmd)
	AWSCmd.AddCommand(ProfilesCmd)
}

// GetProfile returns the resolved AWS profile
func GetProfile() string {
	pm := profile.NewManager()
	return pm.ResolveProfile(awsProfile)
}

// GetRegion returns the AWS region (flag or empty for default)
func GetRegion() string {
	return awsRegion
}

// ValidateCredentials validates AWS credentials before operations
func ValidateCredentials() error {
	resolvedProfile := GetProfile()
	authenticator := auth.NewSSOAuthenticator(resolvedProfile, awsRegion)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !authenticator.IsAuthenticated(ctx) {
		return infraerrors.NewCredentialsNotFoundError(resolvedProfile, nil)
	}

	return nil
}

func runWhoami() error {
	resolvedProfile := GetProfile()
	authenticator := auth.NewSSOAuthenticator(resolvedProfile, awsRegion)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get caller identity (validates credentials and returns account info)
	identity, err := authenticator.GetCallerIdentity(ctx)
	if err != nil {
		return infraerrors.NewCredentialsNotFoundError(resolvedProfile, err)
	}

	creds, err := authenticator.GetCredentials(ctx)
	if err != nil {
		return infraerrors.NewAuthError("CREDS_FAILED", "Failed to get credentials", resolvedProfile, err)
	}

	fmt.Printf("Account: %s\n", identity.Account)
	fmt.Printf("Arn: %s\n", identity.Arn)
	fmt.Printf("UserId: %s\n", identity.UserId)
	fmt.Printf("Profile: %s\n", resolvedProfile)
	fmt.Printf("Region: %s\n", getEffectiveRegion(resolvedProfile))
	fmt.Printf("Credential Source: %s\n", creds.Source)
	if !creds.Expiration.IsZero() {
		fmt.Printf("Expires: %s\n", creds.Expiration.Format(time.RFC3339))
	}

	return nil
}

func runLogin() error {
	resolvedProfile := GetProfile()
	authenticator := auth.NewSSOAuthenticator(resolvedProfile, awsRegion)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Printf("Initiating SSO login for profile '%s'...\n", resolvedProfile)

	creds, err := authenticator.Authenticate(ctx)
	if err != nil {
		return infraerrors.NewAuthError("LOGIN_FAILED", "SSO login failed", resolvedProfile, err)
	}

	fmt.Println("Successfully authenticated!")
	fmt.Printf("Credential Source: %s\n", creds.Source)
	if !creds.Expiration.IsZero() {
		fmt.Printf("Expires: %s\n", creds.Expiration.Format(time.RFC3339))
	}

	return nil
}

func runProfiles() error {
	pm := profile.NewManager()
	profiles, err := pm.ListProfiles()
	if err != nil {
		return infraerrors.NewConfigError("PROFILES_FAILED", "Failed to list profiles", "file", err)
	}

	if len(profiles) == 0 {
		fmt.Println("No profiles found in ~/.aws/config")
		return nil
	}

	fmt.Println("Available AWS profiles:")
	for _, p := range profiles {
		marker := "  "
		if p == GetProfile() {
			marker = "* "
		}
		fmt.Printf("%s%s\n", marker, p)
	}

	return nil
}

func getEffectiveRegion(profileName string) string {
	// If region flag is set, use it
	if awsRegion != "" {
		return awsRegion
	}

	// Try to get region from profile
	pm := profile.NewManager()
	cfg, err := pm.GetProfileConfig(profileName)
	if err == nil && cfg.Region != "" {
		return cfg.Region
	}

	// Check environment variable
	if envRegion := os.Getenv("AWS_REGION"); envRegion != "" {
		return envRegion
	}
	if envRegion := os.Getenv("AWS_DEFAULT_REGION"); envRegion != "" {
		return envRegion
	}

	return "us-east-1" // Default
}
