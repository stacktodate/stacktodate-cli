package versioncheck

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	cacheDirName  = ".stacktodate"
	cacheFileName = "version-cache.json"
	cacheTTL      = 24 * time.Hour
	githubAPIURL  = "https://api.github.com/repos/stacktodate/stacktodate-cli/releases/latest"
	httpTimeout   = 10 * time.Second
)

// getUserHomeDir returns the user's home directory (can be overridden for testing)
var getUserHomeDir = os.UserHomeDir

// VersionCache represents the cached version information
type VersionCache struct {
	Timestamp     time.Time `json:"timestamp"`
	LatestVersion string    `json:"latestVersion"`
	ReleaseURL    string    `json:"releaseUrl"`
}

// GitHubRelease represents the GitHub API response for a release
type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
}

// GetCachePath returns the full path to the version cache file
func GetCachePath() (string, error) {
	home, err := getUserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(home, cacheDirName)
	cachePath := filepath.Join(cacheDir, cacheFileName)
	return cachePath, nil
}

// IsCacheValid checks if a valid cache file exists and is not expired
func IsCacheValid() bool {
	cachePath, err := GetCachePath()
	if err != nil {
		return false
	}

	info, err := os.Stat(cachePath)
	if err != nil {
		return false
	}

	return time.Since(info.ModTime()) < cacheTTL
}

// LoadCache loads the version cache from disk
func LoadCache() (*VersionCache, error) {
	cachePath, err := GetCachePath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache VersionCache
	if err := json.Unmarshal(content, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return &cache, nil
}

// SaveCache saves the version information to cache
func SaveCache(latestVersion, releaseURL string) error {
	cachePath, err := GetCachePath()
	if err != nil {
		return err
	}

	// Ensure cache directory exists
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := VersionCache{
		Timestamp:     time.Now(),
		LatestVersion: latestVersion,
		ReleaseURL:    releaseURL,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// FetchLatestFromGitHub fetches the latest release information from GitHub API
func FetchLatestFromGitHub() (*GitHubRelease, error) {
	client := &http.Client{
		Timeout: httpTimeout,
	}

	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// GitHub API requires User-Agent header
	req.Header.Set("User-Agent", "stacktodate-cli")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching from GitHub: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &release, nil
}

// CompareVersions compares two semantic versions and returns true if latest is newer
// Handles 'dev' versions (always considered older) and different formats
func CompareVersions(current, latest string) (bool, error) {
	// Special case: dev version is always older
	if current == "dev" {
		return true, nil
	}

	// Strip 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Split versions by '.'
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// Compare each numeric part
	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	for i := 0; i < maxLen; i++ {
		currPart := 0
		latPart := 0

		if i < len(currentParts) {
			val, err := strconv.Atoi(strings.TrimSpace(currentParts[i]))
			if err != nil {
				return false, fmt.Errorf("invalid current version format: %s", current)
			}
			currPart = val
		}

		if i < len(latestParts) {
			val, err := strconv.Atoi(strings.TrimSpace(latestParts[i]))
			if err != nil {
				return false, fmt.Errorf("invalid latest version format: %s", latest)
			}
			latPart = val
		}

		if latPart > currPart {
			return true, nil // Newer version available
		} else if currPart > latPart {
			return false, nil // Current is newer
		}
	}

	// All parts are equal
	return false, nil
}

// GetLatestVersion retrieves the latest version, checking cache first and fetching from GitHub if needed
// Implements graceful degradation: uses stale cache if network fails
func GetLatestVersion() (string, string, error) {
	// Check if cache is still valid
	if IsCacheValid() {
		cache, err := LoadCache()
		if err == nil && cache != nil {
			return cache.LatestVersion, cache.ReleaseURL, nil
		}
	}

	// Cache is invalid or missing, fetch new data from GitHub
	release, err := FetchLatestFromGitHub()
	if err != nil {
		// If fetch fails, try to use stale cache as fallback
		cache, cacheErr := LoadCache()
		if cacheErr == nil && cache != nil {
			// Stale cache is better than nothing
			return cache.LatestVersion, cache.ReleaseURL, nil
		}
		// Both fetch and fallback failed
		return "", "", fmt.Errorf("failed to fetch version: %w", err)
	}

	// Successfully fetched, save to cache for future use
	if saveErr := SaveCache(release.TagName, release.HTMLURL); saveErr != nil {
		// Cache save failure is not fatal - we still have the fetched data
		// Silently ignore and continue
	}

	return release.TagName, release.HTMLURL, nil
}
