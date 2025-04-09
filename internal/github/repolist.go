package github

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Repository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	CloneURL    string `json:"clone_url"`
	HTMLURL     string `json:"html_url"`
}

type DownloadTarget struct {
	Owner    string
	Repo     string
	Branch   string
	DirPath  string
	LocalDir string
}

// GetRepositoriesForUser retrieves a list of repositories for a user or organization
func GetRepositoriesForUser(user string, token string) ([]Repository, error) {
	apiURL := fmt.Sprintf("https://api.github.com/users/%s/repos", user)
	return getRepositories(apiURL, token)
}

// GetRepositoriesForOrg retrieves a list of repositories for an organization
func GetRepositoriesForOrg(org string, token string) ([]Repository, error) {
	apiURL := fmt.Sprintf("https://api.github.com/orgs/%s/repos", org)
	return getRepositories(apiURL, token)
}

func getRepositories(apiURL string, token string) ([]Repository, error) {
	req, err := createRequest("GET", apiURL, token)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return repos, nil
}

// ReadTargetsFromFile reads a list of GitHub repository paths from a file
func ReadTargetsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open list file: %w", err)
	}
	defer file.Close()

	var targets []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			targets = append(targets, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading list file: %w", err)
	}

	return targets, nil
}

// ParseTargets parses multiple GitHub paths and converts them to download targets
func ParseTargets(paths []string, baseDir string) ([]DownloadTarget, error) {
	var targets []DownloadTarget

	for _, path := range paths {
		owner, repo, branch, dirPath, err := ParsePath(path)
		if err != nil {
			return nil, fmt.Errorf("invalid path '%s': %w", path, err)
		}

		// Create local directory path
		localDir := baseDir
		if localDir == "" {
			localDir = repo
			if dirPath != "" {
				localDir = fmt.Sprintf("%s-%s", repo, strings.ReplaceAll(dirPath, "/", "-"))
			}
		} else if len(paths) > 1 {
			// When downloading multiple targets to the same base directory,
			// create subdirectories for each target
			subDir := repo
			if dirPath != "" {
				subDir = fmt.Sprintf("%s-%s", repo, strings.ReplaceAll(dirPath, "/", "-"))
			}
			localDir = fmt.Sprintf("%s/%s", baseDir, subDir)
		}

		targets = append(targets, DownloadTarget{
			Owner:    owner,
			Repo:     repo,
			Branch:   branch,
			DirPath:  dirPath,
			LocalDir: localDir,
		})
	}

	return targets, nil
}
