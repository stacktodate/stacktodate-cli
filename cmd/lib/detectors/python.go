package detectors

import (
	"os"
	"regexp"
	"strings"
)

// DetectPython checks multiple sources for Python version
func DetectPython() []Candidate {
	var candidates []Candidate

	// Check .python-version
	if data, err := os.ReadFile(".python-version"); err == nil {
		if version := strings.TrimSpace(string(data)); version != "" {
			candidates = append(candidates, Candidate{
				Value:  version,
				Source: ".python-version",
			})
		}
	}

	// Check pyproject.toml
	if data, err := os.ReadFile("pyproject.toml"); err == nil {
		content := string(data)
		re := regexp.MustCompile(`python\s*=\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			candidates = append(candidates, Candidate{
				Value:  matches[1],
				Source: "pyproject.toml",
			})
		}
	}

	// Check Pipfile
	if data, err := os.ReadFile("Pipfile"); err == nil {
		content := string(data)
		re := regexp.MustCompile(`python_version\s*=\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			candidates = append(candidates, Candidate{
				Value:  matches[1],
				Source: "Pipfile",
			})
		}
	}

	return candidates
}
