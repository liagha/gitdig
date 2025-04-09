package downloader

import (
	"archive/zip"
	"fmt"
	"os"
	"strings"
)

// ZipWriter handles creating zip archives
type ZipWriter struct {
	zipFile *os.File
	writer  *zip.Writer
	baseDir string
}

// NewZipWriter creates a new zip archive writer
func NewZipWriter(outputPath string, baseDir string) (*ZipWriter, error) {
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create zip file: %w", err)
	}

	return &ZipWriter{
		zipFile: zipFile,
		writer:  zip.NewWriter(zipFile),
		baseDir: baseDir,
	}, nil
}

// AddFile adds a file to the zip archive
func (z *ZipWriter) AddFile(data []byte, filePath string) error {
	// Create a relative path for the file in the zip
	relPath := strings.TrimPrefix(filePath, z.baseDir)
	relPath = strings.TrimPrefix(relPath, "/")

	header := &zip.FileHeader{
		Name:   relPath,
		Method: zip.Deflate,
	}

	writer, err := z.writer.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write zip entry: %w", err)
	}

	return nil
}

// Close finalizes the zip archive
func (z *ZipWriter) Close() error {
	if err := z.writer.Close(); err != nil {
		return fmt.Errorf("failed to close zip writer: %w", err)
	}

	if err := z.zipFile.Close(); err != nil {
		return fmt.Errorf("failed to close zip file: %w", err)
	}

	return nil
}

// CreateDirEntry adds a directory entry to the zip
func (z *ZipWriter) CreateDirEntry(dirPath string) error {
	relPath := strings.TrimPrefix(dirPath, z.baseDir)
	relPath = strings.TrimPrefix(relPath, "/")

	// Ensure directory path ends with a slash
	if !strings.HasSuffix(relPath, "/") {
		relPath += "/"
	}

	header := &zip.FileHeader{
		Name:   relPath,
		Method: zip.Store, // Directories are just entries, no compression needed
	}

	_, err := z.writer.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create directory entry: %w", err)
	}

	return nil
}
