package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/stacktodate/stacktodate-cli/cmd/helpers"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// updateCmd updates an existing stacktodate.yml file's stack using autodetect
var updateCmd = &cobra.Command{
    Use:   "update",
    Short: "Update stack in a stacktodate.yml using autodetect",
    Long:  "Run autodetect and update the provided stacktodate.yml's stack, preserving uuid and name.",
    Args:  cobra.NoArgs,
    Run: func(cmd *cobra.Command, args []string) {
        // Load existing config without requiring UUID
        config, err := helpers.LoadConfig(updateConfigFile)
        if err != nil {
            helpers.ExitOnError(err, "failed to load config")
        }

        // Resolve absolute path
        absTargetFile, err := helpers.ResolveAbsPath(updateConfigFile)
        if err != nil {
            if updateConfigFile == "" {
                absTargetFile, _ = helpers.ResolveAbsPath("stacktodate.yml")
            } else {
                helpers.ExitOnError(err, "failed to resolve config path")
            }
        }

        // Get target directory
        targetDir, err := helpers.GetConfigDir(absTargetFile)
        if err != nil {
            helpers.ExitOnError(err, "failed to get config directory")
        }

        fmt.Printf("Updating stack in: %s\n", updateConfigFile)

        // Detect project information in target directory
        reader := bufio.NewReader(os.Stdin)

        var detectedTechs map[string]helpers.StackEntry
        if !skipAutodetect {
            err = helpers.WithWorkingDir(targetDir, func() error {
                info := DetectProjectInfo()
                PrintDetectedInfo(info)
                detectedTechs = selectCandidates(reader, info)
                return nil
            })
            if err != nil {
                helpers.ExitOnError(err, "failed to detect project")
            }
        } else {
            // If autodetect is skipped, keep existing stack
            detectedTechs = config.Stack
        }

        // Preserve UUID and Name, update Stack
        config.Stack = detectedTechs

        // Marshal to YAML
        data, err := yaml.Marshal(&config)
        if err != nil {
            helpers.ExitOnError(err, "failed to create configuration")
        }

        // Write back to the original absolute file path
        if err := os.WriteFile(absTargetFile, data, 0644); err != nil {
            helpers.ExitOnError(err, "failed to write config")
        }

        fmt.Println("\nStack updated successfully!")
        if len(detectedTechs) > 0 {
            fmt.Println("Updated stack:")
            for tech, entry := range detectedTechs {
                fmt.Printf("  %s: %s (from: %s)\n", tech, entry.Version, entry.Source)
            }
        }
    },
}

var updateConfigFile string

func init() {
    rootCmd.AddCommand(updateCmd)
    // Flags for update command
    updateCmd.Flags().StringVarP(&updateConfigFile, "config", "c", "stacktodate.yml", "Path to stacktodate.yml config file (default: stacktodate.yml)")
    updateCmd.Flags().BoolVar(&skipAutodetect, "skip-autodetect", false, "Skip autodetection of project technologies")
    updateCmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Use first candidate by default without prompting")
}
