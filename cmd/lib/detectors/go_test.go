package detectors

import (
	"os"
	"testing"
)

func TestDetectGo(t *testing.T) {
	tests := []struct {
		name        string
		goModContent string
		expectFound bool
		expectValue string
	}{
		{
			name:        "go version simple",
			goModContent: "module example.com\n\ngo 1.21",
			expectFound: true,
			expectValue: "1.21",
		},
		{
			name:        "go version with patch",
			goModContent: "module example.com\n\ngo 1.20.5",
			expectFound: true,
			expectValue: "1.20.5",
		},
		{
			name:        "go version with different spacing",
			goModContent: "go   1.19",
			expectFound: true,
			expectValue: "1.19",
		},
		{
			name:        "go version in complex go.mod",
			goModContent: `module github.com/example/project

go 1.18

require (
	github.com/some/dep v1.0.0
)
`,
			expectFound: true,
			expectValue: "1.18",
		},
		{
			name:        "no go version",
			goModContent: "module example.com",
			expectFound: false,
		},
		{
			name:        "empty go.mod",
			goModContent: "",
			expectFound: false,
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

			if tt.goModContent != "" {
				err := os.WriteFile("go.mod", []byte(tt.goModContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write go.mod: %v", err)
				}
			}

			candidates := DetectGo()

			if tt.expectFound {
				if len(candidates) == 0 {
					t.Errorf("Expected to find Go version, got nothing")
				} else if candidates[0].Value != tt.expectValue {
					t.Errorf("Expected %q, got %q", tt.expectValue, candidates[0].Value)
				} else if candidates[0].Source != "go.mod" {
					t.Errorf("Expected source 'go.mod', got %q", candidates[0].Source)
				}
			} else {
				if len(candidates) > 0 {
					t.Errorf("Expected no Go version, got %v", candidates)
				}
			}
		})
	}
}

func TestDetectGoNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	os.Chdir(tmpDir)

	candidates := DetectGo()
	if len(candidates) != 0 {
		t.Errorf("Expected no candidates when go.mod doesn't exist, got %v", candidates)
	}
}
