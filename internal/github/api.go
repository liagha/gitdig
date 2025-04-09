package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/liagha/gitdig/internal/config"
	"github.com/liagha/gitdig/internal/display"
)

type Content struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

var client = &http.Client{
	Timeout: 30 * time.Second,
}

func GetContents(apiURL, token string) (contents []Content, err error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	req.Header.Set("User-Agent", config.AppName+"/"+config.AppVersion)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close response body: %w", cerr)
		}
	}()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		rateLimitReset := resp.Header.Get("X-RateLimit-Reset")
		if rateLimitReset != "" {
			resetInt, parseErr := time.ParseDuration(rateLimitReset + "s")
			if parseErr == nil {
				waitTime := resetInt - time.Duration(time.Now().Unix())
				display.Yellow("Rate limit exceeded. Reset in %.0f minutes. Waiting...\n", waitTime.Minutes())
				time.Sleep(waitTime)
				return GetContents(apiURL, token)
			}
		}
		return nil, errors.New("GitHub API rate limit exceeded. Try using authentication with --token")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return contents, nil
}

func ParsePath(path string) (owner, repo, branch, dirPath string, err error) {
	branch = "master"

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return parseURL(path)
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 2 {
		return "", "", "", "", errors.New("invalid GitHub path format, must be at least owner/repo")
	}

	owner = parts[0]
	repo = parts[1]

	if len(parts) >= 4 && parts[2] == "tree" {
		branch = parts[3]
		if len(parts) > 4 {
			dirPath = strings.Join(parts[4:], "/")
		}
	} else if len(parts) > 2 {
		dirPath = strings.Join(parts[2:], "/")
	}

	return owner, repo, branch, dirPath, nil
}

func parseURL(rawURL string) (owner, repo, branch, path string, err error) {
	branch = "master"

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", "", fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Host != "github.com" {
		return "", "", "", "", errors.New("not a GitHub URL")
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

	if len(parts) < 2 {
		return "", "", "", "", errors.New("invalid GitHub URL format")
	}

	owner = parts[0]
	repo = parts[1]

	if len(parts) >= 4 {
		if parts[2] == "tree" || parts[2] == "blob" {
			branch = parts[3]
			if len(parts) > 4 {
				path = strings.Join(parts[4:], "/")
			}
		}
	}

	return owner, repo, branch, path, nil
}

func DownloadFileContent(url, token string) (data []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	req.Header.Set("User-Agent", config.AppName+"/"+config.AppVersion)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute download request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close response body: %w", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}
