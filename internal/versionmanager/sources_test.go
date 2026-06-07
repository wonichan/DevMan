package versionmanager

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	if versions == nil {
		t.Fatal("expected non-nil empty versions")
	}
}

func TestParseGoVersionsRejectsInvalidJSON(t *testing.T) {
	if _, err := ParseGoVersions([]byte(`{`)); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestParseBunVersionsReturnsWindowsX64Assets(t *testing.T) {
	data := []byte(`[
		{
			"tag_name": "bun-v1.2.0",
			"published_at": "2025-01-15T12:30:00Z",
			"assets": [
				{"name": "bun-linux-x64.zip", "browser_download_url": "https://example.com/bun-linux-x64.zip"},
				{"name": "bun-windows-x64.zip", "browser_download_url": "https://example.com/bun-windows-x64.zip"}
			]
		},
		{
			"tag_name": "v1.1.0",
			"published_at": "2024-12-01T00:00:00Z",
			"assets": [
				{"name": "bun-windows-aarch64.zip", "browser_download_url": "https://example.com/bun-windows-aarch64.zip"}
			]
		}
	]`)

	versions, err := ParseBunVersions(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	version := versions[0]
	if version.Version != "1.2.0" {
		t.Fatalf("Version = %q", version.Version)
	}
	if !version.Stable {
		t.Fatal("expected Stable")
	}
	if version.Arch != "windows-amd64" {
		t.Fatalf("Arch = %q", version.Arch)
	}
	if version.DownloadURL != "https://example.com/bun-windows-x64.zip" {
		t.Fatalf("DownloadURL = %q", version.DownloadURL)
	}
	if version.ReleaseDate.IsZero() {
		t.Fatal("expected ReleaseDate")
	}
}

func TestParseFlutterVersionsReturnsStableWindowsX64Archives(t *testing.T) {
	data := []byte(`{
		"base_url": "https://storage.googleapis.com/flutter_infra_release/releases",
		"releases": [
			{
				"version": "3.27.1",
				"channel": "stable",
				"archive": "stable/windows/flutter_windows_3.27.1-stable.zip",
				"dart_sdk_arch": "x64",
				"hash": "abc123",
				"release_date": "2024-12-18T17:03:00.000Z"
			},
			{
				"version": "3.28.0-0.1.pre",
				"channel": "beta",
				"archive": "beta/windows/flutter_windows_3.28.0-0.1.pre-beta.zip",
				"dart_sdk_arch": "x64",
				"hash": "def456",
				"release_date": "2025-01-02T00:00:00.000Z"
			},
			{
				"version": "3.27.1",
				"channel": "stable",
				"archive": "stable/windows/flutter_windows_3.27.1-stable-arm64.zip",
				"dart_sdk_arch": "arm64",
				"hash": "ghi789",
				"release_date": "2024-12-18T17:03:00.000Z"
			}
		]
	}`)

	versions, err := ParseFlutterVersions(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	version := versions[0]
	if version.Version != "3.27.1" {
		t.Fatalf("Version = %q", version.Version)
	}
	if !version.Stable {
		t.Fatal("expected Stable")
	}
	if version.Arch != "windows-amd64" {
		t.Fatalf("Arch = %q", version.Arch)
	}
	if version.DownloadURL != "https://storage.googleapis.com/flutter_infra_release/releases/stable/windows/flutter_windows_3.27.1-stable.zip" {
		t.Fatalf("DownloadURL = %q", version.DownloadURL)
	}
	if version.Checksum != "abc123" {
		t.Fatalf("Checksum = %q", version.Checksum)
	}
	if version.ReleaseDate.IsZero() {
		t.Fatal("expected ReleaseDate")
	}
}

func TestHTTPVersionProviderFetchGoUsesOfficialURL(t *testing.T) {
	data, err := os.ReadFile("testdata/go.json")
	if err != nil {
		t.Fatal(err)
	}
	transport := fakeRoundTripper{
		responses: map[string]fakeHTTPResponse{
			officialSourceURL("go"): {
				statusCode: http.StatusOK,
				status:     "200 OK",
				body:       string(data),
			},
		},
	}

	catalog, err := (HTTPVersionProvider{
		Client: &http.Client{Transport: &transport},
	}).Fetch("go")
	if err != nil {
		t.Fatal(err)
	}

	if catalog.ToolKey != "go" {
		t.Fatalf("ToolKey = %q", catalog.ToolKey)
	}
	if catalog.SourceURL != officialSourceURL("go") {
		t.Fatalf("SourceURL = %q", catalog.SourceURL)
	}
	if catalog.Versions == nil {
		t.Fatal("expected non-nil Versions")
	}
	if len(catalog.Versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(catalog.Versions))
	}
	if transport.calls != 1 {
		t.Fatalf("expected 1 HTTP call, got %d", transport.calls)
	}
}

func TestHTTPVersionProviderFetchNon2xxReturnsStatus(t *testing.T) {
	transport := fakeRoundTripper{
		responses: map[string]fakeHTTPResponse{
			officialSourceURL("node"): {
				statusCode: http.StatusTooManyRequests,
				status:     "429 Too Many Requests",
				body:       `rate limited`,
			},
		},
	}

	_, err := (HTTPVersionProvider{
		Client: &http.Client{Transport: &transport},
	}).Fetch("node")
	if err == nil {
		t.Fatal("expected non-2xx error")
	}
	if err.Error() != "official source returned 429 Too Many Requests" {
		t.Fatalf("error = %q", err)
	}
}

func TestHTTPVersionProviderFetchUnsupportedTool(t *testing.T) {
	_, err := (HTTPVersionProvider{}).Fetch("unknown")
	if err == nil {
		t.Fatal("expected unsupported tool error")
	}
	if err.Error() != "unsupported tool: unknown" {
		t.Fatalf("error = %q", err)
	}
}

func TestHTTPVersionProviderFetchMalformedJSONPropagatesError(t *testing.T) {
	transport := fakeRoundTripper{
		responses: map[string]fakeHTTPResponse{
			officialSourceURL("go"): {
				statusCode: http.StatusOK,
				status:     "200 OK",
				body:       `{`,
			},
		},
	}

	_, err := (HTTPVersionProvider{
		Client: &http.Client{Transport: &transport},
	}).Fetch("go")
	if err == nil {
		t.Fatal("expected malformed JSON error")
	}
}

func TestHTTPVersionProviderFetchBunUsesOfficialURL(t *testing.T) {
	transport := fakeRoundTripper{
		responses: map[string]fakeHTTPResponse{
			officialSourceURL("bun"): {
				statusCode: http.StatusOK,
				status:     "200 OK",
				body:       `[{"tag_name":"v1.2.0","assets":[{"name":"bun-windows-x64.zip","browser_download_url":"https://example.com/bun.zip"}]}]`,
			},
		},
	}

	catalog, err := (HTTPVersionProvider{
		Client: &http.Client{Transport: &transport},
	}).Fetch("bun")
	if err != nil {
		t.Fatal(err)
	}
	if catalog.ToolKey != "bun" {
		t.Fatalf("ToolKey = %q", catalog.ToolKey)
	}
	if catalog.SourceURL != officialSourceURL("bun") {
		t.Fatalf("SourceURL = %q", catalog.SourceURL)
	}
	if len(catalog.Versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(catalog.Versions))
	}
	if transport.calls != 1 {
		t.Fatalf("expected 1 HTTP call, got %d", transport.calls)
	}
}

func TestHTTPVersionProviderFetchFlutterUsesOfficialURL(t *testing.T) {
	transport := fakeRoundTripper{
		responses: map[string]fakeHTTPResponse{
			officialSourceURL("flutter"): {
				statusCode: http.StatusOK,
				status:     "200 OK",
				body:       `{"base_url":"https://storage.googleapis.com/flutter_infra_release/releases","releases":[{"version":"3.27.1","channel":"stable","archive":"stable/windows/flutter_windows_3.27.1-stable.zip","dart_sdk_arch":"x64","release_date":"2024-12-18T17:03:00.000Z"}]}`,
			},
		},
	}

	catalog, err := (HTTPVersionProvider{
		Client: &http.Client{Transport: &transport},
	}).Fetch("flutter")
	if err != nil {
		t.Fatal(err)
	}
	if catalog.ToolKey != "flutter" {
		t.Fatalf("ToolKey = %q", catalog.ToolKey)
	}
	if catalog.SourceURL != officialSourceURL("flutter") {
		t.Fatalf("SourceURL = %q", catalog.SourceURL)
	}
	if len(catalog.Versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(catalog.Versions))
	}
	if transport.calls != 1 {
		t.Fatalf("expected 1 HTTP call, got %d", transport.calls)
	}
}

type fakeRoundTripper struct {
	responses map[string]fakeHTTPResponse
	err       error
	calls     int
}

func (t *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	t.calls++
	if t.err != nil {
		return nil, t.err
	}
	response, ok := t.responses[req.URL.String()]
	if !ok {
		return nil, fmt.Errorf("unexpected URL: %s", req.URL.String())
	}
	return &http.Response{
		StatusCode: response.statusCode,
		Status:     response.status,
		Body:       io.NopCloser(strings.NewReader(response.body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

type fakeHTTPResponse struct {
	statusCode int
	status     string
	body       string
}
