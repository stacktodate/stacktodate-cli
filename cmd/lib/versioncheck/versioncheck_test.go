package versioncheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name      string
		current   string
		latest    string
		isNewer   bool
		shouldErr bool
	}{
		// Normal cases
		{"newer patch version", "0.2.2", "0.2.3", true, false},
		{"newer minor version", "0.2.2", "0.3.0", true, false},
		{"newer major version", "0.2.2", "1.0.0", true, false},
		{"same version", "0.2.2", "0.2.2", false, false},
		{"current is newer", "0.3.0", "0.2.2", false, false},

		// Dev version
		{"dev is always older", "dev", "0.2.2", true, false},
		{"dev current with newer", "dev", "1.0.0", true, false},

		// With v prefix
		{"with v prefix - newer", "v0.2.2", "v0.3.0", true, false},
		{"with v prefix - same", "v0.2.2", "v0.2.2", false, false},
		{"mixed v prefix", "v0.2.2", "0.3.0", true, false},

		// Different length versions
		{"shorter current vs longer latest", "0.2", "0.2.1", true, false},
		{"longer current vs shorter latest", "0.2.1", "0.2", false, false},

		// Invalid versions
		{"invalid current", "invalid", "0.2.2", false, true},
		{"invalid latest", "0.2.2", "invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isNewer, err := CompareVersions(tt.current, tt.latest)

			if (err != nil) != tt.shouldErr {
				t.Fatalf("expected error: %v, got: %v", tt.shouldErr, err != nil)
			}

			if isNewer != tt.isNewer {
				t.Fatalf("expected isNewer: %v, got: %v", tt.isNewer, isNewer)
			}
		})
	}
}

func TestCachePath(t *testing.T) {
	cachePath, err := GetCachePath()
	if err != nil {
		t.Fatalf("GetCachePath failed: %v", err)
	}

	// Should contain .stacktodate and version-cache.json
	if !filepath.IsAbs(cachePath) {
		t.Fatalf("cache path should be absolute, got: %s", cachePath)
	}

	if !strings.Contains(cachePath, ".stacktodate") {
		t.Fatalf("cache path should contain .stacktodate, got: %s", cachePath)
	}

	if !strings.Contains(cachePath, "version-cache.json") {
		t.Fatalf("cache path should contain version-cache.json, got: %s", cachePath)
	}
}

func TestSaveAndLoadCache(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override home directory for testing
	originalGetHomeDir := getUserHomeDir
	getUserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() {
		getUserHomeDir = originalGetHomeDir
	}()

	// Test save
	testVersion := "v0.3.0"
	testURL := "https://github.com/stacktodate/stacktodate-cli/releases/tag/v0.3.0"

	if err := SaveCache(testVersion, testURL); err != nil {
		t.Fatalf("SaveCache failed: %v", err)
	}

	// Verify cache file exists
	cachePath, _ := GetCachePath()
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("cache file should exist, got error: %v", err)
	}

	// Test load
	cache, err := LoadCache()
	if err != nil {
		t.Fatalf("LoadCache failed: %v", err)
	}

	if cache.LatestVersion != testVersion {
		t.Fatalf("expected version %s, got %s", testVersion, cache.LatestVersion)
	}

	if cache.ReleaseURL != testURL {
		t.Fatalf("expected URL %s, got %s", testURL, cache.ReleaseURL)
	}
}

func TestIsCacheValid(t *testing.T) {
	tmpDir := t.TempDir()

	originalGetHomeDir := getUserHomeDir
	getUserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() {
		getUserHomeDir = originalGetHomeDir
	}()

	// Test: cache doesn't exist
	if IsCacheValid() {
		t.Fatalf("empty cache should be invalid")
	}

	// Create a valid cache
	if err := SaveCache("v0.3.0", "https://example.com"); err != nil {
		t.Fatalf("SaveCache failed: %v", err)
	}

	// Test: cache exists and is fresh
	if !IsCacheValid() {
		t.Fatalf("fresh cache should be valid")
	}

	// Modify file timestamp to be older than TTL
	cachePath, _ := GetCachePath()
	oldTime := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(cachePath, oldTime, oldTime); err != nil {
		t.Fatalf("failed to modify file times: %v", err)
	}

	// Test: cache exists but is expired
	if IsCacheValid() {
		t.Fatalf("expired cache should be invalid")
	}
}

func TestFetchLatestFromGitHub(t *testing.T) {
	// Create a test server that mimics GitHub API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent header
		if ua := r.Header.Get("User-Agent"); ua != "stacktodate-cli" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := GitHubRelease{
			TagName:     "v0.3.0",
			HTMLURL:     "https://github.com/stacktodate/stacktodate-cli/releases/tag/v0.3.0",
			PublishedAt: "2024-01-01T00:00:00Z",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// TODO: This test would need to mock the HTTP client to test properly
	// For now, we'll test the parsing logic independently
	t.Run("parse release response", func(t *testing.T) {
		jsonData := `{
			"tag_name": "v0.3.0",
			"html_url": "https://github.com/stacktodate/stacktodate-cli/releases/tag/v0.3.0",
			"published_at": "2024-01-01T00:00:00Z"
		}`

		var release GitHubRelease
		if err := json.Unmarshal([]byte(jsonData), &release); err != nil {
			t.Fatalf("failed to parse release: %v", err)
		}

		if release.TagName != "v0.3.0" {
			t.Fatalf("expected tag v0.3.0, got %s", release.TagName)
		}

		if release.HTMLURL != "https://github.com/stacktodate/stacktodate-cli/releases/tag/v0.3.0" {
			t.Fatalf("unexpected URL")
		}
	})
}

func TestGetLatestVersionWithCache(t *testing.T) {
	tmpDir := t.TempDir()

	originalGetHomeDir := getUserHomeDir
	getUserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() {
		getUserHomeDir = originalGetHomeDir
	}()

	// Create a valid cache
	testVersion := "v0.3.0"
	testURL := "https://github.com/stacktodate/stacktodate-cli/releases/tag/v0.3.0"
	if err := SaveCache(testVersion, testURL); err != nil {
		t.Fatalf("SaveCache failed: %v", err)
	}

	// Get latest version should return from cache
	version, url, err := GetLatestVersion()
	if err != nil {
		t.Fatalf("GetLatestVersion failed: %v", err)
	}

	if version != testVersion {
		t.Fatalf("expected version %s, got %s", testVersion, version)
	}

	if url != testURL {
		t.Fatalf("expected URL %s, got %s", testURL, url)
	}
}

func TestLoadCacheNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	originalGetHomeDir := getUserHomeDir
	getUserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() {
		getUserHomeDir = originalGetHomeDir
	}()

	// Try to load non-existent cache
	_, err := LoadCache()
	if err == nil {
		t.Fatalf("LoadCache should fail for non-existent cache")
	}
}
