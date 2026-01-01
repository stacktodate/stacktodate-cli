package installer

import (
	"testing"
)

func TestInstallMethodString(t *testing.T) {
	tests := []struct {
		method   InstallMethod
		expected string
	}{
		{Homebrew, "homebrew"},
		{Binary, "binary"},
		{Unknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.method.String(); got != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestIsHomebrewPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Intel Mac Cellar", "/usr/local/Cellar/stacktodate/0.2.0/bin/stacktodate", true},
		{"Apple Silicon", "/opt/homebrew/Cellar/stacktodate/0.2.0/bin/stacktodate", true},
		{"Apple Silicon bin", "/opt/homebrew/bin/stacktodate", true},
		{"Standard usr local bin", "/usr/local/bin/stacktodate", true},
		{"Binary download", "/Users/username/Downloads/stacktodate", false},
		{"Build from source", "/Users/username/projects/stacktodate-cli/stacktodate", false},
		{"Go workspace", "/home/user/go/bin/stacktodate", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHomebrewPath(tt.path); got != tt.expected {
				t.Fatalf("expected %v, got %v for path %s", tt.expected, got, tt.path)
			}
		})
	}
}

func TestGetUpgradeInstructions(t *testing.T) {
	tests := []struct {
		name     string
		method   InstallMethod
		version  string
		expected string
	}{
		{"Homebrew", Homebrew, "v0.3.0", "Upgrade: brew upgrade stacktodate"},
		{"Homebrew without v", Homebrew, "0.3.0", "Upgrade: brew upgrade stacktodate"},
		{"Binary with v", Binary, "v0.3.0", "Download: https://github.com/stacktodate/stacktodate-cli/releases/tag/v0.3.0"},
		{"Binary without v", Binary, "0.3.0", "Download: https://github.com/stacktodate/stacktodate-cli/releases/tag/0.3.0"},
		{"Unknown", Unknown, "v0.3.0", "Visit: https://github.com/stacktodate/stacktodate-cli/releases/tag/v0.3.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUpgradeInstructions(tt.method, tt.version); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestGetInstallerDownloadURL(t *testing.T) {
	tests := []struct {
		name     string
		method   InstallMethod
		expected string
	}{
		{"Homebrew", Homebrew, "https://github.com/stacktodate/homebrew-stacktodate"},
		{"Binary", Binary, "https://github.com/stacktodate/stacktodate-cli/releases/latest"},
		{"Unknown", Unknown, "https://github.com/stacktodate/stacktodate-cli/releases/latest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInstallerDownloadURL(tt.method); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestDetectInstallMethod(t *testing.T) {
	// This test just verifies the function runs without panic
	// Actual detection result depends on environment
	method := DetectInstallMethod()

	if method != Homebrew && method != Binary && method != Unknown {
		t.Fatalf("unexpected install method: %v", method)
	}
}
