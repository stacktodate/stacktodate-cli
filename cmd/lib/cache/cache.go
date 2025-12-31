package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ProductsCache struct {
	Timestamp time.Time `json:"timestamp"`
	Products  []Product `json:"products"`
}

type Product struct {
	Key      string    `json:"key"`
	Name     string    `json:"name"`
	Releases []Release `json:"releases"`
}

type Release struct {
	ReleaseCycle string `json:"releaseCycle"`
	ReleaseDate  string `json:"releaseDate"`
	Support      string `json:"support,omitempty"`
	Extended     string `json:"extended,omitempty"`
	EOL          string `json:"eol,omitempty"`
	LTS          bool   `json:"lts"`
}

const cacheFileName = "products-cache.json"
const cacheDirName = ".stacktodate"
const cacheTTL = 24 * time.Hour

// GetCachePath returns the full path to the cache file
func GetCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(home, cacheDirName)
	return filepath.Join(cacheDir, cacheFileName), nil
}

// IsCacheValid checks if cache exists and is less than 24 hours old
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

// LoadCache loads cached products from disk
func LoadCache() (*ProductsCache, error) {
	cachePath, err := GetCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache ProductsCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return &cache, nil
}

// SaveCache saves products to cache file with timestamp
func SaveCache(products []Product) error {
	cachePath, err := GetCachePath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := ProductsCache{
		Timestamp: time.Now(),
		Products:  products,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// FetchAndCache downloads products from stacktodate.club API and caches them
func FetchAndCache() error {
	return FetchAndCacheWithURL(GetAPIURL())
}

// FetchAndCacheWithURL is the internal function that fetches with a specific URL
func FetchAndCacheWithURL(apiURL string) error {
	url := fmt.Sprintf("%s/api/v1/products", apiURL)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var products []Product
	if err := json.Unmarshal(body, &products); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	if err := SaveCache(products); err != nil {
		return err
	}

	return nil
}

// GetAPIURL returns the API URL from environment or default
func GetAPIURL() string {
	apiURL := os.Getenv("STD_API_URL")
	if apiURL == "" {
		apiURL = "https://stacktodate.club"
	}
	return apiURL
}

// GetProducts returns cached products, fetching if necessary
// It handles cache expiration and auto-fetches if cache is stale
func GetProducts() ([]Product, error) {
	// Check if cache is valid
	if IsCacheValid() {
		cache, err := LoadCache()
		if err == nil && cache != nil {
			return cache.Products, nil
		}
	}

	// Cache is invalid or missing, fetch new data
	if err := FetchAndCache(); err != nil {
		// If fetch fails, try to return stale cache as fallback
		cache, err2 := LoadCache()
		if err2 == nil && cache != nil {
			return cache.Products, nil
		}
		// Both fetch and fallback failed
		return nil, fmt.Errorf("failed to fetch products and no valid cache available: %w", err)
	}

	// Return the newly cached products
	cache, err := LoadCache()
	if err != nil {
		return nil, err
	}

	return cache.Products, nil
}

// GetProductByKey finds a product in the cache by its key
func GetProductByKey(key string, products []Product) *Product {
	for i := range products {
		if products[i].Key == key {
			return &products[i]
		}
	}
	return nil
}

