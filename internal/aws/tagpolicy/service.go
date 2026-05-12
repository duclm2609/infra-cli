package tagpolicy

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/smithy-go"

	infraerrors "github.com/duclm2609/infra-cli/internal/errors"
)

// OrganizationsClient interface for AWS Organizations operations.
// This interface allows for mocking in tests.
type OrganizationsClient interface {
	DescribeEffectivePolicy(ctx context.Context, params *organizations.DescribeEffectivePolicyInput, optFns ...func(*organizations.Options)) (*organizations.DescribeEffectivePolicyOutput, error)
}

// TagPolicyService handles tag policy operations with AWS Organizations.
type TagPolicyService struct {
	orgClient OrganizationsClient
	parser    *PolicyParser
	profile   string
	region    string
}

// NewTagPolicyService creates a new tag policy service.
// It initializes the AWS Organizations client using the specified profile and region.
func NewTagPolicyService(profile, region string) (*TagPolicyService, error) {
	ctx := context.Background()

	// Build config options
	opts := []func(*config.LoadOptions) error{}

	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Organizations client
	orgClient := organizations.NewFromConfig(cfg)

	return &TagPolicyService{
		orgClient: orgClient,
		parser:    NewPolicyParser(),
		profile:   profile,
		region:    region,
	}, nil
}

// NewTagPolicyServiceWithClient creates a new tag policy service with a custom Organizations client.
// This is useful for testing with mock clients.
func NewTagPolicyServiceWithClient(client OrganizationsClient, profile, region string) *TagPolicyService {
	return &TagPolicyService{
		orgClient: client,
		parser:    NewPolicyParser(),
		profile:   profile,
		region:    region,
	}
}

// mapAWSError converts AWS API errors to infraerrors types for consistent error handling.
// It handles specific AWS Organizations error codes and maps them to appropriate error types.
func mapAWSError(err error) error {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EffectivePolicyNotFoundException":
			return infraerrors.NewNoTagPolicyError()
		case "AWSOrganizationsNotInUseException":
			return infraerrors.NewNotInOrganizationError(err)
		case "AccessDeniedException":
			return infraerrors.NewAWSAPIError("AccessDenied", "Insufficient permissions to describe tag policy", "organizations", "DescribeEffectivePolicy", err)
		default:
			return infraerrors.NewAWSAPIError(apiErr.ErrorCode(), apiErr.ErrorMessage(), "organizations", "DescribeEffectivePolicy", err)
		}
	}
	return infraerrors.NewInternalError("unexpected error from AWS Organizations", err)
}

// GetEffectiveTagPolicy retrieves the effective tag policy for the account.
// It calls the AWS Organizations DescribeEffectivePolicy API with TAG_POLICY type.
//
// Returns the parsed TagPolicy if successful, or an error if:
// - No tag policy exists for the account
// - The account is not part of an AWS Organization
// - The user lacks permissions to describe effective policies
// - The policy content cannot be parsed
func (s *TagPolicyService) GetEffectiveTagPolicy(ctx context.Context) (*TagPolicy, error) {
	// Call DescribeEffectivePolicy API
	// Note: When TargetId is not specified, it defaults to the current account
	input := &organizations.DescribeEffectivePolicyInput{
		PolicyType: types.EffectivePolicyTypeTagPolicy,
	}

	output, err := s.orgClient.DescribeEffectivePolicy(ctx, input)
	if err != nil {
		return nil, mapAWSError(err)
	}

	// Check if policy content exists
	if output.EffectivePolicy == nil || output.EffectivePolicy.PolicyContent == nil {
		return nil, infraerrors.NewNoTagPolicyError()
	}

	policyContent := *output.EffectivePolicy.PolicyContent

	// Handle empty policy content
	if policyContent == "" {
		return nil, infraerrors.NewNoTagPolicyError()
	}

	// Parse the policy content
	tagPolicy, err := s.parser.Parse(policyContent)
	if err != nil {
		return nil, infraerrors.NewInternalError(fmt.Sprintf("failed to parse tag policy: %v", err), err)
	}

	return tagPolicy, nil
}

// GetProfile returns the AWS profile used by this service.
func (s *TagPolicyService) GetProfile() string {
	return s.profile
}

// GetRegion returns the AWS region used by this service.
func (s *TagPolicyService) GetRegion() string {
	return s.region
}
