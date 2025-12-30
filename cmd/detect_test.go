package cmd

import (
	"testing"
)

func TestCleanVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "version with ~> operator",
			input:    "~> 7.1.0",
			expected: "7.1.0",
		},
		{
			name:     "version with >= operator",
			input:    ">= 18.0.0",
			expected: "18.0.0",
		},
		{
			name:     "version with <= operator",
			input:    "<= 3.11",
			expected: "3.11",
		},
		{
			name:     "version with ^ operator",
			input:    "^ 16.0.0",
			expected: "16.0.0",
		},
		{
			name:     "version with ~ operator",
			input:    "~ 2.7.0",
			expected: "2.7.0",
		},
		{
			name:     "version with > operator",
			input:    "> 5.0.0",
			expected: "5.0.0",
		},
		{
			name:     "version with < operator",
			input:    "< 4.0.0",
			expected: "4.0.0",
		},
		{
			name:     "version with = operator",
			input:    "= 3.9.0",
			expected: "3.9.0",
		},
		{
			name:     "plain version",
			input:    "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "version with whitespace",
			input:    "  1.2.3  ",
			expected: "1.2.3",
		},
		{
			name:     "version with multiple operators",
			input:    ">= 2.0.0",
			expected: "2.0.0",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "",
		},
		{
			name:     "version with extra text after numbers",
			input:    "3.11.0-alpine",
			expected: "3.11.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanVersion(tt.input)
			if result != tt.expected {
				t.Errorf("cleanVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractVersionFromDockerImage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ruby with alpine",
			input:    "ruby:3.2.0-alpine",
			expected: "3.2.0",
		},
		{
			name:     "python with slim",
			input:    "python:3.11-slim",
			expected: "3.11",
		},
		{
			name:     "node with alpine",
			input:    "node:18-alpine",
			expected: "18",
		},
		{
			name:     "golang with full version",
			input:    "golang:1.20.5-alpine",
			expected: "1.20.5",
		},
		{
			name:     "postgres with bullseye",
			input:    "postgres:15-bullseye",
			expected: "15",
		},
		{
			name:     "image without tag",
			input:    "ubuntu",
			expected: "ubuntu",
		},
		{
			name:     "image with tag but no suffix",
			input:    "redis:7.0",
			expected: "7.0",
		},
		{
			name:     "registry image",
			input:    "myregistry.azurecr.io/myapp:v1.0",
			expected: "v1.0",
		},
		{
			name:     "complex registry image",
			input:    "docker.io/library/nginx:1.24.0-alpine",
			expected: "1.24.0",
		},
		{
			name:     "version with multiple tags",
			input:    "python:3.10-slim-bullseye",
			expected: "3.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersionFromDockerImage(tt.input)
			if result != tt.expected {
				t.Errorf("extractVersionFromDockerImage(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanCandidateVersions(t *testing.T) {
	tests := []struct {
		name     string
		input    []Candidate
		expected []Candidate
	}{
		{
			name: "multiple versions with operators",
			input: []Candidate{
				{Value: "~> 7.1.0", Source: "Gemfile"},
				{Value: ">= 18.0.0", Source: "package.json"},
				{Value: "3.11.0", Source: ".python-version"},
			},
			expected: []Candidate{
				{Value: "7.1.0", Source: "Gemfile"},
				{Value: "18.0.0", Source: "package.json"},
				{Value: "3.11.0", Source: ".python-version"},
			},
		},
		{
			name:     "empty slice",
			input:    []Candidate{},
			expected: []Candidate{},
		},
		{
			name: "single version",
			input: []Candidate{
				{Value: "^ 2.0.0", Source: "test"},
			},
			expected: []Candidate{
				{Value: "2.0.0", Source: "test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanCandidateVersions(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("cleanCandidateVersions() returned %d items, want %d", len(result), len(tt.expected))
			}
			for i, candidate := range result {
				if candidate.Value != tt.expected[i].Value || candidate.Source != tt.expected[i].Source {
					t.Errorf("cleanCandidateVersions()[%d] = %v, want %v", i, candidate, tt.expected[i])
				}
			}
		})
	}
}
