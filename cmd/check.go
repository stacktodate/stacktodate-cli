package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/stacktodate/stacktodate-cli/cmd/helpers"
	"github.com/spf13/cobra"
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
		// Use default config file if not specified
		if checkConfigFile == "" {
			checkConfigFile = "stacktodate.yml"
		}

		// Load config without requiring UUID
		config, err := helpers.LoadConfig(checkConfigFile)
		if err != nil {
			helpers.ExitWithError(2, "failed to load config: %v", err)
		}

		// Resolve absolute path for directory management
		absConfigPath, err := helpers.ResolveAbsPath(checkConfigFile)
		if err != nil {
			helpers.ExitOnError(err, "failed to resolve config path")
		}

		// Get config directory
		configDir, err := helpers.GetConfigDir(absConfigPath)
		if err != nil {
			helpers.ExitOnError(err, "failed to get config directory")
		}

		// Detect current versions in config directory
		var detectedStack map[string]helpers.StackEntry
		err = helpers.WithWorkingDir(configDir, func() error {
			detectedInfo := DetectProjectInfo()
			detectedStack = normalizeDetectedToStack(detectedInfo)
			return nil
		})
		if err != nil {
			helpers.ExitOnError(err, "failed to detect versions")
		}

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

func normalizeDetectedToStack(info DetectedInfo) map[string]helpers.StackEntry {
	normalized := make(map[string]helpers.StackEntry)

	if len(info.Ruby) > 0 {
		normalized["ruby"] = helpers.StackEntry{
			Version: info.Ruby[0].Value,
			Source:  info.Ruby[0].Source,
		}
	}

	if len(info.Rails) > 0 {
		normalized["rails"] = helpers.StackEntry{
			Version: info.Rails[0].Value,
			Source:  info.Rails[0].Source,
		}
	}

	if len(info.Node) > 0 {
		normalized["nodejs"] = helpers.StackEntry{
			Version: info.Node[0].Value,
			Source:  info.Node[0].Source,
		}
	}

	if len(info.Go) > 0 {
		normalized["go"] = helpers.StackEntry{
			Version: info.Go[0].Value,
			Source:  info.Go[0].Source,
		}
	}

	if len(info.Python) > 0 {
		normalized["python"] = helpers.StackEntry{
			Version: info.Python[0].Value,
			Source:  info.Python[0].Source,
		}
	}

	return normalized
}

func compareStacks(configStack, detectedStack map[string]helpers.StackEntry) CheckResult {
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
