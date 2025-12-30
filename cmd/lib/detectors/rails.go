package detectors

import (
	"os"
	"regexp"
)

// DetectRails checks multiple sources for Rails
func DetectRails() []Candidate {
	var candidates []Candidate

	// Check Gemfile
	if data, err := os.ReadFile("Gemfile"); err == nil {
		content := string(data)
		re := regexp.MustCompile(`gem ['"]rails['"],\s*['"]([^'"]+)['"]`)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			candidates = append(candidates, Candidate{
				Value:  matches[1],
				Source: "Gemfile",
			})
		}
	}

	return candidates
}
