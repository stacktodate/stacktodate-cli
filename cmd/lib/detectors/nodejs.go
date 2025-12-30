package detectors

import (
	"os"
	"regexp"
	"strings"
)

// DetectNode checks multiple sources for Node.js version
func DetectNode() []Candidate {
	var candidates []Candidate

	// Check package.json engines.node
	if data, err := os.ReadFile("package.json"); err == nil {
		content := string(data)
		re := regexp.MustCompile(`"node"\s*:\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			candidates = append(candidates, Candidate{
				Value:  matches[1],
				Source: "package.json",
			})
		}
	}

	// Check .nvmrc
	if data, err := os.ReadFile(".nvmrc"); err == nil {
		if version := strings.TrimSpace(string(data)); version != "" {
			candidates = append(candidates, Candidate{
				Value:  version,
				Source: ".nvmrc",
			})
		}
	}

	return candidates
}
