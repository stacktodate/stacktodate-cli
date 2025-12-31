package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/stacktodate/stacktodate-cli/cmd/lib/cache"
	"github.com/stacktodate/stacktodate-cli/cmd/lib/detectors"
)

// Candidate alias for easier access
type Candidate = detectors.Candidate

type DetectedInfo struct {
	Ruby   []detectors.Candidate
	Rails  []detectors.Candidate
	Node   []detectors.Candidate
	Go     []detectors.Candidate
	Python []detectors.Candidate
	Docker []detectors.Candidate
}

// cleanVersion removes version operators and extracts the core version
// Examples: ~> 7.1.0 -> 7.1.0, >= 18.0.0 -> 18.0.0, <= 3.11 -> 3.11
func cleanVersion(version string) string {
	// Trim whitespace
	version = strings.TrimSpace(version)

	// Remove common version operators
	operators := []string{"~>", ">=", "<=", "!=", "==", "^", "~", ">", "<", "="}
	for _, op := range operators {
		if strings.HasPrefix(version, op) {
			version = strings.TrimPrefix(version, op)
			version = strings.TrimSpace(version)
		}
	}

	// Extract only the version number part (in case there are comments or extra text)
	re := regexp.MustCompile(`^([\d.]+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) > 1 {
		return matches[1]
	}

	return version
}

// cleanCandidateVersions applies cleanVersion to all candidates in a slice
func cleanCandidateVersions(candidates []detectors.Candidate) []detectors.Candidate {
	for i := range candidates {
		candidates[i].Value = cleanVersion(candidates[i].Value)
	}
	return candidates
}

// truncateCandidateVersions truncates all candidate versions to match stacktodate.club API cycles
func truncateCandidateVersions(candidates []detectors.Candidate, product string) []detectors.Candidate {
	for i := range candidates {
		candidates[i].Value = truncateVersionToEOLCycle(product, candidates[i].Value)
	}
	return candidates
}

// extractVersionFromDockerImage extracts version from Docker image names
// Examples: ruby:3.2.0-alpine -> 3.2.0, python:3.11-slim -> 3.11, node:18-alpine -> 18
func extractVersionFromDockerImage(image string) string {
	// Find the version part after the colon
	parts := strings.Split(image, ":")
	if len(parts) < 2 {
		return image
	}

	versionPart := parts[1]

	// Remove common suffixes like -alpine, -slim, -bullseye, etc.
	re := regexp.MustCompile(`^([\d.]+)`)
	matches := re.FindStringSubmatch(versionPart)
	if len(matches) > 1 {
		return matches[1]
	}

	return versionPart
}

func DetectProjectInfo() DetectedInfo {
	info := DetectedInfo{
		Ruby:   detectors.DetectRubyVersion(),
		Rails:  detectors.DetectRails(),
		Node:   detectors.DetectNode(),
		Go:     detectors.DetectGo(),
		Python: detectors.DetectPython(),
		Docker: detectors.DetectDocker(),
	}

	// Clean versions for all detected candidates
	info.Ruby = cleanCandidateVersions(info.Ruby)
	info.Rails = cleanCandidateVersions(info.Rails)
	info.Node = cleanCandidateVersions(info.Node)
	info.Go = cleanCandidateVersions(info.Go)
	info.Python = cleanCandidateVersions(info.Python)

	// Aggregate Docker images into technology stacks
	for _, dockerCandidate := range info.Docker {
		image := dockerCandidate.Value
		version := extractVersionFromDockerImage(image)

		// Ruby
		if strings.Contains(image, "ruby:") {
			info.Ruby = append(info.Ruby, detectors.Candidate{
				Value:  version,
				Source: dockerCandidate.Source,
			})
		}
		// Python
		if strings.Contains(image, "python:") {
			info.Python = append(info.Python, detectors.Candidate{
				Value:  version,
				Source: dockerCandidate.Source,
			})
		}
		// Node.js
		if strings.Contains(image, "node:") {
			info.Node = append(info.Node, detectors.Candidate{
				Value:  version,
				Source: dockerCandidate.Source,
			})
		}
		// Go
		if strings.Contains(image, "golang:") || strings.Contains(image, "go:") {
			info.Go = append(info.Go, detectors.Candidate{
				Value:  version,
				Source: dockerCandidate.Source,
			})
		}
	}

	// Truncate versions to match stacktodate.club API cycles
	info.Ruby = truncateCandidateVersions(info.Ruby, "ruby")
	info.Rails = truncateCandidateVersions(info.Rails, "rails")
	info.Node = truncateCandidateVersions(info.Node, "nodejs")
	info.Go = truncateCandidateVersions(info.Go, "go")
	info.Python = truncateCandidateVersions(info.Python, "python")

	return info
}

// mapProductNameToCacheKey maps internal product names to stacktodate.club API keys
func mapProductNameToCacheKey(product string) string {
	mapping := map[string]string{
		"ruby":   "ruby",
		"rails":  "rails",
		"nodejs": "nodejs",
		"go":     "go",
		"python": "python",
	}

	if key, exists := mapping[product]; exists {
		return key
	}
	return product
}

// truncateVersionToEOLCycle truncates a version to match the format used by stacktodate.club API
// It tries to find the best matching cycle by progressively truncating the version
// Examples: 3.11.0 -> 3.11, 18.0.0 -> 18, 7.1.0 -> 7.1
func truncateVersionToEOLCycle(product, version string) string {
	if product == "" || version == "" {
		return version
	}

	// Get products from cache (auto-fetches if needed or stale)
	products, err := cache.GetProducts()
	if err != nil {
		// Graceful fallback: return original version if cache fetch fails
		return version
	}

	// Map product name to cache key
	cacheKey := mapProductNameToCacheKey(product)
	cachedProduct := cache.GetProductByKey(cacheKey, products)
	if cachedProduct == nil {
		// Product not found in cache, return original version
		return version
	}

	// Build a set of available cycles for quick lookup
	cycles := make(map[string]bool)
	for _, release := range cachedProduct.Releases {
		cycles[release.ReleaseCycle] = true
	}

	// If exact match exists, return as-is
	if cycles[version] {
		return version
	}

	// Split version into parts
	parts := strings.Split(version, ".")

	// Try major.minor (e.g., 3.11 from 3.11.0)
	if len(parts) >= 2 {
		majorMinor := parts[0] + "." + parts[1]
		if cycles[majorMinor] {
			return majorMinor
		}
	}

	// Try major only (e.g., 18 from 18.0.0)
	if len(parts) >= 1 {
		major := parts[0]
		if cycles[major] {
			return major
		}
	}

	// If no match found, return original version
	return version
}

func getEOLStatus(product, version string) string {
	if product == "" || version == "" {
		return ""
	}

	// Get products from cache (auto-fetches if needed or stale)
	products, err := cache.GetProducts()
	if err != nil {
		// Graceful fallback: return empty string if cache fetch fails
		return ""
	}

	// Map product name to cache key
	cacheKey := mapProductNameToCacheKey(product)
	cachedProduct := cache.GetProductByKey(cacheKey, products)
	if cachedProduct == nil {
		// Product not found in cache, return empty string
		return ""
	}

	// Find the matching release cycle
	for _, release := range cachedProduct.Releases {
		if release.ReleaseCycle == version {
			// Check if EOL is empty (still supported)
			if release.EOL == "" {
				return " (supported)"
			}
			// Return EOL date
			return fmt.Sprintf(" (EOL: %s)", release.EOL)
		}
	}

	return ""
}

func PrintDetectedInfo(info DetectedInfo) {
	hasCandidates := len(info.Ruby) > 0 || len(info.Rails) > 0 || len(info.Node) > 0 ||
		len(info.Go) > 0 || len(info.Python) > 0 || len(info.Docker) > 0

	if !hasCandidates {
		fmt.Println("\nNo project files detected in current directory")
		return
	}

	fmt.Println("\n=== Detected Project Information ====")

	// Print Ruby candidates
	if len(info.Ruby) > 0 {
		fmt.Println("Ruby:")
		for _, candidate := range info.Ruby {
			fmt.Printf("  - %s (from: %s)\n", candidate.Value, candidate.Source)
		}
		fmt.Println()
	}

	// Print Rails candidates
	if len(info.Rails) > 0 {
		fmt.Println("Rails:")
		for _, candidate := range info.Rails {
			fmt.Printf("  - %s (from: %s)\n", candidate.Value, candidate.Source)
		}
		fmt.Println()
	}

	// Print Node candidates
	if len(info.Node) > 0 {
		fmt.Println("Node.js:")
		for _, candidate := range info.Node {
			fmt.Printf("  - %s (from: %s)\n", candidate.Value, candidate.Source)
		}
		fmt.Println()
	}

	// Print Go candidates
	if len(info.Go) > 0 {
		fmt.Println("Go:")
		for _, candidate := range info.Go {
			fmt.Printf("  - %s (from: %s)\n", candidate.Value, candidate.Source)
		}
		fmt.Println()
	}

	// Print Python candidates
	if len(info.Python) > 0 {
		fmt.Println("Python:")
		for _, candidate := range info.Python {
			fmt.Printf("  - %s (from: %s)\n", candidate.Value, candidate.Source)
		}
		fmt.Println()
	}

	// Print unclassified Docker candidates (those that don't match Ruby/Python/Node/Go)
	unclassifiedDocker := []detectors.Candidate{}
	for _, dockerCandidate := range info.Docker {
		image := dockerCandidate.Value
		if !strings.Contains(image, "ruby:") && !strings.Contains(image, "python:") &&
			!strings.Contains(image, "node:") && !strings.Contains(image, "golang:") &&
			!strings.Contains(image, "go:") {
			unclassifiedDocker = append(unclassifiedDocker, dockerCandidate)
		}
	}

	if len(unclassifiedDocker) > 0 {
		fmt.Println("Docker:")
		for _, candidate := range unclassifiedDocker {
			fmt.Printf("  - %s (from: %s)\n", candidate.Value, candidate.Source)
		}
		fmt.Println()
	}
}
