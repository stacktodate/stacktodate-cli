package helpers

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the stacktodate.yml structure
type Config struct {
	UUID  string                `yaml:"uuid"`
	Name  string                `yaml:"name"`
	Stack map[string]StackEntry `yaml:"stack,omitempty"`
}

// StackEntry represents a single technology entry in the stack
type StackEntry struct {
	Version string `yaml:"version"`
	Source  string `yaml:"source"`
}

// LoadConfig reads and parses a config file from the given path
// If path is empty, uses "stacktodate.yml" as default
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "stacktodate.yml"
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", configPath, err)
	}

	return &config, nil
}

// LoadConfigWithDefaults loads a config file with optional UUID validation
func LoadConfigWithDefaults(configPath string, requireUUID bool) (*Config, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	if requireUUID && config.UUID == "" {
		return nil, fmt.Errorf("uuid not found in config file")
	}

	return config, nil
}
