# Infra CLI

A cross-platform CLI tool for DevOps/CloudOps engineers to automate daily tasks with AWS and other cloud services.

## Features

- AWS SSO authentication with profile support
- Cross-platform (MacOS, Linux, Windows)
- Multiple output formats (JSON, YAML, table)

## Installation

### From Source

```bash
# Clone and build
git clone https://github.com/user/infra-cli.git
cd infra-cli
make build

# Or build for all platforms
make build-all
```

### Pre-built Binaries

Download from releases and add to your PATH:
- `infra-darwin-amd64` (MacOS Intel)
- `infra-darwin-arm64` (MacOS Apple Silicon)
- `infra-linux-amd64` (Linux x64)
- `infra-linux-arm64` (Linux ARM)
- `infra-windows-amd64.exe` (Windows)

## Usage

### AWS Commands

```bash
# Show current AWS identity
infra aws whoami

# Use specific profile
infra aws whoami --profile my-sso-profile

# Login to AWS SSO
infra aws login --profile my-sso-profile

# List available profiles
infra aws profiles
```

### Global Flags

```bash
--profile, -p    AWS profile to use
--region, -r     AWS region
--output, -o     Output format: json, yaml, table (default: table)
--verbose, -v    Verbose output
--quiet, -q      Suppress non-essential output
```

### Examples

```bash
# Get identity as JSON
infra aws whoami -o json

# Use specific region
infra aws whoami --profile prod --region us-west-2
```

## Requirements

- Go 1.24+ (for building)
- AWS CLI configured with SSO profiles in `~/.aws/config`

## License

MIT
