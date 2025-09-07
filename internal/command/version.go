package command

import (
	"github.com/spf13/cobra"

	"artship/internal/logs"
	"artship/internal/version"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Long:  `Print version info including build details and Go runtime version.`,
	Example: `  # Print version
  artship version`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logs.New(verbose)

		logger.Info("%s", version.Get())
	},
}
