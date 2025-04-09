package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/liagha/gitdig/internal/config"
	"github.com/liagha/gitdig/internal/display"
	"github.com/liagha/gitdig/internal/downloader"
	"github.com/liagha/gitdig/internal/github"
)

func browseRepositories(user string, token string) (string, error) {
	display.Bold("Fetching repositories for %s...\n", user)

	var repos []github.Repository
	var err error

	// Try as organization first
	repos, err = github.GetRepositoriesForOrg(user, token)
	if err != nil {
		display.Info("Trying as user...\n")
		// Try as user
		repos, err = github.GetRepositoriesForUser(user, token)
		if err != nil {
			return "", fmt.Errorf("failed to get repositories: %w", err)
		}
	}

	if len(repos) == 0 {
		return "", fmt.Errorf("no repositories found for %s", user)
	}

	display.BoldCyan("\nRepositories for %s:\n", user)
	for i, repo := range repos {
		desc := repo.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		if desc == "" {
			desc = "(No description)"
		}
		display.Info("[%d] %s - %s\n", i+1, repo.Name, desc)
	}

	selection := promptForSelection(len(repos))
	if selection < 1 || selection > len(repos) {
		return "", fmt.Errorf("invalid selection")
	}

	selectedRepo := repos[selection-1]
	display.Success("Selected: %s\n", selectedRepo.FullName)

	return selectedRepo.FullName, nil
}

func promptForSelection(max int) int {
	reader := bufio.NewReader(os.Stdin)
	for {
		display.Bold("\nEnter repository number (1-%d): ", max)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		num, err := strconv.Atoi(text)
		if err == nil && num >= 1 && num <= max {
			return num
		}
		display.Error("Invalid selection. Please enter a number between 1 and %d.\n", max)
	}
}

func main() {
	// Define flags
	var flags config.AppFlags
	flag.StringVar(&flags.URL, "u", "", "GitHub repository URL or path (can be specified multiple times)")
	flag.StringVar(&flags.Token, "token", "", "GitHub API token for authentication")
	flag.StringVar(&flags.Output, "o", "", "Output directory")
	flag.BoolVar(&flags.Recursive, "r", true, "Download directories recursively")
	flag.IntVar(&flags.Concurrency, "c", 5, "Number of concurrent downloads")
	flag.BoolVar(&flags.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&flags.ZipOutput, "zip", false, "Create ZIP archive instead of extracting files")
	flag.BoolVar(&flags.Preview, "preview", false, "Preview what would be downloaded without downloading")
	flag.BoolVar(&flags.Update, "update", false, "Only download new or changed files")
	flag.StringVar(&flags.ListFile, "list", "", "File containing list of repositories to download")
	flag.IntVar(&flags.Retries, "retries", 3, "Number of retries for failed downloads")
	flag.StringVar(&flags.User, "user", "", "GitHub username or organization for interactive repository selection")
	flag.BoolVar(&flags.Interactive, "i", false, "Interactive mode for selecting repositories")

	flag.Parse()

	// Display banner
	display.BoldCyan("\n%s v%s - GitHub Repository Downloader\n\n", config.AppName, config.AppVersion)

	// Check for token environment variable if not provided via flag
	if flags.Token == "" {
		flags.Token = os.Getenv("GITHUB_TOKEN")
	}

	// Collect all target URLs/paths
	var targets []string

	// Parse command line arguments that are not flags
	extraArgs := flag.Args()
	if len(extraArgs) > 0 {
		targets = append(targets, extraArgs...)
	}

	// If -u flag is provided, add it to targets
	if flags.URL != "" {
		targets = append(targets, flags.URL)
	}

	// If -list flag is provided, read targets from file
	if flags.ListFile != "" {
		fileTargets, err := github.ReadTargetsFromFile(flags.ListFile)
		if err != nil {
			display.Error("Error: %v\n", err)
			os.Exit(1)
		}
		targets = append(targets, fileTargets...)
	}

	// If -user flag is provided, use interactive repository selector
	if flags.User != "" || flags.Interactive {
		var user string
		if flags.User != "" {
			user = flags.User
		} else {
			// Prompt for user or organization name
			display.Bold("Enter GitHub username or organization: ")
			fmt.Scanln(&user)
		}

		repoPath, err := browseRepositories(user, flags.Token)
		if err != nil {
			display.Error("Error: %v\n", err)
			os.Exit(1)
		}
		targets = append(targets, repoPath)
	}

	// Check if we have any targets
	if len(targets) == 0 {
		display.Error("Error: No target specified. Use -u, -list, -user flags or provide a path argument.\n")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create downloader
	dl := downloader.New(
		flags.Token,
		flags.Recursive,
		flags.Concurrency,
		flags.Verbose,
		flags.ZipOutput,
		flags.Preview,
		flags.Update,
		flags.Retries,
	)

	// Process targets
	downloadTargets, err := github.ParseTargets(targets, flags.Output)
	if err != nil {
		display.Error("Error: %v\n", err)
		os.Exit(1)
	}

	// Process each target
	for i, target := range downloadTargets {
		if i > 0 && !flags.Preview {
			display.BoldCyan("\nProcessing next target (%d/%d)...\n", i+1, len(downloadTargets))
		}

		// For ZIP output with multiple targets, add target identifier to filename
		localDir := target.LocalDir
		if flags.ZipOutput && !strings.HasSuffix(localDir, ".zip") {
			localDir += ".zip"
		}

		err := dl.DownloadRepository(
			target.Owner,
			target.Repo,
			target.Branch,
			target.DirPath,
			localDir,
		)

		if err != nil {
			display.Error("Error: %v\n", err)
			// Continue to next target instead of exiting
			if i < len(downloadTargets)-1 {
				display.Warning("Continuing to next target...\n")
			}
		}
	}

	display.BoldCyan("\nAll operations completed.\n")
}
