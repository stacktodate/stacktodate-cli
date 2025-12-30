package detectors

import (
	"os"
	"regexp"
)

// DetectGo checks multiple sources for Go version
func DetectGo() []Candidate {
	var candidates []Candidate

	// Check go.mod
	if data, err := os.ReadFile("go.mod"); err == nil {
		content := string(data)
		re := regexp.MustCompile(`go\s+(\d+\.\d+(?:\.\d+)?)`)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			candidates = append(candidates, Candidate{
				Value:  matches[1],
				Source: "go.mod",
			})
		}
	}

	return candidates
}
