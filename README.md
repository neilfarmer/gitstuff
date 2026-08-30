# GitStuff

[![CI](https://github.com/neilfarmer/gitstuff/actions/workflows/ci.yml/badge.svg)](https://github.com/neilfarmer/gitstuff/actions/workflows/ci.yml)
[![Release](https://github.com/neilfarmer/gitstuff/actions/workflows/release.yml/badge.svg)](https://github.com/neilfarmer/gitstuff/actions/workflows/release.yml)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/neilfarmer/gitstuff)](https://github.com/neilfarmer/gitstuff/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/neilfarmer/gitstuff)](https://golang.org/)

A comprehensive Go CLI application for managing GitLab and GitHub repositories. This tool allows you to list all repositories across multiple SCM providers, clone them individually or all at once, and check their local status including current working branch.

## Quick Start

**Linux x86_64:**
```bash
curl -L https://github.com/neilfarmer/gitstuff/releases/latest/download/gitstuff-linux-amd64 -o gitstuff
chmod +x gitstuff
sudo mv gitstuff /usr/local/bin/
```

**Linux ARM64:**
```bash
curl -L https://github.com/neilfarmer/gitstuff/releases/latest/download/gitstuff-linux-arm64 -o gitstuff
chmod +x gitstuff
sudo mv gitstuff /usr/local/bin/
```

**macOS x86_64:**
```bash
curl -L https://github.com/neilfarmer/gitstuff/releases/latest/download/gitstuff-darwin-amd64 -o gitstuff
chmod +x gitstuff
sudo mv gitstuff /usr/local/bin/
```

**macOS ARM64 (M1/M2):**
```bash
curl -L https://github.com/neilfarmer/gitstuff/releases/latest/download/gitstuff-darwin-arm64 -o gitstuff
chmod +x gitstuff
sudo mv gitstuff /usr/local/bin/
```

**Then configure and use:**
```bash
# Configure your GitLab and/or GitHub connections
gitstuff config

# List your repositories from all providers
gitstuff list

# Clone all repositories
gitstuff clone --all
```

## Features

- **Multi-Provider Support**: Connect to both GitLab and GitHub simultaneously
- **List Repositories**: View all repositories with hierarchical group/organization structure
- **Group Filtering**: Filter repositories by GitLab group or GitHub organization
- **Clone Management**: Download single repositories or all at once from any provider
- **Status Tracking**: See which repos are cloned and their current branch
- **Provider-Aware Display**: Clear indication of which provider each repository comes from
- **Flexible Authentication**: Supports both HTTPS and SSH cloning
- **Update Support**: Pull latest changes for already cloned repositories

## Installation

Choose the appropriate command for your platform:

**Linux x86_64:**

```bash
curl -L https://github.com/neilfarmer/gitstuff/releases/latest/download/gitstuff-linux-amd64 -o gitstuff
chmod +x gitstuff
sudo mv gitstuff /usr/local/bin/
```

**Linux ARM64:**

```bash
curl -L https://github.com/neilfarmer/gitstuff/releases/latest/download/gitstuff-linux-arm64 -o gitstuff
chmod +x gitstuff
sudo mv gitstuff /usr/local/bin/
```

**macOS x86_64:**

```bash
curl -L https://github.com/neilfarmer/gitstuff/releases/latest/download/gitstuff-darwin-amd64 -o gitstuff
chmod +x gitstuff
sudo mv gitstuff /usr/local/bin/
```

**macOS ARM64 (M1/M2):**

```bash
curl -L https://github.com/neilfarmer/gitstuff/releases/latest/download/gitstuff-darwin-arm64 -o gitstuff
chmod +x gitstuff
sudo mv gitstuff /usr/local/bin/
```

Or download directly from the [releases page](https://github.com/neilfarmer/gitstuff/releases/latest).

### Build from Source

**Prerequisites:** Go 1.21 or later

```bash
# Clone the repository
git clone https://github.com/neilfarmer/gitstuff.git
cd gitstuff

# Build using make
make build

# Or build directly
go build -o gitstuff .

# Install system-wide
make install
```

## Configuration

The CLI supports both GitLab and GitHub providers. You can configure one or multiple providers:

```bash
gitstuff config
```

This will present a numbered menu to select which provider to add:

1. **GitLab**
2. **GitHub**

For each provider, you'll be prompted for:

- **Provider Name**: A unique identifier for this provider (e.g., "gitlab-work", "github-personal")
- **Provider URL**: 
  - GitLab: Your GitLab instance URL (e.g., `https://gitlab.com` or `gitlab.example.com`)
  - GitHub: Leave blank for github.com or enter GitHub Enterprise URL
- **Access Token**: Your provider-specific access token
- **Base Directory**: Local directory for cloned repositories (default: `~/gitstuff-repos`)
- **SSL Certificate Verification**: Whether to skip SSL verification for self-signed certificates  
- **Default Group/Organization Filter**: Optional filter for repositories

After configuring one provider, you'll be asked if you want to add another provider.

> **Note**: The CLI automatically adds `https://` to URLs that don't specify a protocol.

### Creating Access Tokens

**For GitLab:**
1. Go to your GitLab instance
2. Navigate to User Settings > Access Tokens
3. Create a token with at least `read_repository` scope
4. Copy the token for use with the CLI

**For GitHub:**
1. Go to GitHub.com (or your GitHub Enterprise instance)
2. Navigate to Settings > Developer settings > Personal access tokens > Tokens (classic)
3. Generate a new token with `repo` scope for private repositories or `public_repo` for public only
4. Copy the token for use with the CLI

### Alternative Configuration

You can also configure using command flags:

```bash
# Configure a GitLab provider
gitstuff config --provider gitlab --name gitlab-work --url https://gitlab.example.com --token your-gitlab-token --base-dir /path/to/repos

# Configure a GitHub provider  
gitstuff config --provider github --name github-personal --url https://github.com --token your-github-token

# For instances with self-signed certificates
gitstuff config --provider gitlab --name gitlab-work --url https://gitlab.example.com --token your-token --insecure
```

## Usage

### List All Repositories

The list command displays repositories from all configured providers:

```bash
# Simple list view (shows folder structure, status, and provider)
gitstuff list

# Tree view with group/organization structure (organized by provider)
gitstuff list --tree

# Show additional details like URLs (info level)
gitstuff list -v

# Show debug information with timing
gitstuff list -vv

# Maximum verbosity with trace information
gitstuff list -vvv

# List without status information
gitstuff list --status=false

# Filter by specific group/organization (works across all providers)
gitstuff list --group my-team
```

**Example output:**
```
Found 15 repositories:

📁 [gitlab] company/backend-api
   Status: ✅ Cloned (branch: main) 🔄 Has uncommitted changes

📁 [github] myuser/personal-project  
   Status: ❌ Not cloned

📁 [gitlab] team/frontend-app
   Status: ✅ Cloned (branch: develop)
```

### Clone Repositories

```bash
# Clone a specific repository
gitstuff clone group/project-name

# Clone all repositories
gitstuff clone --all

# Clone using SSH instead of HTTPS
gitstuff clone --ssh group/project-name

# Update already cloned repositories
gitstuff clone --all --update
```

## Repository Structure

The CLI maintains the exact provider group/organization structure on your filesystem with provider separation:

```text
~/gitstuff-repos/
├── gitlab/                    # GitLab provider
│   ├── gitlab-group1/
│   │   ├── project1/
│   │   ├── project2/
│   │   └── subgroup1/
│   │       └── nested-project/
│   └── gitlab-group2/
│       └── another-project/
└── github/                    # GitHub provider
    ├── github-org/
    │   ├── public-repo/
    │   └── private-repo/
    └── github-user/
        └── personal-project/
```

**Legacy Structure Support**: If you have repositories already cloned without the provider subdirectories (e.g., directly in `~/gitstuff-repos/group/project`), GitStuff will automatically detect and work with them. New clones will use the provider-based structure shown above.

## Repository Status Information

The CLI shows comprehensive status for each repository:

- ✅ **Cloned**: Repository exists locally and is a valid git repository
- ❌ **Not cloned**: Repository doesn't exist locally
- ⚠️ **Directory exists but not git repo**: Directory exists but isn't initialized as git
- 🔄 **Has uncommitted changes**: Repository has local modifications
- **Branch information**: Current working branch name

## Configuration File

Configuration is stored in `~/.gitstuff.yaml` and supports multiple providers:

```yaml
providers:
  - name: "gitlab-work"
    type: "gitlab"
    url: "https://gitlab.company.com"
    token: "your-gitlab-token"
    insecure: false
    group: "backend-team"
  - name: "github-personal"
    type: "github" 
    url: "https://github.com"
    token: "your-github-token"
    insecure: false
    group: "myorg"
local:
  base_dir: "/path/to/gitstuff-repos"
```

## Verbosity Levels

GitStuff supports multiple verbosity levels using the `-v` flag. Each additional `-v` increases the detail level:

- **Normal (default)**: Essential output only
- **`-v` (Info)**: Shows additional details like repository URLs
- **`-vv` (Debug)**: Shows API call timing, internal processing details, and configuration info
- **`-vvv` (Trace)**: Maximum detail including all debug info plus trace-level logging

**Examples:**
```bash
# Normal output
gitstuff list

# Show repository URLs and additional info
gitstuff list -v

# Show debug information with API timing
gitstuff clone --all -vv

# Maximum verbosity for troubleshooting
gitstuff config -vvv
```

The verbosity setting applies globally to all commands and can help with troubleshooting connection issues, understanding performance, and debugging configuration problems.

## Commands Reference

### `gitstuff config`

Configure SCM provider connections (GitLab and/or GitHub).

**Flags:**

- `-p, --provider`: Provider type (`gitlab` or `github`)
- `-n, --name`: Provider name (identifier for multiple providers)
- `-u, --url`: Provider instance URL
- `-t, --token`: Provider access token
- `-d, --base-dir`: Base directory for repositories
- `-k, --insecure`: Skip SSL certificate verification (for self-signed certificates)
- `-g, --group`: Default group/organization to filter repositories (optional)

### `gitstuff list`

List repositories from all configured providers with status information.

**Flags:**

- `-t, --tree`: Display in tree structure organized by provider and groups/organizations
- `-s, --status`: Show local repository status (default: true)
- `-v, --verbose`: Increase verbosity (use -v, -vv, -vvv for info, debug, trace levels)
- `-g, --group`: Filter repositories to only those in the specified group/organization

### `gitstuff clone`

Clone repositories from configured providers.

**Usage:**

- `gitstuff clone [repository-path]`: Clone specific repository
- `gitstuff clone --all`: Clone all repositories from all providers

**Flags:**

- `-a, --all`: Clone all repositories from all providers
- `-s, --ssh`: Use SSH for cloning (default: HTTPS)
- `-u, --update`: Pull latest changes for existing repositories

**Note:** Clone command currently supports GitLab providers only. GitHub support for cloning is coming in a future update.

## Examples

### Basic Workflow

```bash
# 1. Configure your first provider (GitLab or GitHub)
gitstuff config

# 2. Optionally add additional providers
gitstuff config  # This will ask if you want to add another provider

# 3. List all repositories from all providers to see what's available
gitstuff list --tree

# 4. Clone all repositories (currently GitLab only)
gitstuff clone --all

# 5. Later, update all repositories
gitstuff clone --all --update
```

### Group/Organization Filtering

Filter repositories by GitLab group or GitHub organization across all providers:

```bash
# List repositories only in the "backend" group/organization
gitstuff list --group backend

# List repositories in a nested GitLab group
gitstuff list --group team-a/backend

# Use tree view with group/organization filtering
gitstuff list --tree --group team-a

# Set a default group/organization in provider config
gitstuff config --provider gitlab --name work --group team-backend
gitstuff config --provider github --name personal --group myorg

# Override config default with command flag
gitstuff list --group different-team
```

### Working with Specific Repositories

```bash
# Clone a specific project
gitstuff clone mygroup/myproject

# Update a specific project
gitstuff clone mygroup/myproject --update

# Use SSH for cloning
gitstuff clone mygroup/myproject --ssh
```

## Requirements

- Go 1.19 or later
- Git installed and available in PATH
- Access tokens for your SCM providers (GitLab and/or GitHub) with appropriate permissions
- Network access to your SCM provider instances

## Testing

```bash
# Run all tests with clear output
make test

# Run tests with detailed output
make test-verbose
```

## Error Handling

The CLI provides clear error messages for common issues:

- **Missing configuration**: Prompts to run `gitstuff config`
- **Invalid GitLab credentials**: Clear authentication error messages
- **Network issues**: Helpful network connectivity error messages
- **Git errors**: Detailed git operation error messages

## Releases

### Creating a Release

To create a new release:

1. **Tag the release:**

   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```

2. **GitHub Actions automatically:**

   - Runs all tests
   - Builds binaries for all platforms
   - Creates GitHub release with download links
   - Generates install scripts

3. **Release artifacts include:**
   - Cross-platform binaries (Linux, macOS, Windows)
   - Architecture support (x64, ARM64)
   - Automated install scripts
   - SHA256 checksums for verification

### Versioning

This project follows [Semantic Versioning](https://semver.org/):

- `v1.2.3` - Major.Minor.Patch
- `v1.2.0-beta.1` - Pre-release versions

## License

This project is open source and available under the MIT License.
