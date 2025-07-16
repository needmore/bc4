package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/needmore/bc4/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	versionDetailed bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `Display the version of bc4 along with build information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info := version.Get()

		// Check if JSON output is requested
		if viper.GetBool("json") {
			jsonData, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal version info: %w", err)
			}
			fmt.Println(string(jsonData))
			return nil
		}

		// Check if detailed output is requested
		if versionDetailed {
			fmt.Println(info.DetailedString())
		} else {
			fmt.Println(info.String())
		}

		return nil
	},
}

func init() {
	versionCmd.Flags().BoolVarP(&versionDetailed, "detailed", "d", false, "Show detailed version information")
}
