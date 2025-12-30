# StackToDate

Official command-line interface for [Stack To Date](https://stacktodate.club/) — a service that helps development teams track technology lifecycle statuses and plan for end-of-life (EOL) upgrades.

## About Stack To Date

Stack To Date enables teams to:
- **Track lifecycle statuses** of technologies in your tech stack
- **Monitor EOL dates** and support timelines
- **Plan strategic upgrades** ahead of deadlines
- **Manage multiple projects** with different tech stacks

This CLI tool integrates with the Stack To Date platform by automatically detecting your project's technology stack and syncing it to your remote profile.

## Features

- **Auto-detect technologies**: Scans your project and identifies:
  - Programming languages (Go, Python, Node.js, Ruby)
  - Frameworks (Rails, Django, Express, etc.)
  - Container configuration (Docker, Docker Compose)
  - Version information from config files

- **Tech stack management**: Initialize, update, and maintain a `stacktodate.yml` configuration file with your project's tech stack

- **Push to Stack To Date**: Upload your tech stack information to the Stack To Date platform for monitoring and lifecycle tracking

- **Interactive setup**: Prompts for user confirmation when multiple version candidates are found

## Quick Start

1. **Create a project profile** on [Stack To Date](https://stacktodate.club/)

2. **Install stacktodate**:
   Download from [Releases](https://github.com/stacktodate/stacktodate-cli/releases) or build from source:
   ```bash
   go build -o stacktodate
   ```

3. **Initialize your tech stack**:
   ```bash
   stacktodate init --name "My Project"
   ```
   The tool will automatically detect technologies in your project.

4. **Push to Stack To Date**:
   ```bash
   export STD_TOKEN=your_token_from_stack_to_date
   stacktodate push
   ```

5. **View your tech stack** on the Stack To Date platform and monitor EOL dates!

## Installation

### Homebrew (macOS/Linux)

The easiest way to install on macOS or Linux:

```bash
brew tap stacktodate/homebrew-stacktodate
brew install stacktodate
```

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/stacktodate/stacktodate-cli/releases).

#### macOS

```bash
# Intel Macs
curl -L https://github.com/stacktodate/stacktodate-cli/releases/latest/download/stacktodate_darwin_amd64.tar.gz | tar xz
sudo mv stacktodate /usr/local/bin/

# Apple Silicon Macs
curl -L https://github.com/stacktodate/stacktodate-cli/releases/latest/download/stacktodate_darwin_arm64.tar.gz | tar xz
sudo mv stacktodate /usr/local/bin/
```

#### Linux

```bash
# x86_64
curl -L https://github.com/stacktodate/stacktodate-cli/releases/latest/download/stacktodate_linux_amd64.tar.gz | tar xz
sudo mv stacktodate /usr/local/bin/

# ARM64
curl -L https://github.com/stacktodate/stacktodate-cli/releases/latest/download/stacktodate_linux_arm64.tar.gz | tar xz
sudo mv stacktodate /usr/local/bin/
```

#### Windows

Download `stacktodate_windows_amd64.zip` from the [Releases page](https://github.com/stacktodate/stacktodate-cli/releases) and extract to a directory in your PATH.

### Build from Source

Requires Go 1.21 or later:

```bash
git clone https://github.com/stacktodate/stacktodate-cli.git
cd stacktodate-cli
go build -o stacktodate
./stacktodate --help
```

## Usage

### Initialize a new project

Create a `stacktodate.yml` file with autodetection:

```bash
stacktodate init
```

Options:
- `--uuid, -u`: Set project UUID
- `--name, -n`: Set project name
- `--skip-autodetect`: Skip technology detection
- `--no-interactive`: Use first candidate without prompting

### Detect technologies

Scan the current directory and display detected technologies:

```bash
stacktodate autodetect [path]
```

This shows what technologies and versions were detected from:
- `Dockerfile` and `docker-compose.yml` files
- `go.mod` (Go version)
- `package.json` and `.nvmrc` (Node.js version)
- `.python-version`, `pyproject.toml`, `Pipfile` (Python version)
- `.ruby-version` (Ruby version)
- `Gemfile` (Rails version)

### Update existing configuration

Update your `stacktodate.yml` with newly detected technologies:

```bash
stacktodate update
```

Options:
- `--config, -c`: Path to stacktodate.yml file (default: `stacktodate.yml`)
- `--skip-autodetect`: Keep existing stack without detection
- `--no-interactive`: Use first candidate without prompting

### Push to Stack To Date

Upload your detected tech stack to the Stack To Date platform for monitoring and lifecycle tracking:

```bash
stacktodate push
```

This command:
- Reads your `stacktodate.yml` file with the detected tech stack
- Sends it to your Stack To Date profile
- Updates your remote tech stack for EOL monitoring

Requirements:
- Valid `stacktodate.yml` file in current directory (create with `stacktodate init`)
- `STD_TOKEN` environment variable set with your Stack To Date API token

Options:
- `--config, -c`: Path to stacktodate.yml file (default: `stacktodate.yml`)

Configuration:
- API URL can be customized via `STD_API_URL` environment variable (default: `https://stacktodate.club`)

Example:

```bash
export STD_TOKEN=your_stack_to_date_api_token
stacktodate push
```

To find your API token, log in to [Stack To Date](https://stacktodate.club/) and navigate to your account settings.

### View version

```bash
stacktodate version
```

## Configuration File

The `stacktodate.yml` file stores your project's tech stack information:

```yaml
uuid: abc123-def456
name: My Project
stack:
  go:
    version: "1.21"
    source: go.mod
  nodejs:
    version: "18.0.0"
    source: .nvmrc
  python:
    version: "3.11"
    source: .python-version
  rails:
    version: "7.0.0"
    source: Gemfile
```

- `uuid`: Unique identifier for your tech stack
- `name`: Project name
- `stack`: Map of technology names with version and detection source
  - `version`: The detected version of the technology
  - `source`: The file/config where the version was detected from

## Running Tests

Run all tests with verbose output:

```bash
./test.sh
```

Or use Go directly:

```bash
go test -v ./...
```

### Test coverage

The project includes tests for:
- Version detection and parsing
- Docker/Docker Compose detection
- Go version detection
- Node.js version detection
- Python version detection
- Ruby version detection
- Rails detection

### Running specific tests

```bash
# Run tests for a specific package
go test -v ./cmd/lib/detectors

# Run a specific test
go test -v ./cmd/lib/detectors -run TestDetectDocker
```

## Development

### Local Development

```bash
# Clone repository
git clone https://github.com/stacktodate/stacktodate-cli.git
cd stacktodate-cli

# Install dependencies
go mod download

# Run tests
go test -v ./...

# Build for local platform
go build -o stacktodate
```

### Creating a Release

Releases are automated via GitHub Actions and GoReleaser:

1. Ensure all changes are committed and pushed
2. Create and push a version tag:
   ```bash
   git tag -a v0.2.0 -m "Release v0.2.0"
   git push origin v0.2.0
   ```
3. GitHub Actions will automatically:
   - Run tests
   - Build binaries for all platforms
   - Generate release notes
   - Create a GitHub draft release with assets

4. Review the draft release on the [Releases page](https://github.com/stacktodate/stacktodate-cli/releases) and click "Publish release" when ready

### Version Numbering

This project follows [Semantic Versioning](https://semver.org/):
- MAJOR version for incompatible API changes
- MINOR version for new functionality (backwards compatible)
- PATCH version for backwards compatible bug fixes

Tag format: `v{MAJOR}.{MINOR}.{PATCH}` (e.g., v1.2.3)

### Project structure

```
.
├── cmd/                          # Command implementations
│   ├── root.go                  # Root command setup
│   ├── init.go                  # Init command
│   ├── autodetect.go            # Autodetect command
│   ├── update.go                # Update command
│   ├── push.go                  # Push command
│   ├── detect.go                # Detection logic
│   └── lib/
│       └── detectors/           # Language/framework detectors
│           ├── docker.go
│           ├── go.go
│           ├── nodejs.go
│           ├── python.go
│           ├── rails.go
│           └── ruby.go
├── main.go                       # Entry point
├── build.sh                      # Build script for all platforms
├── test.sh                       # Test runner script
└── README.md                     # This file
```

### Adding a new detector

1. Create a new file in `cmd/lib/detectors/` (e.g., `java.go`)
2. Implement the detector function following the existing pattern
3. Add tests in `cmd/lib/detectors/java_test.go`
4. Integrate into `cmd/detect.go` in the `DetectProjectInfo()` function

## Contributing

We welcome contributions! Here are some ways you can help:

### Improving Existing Detectors

If you notice a detector isn't working correctly for your project, please:

1. Open a [GitHub Issue](https://github.com/stacktodate/stacktodate-cli/issues) with details about the problem
2. Include example files or project structures where the detection fails
3. Provide the expected behavior

### Requesting New Detectors

Want to add support for a language or framework we don't currently detect?

1. Open a [GitHub Issue](https://github.com/stacktodate/stacktodate-cli/issues) with the title "Feature: Add detector for [Technology Name]"
2. Attach example configuration files that contain version information (e.g., `package.json`, `requirements.txt`, `composer.json`, `pom.xml`, etc.)
3. Explain how version information is typically stored in projects using that technology

Your contributions and feedback help us improve the detector accuracy and coverage!

## Environment Variables

- `STD_TOKEN`: Stack To Date API authentication token (required for `push` command). Get your token from your Stack To Date account settings at https://stacktodate.club
- `STD_API_URL`: API base URL (optional, defaults to `https://stacktodate.club`)

## Credits

This project was built with the assistance of large language models. We're grateful to the AI community for enabling modern development practices.

## License

MIT License — See [LICENSE](LICENSE) file for details.
