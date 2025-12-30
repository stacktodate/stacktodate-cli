package detectors

// Candidate represents a detected value with its source
type Candidate struct {
	Value  string // The detected version/value
	Source string // Where it was found (e.g., ".ruby-version", "Gemfile", "Dockerfile")
}
