package helpers

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveAbsPath resolves a path to its absolute form
func ResolveAbsPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolving path %s: %w", path, err)
	}
	return absPath, nil
}

// GetConfigDir returns the directory containing the config file
func GetConfigDir(configPath string) (string, error) {
	absPath, err := ResolveAbsPath(configPath)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(absPath)
	if dir == "" {
		dir = "."
	}

	return dir, nil
}

// WithWorkingDir executes the given function in the specified directory,
// then restores the original working directory. Errors from either
// directory change or the function are returned.
func WithWorkingDir(targetDir string, fn func() error) error {
	if targetDir == "." || targetDir == "" {
		// No need to change directory
		return fn()
	}

	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	if err := os.Chdir(targetDir); err != nil {
		return fmt.Errorf("changing to directory %s: %w", targetDir, err)
	}

	defer os.Chdir(originalDir)

	return fn()
}
