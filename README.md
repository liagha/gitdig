# gitdig 🕳️

<div align="center">

**A lightweight and blazing-fast CLI tool for downloading GitHub directories**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/liagha/gitdig)](https://goreportcard.com/report/github.com/liagha/gitdig)
[![Go Version](https://img.shields.io/github/go-mod/go-version/liagha/gitdig)](https://github.com/liagha/gitdig)
[![Latest Release](https://img.shields.io/github/v/release/liagha/gitdig)](https://github.com/liagha/gitdig/releases/latest)

</div>

## 📚 Overview

`gitdig` is a command-line utility that lets you download specific directories or entire repositories from GitHub without cloning the whole repository. Perfect for when you need just a portion of a large codebase.

## ✨ Features

- 📂 Download **specific directories** or **entire repositories**
- 🔁 Support for **recursive** subdirectory downloads
- 🔐 **GitHub authentication** to bypass API rate limits
- ⚡ **Concurrent file operations** for maximum performance
- 🎨 **Colorized terminal output** with automatic Windows compatibility detection
- 📊 **Progress indicators** and download statistics
- 🔍 Support for both **full GitHub URLs** and shorthand notation (`username/repo/path`)
- 🧩 Clean and **composable command-line interface**

## 🚀 Installation

### Binary Releases

Download pre-built binaries from the [releases page](https://github.com/liagha/gitdig/releases).

### Using Go Install

```bash
go install github.com/liagha/gitdig@latest
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/liagha/gitdig.git

# Navigate to the project directory
cd gitdig

# Build the project
go build -o gitdig

# Optionally install globally
go install
```

## 📋 Usage

### Interactive Mode

Run `gitdig` without arguments to launch interactive mode:

```bash
gitdig
```

## 🔧 Command-Line Options

```
Usage: gitdig [options] [username/repo/path]

Options:
  -url, -u string        GitHub repository path (URL or user/repo/path format)
  -token, -t string      GitHub personal access token for authentication
  -output, -o string     Output directory (default: last part of path)
  -recursive, -r         Download directories recursively (default: false)
  -concurrency, -n int   Number of concurrent downloads (default: 5)
  -verbose, -v           Enable verbose output (default: false)
  -help, -h              Display help information
```

## 📖 Examples

### Download a Specific Directory

```bash
gitdig -u https://github.com/golang/go/tree/master/src/encoding/json
```

Or using the shorthand format:

```bash
gitdig golang/go/src/encoding/json
```

### Recursive Download with Custom Output Directory

```bash
gitdig -u https://github.com/golang/go/tree/master/src/encoding -r -o ./my-encoding-folder
```

### Using GitHub Authentication

To avoid GitHub API rate limits, use a personal access token:

#### Option 1: Command-line Flag

```bash
gitdig -u https://github.com/golang/go/tree/master/src/encoding/json -t YOUR_GITHUB_TOKEN
```

#### Option 2: Environment Variable

```bash
export GITHUB_TOKEN=YOUR_GITHUB_TOKEN
gitdig golang/go/src/encoding/json
```

### Adjust Concurrency for Faster Downloads

```bash
gitdig -u https://github.com/golang/go/tree/master/src/encoding -n 10
```

### Enable Verbose Output

```bash
gitdig -u https://github.com/golang/go/tree/master/src/encoding -v
```

## 🔒 GitHub Authentication

For frequent use or downloading from private repositories, it's recommended to use GitHub authentication:

1. Create a [Personal Access Token](https://github.com/settings/tokens) on GitHub
2. Use it with the `-t` flag or set it as an environment variable

## 🧠 Advanced Usage

### Combined Options Example

```bash
gitdig -u https://github.com/golang/go/tree/master/src/encoding \
       -r \
       -o ./golang-encoding \
       -n 8 \
       -t YOUR_GITHUB_TOKEN \
       -v
```

This will:
- Download the `encoding` directory from the Go repository
- Download all subdirectories recursively
- Save to `./golang-encoding`
- Use 8 concurrent downloads
- Use your GitHub token for authentication
- Show verbose output during the process

## 💡 Tips

- For large directories, increase concurrency (`-n`) for faster downloads
- Set your GitHub token as an environment variable to avoid exposing it in your command history
- Use the recursive flag (`-r`) with caution on large repositories

## 🛠️ Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👏 Acknowledgements

- [Go programming language](https://golang.org/)

---

Built with ❤️ by [liagha](https://github.com/liagha)