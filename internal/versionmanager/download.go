package versionmanager

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Downloader interface {
	DownloadAndExtract(plan VersionInstallPlan) error
}

type HTTPDownloader struct{}

func (HTTPDownloader) DownloadAndExtract(plan VersionInstallPlan) error {
	if strings.TrimSpace(plan.DownloadURL) == "" {
		return fmt.Errorf("download URL is required")
	}
	if strings.TrimSpace(plan.TargetDir) == "" {
		return fmt.Errorf("target directory is required")
	}

	resp, err := http.Get(plan.DownloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download returned %d", resp.StatusCode)
	}

	tempDir, err := os.MkdirTemp("", "devman-download-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, archiveFileName(plan))
	archive, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(archive, resp.Body); err != nil {
		closeErr := archive.Close()
		if closeErr != nil {
			return fmt.Errorf("%w; close archive: %v", err, closeErr)
		}
		return err
	}
	if err := archive.Close(); err != nil {
		return err
	}

	return extractZip(archivePath, plan.TargetDir)
}

func archiveFileName(plan VersionInstallPlan) string {
	if name := safeBaseName(plan.ArchiveName); name != "" {
		return name
	}
	if parsed, err := url.Parse(plan.DownloadURL); err == nil {
		if name := safeBaseName(parsed.Path); name != "" {
			return name
		}
	}
	name := strings.TrimSpace(plan.ToolKey)
	if name == "" {
		name = "archive"
	}
	version := strings.TrimSpace(plan.Version)
	if version != "" {
		name += "-" + version
	}
	return name + ".zip"
}

func safeBaseName(value string) string {
	base := filepath.Base(strings.TrimSpace(value))
	if base == "" || base == "." || base == string(filepath.Separator) {
		return ""
	}
	return base
}

func extractZip(archivePath string, targetDir string) error {
	targetDir = filepath.Clean(targetDir)
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	for _, entry := range reader.File {
		targetPath, err := zipEntryTarget(targetDir, entry.Name)
		if err != nil {
			return err
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, entry.Mode()); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		source, err := entry.Open()
		if err != nil {
			return err
		}
		dest, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, entry.Mode())
		if err != nil {
			closeErr := source.Close()
			if closeErr != nil {
				return fmt.Errorf("%w; close archive entry: %v", err, closeErr)
			}
			return err
		}
		if _, err := io.Copy(dest, source); err != nil {
			closeSourceErr := source.Close()
			closeDestErr := dest.Close()
			if closeSourceErr != nil {
				return fmt.Errorf("%w; close archive entry: %v", err, closeSourceErr)
			}
			if closeDestErr != nil {
				return fmt.Errorf("%w; close destination: %v", err, closeDestErr)
			}
			return err
		}
		if err := source.Close(); err != nil {
			closeDestErr := dest.Close()
			if closeDestErr != nil {
				return fmt.Errorf("%w; close destination: %v", err, closeDestErr)
			}
			return err
		}
		if err := dest.Close(); err != nil {
			return err
		}
	}
	return nil
}

func zipEntryTarget(targetDir string, entryName string) (string, error) {
	if strings.TrimSpace(entryName) == "" {
		return "", fmt.Errorf("invalid archive entry: %s", entryName)
	}
	normalized := strings.ReplaceAll(entryName, "\\", "/")
	if filepath.IsAbs(normalized) || strings.HasPrefix(normalized, "/") {
		return "", fmt.Errorf("invalid archive entry: %s", entryName)
	}
	cleanEntry := filepath.Clean(normalized)
	if cleanEntry == "." || strings.HasPrefix(cleanEntry, ".."+string(filepath.Separator)) || cleanEntry == ".." {
		return "", fmt.Errorf("invalid archive entry: %s", entryName)
	}
	targetPath := filepath.Join(targetDir, cleanEntry)
	rel, err := filepath.Rel(targetDir, targetPath)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("invalid archive entry: %s", entryName)
	}
	return targetPath, nil
}
