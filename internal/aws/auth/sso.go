package auth

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Credentials represents AWS credentials
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
	Source          string
}

// SSOAuthenticator manages AWS SSO authentication
type SSOAuthenticator struct {
	profile string
	region  string
}

// NewSSOAuthenticator creates a new SSO authenticator
func NewSSOAuthenticator(profile, region string) *SSOAuthenticator {
	return &SSOAuthenticator{
		profile: profile,
		region:  region,
	}
}

// Authenticate performs SSO authentication for the given profile
func (a *SSOAuthenticator) Authenticate(ctx context.Context) (*Credentials, error) {
	// Load AWS config with the specified profile
	cfg, err := a.loadConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Try to get credentials
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		// If credentials are expired or invalid, try SSO login
		if err := a.initiateLogin(); err != nil {
			return nil, fmt.Errorf("SSO login failed: %w", err)
		}

		// Reload config after login
		cfg, err = a.loadConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to reload AWS config after SSO login: %w", err)
		}

		creds, err = cfg.Credentials.Retrieve(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve credentials after SSO login: %w", err)
		}
	}

	return &Credentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Expiration:      creds.Expires,
		Source:          creds.Source,
	}, nil
}

// IsAuthenticated checks if valid credentials exist
func (a *SSOAuthenticator) IsAuthenticated(ctx context.Context) bool {
	cfg, err := a.loadConfig(ctx)
	if err != nil {
		return false
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return false
	}

	// Check if credentials are expired
	if !creds.Expires.IsZero() && creds.Expires.Before(time.Now()) {
		return false
	}

	return true
}

// GetCredentials returns cached credentials if valid
func (a *SSOAuthenticator) GetCredentials(ctx context.Context) (*Credentials, error) {
	cfg, err := a.loadConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	return &Credentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Expiration:      creds.Expires,
		Source:          creds.Source,
	}, nil
}

// CallerIdentity represents the AWS caller identity
type CallerIdentity struct {
	Account string
	Arn     string
	UserId  string
}

// ValidateCredentials validates that credentials are available and valid
func (a *SSOAuthenticator) ValidateCredentials(ctx context.Context) error {
	cfg, err := a.loadConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Use STS GetCallerIdentity to validate credentials
	stsClient := sts.NewFromConfig(cfg)
	_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}

	return nil
}

// GetCallerIdentity returns the AWS caller identity (account, ARN, user ID)
func (a *SSOAuthenticator) GetCallerIdentity(ctx context.Context) (*CallerIdentity, error) {
	cfg, err := a.loadConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	result, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}

	return &CallerIdentity{
		Account: aws.ToString(result.Account),
		Arn:     aws.ToString(result.Arn),
		UserId:  aws.ToString(result.UserId),
	}, nil
}

// GetConfig returns the AWS config for the profile
func (a *SSOAuthenticator) GetConfig(ctx context.Context) (aws.Config, error) {
	return a.loadConfig(ctx)
}

// loadConfig loads AWS configuration with the specified profile
func (a *SSOAuthenticator) loadConfig(ctx context.Context) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithSharedConfigProfile(a.profile),
	}

	if a.region != "" {
		opts = append(opts, config.WithRegion(a.region))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}

// initiateLogin opens the browser for SSO login
func (a *SSOAuthenticator) initiateLogin() error {
	// Use AWS CLI to perform SSO login
	cmd := exec.Command("aws", "sso", "login", "--profile", a.profile)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("aws sso login command failed: %w", err)
	}

	return nil
}

// openBrowser opens the default browser with the given URL
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// AssumeRoleAuthenticator handles role assumption
type AssumeRoleAuthenticator struct {
	baseAuth      *SSOAuthenticator
	roleARN       string
	sessionName   string
	externalID    string
	durationSecs  int32
}

// NewAssumeRoleAuthenticator creates a new role assumption authenticator
func NewAssumeRoleAuthenticator(baseAuth *SSOAuthenticator, roleARN string) *AssumeRoleAuthenticator {
	return &AssumeRoleAuthenticator{
		baseAuth:     baseAuth,
		roleARN:      roleARN,
		sessionName:  "infra-cli-session",
		durationSecs: 3600,
	}
}

// Authenticate assumes the role and returns credentials
func (a *AssumeRoleAuthenticator) Authenticate(ctx context.Context) (*Credentials, error) {
	baseCfg, err := a.baseAuth.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get base config: %w", err)
	}

	stsClient := sts.NewFromConfig(baseCfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, a.roleARN, func(o *stscreds.AssumeRoleOptions) {
		o.RoleSessionName = a.sessionName
		if a.externalID != "" {
			o.ExternalID = &a.externalID
		}
		o.Duration = time.Duration(a.durationSecs) * time.Second
	})

	creds, err := provider.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %w", err)
	}

	return &Credentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Expiration:      creds.Expires,
		Source:          "AssumeRole",
	}, nil
}
