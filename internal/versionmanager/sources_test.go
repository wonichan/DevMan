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

func TestHTTPVersionProviderFetchBunAndFlutterShortCircuitWithoutHTTP(t *testing.T) {
	for _, toolKey := range []string{"bun", "flutter"} {
		t.Run(toolKey, func(t *testing.T) {
			transport := fakeRoundTripper{
				err: fmt.Errorf("unexpected HTTP call"),
			}

			catalog, err := (HTTPVersionProvider{
				Client: &http.Client{Transport: &transport},
			}).Fetch(toolKey)
			if err != nil {
				t.Fatal(err)
			}
			if catalog.ToolKey != toolKey {
				t.Fatalf("ToolKey = %q", catalog.ToolKey)
			}
			if catalog.SourceURL != officialSourceURL(toolKey) {
				t.Fatalf("SourceURL = %q", catalog.SourceURL)
			}
			if catalog.FetchedAt.IsZero() {
				t.Fatal("expected FetchedAt")
			}
			if catalog.Versions == nil {
				t.Fatal("expected non-nil Versions")
			}
			if len(catalog.Versions) != 0 {
				t.Fatalf("expected no versions, got %d", len(catalog.Versions))
			}
			if transport.calls != 0 {
				t.Fatalf("expected no HTTP calls, got %d", transport.calls)
			}
		})
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
