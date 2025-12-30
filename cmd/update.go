package cmd

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"

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
        // Read config path from flag or default
        targetFile := updateConfigFile
        if targetFile == "" {
            targetFile = "stacktodate.yml"
        }

        // Resolve absolute path for robust read/write regardless of chdir
        absTargetFile, err := filepath.Abs(targetFile)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error resolving absolute path for %s: %v\n", targetFile, err)
            os.Exit(1)
        }

        // Read existing config
        content, err := os.ReadFile(absTargetFile)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error reading config file %s: %v\n", targetFile, err)
            os.Exit(1)
        }

        var config Config
        if err := yaml.Unmarshal(content, &config); err != nil {
            fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", targetFile, err)
            os.Exit(1)
        }

        // Change to directory of the target file to run detection there
        originalDir, err := os.Getwd()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
            os.Exit(1)
        }

        targetDir := filepath.Dir(absTargetFile)
        if targetDir == "" {
            targetDir = "."
        }

        if targetDir != "." {
            if err := os.Chdir(targetDir); err != nil {
                fmt.Fprintf(os.Stderr, "Error changing to directory %s: %v\n", targetDir, err)
                os.Exit(1)
            }
            defer os.Chdir(originalDir)
        }

        fmt.Printf("Updating stack in: %s\n", targetFile)

        // Detect project information
        reader := bufio.NewReader(os.Stdin)

        var info DetectedInfo
        var detectedTechs map[string]StackEntry
        if !skipAutodetect {
            info = DetectProjectInfo()
            PrintDetectedInfo(info)
            detectedTechs = selectCandidates(reader, info)
        } else {
            // If autodetect is skipped, keep existing stack
            detectedTechs = config.Stack
        }

        // Preserve UUID and Name, update Stack
        config.Stack = detectedTechs

        // Marshal to YAML
        data, err := yaml.Marshal(&config)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error creating configuration: %v\n", err)
            os.Exit(1)
        }

        // Write back to the original absolute file path
        if err := os.WriteFile(absTargetFile, data, 0644); err != nil {
            fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", targetFile, err)
            os.Exit(1)
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
    // Flags for update command
    updateCmd.Flags().StringVarP(&updateConfigFile, "config", "c", "stacktodate.yml", "Path to stacktodate.yml config file (default: stacktodate.yml)")
    updateCmd.Flags().BoolVar(&skipAutodetect, "skip-autodetect", false, "Skip autodetection of project technologies")
    updateCmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Use first candidate by default without prompting")
}
