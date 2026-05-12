# Infra CLI

A cross-platform CLI tool for DevOps/CloudOps engineers to automate daily tasks with AWS and other cloud services.

## Features

- AWS SSO authentication with profile support
- Cross-platform (MacOS, Linux, Windows)
- Multiple output formats (JSON, YAML, table)

## Installation

### macOS / Linux (one-liner)

```bash
curl -fsSL https://raw.githubusercontent.com/duclm2609/infra-cli/main/install.sh | bash
```

Installs to `/usr/local/bin/infra`. Auto-detects OS and architecture.

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/duclm2609/infra-cli/main/install.ps1 | iex
```

Installs to `$env:LOCALAPPDATA\infra\infra.exe` and adds it to your PATH.

### From Source

```bash
git clone https://github.com/duclm2609/infra-cli.git
cd infra-cli
make build
```

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
