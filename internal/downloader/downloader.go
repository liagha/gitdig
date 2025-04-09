package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/liagha/gitdig/internal/display"
	"github.com/liagha/gitdig/internal/github"
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
	ZipOutput   bool
	Preview     bool
	Update      bool
	Retries     int
	Stats       Stats
	wg          sync.WaitGroup
	sem         chan struct{}
	zipWriter   *ZipWriter
}

func New(token string, recursive bool, concurrency int, verbose bool, zipOutput bool, preview bool, update bool, retries int) *Downloader {
	return &Downloader{
		Token:       token,
		Recursive:   recursive,
		Concurrency: concurrency,
		Verbose:     verbose,
		ZipOutput:   zipOutput,
		Preview:     preview,
		Update:      update,
		Retries:     retries,
		sem:         make(chan struct{}, concurrency),
	}
}

func (d *Downloader) DownloadRepository(owner, repo, branch, dirPath, localDir string) error {
	if d.Preview {
		display.Bold("PREVIEW MODE: Showing what would be downloaded from %s/%s (branch: %s, path: %s)\n", owner, repo, branch, dirPath)
		display.Info("Would save to: %s\n", localDir)
		err := d.previewDirectory(owner, repo, branch, dirPath, "")
		if err != nil {
			return err
		}

		display.BoldCyan("\nPreview Summary\n")
		display.Info("Files: %d\n", d.Stats.Files)
		display.Info("Directories: %d\n", d.Stats.Dirs)

		return nil
	}

	if !d.ZipOutput {
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	} else {
		// Create zip file
		zipPath := localDir
		if !strings.HasSuffix(zipPath, ".zip") {
			zipPath += ".zip"
		}

		// Create parent directory for zip file if needed
		parentDir := filepath.Dir(zipPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for zip file: %w", err)
		}

		var err error
		d.zipWriter, err = NewZipWriter(zipPath, "")
		if err != nil {
			return fmt.Errorf("failed to create zip archive: %w", err)
		}
		defer d.zipWriter.Close()

		display.Bold("Downloading from %s/%s (branch: %s, path: %s)\n", owner, repo, branch, dirPath)
		display.Info("Saving to zip archive: %s\n", zipPath)
	}

	if !d.ZipOutput {
		display.Bold("Downloading from %s/%s (branch: %s, path: %s)\n", owner, repo, branch, dirPath)
		display.Info("Saving to: %s\n", localDir)
	}

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

func (d *Downloader) previewDirectory(owner, repo, branch, dirPath, prefix string) error {
	d.Stats.Lock()
	d.Stats.Dirs++
	d.Stats.Unlock()

	display.Info("%s└── %s/\n", prefix, filepath.Base(dirPath))
	newPrefix := prefix + "    "

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, dirPath, branch)
	contents, err := github.GetContents(apiURL, d.Token)
	if err != nil {
		return fmt.Errorf("failed to get directory contents: %w", err)
	}

	for i, content := range contents {
		if content.Type == "file" {
			d.Stats.Lock()
			d.Stats.Files++
			d.Stats.Unlock()

			isLast := i == len(contents)-1
			if isLast {
				display.Info("%s└── %s\n", newPrefix, content.Name)
			} else {
				display.Info("%s├── %s\n", newPrefix, content.Name)
			}
		} else if content.Type == "dir" && d.Recursive {
			if i == len(contents)-1 {
				err := d.previewDirectory(owner, repo, branch, filepath.Join(dirPath, content.Name), newPrefix)
				if err != nil {
					return err
				}
			} else {
				err := d.previewDirectory(owner, repo, branch, filepath.Join(dirPath, content.Name), newPrefix+"│   ")
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (d *Downloader) downloadDirectory(owner, repo, branch, dirPath, localDir string) error {
	d.Stats.Lock()
	d.Stats.Dirs++
	d.Stats.Unlock()

	if d.ZipOutput && dirPath != "" {
		// Add directory entry to zip
		err := d.zipWriter.CreateDirEntry(dirPath)
		if err != nil && d.Verbose {
			display.Warning("Warning: Could not create zip directory entry: %v\n", err)
		}
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, dirPath, branch)
	contents, err := github.GetContents(apiURL, d.Token)
	if err != nil {
		return fmt.Errorf("failed to get directory contents: %w", err)
	}

	for _, content := range contents {
		if content.Type == "file" {
			d.sem <- struct{}{}
			d.wg.Add(1)

			go func(content github.Content) {
				defer d.wg.Done()
				defer func() { <-d.sem }()

				filePath := filepath.Join(localDir, content.Name)

				// Check if updating and file already exists
				if d.Update && !d.ZipOutput {
					if stat, err := os.Stat(filePath); err == nil {
						// File exists, check if we need to update it
						if !d.shouldUpdate(content, stat) {
							if d.Verbose {
								display.Info("Skipped (up-to-date): %s\n", content.Path)
							}
							return
						}
					}
				}

				var size int64
				var err error
				attempts := 0
				maxAttempts := d.Retries + 1

				for attempts < maxAttempts {
					attempts++
					if attempts > 1 && d.Verbose {
						display.Warning("Retry %d/%d: %s\n", attempts-1, d.Retries, content.Path)
					}

					if d.ZipOutput {
						size, err = d.downloadFileToZip(content.DownloadURL, content.Path)
					} else {
						size, err = d.downloadFile(content.DownloadURL, filePath)
					}

					if err == nil {
						break
					}

					if attempts < maxAttempts {
						// Exponential backoff: wait 2^attempt * 100ms
						backoff := (1 << (attempts - 1)) * 100
						time.Sleep(time.Duration(backoff) * time.Millisecond)
					}
				}

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
			if !d.ZipOutput {
				if err := os.MkdirAll(subDir, 0755); err != nil {
					display.Error("Error creating subdirectory %s: %v\n", subDir, err)
					d.Stats.Lock()
					d.Stats.Failures++
					d.Stats.Unlock()
					continue
				}
			}

			subPath := filepath.Join(dirPath, content.Name)
			if err := d.downloadDirectory(owner, repo, branch, subPath, subDir); err != nil {
				display.Warning("Warning: Error in subdirectory %s: %v\n", content.Path, err)
			}
		}
	}

	return nil
}

// shouldUpdate determines if a file needs to be updated based on the update mode
func (d *Downloader) shouldUpdate(content github.Content, stat os.FileInfo) bool {
	// For now, just update based on file size
	// In a real implementation, you might use ETag, modified timestamp, or file hash
	return stat.Size() == 0 || content.Size != stat.Size()
}

func (d *Downloader) downloadFileToZip(url, zipPath string) (int64, error) {
	data, err := github.DownloadFileContent(url, d.Token)
	if err != nil {
		return 0, err
	}

	err = d.zipWriter.AddFile(data, zipPath)
	if err != nil {
		return 0, err
	}

	return int64(len(data)), nil
}

func (d *Downloader) downloadFile(url, filePath string) (int64, error) {
	data, err := github.DownloadFileContent(url, d.Token)
	if err != nil {
		return 0, err
	}

	out, err := os.Create(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}

	// Use a named return to capture close error properly
	var writeErr error
	defer func() {
		if cerr := out.Close(); cerr != nil {
			if writeErr == nil {
				writeErr = fmt.Errorf("failed to close file: %w", cerr)
			}
		}
	}()

	n, err := out.Write(data)
	if err != nil {
		writeErr = fmt.Errorf("failed to write file data: %w", err)
		return 0, writeErr
	}

	return int64(n), writeErr
}
