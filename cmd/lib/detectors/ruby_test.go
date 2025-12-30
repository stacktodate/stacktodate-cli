package detectors

import (
	"os"
	"testing"
)

func TestDetectRubyVersion(t *testing.T) {
	tests := []struct {
		name      string
		fileContent string
		expectFound bool
		expectValue string
	}{
		{
			name:        "ruby version detected",
			fileContent: "3.2.0",
			expectFound: true,
			expectValue: "3.2.0",
		},
		{
			name:        "ruby version with newlines",
			fileContent: "3.1.0\n",
			expectFound: true,
			expectValue: "3.1.0",
		},
		{
			name:        "ruby version with spaces",
			fileContent: "  2.7.0  ",
			expectFound: true,
			expectValue: "2.7.0",
		},
		{
			name:        "empty file",
			fileContent: "",
			expectFound: false,
		},
		{
			name:        "whitespace only",
			fileContent: "   \n  ",
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

			if tt.fileContent != "" || !tt.expectFound {
				err := os.WriteFile(".ruby-version", []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			}

			candidates := DetectRubyVersion()

			if tt.expectFound {
				if len(candidates) == 0 {
					t.Errorf("Expected to find Ruby version, got nothing")
				} else if candidates[0].Value != tt.expectValue {
					t.Errorf("Expected %q, got %q", tt.expectValue, candidates[0].Value)
				} else if candidates[0].Source != ".ruby-version" {
					t.Errorf("Expected source '.ruby-version', got %q", candidates[0].Source)
				}
			} else {
				if len(candidates) > 0 {
					t.Errorf("Expected no Ruby version, got %v", candidates)
				}
			}
		})
	}
}

func TestDetectRubyVersionNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	os.Chdir(tmpDir)

	candidates := DetectRubyVersion()
	if len(candidates) != 0 {
		t.Errorf("Expected no candidates when .ruby-version doesn't exist, got %v", candidates)
	}
}
