package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/duclm2609/infra-cli/internal/aws/tagpolicy"
)

// TagPolicyCmd represents the tag-policy subcommand
var TagPolicyCmd = &cobra.Command{
	Use:   "tag-policy",
	Short: "Display effective tag policy for the AWS account",
	Long: `Retrieves and displays the effective tag policy from AWS Organizations.

Use arrow keys to navigate between tag keys, Enter/Space to expand/collapse,
and 'q' to quit.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTagPolicy()
	},
}

// runTagPolicy executes the tag-policy command logic.
// It validates credentials, retrieves the effective tag policy from AWS Organizations,
// and displays an interactive TUI for navigating tag keys and their allowed values.
//
// Requirements: 1.1, 1.4, 5.3, 5.4
func runTagPolicy() error {
	// 1. Validate credentials before making API calls
	// Requirement 5.4: Use existing ValidateCredentials() function
	if err := ValidateCredentials(); err != nil {
		return err
	}

	// 2. Create service with profile and region
	// Requirement 5.3: Use existing GetProfile() and GetRegion() helper functions
	service, err := tagpolicy.NewTagPolicyService(GetProfile(), GetRegion())
	if err != nil {
		return err
	}

	// 3. Get effective tag policy with timeout
	// Requirement 1.1: Retrieve the effective tag policy from AWS Organizations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	policy, err := service.GetEffectiveTagPolicy(ctx)
	if err != nil {
		return err
	}

	// 4. Handle empty policy case
	// Requirement 1.4: Display message if no tag policy exists
	if len(policy.Tags) == 0 {
		fmt.Println("No tag keys found in the effective tag policy.")
		return nil
	}

	// 5. Create and run TUI for interactive display
	view := tagpolicy.NewTagPolicyView(policy)
	return view.Run()
}
