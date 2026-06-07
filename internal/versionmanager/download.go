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
	if err := validateDownloadURL(plan.DownloadURL); err != nil {
		return err
	}
	if err := validateInstallPath(plan.TargetDir); err != nil {
		return err
	}

	resp, err := http.Get(strings.TrimSpace(plan.DownloadURL))
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

	archiveName, err := archiveFileName(plan)
	if err != nil {
		return err
	}
	archivePath := filepath.Join(tempDir, archiveName)
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

func validateDownloadURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return fmt.Errorf("download URL is required")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid download URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("download URL must use http or https")
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("download URL host is required")
	}
	return nil
}

func archiveFileName(plan VersionInstallPlan) (string, error) {
	if strings.TrimSpace(plan.ArchiveName) != "" {
		return safeArchiveBaseName(plan.ArchiveName)
	}
	if parsed, err := url.Parse(plan.DownloadURL); err == nil {
		path := strings.TrimSpace(parsed.Path)
		if path != "" && path != "/" {
			return safeURLArchiveBaseName(path)
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
	return safeArchiveBaseName(name + ".zip")
}

func safeArchiveBaseName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("archive name is required")
	}
	if strings.ContainsAny(value, `/\`) {
		return "", fmt.Errorf("invalid archive name: %s", value)
	}
	if filepath.VolumeName(value) != "" || windowsVolumeName(value) != "" {
		return "", fmt.Errorf("invalid archive name: %s", value)
	}
	if value == "." || value == ".." {
		return "", fmt.Errorf("invalid archive name: %s", value)
	}
	return value, nil
}

func safeURLArchiveBaseName(value string) (string, error) {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	segments := strings.Split(normalized, "/")
	base := ""
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		if segment == "." || segment == ".." || strings.Contains(segment, ":") || windowsVolumeName(segment) != "" {
			return "", fmt.Errorf("invalid archive name: %s", value)
		}
		base = segment
	}
	if base == "" {
		return "", fmt.Errorf("invalid archive name: %s", value)
	}
	return safeArchiveBaseName(base)
}

func extractZip(archivePath string, targetDir string) error {
	targetDir = filepath.Clean(targetDir)
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	targets, err := validateZipEntries(targetDir, reader.File)
	if err != nil {
		return err
	}

	parent := filepath.Dir(targetDir)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}
	stageDir, err := os.MkdirTemp(parent, ".devman-extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stageDir)

	for _, entry := range reader.File {
		targetPath := filepath.Join(stageDir, targets[entry.Name])
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

	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}
	if err := os.Rename(stageDir, targetDir); err != nil {
		return err
	}
	return nil
}

func validateZipEntries(targetDir string, entries []*zip.File) (map[string]string, error) {
	targets := make(map[string]string, len(entries))
	for _, entry := range entries {
		rel, err := zipEntryRel(entry.Name)
		if err != nil {
			return nil, err
		}
		targetPath := filepath.Join(targetDir, rel)
		relToTarget, err := filepath.Rel(targetDir, targetPath)
		if err != nil {
			return nil, err
		}
		if relToTarget == ".." || strings.HasPrefix(relToTarget, ".."+string(filepath.Separator)) || filepath.IsAbs(relToTarget) {
			return nil, fmt.Errorf("invalid archive entry: %s", entry.Name)
		}
		targets[entry.Name] = rel
	}
	return targets, nil
}

func zipEntryTarget(targetDir string, entryName string) (string, error) {
	rel, err := zipEntryRel(entryName)
	if err != nil {
		return "", err
	}
	targetPath := filepath.Join(targetDir, rel)
	relToTarget, err := filepath.Rel(targetDir, targetPath)
	if err != nil {
		return "", err
	}
	if relToTarget == ".." || strings.HasPrefix(relToTarget, ".."+string(filepath.Separator)) || filepath.IsAbs(relToTarget) {
		return "", fmt.Errorf("invalid archive entry: %s", entryName)
	}
	return targetPath, nil
}

func zipEntryRel(entryName string) (string, error) {
	if strings.TrimSpace(entryName) == "" {
		return "", fmt.Errorf("invalid archive entry: %s", entryName)
	}
	normalized := strings.ReplaceAll(entryName, "\\", "/")
	if strings.HasPrefix(normalized, "/") || strings.HasPrefix(normalized, `\`) || filepath.IsAbs(normalized) || windowsVolumeName(normalized) != "" {
		return "", fmt.Errorf("invalid archive entry: %s", entryName)
	}
	cleanEntry := filepath.Clean(normalized)
	cleanEntry = strings.ReplaceAll(cleanEntry, "\\", "/")
	if cleanEntry == "." || cleanEntry == ".." || strings.HasPrefix(cleanEntry, "../") || strings.Contains(cleanEntry, "/../") {
		return "", fmt.Errorf("invalid archive entry: %s", entryName)
	}
	return filepath.FromSlash(cleanEntry), nil
}

func windowsVolumeName(path string) string {
	if len(path) >= 2 && path[1] == ':' {
		first := path[0]
		if (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') {
			return path[:2]
		}
	}
	if strings.HasPrefix(path, "//") || strings.HasPrefix(path, `\\`) {
		return string(path[0]) + string(path[1])
	}
	return ""
}
