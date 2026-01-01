package installer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InstallMethod represents how stacktodate was installed
type InstallMethod int

const (
	Unknown InstallMethod = iota
	Homebrew
	Binary
)

// String returns a string representation of the install method
func (m InstallMethod) String() string {
	switch m {
	case Homebrew:
		return "homebrew"
	case Binary:
		return "binary"
	default:
		return "unknown"
	}
}

// DetectInstallMethod attempts to determine how stacktodate was installed
func DetectInstallMethod() InstallMethod {
	// Try to detect Homebrew installation first
	if IsHomebrew() {
		return Homebrew
	}

	// Default to binary download
	return Binary
}

// IsHomebrew checks if stacktodate was installed via Homebrew
func IsHomebrew() bool {
	// Method 1: Check executable path for Homebrew-specific directories
	executable, err := os.Executable()
	if err == nil {
		if isHomebrewPath(executable) {
			return true
		}
	}

	// Method 2: Verify with brew command (silent check)
	if isBrewInstalled() {
		return true
	}

	return false
}

// isHomebrewPath checks if the executable path looks like a Homebrew installation
func isHomebrewPath(execPath string) bool {
	// Common Homebrew paths
	homebrewPatterns := []string{
		"/Cellar/stacktodate/",           // Intel Macs, Linux
		"/opt/homebrew/Cellar/stacktodate", // Apple Silicon Macs
		"/opt/homebrew/bin/stacktodate",
		"/usr/local/bin/stacktodate",
		"/usr/local/Cellar/stacktodate/",
	}

	for _, pattern := range homebrewPatterns {
		if strings.Contains(execPath, pattern) {
			return true
		}
	}

	return false
}

// isBrewInstalled checks if the brew command recognizes stacktodate
func isBrewInstalled() bool {
	// Run: brew list stacktodate
	// This will succeed (exit code 0) if stacktodate is installed via Homebrew
	cmd := exec.Command("brew", "list", "stacktodate")

	// Redirect output to /dev/null (we don't need the output)
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Silent execution - we only care about the exit code
	return cmd.Run() == nil
}

// GetUpgradeInstructions returns the appropriate upgrade instructions based on install method
func GetUpgradeInstructions(method InstallMethod, version string) string {
	switch method {
	case Homebrew:
		return "Upgrade: brew upgrade stacktodate"

	case Binary:
		return fmt.Sprintf("Download: https://github.com/stacktodate/stacktodate-cli/releases/tag/%s", version)

	default:
		return fmt.Sprintf("Visit: https://github.com/stacktodate/stacktodate-cli/releases/tag/%s", version)
	}
}

// GetInstallerDownloadURL returns the download URL appropriate for the install method
func GetInstallerDownloadURL(method InstallMethod) string {
	switch method {
	case Homebrew:
		return "https://github.com/stacktodate/homebrew-stacktodate"

	default:
		return "https://github.com/stacktodate/stacktodate-cli/releases/latest"
	}
}
