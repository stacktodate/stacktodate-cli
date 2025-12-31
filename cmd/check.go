package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type CheckResult struct {
	Status  string                 `json:"status"`
	Summary CheckSummary           `json:"summary"`
	Results CheckResults           `json:"results"`
}

type CheckSummary struct {
	Matches       int `json:"matches"`
	Mismatches    int `json:"mismatches"`
	MissingConfig int `json:"missing_config"`
}

type CheckResults struct {
	Matched       []ComparisonEntry `json:"matched"`
	Mismatched    []ComparisonEntry `json:"mismatched"`
	MissingConfig []ComparisonEntry `json:"missing_config"`
}

type ComparisonEntry struct {
	Name     string `json:"name"`
	Version  string `json:"version,omitempty"`
	Detected string `json:"detected,omitempty"`
	Source   string `json:"source,omitempty"`
}

var (
	checkConfigFile string
	checkFormat     string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if detected versions match stacktodate.yml",
	Long:  `Verify that the versions in stacktodate.yml match the currently detected versions in your project. Useful for CI/CD pipelines.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Determine config file
		configPath := checkConfigFile
		if configPath == "" {
			configPath = "stacktodate.yml"
		}

		// Resolve absolute path
		absConfigPath, err := filepath.Abs(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving config path: %v\n", err)
			os.Exit(2)
		}

		// Read config file
		content, err := os.ReadFile(absConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading config file %s: %v\n", configPath, err)
			os.Exit(2)
		}

		// Parse YAML
		var config Config
		if err := yaml.Unmarshal(content, &config); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", configPath, err)
			os.Exit(2)
		}

		// Change to config directory for detection
		originalDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(2)
		}

		configDir := filepath.Dir(absConfigPath)
		if configDir == "" {
			configDir = "."
		}

		if configDir != "." {
			if err := os.Chdir(configDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error changing to directory: %v\n", err)
				os.Exit(2)
			}
			defer os.Chdir(originalDir)
		}

		// Detect current versions
		detectedInfo := DetectProjectInfo()
		detectedStack := normalizeDetectedToStack(detectedInfo)

		// Compare stacks
		result := compareStacks(config.Stack, detectedStack)

		// Output results
		if checkFormat == "json" {
			outputJSON(result)
		} else {
			outputText(result)
		}

		// Exit with appropriate code
		if result.Status != "match" {
			os.Exit(1)
		}
	},
}

func normalizeDetectedToStack(info DetectedInfo) map[string]StackEntry {
	normalized := make(map[string]StackEntry)

	if len(info.Ruby) > 0 {
		normalized["ruby"] = StackEntry{
			Version: info.Ruby[0].Value,
			Source:  info.Ruby[0].Source,
		}
	}

	if len(info.Rails) > 0 {
		normalized["rails"] = StackEntry{
			Version: info.Rails[0].Value,
			Source:  info.Rails[0].Source,
		}
	}

	if len(info.Node) > 0 {
		normalized["nodejs"] = StackEntry{
			Version: info.Node[0].Value,
			Source:  info.Node[0].Source,
		}
	}

	if len(info.Go) > 0 {
		normalized["go"] = StackEntry{
			Version: info.Go[0].Value,
			Source:  info.Go[0].Source,
		}
	}

	if len(info.Python) > 0 {
		normalized["python"] = StackEntry{
			Version: info.Python[0].Value,
			Source:  info.Python[0].Source,
		}
	}

	return normalized
}

func compareStacks(configStack, detectedStack map[string]StackEntry) CheckResult {
	result := CheckResult{
		Results: CheckResults{
			Matched:       []ComparisonEntry{},
			Mismatched:    []ComparisonEntry{},
			MissingConfig: []ComparisonEntry{},
		},
	}

	// Check all items in config
	for tech, configEntry := range configStack {
		if detectedEntry, exists := detectedStack[tech]; exists {
			if configEntry.Version == detectedEntry.Version {
				result.Results.Matched = append(result.Results.Matched, ComparisonEntry{
					Name:     tech,
					Version:  configEntry.Version,
					Detected: detectedEntry.Version,
					Source:   detectedEntry.Source,
				})
				result.Summary.Matches++
			} else {
				result.Results.Mismatched = append(result.Results.Mismatched, ComparisonEntry{
					Name:     tech,
					Version:  configEntry.Version,
					Detected: detectedEntry.Version,
					Source:   detectedEntry.Source,
				})
				result.Summary.Mismatches++
			}
		} else {
			result.Results.MissingConfig = append(result.Results.MissingConfig, ComparisonEntry{
				Name:    tech,
				Version: configEntry.Version,
				Source:  configEntry.Source,
			})
			result.Summary.MissingConfig++
		}
	}

	// Determine overall status
	if result.Summary.Mismatches == 0 && result.Summary.MissingConfig == 0 {
		result.Status = "match"
	} else {
		result.Status = "mismatch"
	}

	return result
}

func outputText(result CheckResult) {
	fmt.Println("Technology Check Results")
	fmt.Println("========================")
	fmt.Println()

	if len(result.Results.Matched) > 0 {
		fmt.Printf("MATCH (%d):\n", len(result.Results.Matched))
		for _, entry := range result.Results.Matched {
			fmt.Printf("  %-12s %s == %s   ✓\n", entry.Name+":", entry.Version, entry.Detected)
		}
		fmt.Println()
	}

	if len(result.Results.Mismatched) > 0 {
		fmt.Printf("MISMATCH (%d):\n", len(result.Results.Mismatched))
		for _, entry := range result.Results.Mismatched {
			fmt.Printf("  %-12s %s != %s   (config has %s)\n", entry.Name+":", entry.Detected, entry.Version, entry.Version)
		}
		fmt.Println()
	}

	if len(result.Results.MissingConfig) > 0 {
		fmt.Printf("MISSING FROM DETECTION (%d):\n", len(result.Results.MissingConfig))
		for _, entry := range result.Results.MissingConfig {
			fmt.Printf("  %-12s %s   (in config but not detected)\n", entry.Name+":", entry.Version)
		}
		fmt.Println()
	}

	fmt.Printf("Summary: %d match, %d mismatch, %d missing\n",
		result.Summary.Matches,
		result.Summary.Mismatches,
		result.Summary.MissingConfig)

	if result.Status == "mismatch" {
		fmt.Println("Exit code: 1 (has differences)")
	} else {
		fmt.Println("Exit code: 0 (all match)")
	}
}

func outputJSON(result CheckResult) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(2)
	}
	fmt.Println(string(data))
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVarP(&checkConfigFile, "config", "c", "", "Path to stacktodate.yml config file (default: stacktodate.yml)")
	checkCmd.Flags().StringVarP(&checkFormat, "format", "f", "text", "Output format: text or json (default: text)")
}
