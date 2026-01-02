package helpers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const (
	serviceName = "stacktodate"
	username    = "token"
)

// CredentialSource indicates where a credential came from
type CredentialSource string

const (
	SourceEnvVar CredentialSource = "environment variable"
	SourceKeyring CredentialSource = "OS keychain"
	SourceFile   CredentialSource = "config file"
)

// CredentialInfo contains information about stored credentials
type CredentialInfo struct {
	Token  string
	Source CredentialSource
}

// credentialsFile represents the structure of the credentials YAML file
type credentialsFile struct {
	Token string `yaml:"token"`
}

// GetToken retrieves the API token using the priority order:
// 1. STD_TOKEN environment variable (highest priority)
// 2. OS Keychain (macOS/Linux/Windows)
// 3. Returns error if not found (Option B - fail securely)
func GetToken() (string, error) {
	// Check environment variable first
	if token := os.Getenv("STD_TOKEN"); token != "" {
		return token, nil
	}

	// Try to get from keychain
	token, err := keyring.Get(serviceName, username)
	if err == nil && token != "" {
		return token, nil
	}

	// Try to get from fallback file (for migration purposes, but don't use it by default)
	if token, err := getTokenFromFile(); err == nil && token != "" {
		return token, nil
	}

	// No token found anywhere
	return "", fmt.Errorf("no authentication token found\n\nSetup your token with one of these methods:\n  1. Interactive setup: stacktodate global-config set\n  2. Environment variable: export STD_TOKEN=<your_token>\n\nFor more help: stacktodate global-config --help")
}

// GetTokenWithSource retrieves the token and returns information about its source
func GetTokenWithSource() (*CredentialInfo, error) {
	// Check environment variable first
	if token := os.Getenv("STD_TOKEN"); token != "" {
		return &CredentialInfo{
			Token:  token,
			Source: SourceEnvVar,
		}, nil
	}

	// Try to get from keychain
	token, err := keyring.Get(serviceName, username)
	if err == nil && token != "" {
		return &CredentialInfo{
			Token:  token,
			Source: SourceKeyring,
		}, nil
	}

	// Try to get from fallback file
	if token, err := getTokenFromFile(); err == nil && token != "" {
		return &CredentialInfo{
			Token:  token,
			Source: SourceFile,
		}, nil
	}

	// No token found anywhere
	return nil, fmt.Errorf("no authentication token found")
}

// SetToken stores the token in the OS keychain
// Falls back to file storage if keychain is unavailable
// Per Option B: Fails if keychain is unavailable and no fallback
func SetToken(token string) error {
	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Try to store in keychain first
	err := keyring.Set(serviceName, username, token)
	if err == nil {
		return nil
	}

	// If keychain fails, also try file storage as a fallback
	// This allows local development to work
	if err := setTokenInFile(token); err != nil {
		return fmt.Errorf("failed to store token securely:\n  Keychain error: %v\n  File storage error: %v\n\nFor CI/headless environments, use: export STD_TOKEN=<your_token>", err, err)
	}

	fmt.Println("⚠️  Warning: Token stored in plain text file at ~/.stacktodate/credentials.yaml")
	fmt.Println("For better security, consider using a system with OS keychain support")
	return nil
}

// DeleteToken removes the token from keychain and file storage
func DeleteToken() error {
	var keychainErr error
	var fileErr error

	// Try to delete from keychain
	keychainErr = keyring.Delete(serviceName, username)

	// Try to delete from file
	fileErr = deleteTokenFromFile()

	// If both failed, return error
	if keychainErr != nil && fileErr != nil {
		return fmt.Errorf("failed to delete token: keychain error: %v, file error: %v", keychainErr, fileErr)
	}

	return nil
}

// GetTokenSource returns information about where the token is currently stored
func GetTokenSource() (string, bool, error) {
	// Check environment variable
	if os.Getenv("STD_TOKEN") != "" {
		return "STD_TOKEN environment variable", true, nil
	}

	// Check keychain
	_, err := keyring.Get(serviceName, username)
	if err == nil {
		return "OS keychain", true, nil
	}

	// Check file
	if _, err := getTokenFromFile(); err == nil {
		return "credentials file (~/.stacktodate/credentials.yaml)", false, nil
	}

	return "not configured", false, fmt.Errorf("no token found")
}

// EnsureConfigDir creates the ~/.stacktodate directory if it doesn't exist
func EnsureConfigDir() error {
	configDir := getConfigDir()
	return os.MkdirAll(configDir, 0700)
}

// Helper functions

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home can't be determined
		return ".stacktodate"
	}
	return filepath.Join(home, ".stacktodate")
}

func getCredentialsFilePath() string {
	return filepath.Join(getConfigDir(), "credentials.yaml")
}

func getTokenFromFile() (string, error) {
	filePath := getCredentialsFilePath()

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds credentialsFile
	if err := yaml.Unmarshal(content, &creds); err != nil {
		return "", fmt.Errorf("failed to parse credentials file: %w", err)
	}

	return creds.Token, nil
}

func setTokenInFile(token string) error {
	filePath := getCredentialsFilePath()

	creds := credentialsFile{
		Token: token,
	}

	content, err := yaml.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Write with restricted permissions (0600 = read/write for owner only)
	if err := os.WriteFile(filePath, content, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

func deleteTokenFromFile() error {
	filePath := getCredentialsFilePath()
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete credentials file: %w", err)
	}
	return nil
}
