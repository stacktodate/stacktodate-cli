package detectors

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DetectDocker extracts Docker base images from Dockerfiles and docker-compose.yml
func DetectDocker() []Candidate {
	var candidates []Candidate

	// Find all Dockerfiles
	files, err := filepath.Glob("*Dockerfile*")
	if err == nil && len(files) > 0 {
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			content := string(data)

			// Extract FROM statements
			re := regexp.MustCompile(`(?m)^FROM\s+(.+)`)
			matches := re.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				if len(match) > 1 {
					image := strings.Split(match[1], " ")[0]
					candidates = append(candidates, Candidate{
						Value:  image,
						Source: file,
					})
				}
			}
		}
	}

	// Check docker-compose.yml
	if data, err := os.ReadFile("docker-compose.yml"); err == nil {
		content := string(data)

		// Extract image references
		re := regexp.MustCompile(`(?m)^\s*image:\s*(.+)`)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				candidates = append(candidates, Candidate{
					Value:  strings.TrimSpace(match[1]),
					Source: "docker-compose.yml",
				})
			}
		}
	}

	return candidates
}
