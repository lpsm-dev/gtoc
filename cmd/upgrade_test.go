package cmd

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShouldUpgrade(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"newer minor available", "v0.9.0", "v0.10.0", true},
		{"current already newest", "v0.10.0", "v0.9.0", false},
		{"equal versions", "v1.2.3", "v1.2.3", false},
		{"dev always upgrades", "dev", "v1.0.0", true},
		{"without v prefix", "1.0.0", "1.1.0", true},
		{"unparsable current always upgrades", "not-a-version", "v1.0.0", true},
		{"unparsable latest never upgrades", "v1.0.0", "not-a-version", false},
		{"unequal segment lengths, latest has extra patch", "v1.2", "v1.2.1", true},
		{"unequal segment lengths, equal after padding", "v1.2.0", "v1.2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldUpgrade(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("shouldUpgrade(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestSelectAsset(t *testing.T) {
	release := &Release{
		TagName: "v1.2.3",
		Assets: []ReleaseAsset{
			{Name: "gtoc_Darwin_x86_64.tar.gz", BrowserDownloadURL: "https://example.com/darwin-amd64"},
			{Name: "gtoc_Darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin-arm64"},
			{Name: "gtoc_Linux_x86_64.tar.gz", BrowserDownloadURL: "https://example.com/linux-amd64"},
			{Name: "gtoc_Linux_arm64.tar.gz", BrowserDownloadURL: "https://example.com/linux-arm64"},
			{Name: "gtoc_Windows_x86_64.zip", BrowserDownloadURL: "https://example.com/windows-amd64"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums"},
		},
	}

	tests := []struct {
		name    string
		goos    string
		goarch  string
		wantURL string
		wantErr bool
	}{
		{"darwin amd64", "darwin", "amd64", "https://example.com/darwin-amd64", false},
		{"darwin arm64", "darwin", "arm64", "https://example.com/darwin-arm64", false},
		{"linux amd64", "linux", "amd64", "https://example.com/linux-amd64", false},
		{"linux arm64", "linux", "arm64", "https://example.com/linux-arm64", false},
		{"windows amd64 picks zip", "windows", "amd64", "https://example.com/windows-amd64", false},
		{"unsupported os", "plan9", "amd64", "", true},
		{"unsupported arch", "linux", "mips", "", true},
		{"missing asset for platform", "linux", "arm", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset, err := selectAsset(release, tt.goos, tt.goarch)
			if (err != nil) != tt.wantErr {
				t.Fatalf("selectAsset() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if asset.BrowserDownloadURL != tt.wantURL {
				t.Errorf("selectAsset() URL = %q, want %q", asset.BrowserDownloadURL, tt.wantURL)
			}
		})
	}
}

func TestGetLatestReleaseSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.3","name":"v1.2.3","published_at":"2024-01-01T00:00:00Z","assets":[]}`))
	}))
	defer server.Close()

	release, err := getLatestRelease(server.URL)
	if err != nil {
		t.Fatalf("getLatestRelease() error = %v", err)
	}
	if release.TagName != "v1.2.3" {
		t.Errorf("release.TagName = %q, want %q", release.TagName, "v1.2.3")
	}
}

func TestGetLatestReleaseNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := getLatestRelease(server.URL)
	if !errors.Is(err, errNoReleases) {
		t.Fatalf("getLatestRelease() error = %v, want errNoReleases", err)
	}
}
