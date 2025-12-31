package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/stacktodate/stacktodate-cli/cmd/helpers"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	uuid           string
	name           string
	skipAutodetect bool
	noInteractive  bool
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new project",
	Long:  `Initialize a new project with default configuration`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine target directory
		targetDir := "."
		if len(args) > 0 {
			targetDir = args[0]
		}

		fmt.Printf("Initializing project in: %s\n", targetDir)

		reader := bufio.NewReader(os.Stdin)

		// Detect project information in target directory
		var detectedTechs map[string]helpers.StackEntry
		if !skipAutodetect {
			err := helpers.WithWorkingDir(targetDir, func() error {
				info := DetectProjectInfo()
				PrintDetectedInfo(info)
				detectedTechs = selectCandidates(reader, info)
				return nil
			})
			if err != nil {
				helpers.ExitOnError(err, "failed to detect project")
			}
		}

		// Get UUID
		if uuid == "" {
			fmt.Print("Enter UUID: ")
			input, _ := reader.ReadString('\n')
			uuid = strings.TrimSpace(input)
		}

		// Get Name
		if name == "" {
			fmt.Print("Enter name: ")
			input, _ := reader.ReadString('\n')
			name = strings.TrimSpace(input)
		}

		// Create config
		config := helpers.Config{
			UUID:  uuid,
			Name:  name,
			Stack: detectedTechs,
		}

		// Marshal to YAML
		data, err := yaml.Marshal(&config)
		if err != nil {
			helpers.ExitOnError(err, "failed to create configuration")
		}

		// Write to file
		err = os.WriteFile("stacktodate.yml", data, 0644)
		if err != nil {
			helpers.ExitOnError(err, "failed to write stacktodate.yml")
		}

		fmt.Println("\nProject initialized successfully!")
		fmt.Println("Created stacktodate.yml with:")
		fmt.Printf("  UUID: %s\n", uuid)
		fmt.Printf("  Name: %s\n", name)
		if len(detectedTechs) > 0 {
			fmt.Println("  Stack:")
			for tech, entry := range detectedTechs {
				fmt.Printf("    %s: %s (from: %s)\n", tech, entry.Version, entry.Source)
			}
		}
	},
}

// selectCandidates allows user to select from detected candidates
func selectCandidates(reader *bufio.Reader, info DetectedInfo) map[string]helpers.StackEntry {
	selected := make(map[string]helpers.StackEntry)

	// Ruby
	if len(info.Ruby) > 0 {
		choice := selectFromCandidates(reader, "ruby", info.Ruby)
		if choice.Version != "" {
			selected["ruby"] = choice
		}
	}

	// Rails
	if len(info.Rails) > 0 {
		choice := selectFromCandidates(reader, "rails", info.Rails)
		if choice.Version != "" {
			selected["rails"] = choice
		}
	}

	// Node.js
	if len(info.Node) > 0 {
		choice := selectFromCandidates(reader, "nodejs", info.Node)
		if choice.Version != "" {
			selected["nodejs"] = choice
		}
	}

	// Go
	if len(info.Go) > 0 {
		choice := selectFromCandidates(reader, "go", info.Go)
		if choice.Version != "" {
			selected["go"] = choice
		}
	}

	// Python
	if len(info.Python) > 0 {
		choice := selectFromCandidates(reader, "python", info.Python)
		if choice.Version != "" {
			selected["python"] = choice
		}
	}

	return selected
}

// selectFromCandidates lets user choose one candidate or none
func selectFromCandidates(reader *bufio.Reader, tech string, candidates []Candidate) helpers.StackEntry {
	if noInteractive {
		// In non-interactive mode, use the first candidate
		return helpers.StackEntry{
			Version: candidates[0].Value,
			Source:  candidates[0].Source,
		}
	}

	fmt.Printf("\nSelect %s version (or press Enter to skip):\n", tech)
	for i, candidate := range candidates {
		fmt.Printf("  %d) %s (from: %s)\n", i+1, candidate.Value, candidate.Source)
	}
	fmt.Printf("  0) Skip\n")

	for {
		fmt.Print("Your choice: ")
		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		if choice == "0" || choice == "" {
			return helpers.StackEntry{}
		}

		idx, err := strconv.Atoi(choice)
		if err != nil || idx < 1 || idx > len(candidates) {
			fmt.Println("Invalid choice. Please try again.")
			continue
		}

		return helpers.StackEntry{
			Version: candidates[idx-1].Value,
			Source:  candidates[idx-1].Source,
		}
	}
}

func init() {
	initCmd.Flags().StringVarP(&uuid, "uuid", "u", "", "UUID for the project")
	initCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the project")
	initCmd.Flags().BoolVar(&skipAutodetect, "skip-autodetect", false, "Skip autodetection of project technologies")
	initCmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Use first candidate by default without prompting")
}
