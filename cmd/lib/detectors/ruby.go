package detectors

import (
	"os"
	"strings"
)

// DetectRubyVersion checks multiple sources for Ruby version
func DetectRubyVersion() []Candidate {
	var candidates []Candidate

	// Check .ruby-version
	if data, err := os.ReadFile(".ruby-version"); err == nil {
		if version := strings.TrimSpace(string(data)); version != "" {
			candidates = append(candidates, Candidate{
				Value:  version,
				Source: ".ruby-version",
			})
		}
	}

	return candidates
}
