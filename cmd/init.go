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
		reader := bufio.NewReader(os.Stdin)

		// Check if token is configured, prompt if not
		token, err := helpers.GetToken()
		if err != nil {
			fmt.Println("Authentication token not configured.")
			fmt.Print("Would you like to set one up now? (y/n): ")
			response, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(response)) == "y" {
				fmt.Println("\nRun: stacktodate global-config set")
				return
			}
		}

		// Determine target directory
		targetDir := "."
		if len(args) > 0 {
			targetDir = args[0]
		}

		fmt.Printf("Initializing project in: %s\n", targetDir)

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

		// NEW: Menu-based project selection (create new or link existing)
		var projUUID, projName string
		if uuid == "" && name == "" {
			// Interactive mode: prompt user for choice
			choice := promptProjectChoice(reader)

			if choice == 1 {
				// Create new project on API
				var createErr error
				projUUID, projName, createErr = createNewProject(reader, detectedTechs, token)
				if createErr != nil {
					helpers.ExitOnError(createErr, "failed to create project")
				}
			} else {
				// Link to existing project on API
				var linkErr error
				projUUID, projName, linkErr = linkExistingProject(reader, token)
				if linkErr != nil {
					helpers.ExitOnError(linkErr, "failed to link project")
				}
			}
		} else {
			// Non-interactive mode: use provided flags or fallback to old prompts
			if uuid == "" {
				fmt.Print("Enter UUID: ")
				input, _ := reader.ReadString('\n')
				projUUID = strings.TrimSpace(input)
			} else {
				projUUID = uuid
			}

			if name == "" {
				fmt.Print("Enter name: ")
				input, _ := reader.ReadString('\n')
				projName = strings.TrimSpace(input)
			} else {
				projName = name
			}
		}

		// Create config
		config := helpers.Config{
			UUID:  projUUID,
			Name:  projName,
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
		fmt.Printf("  UUID: %s\n", projUUID)
		fmt.Printf("  Name: %s\n", projName)
		if len(detectedTechs) > 0 {
			fmt.Println("  Stack:")
			for tech, entry := range detectedTechs {
				fmt.Printf("    %s: %s (from: %s)\n", tech, entry.Version, entry.Source)
			}
		}
	},
}

// promptProjectChoice displays a menu for choosing between creating a new project or linking an existing one
func promptProjectChoice(reader *bufio.Reader) int {
	for {
		fmt.Println("\nDo you want to:")
		fmt.Println("  1) Create a new project on StackToDate")
		fmt.Println("  2) Link to an existing project")
		fmt.Print("\nEnter your choice (1 or 2): ")

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		if choice == "1" {
			return 1
		} else if choice == "2" {
			return 2
		}

		fmt.Println("Invalid choice. Please enter 1 or 2.")
	}
}

// createNewProject prompts for project name and creates a new project via API
func createNewProject(reader *bufio.Reader, detectedTechs map[string]helpers.StackEntry, token string) (uuid, projName string, err error) {
	fmt.Print("\nEnter project name: ")
	input, _ := reader.ReadString('\n')
	projName = strings.TrimSpace(input)

	if projName == "" {
		return "", "", fmt.Errorf("project name cannot be empty")
	}

	// Convert detected technologies to API components
	components := helpers.ConvertStackToComponents(detectedTechs)

	if len(components) == 0 {
		fmt.Println("⚠️  Warning: No technologies detected")
		fmt.Println("You can add them later by editing stacktodate.yml and running 'stacktodate push'")
	}

	fmt.Println("\nCreating project on StackToDate...")

	// Call API to create project
	response, err := helpers.CreateTechStack(token, projName, components)
	if err != nil {
		return "", "", err
	}

	uuid = response.TechStack.ID
	fmt.Println("✓ Project created successfully!")
	fmt.Printf("  UUID: %s\n", uuid)
	fmt.Printf("  Name: %s\n\n", projName)

	return uuid, projName, nil
}

// linkExistingProject prompts for UUID and links to an existing project via API
func linkExistingProject(reader *bufio.Reader, token string) (projUUID, projName string, err error) {
	fmt.Print("\nEnter project UUID: ")
	input, _ := reader.ReadString('\n')
	projUUID = strings.TrimSpace(input)

	if projUUID == "" {
		return "", "", fmt.Errorf("UUID cannot be empty")
	}

	fmt.Println("\nValidating project UUID...")

	// Call API to fetch project details
	response, err := helpers.GetTechStack(token, projUUID)
	if err != nil {
		return "", "", err
	}

	projName = response.TechStack.Name
	fmt.Printf("✓ Linked to existing project: %s\n\n", projName)

	return projUUID, projName, nil
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
