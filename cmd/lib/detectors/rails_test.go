package detectors

import (
	"os"
	"testing"
)

func TestDetectRails(t *testing.T) {
	tests := []struct {
		name        string
		gemfileContent string
		expectFound bool
		expectValue string
	}{
		{
			name:           "rails with double quotes",
			gemfileContent: `gem "rails", "7.0.0"`,
			expectFound:    true,
			expectValue:    "7.0.0",
		},
		{
			name:           "rails with single quotes",
			gemfileContent: `gem 'rails', '6.1.5'`,
			expectFound:    true,
			expectValue:    "6.1.5",
		},
		{
			name:           "rails in multiline Gemfile",
			gemfileContent: `source "https://rubygems.org"

gem 'rails', '~> 7.0.0'
gem 'sqlite3'
`,
			expectFound: true,
			expectValue: "~> 7.0.0",
		},
		{
			name:           "no rails gem",
			gemfileContent: `gem 'sqlite3'`,
			expectFound:    false,
		},
		{
			name:           "empty Gemfile",
			gemfileContent: "",
			expectFound:    false,
		},
		{
			name:           "rails with different whitespace",
			gemfileContent: `gem "rails", "5.0.0"`,
			expectFound:    true,
			expectValue:    "5.0.0",
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

			if tt.gemfileContent != "" {
				err := os.WriteFile("Gemfile", []byte(tt.gemfileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write Gemfile: %v", err)
				}
			}

			candidates := DetectRails()

			if tt.expectFound {
				if len(candidates) == 0 {
					t.Errorf("Expected to find Rails, got nothing")
				} else if candidates[0].Value != tt.expectValue {
					t.Errorf("Expected %q, got %q", tt.expectValue, candidates[0].Value)
				} else if candidates[0].Source != "Gemfile" {
					t.Errorf("Expected source 'Gemfile', got %q", candidates[0].Source)
				}
			} else {
				if len(candidates) > 0 {
					t.Errorf("Expected no Rails, got %v", candidates)
				}
			}
		})
	}
}

func TestDetectRailsNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	os.Chdir(tmpDir)

	candidates := DetectRails()
	if len(candidates) != 0 {
		t.Errorf("Expected no candidates when Gemfile doesn't exist, got %v", candidates)
	}
}
