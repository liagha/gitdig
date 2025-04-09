package downloader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/liagha/grawl/internal/display"
	"github.com/liagha/grawl/internal/github"
)

type Stats struct {
	Files    int
	Dirs     int
	Failures int
	Bytes    int64
	sync.Mutex
}

type Downloader struct {
	Token       string
	Recursive   bool
	Concurrency int
	Verbose     bool
	Stats       Stats
	wg          sync.WaitGroup
	sem         chan struct{}
}

func New(token string, recursive bool, concurrency int, verbose bool) *Downloader {
	return &Downloader{
		Token:       token,
		Recursive:   recursive,
		Concurrency: concurrency,
		Verbose:     verbose,
		sem:         make(chan struct{}, concurrency),
	}
}

func (d *Downloader) DownloadRepository(owner, repo, branch, dirPath, localDir string) error {
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	display.Bold("Downloading from %s/%s (branch: %s, path: %s)\n", owner, repo, branch, dirPath)
	display.Info("Saving to: %s\n", localDir)

	startTime := time.Now()
	err := d.downloadDirectory(owner, repo, branch, dirPath, localDir)
	if err != nil {
		return err
	}

	d.wg.Wait()

	elapsed := time.Since(startTime).Seconds()
	display.BoldCyan("\nDownload Summary\n")
	display.Info("Time: %.1f seconds\n", elapsed)
	display.Info("Files: %d\n", d.Stats.Files)
	display.Info("Directories: %d\n", d.Stats.Dirs)
	display.Info("Size: %.2f MB\n", float64(d.Stats.Bytes)/(1024*1024))

	if d.Stats.Failures > 0 {
		display.Warning("Failures: %d\n", d.Stats.Failures)
		return fmt.Errorf("%d files failed to download", d.Stats.Failures)
	} else {
		display.Success("All files downloaded successfully!\n")
	}

	return nil
}

func (d *Downloader) downloadDirectory(owner, repo, branch, dirPath, localDir string) error {
	d.Stats.Lock()
	d.Stats.Dirs++
	d.Stats.Unlock()

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, dirPath, branch)
	contents, err := github.GetContents(apiURL, d.Token)
	if err != nil {
		return fmt.Errorf("failed to get directory contents: %w", err)
	}

	for _, content := range contents {
		if content.Type == "file" {
			d.sem <- struct{}{} // Acquire semaphore
			d.wg.Add(1)

			go func(content github.GitHubContent) {
				defer d.wg.Done()
				defer func() { <-d.sem }() // Release semaphore

				filePath := filepath.Join(localDir, content.Name)
				size, err := d.downloadFile(content.DownloadURL, filePath)

				d.Stats.Lock()
				defer d.Stats.Unlock()

				if err != nil {
					display.Error("Failed: %s (%v)\n", content.Path, err)
					d.Stats.Failures++
				} else {
					if d.Verbose {
						display.Success("Downloaded: %s (%.2f KB)\n", content.Path, float64(size)/1024)
					}
					d.Stats.Files++
					d.Stats.Bytes += size
				}
			}(content)
		} else if content.Type == "dir" && d.Recursive {
			subDir := filepath.Join(localDir, content.Name)
			if err := os.MkdirAll(subDir, 0755); err != nil {
				display.Error("Error creating subdirectory %s: %v\n", subDir, err)
				d.Stats.Lock()
				d.Stats.Failures++
				d.Stats.Unlock()
				continue
			}

			subPath := filepath.Join(dirPath, content.Name)
			if err := d.downloadDirectory(owner, repo, branch, subPath, subDir); err != nil {
				display.Warning("Warning: Error in subdirectory %s: %v\n", content.Path, err)
			}
		}
	}

	return nil
}

func (d *Downloader) downloadFile(url, filePath string) (int64, error) {
	data, err := github.DownloadFileContent(url, d.Token)
	if err != nil {
		return 0, err
	}

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Write the data to file
	n, err := out.Write(data)
	if err != nil {
		return 0, fmt.Errorf("failed to write file data: %w", err)
	}

	return int64(n), nil
}
