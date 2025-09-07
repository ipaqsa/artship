package command

import (
	"fmt"

	"artship/internal/client"

	"github.com/spf13/cobra"

	"artship/internal/logs"
)

func init() {
	authCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	rootCmd.AddCommand(authCmd)
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Show Docker registry authentication information",
	Long: `Auth displays the current Docker registry authentication information,
showing which registries you're logged into and with which usernames.

This command reads the Docker configuration file to show your current
authentication state without exposing sensitive credentials.`,
	Example: `  # Show authentication info
  artship auth`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logs.New(verbose)

		cli := client.New(&client.Options{Logger: logger})

		auth, err := cli.GetAuth()
		if err != nil {
			return fmt.Errorf("failed to get authentication info: %w", err)
		}

		logger.Info("Authentication info: \n%s", auth.String())

		return nil
	},
}
