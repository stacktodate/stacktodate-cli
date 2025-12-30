package detectors

import (
	"os"
	"testing"
)

func TestDetectDocker(t *testing.T) {
	tests := []struct {
		name              string
		dockerfileContent string
		composeContent    string
		expectFound       bool
		expectValues      []string
		expectSources     []string
	}{
		{
			name:              "single Dockerfile with FROM",
			dockerfileContent: "FROM ruby:3.2.0-alpine",
			expectFound:       true,
			expectValues:      []string{"ruby:3.2.0-alpine"},
			expectSources:     []string{"Dockerfile"},
		},
		{
			name:              "Dockerfile with multiple FROM (multi-stage)",
			dockerfileContent: `FROM node:18-alpine as builder
RUN npm install
FROM node:18-alpine
COPY --from=builder .`,
			expectFound:   true,
			expectValues:  []string{"node:18-alpine", "node:18-alpine"},
			expectSources: []string{"Dockerfile", "Dockerfile"},
		},
		{
			name:           "docker-compose.yml with image",
			composeContent: `version: '3'
services:
  web:
    image: python:3.11-slim
  db:
    image: postgres:15
`,
			expectFound:   true,
			expectValues:  []string{"python:3.11-slim", "postgres:15"},
			expectSources: []string{"docker-compose.yml", "docker-compose.yml"},
		},
		{
			name:              "Dockerfile and docker-compose",
			dockerfileContent: "FROM ruby:3.1.0",
			composeContent: `version: '3'
services:
  web:
    image: golang:1.20`,
			expectFound:   true,
			expectValues:  []string{"ruby:3.1.0", "golang:1.20"},
			expectSources: []string{"Dockerfile", "docker-compose.yml"},
		},
		{
			name:              "Dockerfile.prod variant",
			dockerfileContent: "FROM alpine:3.18",
			composeContent:    "",
			expectFound:       true,
			expectValues:      []string{"alpine:3.18"},
			expectSources:     []string{"Dockerfile"},
		},
		{
			name:              "empty Dockerfile",
			dockerfileContent: "",
			expectFound:       false,
		},
		{
			name:              "Dockerfile without FROM",
			dockerfileContent: "RUN echo 'test'",
			expectFound:       false,
		},
		{
			name:           "docker-compose with tag",
			composeContent: `services:
  app:
    image: myregistry.azurecr.io/myapp:v1.0`,
			expectFound:   true,
			expectValues:  []string{"myregistry.azurecr.io/myapp:v1.0"},
			expectSources: []string{"docker-compose.yml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}
			defer os.Chdir(oldWd)

			os.Chdir(tmpDir)

			if tt.dockerfileContent != "" {
				err := os.WriteFile("Dockerfile", []byte(tt.dockerfileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write Dockerfile: %v", err)
				}
			}

			if tt.composeContent != "" {
				err := os.WriteFile("docker-compose.yml", []byte(tt.composeContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write docker-compose.yml: %v", err)
				}
			}

			candidates := DetectDocker()

			if tt.expectFound {
				if len(candidates) != len(tt.expectValues) {
					t.Errorf("Expected %d candidates, got %d: %v", len(tt.expectValues), len(candidates), candidates)
				}
				for i, expected := range tt.expectValues {
					if i >= len(candidates) {
						break
					}
					if candidates[i].Value != expected {
						t.Errorf("Expected value %q at index %d, got %q", expected, i, candidates[i].Value)
					}
					if candidates[i].Source != tt.expectSources[i] {
						t.Errorf("Expected source %q at index %d, got %q", tt.expectSources[i], i, candidates[i].Source)
					}
				}
			} else {
				if len(candidates) > 0 {
					t.Errorf("Expected no Docker images, got %v", candidates)
				}
			}
		})
	}
}

func TestDetectDockerNoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	os.Chdir(tmpDir)

	candidates := DetectDocker()
	if len(candidates) != 0 {
		t.Errorf("Expected no candidates when Docker files don't exist, got %v", candidates)
	}
}
