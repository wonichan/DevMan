package versionmanager

import (
	"os"
	"testing"
)

func TestParseNodeVersions(t *testing.T) {
	data, err := os.ReadFile("testdata/node-index.json")
	if err != nil {
		t.Fatal(err)
	}

	versions, err := ParseNodeVersions(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].Version != "22.11.0" {
		t.Fatalf("expected version 22.11.0, got %q", versions[0].Version)
	}
}

func TestParseNodeVersionsFiltersMissingWindowsZip(t *testing.T) {
	data := []byte(`[{"version":"v22.11.0","date":"2024-10-29","files":["linux-x64"]}]`)

	versions, err := ParseNodeVersions(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) != 0 {
		t.Fatalf("expected no versions, got %d", len(versions))
	}
}

func TestParseNodeVersionsRejectsInvalidJSON(t *testing.T) {
	if _, err := ParseNodeVersions([]byte(`{`)); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestParseGoVersions(t *testing.T) {
	data, err := os.ReadFile("testdata/go.json")
	if err != nil {
		t.Fatal(err)
	}

	versions, err := ParseGoVersions(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].Version != "1.25.0" {
		t.Fatalf("expected version 1.25.0, got %q", versions[0].Version)
	}
	if versions[0].DownloadURL == "" {
		t.Fatal("expected non-empty download URL")
	}
	if versions[0].Checksum != "abc123" {
		t.Fatalf("expected checksum abc123, got %q", versions[0].Checksum)
	}
}

func TestParseGoVersionsFiltersNonWindowsAMD64(t *testing.T) {
	data := []byte(`[{"version":"go1.25.0","stable":true,"files":[{"filename":"go1.25.0.linux-amd64.tar.gz","os":"linux","arch":"amd64","sha256":"abc123"},{"filename":"go1.25.0.windows-arm64.zip","os":"windows","arch":"arm64","sha256":"def456"}]}]`)

	versions, err := ParseGoVersions(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) != 0 {
		t.Fatalf("expected no versions, got %d", len(versions))
	}
}

func TestParseGoVersionsRejectsInvalidJSON(t *testing.T) {
	if _, err := ParseGoVersions([]byte(`{`)); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
