package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/liagha/gitdig/internal/config"
	"github.com/liagha/gitdig/internal/display"
	"github.com/liagha/gitdig/internal/downloader"
	"github.com/liagha/gitdig/internal/github"
)

func main() {
	var flags config.AppFlags

	flag.StringVar(&flags.URL, "url", "", "GitHub repository path (e.g., username/repo/path or full URL)")
	flag.StringVar(&flags.Token, "token", "", "GitHub personal access token")
	flag.StringVar(&flags.Output, "output", "", "Output directory (default: last part of GitHub path)")
	flag.BoolVar(&flags.Recursive, "recursive", false, "Download directories recursively")
	flag.IntVar(&flags.Concurrency, "concurrency", 5, "Number of concurrent downloads")
	flag.BoolVar(&flags.Verbose, "verbose", false, "Verbose output")

	flag.StringVar(&flags.URL, "u", "", "GitHub repository path (shorthand)")
	flag.StringVar(&flags.Token, "t", "", "GitHub personal access token (shorthand)")
	flag.StringVar(&flags.Output, "o", "", "Output directory (shorthand)")
	flag.BoolVar(&flags.Recursive, "r", false, "Download directories recursively (shorthand)")
	flag.IntVar(&flags.Concurrency, "n", 5, "Number of concurrent downloads (shorthand)")
	flag.BoolVar(&flags.Verbose, "v", false, "Verbose output (shorthand)")

	flag.Parse()

	if flags.Token == "" {
		flags.Token = os.Getenv("GITHUB_TOKEN")
	}

	if flags.URL == "" && len(flag.Args()) > 0 {
		flags.URL = flag.Args()[0]
	}

	if flags.URL == "" {
		var err error
		flags.URL, err = display.Prompt("Enter GitHub repository path or URL: ")
		if err != nil {
			display.Error("Failed to read input: %v\n", err)
			os.Exit(1)
		}
	}

	owner, repo, branch, dirPath, err := github.ParsePath(flags.URL)
	if err != nil {
		display.Error("Error parsing path: %v\n", err)
		os.Exit(1)
	}

	localDir := flags.Output
	if localDir == "" {
		localDir = filepath.Base(dirPath)
		if localDir == "." || localDir == "/" {
			localDir = repo
		}
	}

	if err := os.MkdirAll(localDir, 0755); err != nil {
		display.Error("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	dl := downloader.New(flags.Token, flags.Recursive, flags.Concurrency, flags.Verbose)

	if err := dl.DownloadRepository(owner, repo, branch, dirPath, localDir); err != nil {
		display.Error("Error: %v\n", err)
		os.Exit(1)
	}
}
