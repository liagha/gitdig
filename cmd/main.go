package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/user/github-dir-dl/internal/config"
	"github.com/user/github-dir-dl/internal/display"
	"github.com/user/github-dir-dl/internal/downloader"
	"github.com/user/github-dir-dl/internal/github"
)

func main() {
	var flags config.AppFlags

	// Set up command line flags
	flag.StringVar(&flags.URL, "url", "", "GitHub repository path (e.g., username/repo/path or full URL)")
	flag.StringVar(&flags.Token, "token", "", "GitHub personal access token")
	flag.StringVar(&flags.Output, "output", "", "Output directory (default: last part of GitHub path)")
	flag.BoolVar(&flags.Recursive, "recursive", false, "Download directories recursively")
	flag.IntVar(&flags.Concurrency, "concurrency", 5, "Number of concurrent downloads")
	flag.BoolVar(&flags.Verbose, "verbose", false, "Verbose output")

	// Define shorthand flags
	flag.StringVar(&flags.URL, "u", "", "GitHub repository path (shorthand)")
	flag.StringVar(&flags.Token, "t", "", "GitHub personal access token (shorthand)")
	flag.StringVar(&flags.Output, "o", "", "Output directory (shorthand)")
	flag.BoolVar(&flags.Recursive, "r", false, "Download directories recursively (shorthand)")
	flag.IntVar(&flags.Concurrency, "n", 5, "Number of concurrent downloads (shorthand)")
	flag.BoolVar(&flags.Verbose, "v", false, "Verbose output (shorthand)")

	// Parse the flags
	flag.Parse()

	// Check for environment variable token if not provided
	if flags.Token == "" {
		flags.Token = os.Getenv("GITHUB_TOKEN")
	}

	// If URL wasn't provided as a flag, check for positional arguments
	if flags.URL == "" && len(flag.Args()) > 0 {
		flags.URL = flag.Args()[0]
	}

	// Still no URL? Prompt for it
	if flags.URL == "" {
		var err error
		flags.URL, err = display.Prompt("Enter GitHub repository path or URL: ")
		if err != nil {
			display.Error("Failed to read input: %v\n", err)
			os.Exit(1)
		}
	}

	// Parse GitHub URL or path
	owner, repo, branch, dirPath, err := github.ParsePath(flags.URL)
	if err != nil {
		display.Error("Error parsing path: %v\n", err)
		os.Exit(1)
	}

	// Determine local directory for downloads
	localDir := flags.Output
	if localDir == "" {
		// Use the last part of the path as directory name
		localDir = filepath.Base(dirPath)
		if localDir == "." || localDir == "/" {
			localDir = repo
		}
	}

	// Create output directory
	if err := os.MkdirAll(localDir, 0755); err != nil {
		display.Error("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Create and configure downloader
	dl := downloader.New(flags.Token, flags.Recursive, flags.Concurrency, flags.Verbose)

	// Start download
	if err := dl.DownloadRepository(owner, repo, branch, dirPath, localDir); err != nil {
		display.Error("Error: %v\n", err)
		os.Exit(1)
	}
}
