package versionmanager

import (
	"archive/zip"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeDownloader struct {
	err   error
	plans []VersionInstallPlan
}

func (f *fakeDownloader) DownloadAndExtract(plan VersionInstallPlan) error {
	f.plans = append(f.plans, plan)
	return f.err
}

func TestInstallVersionPersistsDevManSource(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\production\go1.26`
	env.dirs[`D:\production\go1.26`] = true
	reg := newFakeVersionRegistry(nil)
	downloader := &fakeDownloader{}
	service := NewService(reg, env)
	service.downloader = downloader

	result, err := service.InstallVersion("go", "1.25.0", `D:\production\go1.25.0`)
	if err != nil {
		t.Fatalf("InstallVersion returned error: %v", err)
	}
	if result == nil || !result.Success {
		t.Fatalf("result.Success = %v, want true", result)
	}
	if result.Message != "version installed" {
		t.Fatalf("result.Message = %q, want %q", result.Message, "version installed")
	}
	if len(downloader.plans) != 1 {
		t.Fatalf("downloader plans = %d, want 1", len(downloader.plans))
	}
	if downloader.plans[0].TargetDir != `D:\production\go1.25.0` {
		t.Fatalf("download target = %q, want override target", downloader.plans[0].TargetDir)
	}
	if len(reg.saved) != 1 {
		t.Fatalf("saved versions = %d, want 1", len(reg.saved))
	}
	saved := reg.saved[0]
	if saved.Source != SourceDevMan {
		t.Fatalf("saved Source = %q, want %q", saved.Source, SourceDevMan)
	}
	if saved.ToolKey != "go" || saved.Version != "1.25.0" || saved.InstallPath != `D:\production\go1.25.0` {
		t.Fatalf("saved version = %#v, want go 1.25.0 at target", saved)
	}
	if saved.BinPath != filepath.Join(`D:\production\go1.25.0`, "bin", "go.exe") {
		t.Fatalf("saved BinPath = %q, want primary shim target", saved.BinPath)
	}
	if saved.IsDefault || saved.IsActive || !saved.CanDelete || saved.DeletePolicy != DeletePolicyDirect {
		t.Fatalf("saved flags = %#v, want inactive deletable direct DevMan version", saved)
	}
	if len(reg.savedStrategies) != 1 {
		t.Fatalf("saved strategies = %d, want 1", len(reg.savedStrategies))
	}
	if reg.savedStrategies[0].ToolKey != "go" || reg.savedStrategies[0].RootDir != `D:\production` {
		t.Fatalf("saved strategy = %#v, want go root D:\\production", reg.savedStrategies[0])
	}
	if got := result.AffectedEnvironment["GOROOT"]; got != `D:\production\go1.25.0` {
		t.Fatalf("affected GOROOT = %q, want target dir", got)
	}
}

func TestInstallVersionDownloaderFailureDoesNotSaveVersion(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\production\go1.26`
	env.dirs[`D:\production\go1.26`] = true
	reg := newFakeVersionRegistry(nil)
	service := NewService(reg, env)
	service.downloader = &fakeDownloader{err: errors.New("download failed")}

	_, err := service.InstallVersion("go", "1.25.0", `D:\production\go1.25.0`)
	if err == nil {
		t.Fatal("InstallVersion error = nil, want downloader error")
	}
	if len(reg.saved) != 0 {
		t.Fatalf("saved versions = %d, want 0 after downloader failure", len(reg.saved))
	}
}

func TestInstallVersionUnsafeOverrideTargetDoesNotDownloadOrSave(t *testing.T) {
	tests := []struct {
		name      string
		targetDir string
	}{
		{name: "relative", targetDir: `go1.25.0`},
		{name: "quote", targetDir: `D:\production\bad"path`},
		{name: "drive root", targetDir: `D:\`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newFakeEnvironment()
			env.vars["GOROOT"] = `D:\production\go1.26`
			env.dirs[`D:\production\go1.26`] = true
			reg := newFakeVersionRegistry(nil)
			downloader := &fakeDownloader{}
			service := NewService(reg, env)
			service.downloader = downloader

			_, err := service.InstallVersion("go", "1.25.0", tt.targetDir)
			if err == nil {
				t.Fatal("InstallVersion error = nil, want unsafe target error")
			}
			if len(downloader.plans) != 0 {
				t.Fatalf("downloader plans = %d, want 0 for unsafe target", len(downloader.plans))
			}
			if len(reg.saved) != 0 {
				t.Fatalf("saved versions = %d, want 0 for unsafe target", len(reg.saved))
			}
		})
	}
}

func TestInstallVersionStrategyFailureDoesNotSaveVersion(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\production\go1.26`
	env.dirs[`D:\production\go1.26`] = true
	reg := newFakeVersionRegistry(nil)
	reg.saveStrategyErr = errors.New("strategy failed")
	service := NewService(reg, env)
	service.downloader = &fakeDownloader{}

	_, err := service.InstallVersion("go", "1.25.0", `D:\production\go1.25.0`)
	if err == nil {
		t.Fatal("InstallVersion error = nil, want strategy save error")
	}
	if len(reg.saved) != 0 {
		t.Fatalf("saved versions = %d, want 0 when strategy save fails", len(reg.saved))
	}
}

func TestInstallVersionUnsupportedToolUsesPreviewValidation(t *testing.T) {
	env := newFakeEnvironment()
	reg := newFakeVersionRegistry(nil)
	downloader := &fakeDownloader{}
	service := NewService(reg, env)
	service.downloader = downloader

	_, err := service.InstallVersion("ruby", "3.3.0", `D:\production\ruby3.3.0`)
	if err == nil || !strings.Contains(err.Error(), "unsupported tool: ruby") {
		t.Fatalf("InstallVersion error = %v, want unsupported tool error", err)
	}
	if len(downloader.plans) != 0 {
		t.Fatalf("downloader plans = %d, want 0 for preview validation failure", len(downloader.plans))
	}
	if len(reg.saved) != 0 {
		t.Fatalf("saved versions = %d, want 0", len(reg.saved))
	}
}

func TestInstallVersionConflictUsesPreviewValidation(t *testing.T) {
	env := newFakeEnvironment()
	env.vars["GOROOT"] = `D:\production\go1.26`
	env.paths["gvm"] = `D:\gvm\gvm.exe`
	env.dirs[`D:\production\go1.26`] = true
	reg := newFakeVersionRegistry(nil)
	downloader := &fakeDownloader{}
	service := NewService(reg, env)
	service.downloader = downloader

	_, err := service.InstallVersion("go", "1.25.0", `D:\production\go1.25.0`)
	if err == nil || !strings.Contains(err.Error(), "managed by gvm") {
		t.Fatalf("InstallVersion error = %v, want conflict error", err)
	}
	if len(downloader.plans) != 0 {
		t.Fatalf("downloader plans = %d, want 0 for conflict", len(downloader.plans))
	}
}

func TestExtractZipRejectsZipSlip(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "escape.zip")
	createZip(t, archivePath, map[string]string{
		"../evil.exe": "bad",
	})

	err := extractZip(archivePath, filepath.Join(t.TempDir(), "target"))
	if err == nil {
		t.Fatal("extractZip error = nil, want zip-slip rejection")
	}
	if !strings.Contains(err.Error(), "invalid archive entry") {
		t.Fatalf("extractZip error = %v, want invalid archive entry", err)
	}
}

func TestExtractZipPrevalidatesBeforeWritingFiles(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "partial.zip")
	createZip(t, archivePath, map[string]string{
		"safe.txt":       "safe",
		`C:\evil.exe`:    "bad",
		`dir\..\evil.go`: "bad",
	})
	targetDir := filepath.Join(t.TempDir(), "target")

	err := extractZip(archivePath, targetDir)
	if err == nil {
		t.Fatal("extractZip error = nil, want invalid archive entry")
	}
	if _, statErr := os.Stat(filepath.Join(targetDir, "safe.txt")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("safe file stat error = %v, want not exist after prevalidation failure", statErr)
	}
}

func TestHTTPDownloaderValidatesDownloadURL(t *testing.T) {
	tests := []struct {
		name        string
		downloadURL string
	}{
		{name: "blank", downloadURL: ""},
		{name: "malformed", downloadURL: "http://[::1"},
		{name: "unsupported scheme", downloadURL: "file:///tmp/go.zip"},
		{name: "missing host", downloadURL: "https:///go.zip"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (HTTPDownloader{}).DownloadAndExtract(VersionInstallPlan{
				ToolKey:     "go",
				Version:     "1.25.0",
				TargetDir:   filepath.Join(t.TempDir(), "go1.25.0"),
				DownloadURL: tt.downloadURL,
				ArchiveName: "go.zip",
			})
			if err == nil {
				t.Fatal("DownloadAndExtract error = nil, want URL validation error")
			}
		})
	}
}

