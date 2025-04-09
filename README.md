# GitHub Directory Downloader

A lightweight command-line tool to download directories from GitHub repositories with support for authentication, recursive downloads, and concurrent operations. Built using standard Go libraries.

## Features

- 📂 Download single directories or entire repos from GitHub
- 🔄 Recursive downloading of subdirectories
- 🔑 GitHub authentication support to avoid rate limits
- ⚡ Concurrent downloading for faster speeds
- 🎨 Colored terminal output (auto-detects Windows compatibility)
- 📊 Download statistics and progress reporting
- ✅ Command-line flags for easy configuration
- 🔍 Support for both URL and shorthand formats (username/repo/path)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/github-dir-dl.git
cd github-dir-dl

# Build the binary
go build -o github-dir-dl

# Optional: Install to your path
go install
```

### Using Go Install

```bash
go install github.com/yourusername/github-dir-dl@latest
```

## Usage

### Basic Usage

```bash
# Download a directory using interactive prompt
github-dir-dl

# Download a directory by providing URL directly
github-dir-dl -u https://github.com/golang/go/tree/master/src/encoding/json
```

### Command Line Arguments

```
Usage: github-dir-dl [options] [username/repo/path]

Options:
  -url, -u string        GitHub repository path (URL or username/repo/path format)
  -token, -t string      GitHub personal access token
  -output, -o string     Output directory (default: last part of path)
  -recursive, -r         Download directories recursively
  -concurrency, -n int   Number of concurrent downloads (default 5)
  -verbose, -v           Verbose output
  -help, -h              Display help information
```

You can also set the GitHub token via the GITHUB_TOKEN environment variable.


### Authentication

To avoid GitHub API rate limits, you can provide a personal access token:

```bash
# Using command-line flag
github-dir-dl -u https://github.com/golang/go/tree/master/src/encoding/json -t YOUR_GITHUB_TOKEN

# Using config file
github-dir-dl -u https://github.com/golang/go/tree/master/src/encoding/json
```

To create a GitHub personal access token, visit: https://github.com/settings/tokens

### Authentication

To avoid GitHub API rate limits, you can provide a personal access token:

```bash
# Using command-line flag
github-dir-dl -t YOUR_GITHUB_TOKEN username/repo/path

# Using environment variable
export GITHUB_TOKEN=YOUR_GITHUB_TOKEN
github-dir-dl username/repo/path
```

To create a GitHub personal access token, visit: https://github.com/settings/tokens

## Example Use Cases

### Download a specific directory

```bash
github-dir-dl -u https://github.com/golang/go/tree/master/src/encoding/json
```

### Recursive download with authentication

```bash
github-dir-dl -u https://github.com/golang/go/tree/master/src/encoding -r -t YOUR_GITHUB_TOKEN
```

### Download to a custom location with verbose output

```bash
github-dir-dl -u https://github.com/golang/go/tree/master/src/encoding/json -o ./my-json-dir -v
```

## License

MIT