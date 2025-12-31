package cmd

import (
	"testing"
)

func TestNormalizeDetectedToStack(t *testing.T) {
	info := DetectedInfo{
		Ruby: []Candidate{
			{Value: "3.2.0", Source: ".ruby-version"},
		},
		Rails: []Candidate{
			{Value: "7.0.0", Source: "Gemfile"},
		},
		Node: []Candidate{
			{Value: "18.0.0", Source: ".nvmrc"},
		},
		Go: []Candidate{
			{Value: "1.21", Source: "go.mod"},
		},
		Python: []Candidate{
			{Value: "3.11", Source: ".python-version"},
		},
	}

	result := normalizeDetectedToStack(info)

	tests := []struct {
		tech    string
		version string
		source  string
	}{
		{"ruby", "3.2.0", ".ruby-version"},
		{"rails", "7.0.0", "Gemfile"},
		{"nodejs", "18.0.0", ".nvmrc"},
		{"go", "1.21", "go.mod"},
		{"python", "3.11", ".python-version"},
	}

	for _, tt := range tests {
		entry, exists := result[tt.tech]
		if !exists {
			t.Errorf("normalizeDetectedToStack: expected %s to exist in result", tt.tech)
			continue
		}

		if entry.Version != tt.version {
			t.Errorf("normalizeDetectedToStack for %s: expected version %s, got %s", tt.tech, tt.version, entry.Version)
		}

		if entry.Source != tt.source {
			t.Errorf("normalizeDetectedToStack for %s: expected source %s, got %s", tt.tech, tt.source, entry.Source)
		}
	}
}

func TestCompareStacks_AllMatch(t *testing.T) {
	configStack := map[string]StackEntry{
		"ruby":   {Version: "3.2.0", Source: ".ruby-version"},
		"nodejs": {Version: "18.0.0", Source: ".nvmrc"},
	}

	detectedStack := map[string]StackEntry{
		"ruby":   {Version: "3.2.0", Source: ".ruby-version"},
		"nodejs": {Version: "18.0.0", Source: ".nvmrc"},
	}

	result := compareStacks(configStack, detectedStack)

	if result.Status != "match" {
		t.Errorf("compareStacks: expected status 'match', got %s", result.Status)
	}

	if result.Summary.Matches != 2 {
		t.Errorf("compareStacks: expected 2 matches, got %d", result.Summary.Matches)
	}

	if result.Summary.Mismatches != 0 {
		t.Errorf("compareStacks: expected 0 mismatches, got %d", result.Summary.Mismatches)
	}
}

func TestCompareStacks_WithMismatch(t *testing.T) {
	configStack := map[string]StackEntry{
		"ruby":   {Version: "3.2.0", Source: ".ruby-version"},
		"rails":  {Version: "7.1.0", Source: "Gemfile"},
	}

	detectedStack := map[string]StackEntry{
		"ruby":   {Version: "3.2.0", Source: ".ruby-version"},
		"rails":  {Version: "7.0.0", Source: "Gemfile"},
	}

	result := compareStacks(configStack, detectedStack)

	if result.Status != "mismatch" {
		t.Errorf("compareStacks: expected status 'mismatch', got %s", result.Status)
	}

	if result.Summary.Matches != 1 {
		t.Errorf("compareStacks: expected 1 match, got %d", result.Summary.Matches)
	}

	if result.Summary.Mismatches != 1 {
		t.Errorf("compareStacks: expected 1 mismatch, got %d", result.Summary.Mismatches)
	}
}

func TestCompareStacks_MissingConfig(t *testing.T) {
	configStack := map[string]StackEntry{
		"ruby":   {Version: "3.2.0", Source: ".ruby-version"},
		"nodejs": {Version: "18.0.0", Source: ".nvmrc"},
	}

	detectedStack := map[string]StackEntry{
		"ruby": {Version: "3.2.0", Source: ".ruby-version"},
	}

	result := compareStacks(configStack, detectedStack)

	if result.Status != "mismatch" {
		t.Errorf("compareStacks: expected status 'mismatch', got %s", result.Status)
	}

	if result.Summary.MissingConfig != 1 {
		t.Errorf("compareStacks: expected 1 missing config, got %d", result.Summary.MissingConfig)
	}

	if len(result.Results.MissingConfig) != 1 {
		t.Errorf("compareStacks: expected 1 item in MissingConfig, got %d", len(result.Results.MissingConfig))
	}
}

func TestCompareStacks_Complex(t *testing.T) {
	configStack := map[string]StackEntry{
		"ruby":   {Version: "3.2.0", Source: ".ruby-version"},
		"rails":  {Version: "7.1.0", Source: "Gemfile"},
		"nodejs": {Version: "18.0.0", Source: ".nvmrc"},
		"python": {Version: "3.10", Source: ".python-version"},
	}

	detectedStack := map[string]StackEntry{
		"ruby":   {Version: "3.2.0", Source: ".ruby-version"},
		"rails":  {Version: "7.0.0", Source: "Gemfile"},
		"nodejs": {Version: "18.0.0", Source: ".nvmrc"},
		"go":     {Version: "1.21", Source: "go.mod"},
	}

	result := compareStacks(configStack, detectedStack)

	if result.Status != "mismatch" {
		t.Errorf("compareStacks: expected status 'mismatch', got %s", result.Status)
	}

	if result.Summary.Matches != 2 {
		t.Errorf("compareStacks: expected 2 matches, got %d", result.Summary.Matches)
	}

	if result.Summary.Mismatches != 1 {
		t.Errorf("compareStacks: expected 1 mismatch, got %d", result.Summary.Mismatches)
	}

	if result.Summary.MissingConfig != 1 {
		t.Errorf("compareStacks: expected 1 missing config, got %d", result.Summary.MissingConfig)
	}
}

func TestCompareStacks_Empty(t *testing.T) {
	configStack := map[string]StackEntry{}
	detectedStack := map[string]StackEntry{}

	result := compareStacks(configStack, detectedStack)

	if result.Status != "match" {
		t.Errorf("compareStacks: expected status 'match', got %s", result.Status)
	}

	if result.Summary.Matches != 0 {
		t.Errorf("compareStacks: expected 0 matches, got %d", result.Summary.Matches)
	}
}