func TestHTTPDownloaderValidatesTargetDir(t *testing.T) {
	tests := []struct {
		name      string
		targetDir string
	}{
		{name: "blank", targetDir: ""},
		{name: "relative", targetDir: "go1.25.0"},
		{name: "quote", targetDir: `D:\production\bad"path`},
		{name: "drive root", targetDir: `D:\`},
		{name: "filesystem root", targetDir: filepath.VolumeName(os.TempDir()) + string(filepath.Separator)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (HTTPDownloader{}).DownloadAndExtract(VersionInstallPlan{
				ToolKey:     "go",
				Version:     "1.25.0",
				TargetDir:   tt.targetDir,
				DownloadURL: "https://example.com/go.zip",
				ArchiveName: "go.zip",
			})
			if err == nil {
				t.Fatal("DownloadAndExtract error = nil, want target validation error")
			}
		})
	}
}

func TestArchiveFileNameRejectsUnsafeFallbacks(t *testing.T) {
	tests := []struct {
		name string
		plan VersionInstallPlan
	}{
		{name: "unsafe archive name", plan: VersionInstallPlan{ArchiveName: `..\go.zip`, DownloadURL: "https://example.com/"}},
		{name: "unsafe URL basename", plan: VersionInstallPlan{DownloadURL: "https://example.com/C:/go.zip"}},
		{name: "unsafe fallback tool", plan: VersionInstallPlan{ToolKey: `go\evil`, Version: "1.25.0"}},
		{name: "unsafe fallback version", plan: VersionInstallPlan{ToolKey: "go", Version: `..\1.25.0`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if name, err := archiveFileName(tt.plan); err == nil {
				t.Fatalf("archiveFileName = %q, nil error; want unsafe name error", name)
			}
		})
	}
}

func TestArchiveFileNameReturnsSafeBasename(t *testing.T) {
	name, err := archiveFileName(VersionInstallPlan{
		ToolKey:     "go",
		Version:     "1.25.0",
		DownloadURL: "https://example.com/downloads/go.zip",
	})
	if err != nil {
		t.Fatalf("archiveFileName returned error: %v", err)
	}
	if name != "go.zip" {
		t.Fatalf("archiveFileName = %q, want go.zip", name)
	}
	if strings.ContainsAny(name, `/\`) || filepath.VolumeName(name) != "" || name == "." || name == ".." {
		t.Fatalf("archiveFileName = %q, want filename-only safe basename", name)
	}
}

func TestHTTPDownloaderRejectsNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusTeapot)
	}))
	defer server.Close()

	err := (HTTPDownloader{}).DownloadAndExtract(VersionInstallPlan{
		ToolKey:     "go",
		Version:     "1.25.0",
		TargetDir:   filepath.Join(t.TempDir(), "go1.25.0"),
		DownloadURL: server.URL + "/go.zip",
		ArchiveName: "go.zip",
	})
	if err == nil || !strings.Contains(err.Error(), "download returned 418") {
		t.Fatalf("DownloadAndExtract error = %v, want non-2xx status", err)
	}
}

func TestHTTPDownloaderDownloadsAndExtractsZip(t *testing.T) {
	zipBytes := zipBytes(t, map[string]string{
		"go/bin/go.exe": "binary",
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(zipBytes)
	}))
	defer server.Close()

	targetDir := filepath.Join(t.TempDir(), "go1.25.0")
	err := (HTTPDownloader{}).DownloadAndExtract(VersionInstallPlan{
		ToolKey:     "go",
		Version:     "1.25.0",
		TargetDir:   targetDir,
		DownloadURL: server.URL + "/go.zip",
		ArchiveName: "go.zip",
	})
	if err != nil {
		t.Fatalf("DownloadAndExtract returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "go", "bin", "go.exe")); err != nil {
		t.Fatalf("extracted file stat error: %v", err)
	}
}

func createZip(t *testing.T, archivePath string, files map[string]string) {
	t.Helper()
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create zip file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatalf("close zip file: %v", err)
		}
	}()

	writer := zip.NewWriter(file)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %q: %v", name, err)
		}
		if _, err := fmt.Fprint(entry, content); err != nil {
			t.Fatalf("write zip entry %q: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
}

func zipBytes(t *testing.T, files map[string]string) []byte {
	t.Helper()
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "archive.zip")
	createZip(t, archivePath, files)
	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read zip bytes: %v", err)
	}
	return data
}
