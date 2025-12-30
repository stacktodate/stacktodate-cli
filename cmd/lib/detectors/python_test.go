package detectors

import (
	"os"
	"testing"
)

func TestDetectPython(t *testing.T) {
	tests := []struct {
		name               string
		pythonVersionFile  string
		pyprojectContent   string
		pipfileContent     string
		expectFound        bool
		expectValues       []string
		expectSources      []string
	}{
		{
			name:              "python version from .python-version",
			pythonVersionFile: "3.11.0",
			expectFound:       true,
			expectValues:      []string{"3.11.0"},
			expectSources:     []string{".python-version"},
		},
		{
			name:             "python version from pyproject.toml",
			pyprojectContent: `[tool.poetry]
python = "3.10.0"
`,
			expectFound:   true,
			expectValues:  []string{"3.10.0"},
			expectSources: []string{"pyproject.toml"},
		},
		{
			name:           "python version from Pipfile",
			pipfileContent: `[requires]
python_version = "3.9.0"
`,
			expectFound:   true,
			expectValues:  []string{"3.9.0"},
			expectSources: []string{"Pipfile"},
		},
		{
			name:              "python version from all sources",
			pythonVersionFile: "3.11.0",
			pyprojectContent: `[tool.poetry]
python = "3.10.0"
`,
			pipfileContent: `[requires]
python_version = "3.9.0"
`,
			expectFound:   true,
			expectValues:  []string{"3.11.0", "3.10.0", "3.9.0"},
			expectSources: []string{".python-version", "pyproject.toml", "Pipfile"},
		},
		{
			name:              "python version with whitespace",
			pythonVersionFile: "  3.8.0  \n",
			expectFound:       true,
			expectValues:      []string{"3.8.0"},
			expectSources:     []string{".python-version"},
		},
		{
			name:             "no python version",
			pyprojectContent: "[tool.poetry]\nname = 'test'",
			expectFound:      false,
		},
		{
			name:             "python version with double quotes in pyproject",
			pyprojectContent: `[tool.poetry]
python = "3.7.0"
`,
			expectFound:   true,
			expectValues:  []string{"3.7.0"},
			expectSources: []string{"pyproject.toml"},
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

			if tt.pythonVersionFile != "" {
				err := os.WriteFile(".python-version", []byte(tt.pythonVersionFile), 0644)
				if err != nil {
					t.Fatalf("Failed to write .python-version: %v", err)
				}
			}

			if tt.pyprojectContent != "" {
				err := os.WriteFile("pyproject.toml", []byte(tt.pyprojectContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write pyproject.toml: %v", err)
				}
			}

			if tt.pipfileContent != "" {
				err := os.WriteFile("Pipfile", []byte(tt.pipfileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write Pipfile: %v", err)
				}
			}

			candidates := DetectPython()

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
					t.Errorf("Expected no Python, got %v", candidates)
				}
			}
		})
	}
}

func TestDetectPythonNoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	os.Chdir(tmpDir)

	candidates := DetectPython()
	if len(candidates) != 0 {
		t.Errorf("Expected no candidates when files don't exist, got %v", candidates)
	}
}
