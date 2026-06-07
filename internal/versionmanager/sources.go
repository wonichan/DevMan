package versionmanager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OfficialVersionProvider interface {
	Fetch(toolKey string) (*ToolVersionCatalog, error)
}

type HTTPVersionProvider struct {
	Client *http.Client
}

func (p HTTPVersionProvider) Fetch(toolKey string) (*ToolVersionCatalog, error) {
	sourceURL := officialSourceURL(toolKey)
	if sourceURL == "" {
		return nil, fmt.Errorf("unsupported tool: %s", toolKey)
	}
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Get(sourceURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("official source returned %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var versions []AvailableVersion
	switch toolKey {
	case "go":
		versions, err = ParseGoVersions(data)
	case "node":
		versions, err = ParseNodeVersions(data)
	case "bun":
		versions, err = ParseBunVersions(data)
	case "flutter":
		versions, err = ParseFlutterVersions(data)
	}
	if err != nil {
		return nil, err
	}

	return &ToolVersionCatalog{
		ToolKey:   toolKey,
		Versions:  versions,
		FetchedAt: time.Now(),
		SourceURL: sourceURL,
	}, nil
}

func officialSourceURL(toolKey string) string {
	switch toolKey {
	case "go":
		return "https://go.dev/dl/?mode=json"
	case "node":
		return "https://nodejs.org/dist/index.json"
	case "bun":
		return "https://api.github.com/repos/oven-sh/bun/releases"
	case "flutter":
		return "https://storage.googleapis.com/flutter_infra_release/releases/releases_windows.json"
	default:
		return ""
	}
}

func ParseNodeVersions(data []byte) ([]AvailableVersion, error) {
	var releases []struct {
		Version string   `json:"version"`
		Date    string   `json:"date"`
		Files   []string `json:"files"`
	}
	if err := json.Unmarshal(data, &releases); err != nil {
		return nil, err
	}

	versions := make([]AvailableVersion, 0, len(releases))
	for _, release := range releases {
		if !contains(release.Files, "win-x64-zip") {
			continue
		}
		version := strings.TrimPrefix(release.Version, "v")
		releaseDate, err := time.Parse("2006-01-02", release.Date)
		if err != nil {
			return nil, err
		}
		versions = append(versions, AvailableVersion{
			Version:     version,
			Stable:      true,
			ReleaseDate: releaseDate,
			Arch:        "windows-amd64",
			DownloadURL: fmt.Sprintf("https://nodejs.org/dist/v%s/node-v%s-win-x64.zip", version, version),
		})
	}
	return versions, nil
}

func ParseGoVersions(data []byte) ([]AvailableVersion, error) {
	var releases []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
		Files   []struct {
			Filename string `json:"filename"`
			OS       string `json:"os"`
			Arch     string `json:"arch"`
			SHA256   string `json:"sha256"`
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &releases); err != nil {
		return nil, err
	}

	versions := make([]AvailableVersion, 0, len(releases))
	for _, release := range releases {
		version := strings.TrimPrefix(release.Version, "go")
		for _, file := range release.Files {
			if file.OS != "windows" || file.Arch != "amd64" {
				continue
			}
			versions = append(versions, AvailableVersion{
				Version:     version,
				Stable:      release.Stable,
				Arch:        "windows-amd64",
				DownloadURL: "https://go.dev/dl/" + file.Filename,
				Checksum:    file.SHA256,
			})
		}
	}
	return versions, nil
}

func ParseBunVersions(data []byte) ([]AvailableVersion, error) {
	var releases []struct {
		TagName     string `json:"tag_name"`
		Prerelease  bool   `json:"prerelease"`
		PublishedAt string `json:"published_at"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(data, &releases); err != nil {
		return nil, err
	}

	versions := make([]AvailableVersion, 0, len(releases))
	for _, release := range releases {
		version := strings.TrimPrefix(release.TagName, "bun-v")
		version = strings.TrimPrefix(version, "v")
		if strings.TrimSpace(version) == "" {
			continue
		}
		releaseDate, err := parseOptionalReleaseDate(release.PublishedAt)
		if err != nil {
			return nil, err
		}
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)
			if !strings.Contains(name, "windows-x64") || !strings.HasSuffix(name, ".zip") {
				continue
			}
			if strings.TrimSpace(asset.BrowserDownloadURL) == "" {
				continue
			}
			versions = append(versions, AvailableVersion{
				Version:     version,
				Stable:      !release.Prerelease,
				ReleaseDate: releaseDate,
				Arch:        "windows-amd64",
				DownloadURL: asset.BrowserDownloadURL,
			})
		}
	}
	return versions, nil
}

func ParseFlutterVersions(data []byte) ([]AvailableVersion, error) {
	var catalog struct {
		BaseURL  string `json:"base_url"`
		Releases []struct {
			Version     string `json:"version"`
			Channel     string `json:"channel"`
			Archive     string `json:"archive"`
			DartSDKArch string `json:"dart_sdk_arch"`
			Hash        string `json:"hash"`
			ReleaseDate string `json:"release_date"`
		} `json:"releases"`
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, err
	}

	versions := make([]AvailableVersion, 0, len(catalog.Releases))
	for _, release := range catalog.Releases {
		if release.Channel != "stable" || release.DartSDKArch != "x64" {
			continue
		}
		if strings.TrimSpace(release.Version) == "" || strings.TrimSpace(release.Archive) == "" {
			continue
		}
		downloadURL, err := flutterDownloadURL(catalog.BaseURL, release.Archive)
		if err != nil {
			return nil, err
		}
		releaseDate, err := parseOptionalReleaseDate(release.ReleaseDate)
		if err != nil {
			return nil, err
		}
		versions = append(versions, AvailableVersion{
			Version:     release.Version,
			Stable:      true,
			ReleaseDate: releaseDate,
			Arch:        "windows-amd64",
			DownloadURL: downloadURL,
			Checksum:    release.Hash,
		})
	}
	return versions, nil
}

func parseOptionalReleaseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02", value)
}

func flutterDownloadURL(baseURL string, archive string) (string, error) {
	archive = strings.TrimSpace(archive)
	parsed, err := url.Parse(archive)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() {
		return archive, nil
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return "", fmt.Errorf("flutter base_url is required for archive %s", archive)
	}
	return baseURL + "/" + strings.TrimLeft(archive, "/"), nil
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
