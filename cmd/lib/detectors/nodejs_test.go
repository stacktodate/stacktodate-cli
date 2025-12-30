package detectors

import (
	"os"
	"testing"
)

func TestDetectNode(t *testing.T) {
	tests := []struct {
		name             string
		packageJsonContent string
		nvmrcContent     string
		expectFound      bool
		expectValues     []string
		expectSources    []string
	}{
		{
			name:             "node version from package.json",
			packageJsonContent: `{"engines": {"node": "18.0.0"}}`,
			expectFound:      true,
			expectValues:     []string{"18.0.0"},
			expectSources:    []string{"package.json"},
		},
		{
			name:             "node version from .nvmrc",
			nvmrcContent:     "16.0.0",
			expectFound:      true,
			expectValues:     []string{"16.0.0"},
			expectSources:    []string{".nvmrc"},
		},
		{
			name:             "node version from both sources",
			packageJsonContent: `{"engines": {"node": "18.0.0"}}`,
			nvmrcContent:     "16.0.0",
			expectFound:      true,
			expectValues:     []string{"18.0.0", "16.0.0"},
			expectSources:    []string{"package.json", ".nvmrc"},
		},
		{
			name:             "node version with whitespace in .nvmrc",
			nvmrcContent:     "  14.0.0  \n",
			expectFound:      true,
			expectValues:     []string{"14.0.0"},
			expectSources:    []string{".nvmrc"},
		},
		{
			name:             "no node version",
			packageJsonContent: `{"name": "test"}`,
			expectFound:      false,
		},
		{
			name:             "node in package.json with complex structure",
			packageJsonContent: `{
  "name": "my-app",
  "version": "1.0.0",
  "engines": {
    "node": "20.0.0",
    "npm": "9.0.0"
  }
}`,
			expectFound:   true,
			expectValues:  []string{"20.0.0"},
			expectSources: []string{"package.json"},
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

			if tt.packageJsonContent != "" {
				err := os.WriteFile("package.json", []byte(tt.packageJsonContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write package.json: %v", err)
				}
			}

			if tt.nvmrcContent != "" {
				err := os.WriteFile(".nvmrc", []byte(tt.nvmrcContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write .nvmrc: %v", err)
				}
			}

			candidates := DetectNode()

			if tt.expectFound {
				if len(candidates) != len(tt.expectValues) {
					t.Errorf("Expected %d candidates, got %d", len(tt.expectValues), len(candidates))
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
					t.Errorf("Expected no Node.js, got %v", candidates)
				}
			}
		})
	}
}

func TestDetectNodeNoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	os.Chdir(tmpDir)

	candidates := DetectNode()
	if len(candidates) != 0 {
		t.Errorf("Expected no candidates when files don't exist, got %v", candidates)
	}
}
