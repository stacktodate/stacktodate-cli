package globalconfig

import "github.com/spf13/cobra"

var GlobalConfigCmd = &cobra.Command{
	Use:   "global-config",
	Short: "Manage global configuration and authentication",
	Long:  `Configure authentication tokens and other global settings for stacktodate-cli`,
}

func init() {
	GlobalConfigCmd.AddCommand(setCmd)
	GlobalConfigCmd.AddCommand(getCmd)
	GlobalConfigCmd.AddCommand(deleteCmd)
}
